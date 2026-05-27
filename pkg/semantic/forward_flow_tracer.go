package semantic

// Package semantic — forward_flow_tracer.go
//
// forwardFlowTracer has a single responsibility: tracing how taint flows
// forward through a parsed codebase.  Starting from a set of input sources
// (FlowNode values of type NodeSource) it follows assignments, function calls,
// and cross-file parameter passing up to a configurable depth.
//
// It does NOT discover files, build symbol tables, or collect sources — those
// concerns live in fileDiscoverer, symbolTableBuilder, and Tracer respectively.

import (
	"fmt"
	"runtime"
	"strings"
	"sync"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
)

// forwardFlowTracer traces taint forwards through assignments and function
// calls.  It holds references (not copies) to the Tracer's shared data
// structures so that it operates on exactly the same state without
// duplicating large maps.
type forwardFlowTracer struct {
	config      *Config
	files       map[string]*FileInfo // shared with Tracer (read-only during tracing)
	mu          *sync.RWMutex        // protects files and symbolTable
	stats       *TraceStats          // shared with Tracer (written under flowMu in workers)
	symbolTable *types.SymbolTable   // shared with Tracer (read-only during tracing)
}

// newForwardFlowTracer constructs a forwardFlowTracer that references the
// Tracer's own data.  Call this after all parsing and symbol-table building
// is complete so that the shared maps are stable.
func newForwardFlowTracer(t *Tracer) *forwardFlowTracer {
	return &forwardFlowTracer{
		config:      t.config,
		files:       t.files,
		mu:          &t.mu,
		stats:       t.stats,
		symbolTable: t.symbolTable,
	}
}

// flowTraceCtx bundles the three positional arguments that every inner trace
// function receives so that call sites stay concise.
type flowTraceCtx struct {
	flowMap  *types.FlowMap
	rootPath string
	depth    int
}

// traceAllFlows traces flows from all sources using parallel workers and
// returns the merged FlowMap.
func (fft *forwardFlowTracer) traceAllFlows(sources []*types.FlowNode, rootPath string) *types.FlowMap {
	// Use NewFlowMapWithLimits for O(1) deduplication support with configurable limits
	flowMap := types.NewFlowMapWithLimits(fft.config.MaxFlowNodes, fft.config.MaxFlowEdges)

	// Add all sources as nodes using O(1) AddNode
	for _, source := range sources {
		flowMap.AddNode(*source)
	}

	// Limit number of sources to trace
	maxSources := 200
	if len(sources) > maxSources {
		if fft.config.Verbose {
			fmt.Printf("  Limiting flow analysis to %d sources (of %d) for memory safety\n", maxSources, len(sources))
		}
		sources = sources[:maxSources]
	}

	// Run GC before flow tracing to start with clean slate
	runtime.GC()

	// If few sources, trace sequentially to avoid overhead
	if len(sources) <= 2 {
		for _, source := range sources {
			fft.traceSource(source, flowMap, rootPath, nil)
		}
		return flowMap
	}

	numWorkers := fft.config.Workers
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}
	if numWorkers > len(sources) && len(sources) > 0 {
		numWorkers = len(sources)
	}

	// Create source channel
	sourceChan := make(chan *types.FlowNode, len(sources))
	for _, s := range sources {
		sourceChan <- s
	}
	close(sourceChan)

	// Worker results channel
	type workerResult struct {
		nodes []types.FlowNode
		edges []types.FlowEdge
	}
	results := make(chan workerResult, numWorkers)

	// Mutex for protecting flowMap writes (for stats and intermediate writes)
	var flowMu sync.Mutex

	// Memory tracking for flow tracing
	var memoryExceeded bool
	var memCheckMu sync.Mutex
	sourcesProcessed := 0
	memCheckInterval := 20 // Check memory every 20 sources

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Worker-local storage for results
			localNodes := make([]types.FlowNode, 0, 64)
			localEdges := make([]types.FlowEdge, 0, 128)

			// Create worker-local flowMap for collection with configurable limits
			localFlowMap := types.NewFlowMapWithLimits(fft.config.MaxFlowNodes, fft.config.MaxFlowEdges)

			for source := range sourceChan {
				// Check if memory limit exceeded
				memCheckMu.Lock()
				if memoryExceeded {
					memCheckMu.Unlock()
					continue // Skip remaining sources
				}
				sourcesProcessed++
				localCount := sourcesProcessed
				memCheckMu.Unlock()

				// Trace into local flowMap
				fft.traceSource(source, localFlowMap, rootPath, &flowMu)

				// Periodic memory check
				if fft.config.MaxMemoryMB > 0 && localCount%memCheckInterval == 0 {
					runtime.GC()
					memMB := getMemoryUsageMB()
					maxMB := uint64(fft.config.MaxMemoryMB)
					if memMB > maxMB {
						memCheckMu.Lock()
						memoryExceeded = true
						memCheckMu.Unlock()
						if fft.config.Verbose {
							fmt.Printf("  [Memory] Flow tracing stopped at %d MB (limit: %d MB)\n", memMB, maxMB)
						}
					}
				}
			}

			// Collect local results
			localNodes = append(localNodes, localFlowMap.AllNodes...)
			localEdges = append(localEdges, localFlowMap.AllEdges...)

			results <- workerResult{localNodes, localEdges}
		}()
	}

	// Close results channel when workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Merge results (dedup happens automatically via AddNode/AddEdge)
	for result := range results {
		for _, n := range result.nodes {
			flowMap.AddNode(n)
		}
		for _, e := range result.edges {
			flowMap.AddEdge(e)
		}
	}

	return flowMap
}

// traceSource traces flows from a single source.
// When statsMu is non-nil the call is treated as parallel and FlowsTraced
// increments are protected by that mutex; when nil they happen directly
// (sequential path, no contention).
func (fft *forwardFlowTracer) traceSource(source *types.FlowNode, flowMap *types.FlowMap, rootPath string, statsMu *sync.Mutex) {
	// Get file info with lock to prevent concurrent map access
	fft.mu.RLock()
	fileInfo := fft.files[source.FilePath]
	fft.mu.RUnlock()

	if fileInfo == nil {
		return
	}

	// Get analyzer (read-only, safe)
	langAnalyzer := analyzer.DefaultRegistry.Get(fileInfo.Language)
	if langAnalyzer == nil {
		return
	}

	// Create initial taint chain for this source
	initialChain := types.NewTaintChain(
		source.Snippet,
		string(source.SourceType),
		source.FilePath,
		source.Line,
	)

	// incFlows increments FlowsTraced in a goroutine-safe way.
	incFlows := func() {
		if statsMu != nil {
			statsMu.Lock()
			fft.stats.FlowsTraced++
			statsMu.Unlock()
		} else {
			fft.stats.FlowsTraced++
		}
	}

	// Use cached assignments extracted during initial parsing
	assignments := fileInfo.Assignments
	if assignments == nil {
		return // No assignments cached (file had no sources during parsing)
	}

	// Find assignments that use this source
	for _, assign := range assignments {
		if assign.IsTainted && containsSourceName(assign.Source, source.Name) {
			varNode := types.FlowNode{
				ID:         fmt.Sprintf("%s:%d:%d", source.FilePath, assign.Line, assign.Column),
				Type:       types.NodeVariable,
				Language:   fileInfo.Language,
				FilePath:   source.FilePath,
				Line:       assign.Line,
				Column:     assign.Column,
				Name:       assign.Target,
				Snippet:    fmt.Sprintf("%s = %s", assign.Target, assign.Source),
				SourceType: source.SourceType,
			}
			flowMap.AddNode(varNode)

			edge := types.FlowEdge{
				From:        source.ID,
				To:          varNode.ID,
				Type:        types.EdgeAssignment,
				Description: "assigned to",
			}
			flowMap.AddEdge(edge)
			incFlows()

			// Clone and extend taint chain for this assignment
			varChain := initialChain.Clone()
			varChain.AddStep("assignment", assign.Target, source.FilePath, assign.Line,
				fmt.Sprintf("%s assigned from %s", assign.Target, source.Name))

			// Recursively trace this variable with taint chain
			fft.traceVariable(&varNode, varChain, flowTraceCtx{flowMap: flowMap, rootPath: rootPath, depth: 1}, fileInfo, langAnalyzer)
		}
	}

	// Use cached calls extracted during initial parsing
	calls := fileInfo.Calls
	if calls == nil {
		return // No calls cached
	}

	for _, call := range calls {
		if call.HasTaintedArgs {
			for _, argIdx := range call.TaintedArgIndices {
				if argIdx < len(call.Arguments) {
					arg := call.Arguments[argIdx]
					if containsSourceName(arg.Value, source.Name) {
						fft.traceCall(source, call, nil, flowTraceCtx{flowMap: flowMap, rootPath: rootPath, depth: 1})
					}
				}
			}
		}
	}
}

// traceVariable traces flows from a tainted variable.
// When chain is non-nil the full taint chain is tracked through each step.
func (fft *forwardFlowTracer) traceVariable(varNode *types.FlowNode, chain *types.TaintChain, ftc flowTraceCtx, fileInfo *FileInfo, langAnalyzer analyzer.LanguageAnalyzer) {
	if ftc.depth > fft.config.MaxDepth {
		return
	}

	// Use cached assignments extracted during initial parsing
	assignments := fileInfo.Assignments
	if assignments == nil {
		return
	}

	for _, assign := range assignments {
		if assign.Line > varNode.Line && containsSourceName(assign.Source, varNode.Name) {
			// Create node for new variable
			newVarNode := types.FlowNode{
				ID:         fmt.Sprintf("%s:%d:%d", varNode.FilePath, assign.Line, assign.Column),
				Type:       types.NodeVariable,
				Language:   fileInfo.Language,
				FilePath:   varNode.FilePath,
				Line:       assign.Line,
				Column:     assign.Column,
				Name:       assign.Target,
				Snippet:    fmt.Sprintf("%s = %s", assign.Target, assign.Source),
				SourceType: varNode.SourceType,
			}

			// Use O(1) AddNode with built-in deduplication
			if ftc.flowMap.AddNode(newVarNode) {
				edge := types.FlowEdge{
					From:        varNode.ID,
					To:          newVarNode.ID,
					Type:        types.EdgeAssignment,
					Description: "assigned to",
				}
				ftc.flowMap.AddEdge(edge)
				fft.stats.FlowsTraced++

				var newChain *types.TaintChain
				if chain != nil {
					// Clone and extend taint chain
					newChain = chain.Clone()
					newChain.AddStep("assignment", assign.Target, varNode.FilePath, assign.Line,
						fmt.Sprintf("%s assigned from %s", assign.Target, varNode.Name))
				}

				// Recursively trace
				fft.traceVariable(&newVarNode, newChain, flowTraceCtx{flowMap: ftc.flowMap, rootPath: ftc.rootPath, depth: ftc.depth + 1}, fileInfo, langAnalyzer)
			}
		}
	}

	// Use cached calls extracted during initial parsing
	calls := fileInfo.Calls
	if calls == nil {
		return
	}

	for _, call := range calls {
		if call.Line > varNode.Line {
			for i, arg := range call.Arguments {
				if containsSourceName(arg.Value, varNode.Name) {
					// Create copy with taint info for this specific call
					callCopy := *call
					callCopy.HasTaintedArgs = true
					callCopy.TaintedArgIndices = []int{i}

					if chain != nil {
						// Attach taint chain to the argument
						if i < len(callCopy.Arguments) {
							argChain := chain.Clone()
							argChain.AddStep("parameter", fmt.Sprintf("arg[%d] = %s", i, arg.Value),
								varNode.FilePath, call.Line,
								fmt.Sprintf("passed as argument %d to %s", i, call.FunctionName))
							callCopy.Arguments[i].TaintChain = argChain
						}
					}

					fft.traceCall(varNode, &callCopy, chain, flowTraceCtx{flowMap: ftc.flowMap, rootPath: ftc.rootPath, depth: ftc.depth})
				}
			}
		}
	}
}

// traceCall traces a function call with a tainted argument.
// When chain is non-nil the full taint chain is forwarded into the callee.
func (fft *forwardFlowTracer) traceCall(source *types.FlowNode, call *types.CallSite, chain *types.TaintChain, ftc flowTraceCtx) {
	if ftc.depth > fft.config.MaxDepth {
		return
	}

	// Create node for the function call
	callNode := types.FlowNode{
		ID:         fmt.Sprintf("%s:%d:%d:call", source.FilePath, call.Line, call.Column),
		Type:       types.NodeFunction,
		Language:   source.Language,
		FilePath:   source.FilePath,
		Line:       call.Line,
		Column:     call.Column,
		Name:       call.FunctionName,
		Snippet:    call.FunctionName + "()",
		SourceType: source.SourceType,
	}

	// Use O(1) AddNode with built-in deduplication
	if ftc.flowMap.AddNode(callNode) {
		argStr := "arg"
		if len(call.TaintedArgIndices) > 0 {
			argStr = fmt.Sprintf("arg[%d]", call.TaintedArgIndices[0])
		}

		edge := types.FlowEdge{
			From:        source.ID,
			To:          callNode.ID,
			Type:        types.EdgeCall,
			Description: argStr,
		}
		ftc.flowMap.AddEdge(edge)
		fft.stats.FlowsTraced++
	}

	// If cross-file tracing is enabled, find the function definition and trace into it
	if fft.config.FollowImports {
		fft.traceIntoFunction(&callNode, call, chain, flowTraceCtx{flowMap: ftc.flowMap, rootPath: ftc.rootPath, depth: ftc.depth + 1})
	}
}

// traceIntoFunction traces execution into a called function.
// When chain is non-nil the full taint chain is tracked through parameters.
func (fft *forwardFlowTracer) traceIntoFunction(callNode *types.FlowNode, call *types.CallSite, chain *types.TaintChain, ftc flowTraceCtx) {
	if ftc.depth > fft.config.MaxDepth {
		return
	}

	// Find function definition in global symbol table
	fft.mu.RLock()
	var funcDef *types.FunctionDef
	var funcFile string

	// Try different name patterns
	funcNames := []string{call.FunctionName, call.MethodName}
	if call.ClassName != "" {
		funcNames = append(funcNames, call.ClassName+"::"+call.MethodName)
	}

	for _, name := range funcNames {
		if fn, ok := fft.symbolTable.Functions[name]; ok {
			funcDef = fn
			funcFile = fn.FilePath
			break
		}
		// Also search with file prefix
		for key, fn := range fft.symbolTable.Functions {
			if strings.HasSuffix(key, "::"+name) {
				funcDef = fn
				funcFile = fn.FilePath
				break
			}
		}
	}
	fft.mu.RUnlock()

	if funcDef == nil {
		return
	}

	// Create node for the function definition
	funcNode := types.FlowNode{
		ID:       fmt.Sprintf("%s:%d:func", funcFile, funcDef.Line),
		Type:     types.NodeFunction,
		Language: callNode.Language,
		FilePath: funcFile,
		Line:     funcDef.Line,
		Name:     funcDef.Name,
		Snippet:  funcDef.Name + "()",
	}

	// Use O(1) AddNode with built-in deduplication
	if ftc.flowMap.AddNode(funcNode) {
		edge := types.FlowEdge{
			From:        callNode.ID,
			To:          funcNode.ID,
			Type:        types.EdgeCall,
			Description: "calls",
		}
		ftc.flowMap.AddEdge(edge)
		fft.stats.FlowsTraced++

		if callNode.FilePath != funcFile {
			fft.stats.CrossFileFlows++
		}
	}

	// If function has parameters and we have tainted args, trace parameter
	if len(call.TaintedArgIndices) > 0 && len(funcDef.Parameters) > 0 {
		for _, argIdx := range call.TaintedArgIndices {
			if argIdx < len(funcDef.Parameters) {
				param := funcDef.Parameters[argIdx]

				nodeType := types.NodeVariable
				edgeType := types.EdgeDataFlow
				edgeDesc := "param"
				if chain != nil {
					nodeType = types.NodeParam
					edgeType = types.EdgeParameter
					edgeDesc = fmt.Sprintf("param[%d]", argIdx)
				}

				paramNode := types.FlowNode{
					ID:       fmt.Sprintf("%s:%d:param:%s", funcFile, funcDef.Line, param.Name),
					Type:     nodeType,
					Language: callNode.Language,
					FilePath: funcFile,
					Line:     funcDef.Line,
					Name:     param.Name,
					Snippet:  fmt.Sprintf("param $%s", param.Name),
				}

				// Use O(1) AddNode with built-in deduplication
				if ftc.flowMap.AddNode(paramNode) {
					edge := types.FlowEdge{
						From:        funcNode.ID,
						To:          paramNode.ID,
						Type:        edgeType,
						Description: edgeDesc,
					}
					ftc.flowMap.AddEdge(edge)
					fft.stats.FlowsTraced++

					// Compute parameter taint chain when chain tracking is active
					var paramChain *types.TaintChain
					if chain != nil {
						if argIdx < len(call.Arguments) && call.Arguments[argIdx].TaintChain != nil {
							paramChain = call.Arguments[argIdx].TaintChain.Clone()
						} else {
							paramChain = chain.Clone()
						}
						paramChain.AddStep("parameter", param.Name, funcFile, funcDef.Line,
							fmt.Sprintf("received as parameter %s in %s", param.Name, funcDef.Name))
					}

					// Continue tracing inside the function
					fft.mu.RLock()
					fileInfo := fft.files[funcFile]
					fft.mu.RUnlock()

					if fileInfo != nil {
						langAnalyzer := analyzer.DefaultRegistry.Get(fileInfo.Language)
						if langAnalyzer != nil {
							fft.traceVariable(&paramNode, paramChain, flowTraceCtx{flowMap: ftc.flowMap, rootPath: ftc.rootPath, depth: ftc.depth}, fileInfo, langAnalyzer)
						}
					}
				}
			}
		}
	}
}

// containsSourceName checks if an expression contains a source reference.
func containsSourceName(expr, sourceName string) bool {
	return strings.Contains(expr, sourceName)
}
