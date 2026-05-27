package semantic

// tracer_backward.go — backward taint analysis methods on *Tracer.
//
// Backward analysis starts from a target variable or expression and walks
// backwards through assignments to find the original input sources.  The
// public entry points are TraceBackward and TraceBackwardBatch; all other
// functions in this file are unexported helpers used only by those two methods.

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	"github.com/hatlesswizard/inputtracer/pkg/sources"
	phpPatterns "github.com/hatlesswizard/inputtracer/pkg/sources/php"
)

// TraceBackwardBatch performs backward taint analysis for MULTIPLE target expressions in a SINGLE pass.
// This is CRITICAL for performance: instead of N × files reads (for N variables),
// we do a single pass through all files, checking all variables at once.
// PERF: Shares TraceContext and assignment cache across all variables.
func (t *Tracer) TraceBackwardBatch(targets []string, codebasePath string) (*types.BatchTraceResult, error) {
	startTime := time.Now()

	if len(targets) == 0 {
		return &types.BatchTraceResult{
			HasUserInput:   false,
			PerVariable:    make(map[string]*types.BackwardTraceResult),
			TotalDuration:  time.Since(startTime),
			AnalyzedFiles:  0,
			VariablesFound: 0,
		}, nil
	}

	// Ensure the codebase has been parsed before tracing.
	if err := t.ensureParsed(codebasePath); err != nil {
		return nil, fmt.Errorf("failed to parse codebase: %w", err)
	}

	result := &types.BatchTraceResult{
		HasUserInput:   false,
		PerVariable:    make(map[string]*types.BackwardTraceResult),
		TotalDuration:  0,
		AnalyzedFiles:  len(t.files),
		VariablesFound: 0,
	}

	// Initialize per-variable results
	for _, target := range targets {
		result.PerVariable[target] = &types.BackwardTraceResult{
			TargetExpression: target,
			Paths:            make([]types.BackwardPath, 0),
			Sources:          make([]types.SourceInfo, 0),
			AnalyzedFiles:    len(t.files),
		}
	}

	// Clean target variable names - build lookup map
	targetVars := make(map[string]string) // cleaned -> original
	for _, target := range targets {
		targetVar := strings.TrimSpace(target)
		targetVar = strings.TrimPrefix(targetVar, "$")
		targetVars[targetVar] = target
	}

	// Collect all file paths for processing
	t.mu.RLock()
	filePaths := make([]string, 0, len(t.files))
	for filePath := range t.files {
		filePaths = append(filePaths, filePath)
	}
	t.mu.RUnlock()

	// CRITICAL: Create ONE shared TraceContext for ALL variables
	// This is the key optimization - the assignment cache is shared!
	ctx := newTraceContext()
	defer ctx.Close()

	// Global dedup map for sources
	seenSources := make(map[string]map[string]bool) // variable -> sourceKey -> seen
	for _, target := range targets {
		seenSources[target] = make(map[string]bool)
	}

	// Process ALL files in a SINGLE pass
	for _, filePath := range filePaths {
		// Get file info
		t.mu.RLock()
		fileInfo := t.files[filePath]
		t.mu.RUnlock()
		if fileInfo == nil {
			continue
		}

		// Get cached assignments (parses → extracts → immediately discards AST)
		// This is the expensive operation that's now SHARED across all variables
		assignments := ctx.getAssignmentsDirectly(filePath, fileInfo.Language)
		if assignments == nil {
			continue
		}

		// Check ALL target variables against these assignments in ONE loop
		for _, assign := range assignments {
			assignTarget := strings.TrimPrefix(assign.Target, "$")

			// Check if this assignment is to ANY of our target variables
			originalTarget, isTarget := targetVars[assignTarget]
			if !isTarget {
				continue
			}

			// Found an assignment to one of our targets
			varResult := result.PerVariable[originalTarget]

			path := types.BackwardPath{
				Steps:     make([]types.BackwardStep, 0),
				CrossFile: false,
			}

			path.Steps = append(path.Steps, types.BackwardStep{
				StepNumber:  1,
				Expression:  fmt.Sprintf("$%s = %s", assignTarget, assign.Source),
				FilePath:    filePath,
				Line:        assign.Line,
				StepType:    "assignment",
				Description: fmt.Sprintf("$%s assigned from %s", assignTarget, assign.Source),
			})

			// Check if the source is a superglobal
			sourceInfo := t.identifySource(assign.Source, filePath, assign.Line)
			if sourceInfo != nil {
				path.Source = *sourceInfo
				path.Steps = append([]types.BackwardStep{{
					StepNumber:  0,
					Expression:  sourceInfo.Expression,
					FilePath:    sourceInfo.FilePath,
					Line:        sourceInfo.Line,
					StepType:    "source",
					Description: fmt.Sprintf("Input source: %s (%s)", sourceInfo.Expression, sourceInfo.Type),
				}}, path.Steps...)

				varResult.Paths = append(varResult.Paths, path)

				sourceKey := fmt.Sprintf("%s:%s", sourceInfo.Type, sourceInfo.Expression)
				if !seenSources[originalTarget][sourceKey] {
					seenSources[originalTarget][sourceKey] = true
					varResult.Sources = append(varResult.Sources, *sourceInfo)
				}

				result.HasUserInput = true
				result.VariablesFound++
			} else if strings.HasPrefix(assign.Source, "$") {
				// The source is another variable - trace recursively WITH SHARED CONTEXT
				innerSources := t.traceBackwardRecursiveWithContext(ctx, assign.Source, filePath, make(map[string]bool), 0)
				for _, innerSource := range innerSources {
					innerPath := types.BackwardPath{
						Source:    innerSource,
						Steps:     make([]types.BackwardStep, 0),
						CrossFile: innerSource.FilePath != filePath,
					}
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  0,
						Expression:  innerSource.Expression,
						FilePath:    innerSource.FilePath,
						Line:        innerSource.Line,
						StepType:    "source",
						Description: fmt.Sprintf("Input source: %s", innerSource.Expression),
					})
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  1,
						Expression:  assign.Source,
						FilePath:    filePath,
						Line:        assign.Line,
						StepType:    "intermediate",
						Description: fmt.Sprintf("Via %s", assign.Source),
					})
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  2,
						Expression:  fmt.Sprintf("$%s = %s", assignTarget, assign.Source),
						FilePath:    filePath,
						Line:        assign.Line,
						StepType:    "assignment",
						Description: fmt.Sprintf("Assigned to $%s", assignTarget),
					})

					varResult.Paths = append(varResult.Paths, innerPath)

					sourceKey := fmt.Sprintf("%s:%s", innerSource.Type, innerSource.Expression)
					if !seenSources[originalTarget][sourceKey] {
						seenSources[originalTarget][sourceKey] = true
						varResult.Sources = append(varResult.Sources, innerSource)
					}

					result.HasUserInput = true
					result.VariablesFound++
				}
			}
		}
	}

	// Set durations for all per-variable results
	totalDuration := time.Since(startTime)
	for _, varResult := range result.PerVariable {
		varResult.Duration = totalDuration
	}
	result.TotalDuration = totalDuration

	return result, nil
}

// TraceBackward performs backward taint analysis from a target expression.
// This traces from a target variable/expression back to its input sources.
func (t *Tracer) TraceBackward(target string, codebasePath string) (*types.BackwardTraceResult, error) {
	startTime := time.Now()

	// Ensure the codebase has been parsed before tracing.
	if err := t.ensureParsed(codebasePath); err != nil {
		return nil, fmt.Errorf("failed to parse codebase: %w", err)
	}

	result := &types.BackwardTraceResult{
		TargetExpression: target,
		Paths:            make([]types.BackwardPath, 0),
		Sources:          make([]types.SourceInfo, 0),
		AnalyzedFiles:    len(t.files),
	}

	// Clean target variable name
	targetVar := strings.TrimSpace(target)
	targetVar = strings.TrimPrefix(targetVar, "$")

	// Collect all file paths for parallel processing (with lock)
	t.mu.RLock()
	filePaths := make([]string, 0, len(t.files))
	for filePath := range t.files {
		filePaths = append(filePaths, filePath)
	}
	t.mu.RUnlock()

	// If few files, process sequentially with single context
	if len(filePaths) <= 4 {
		ctx := newTraceContext()
		defer ctx.Close()

		seenSources := make(map[string]bool)
		for _, filePath := range filePaths {
			paths, srcs := t.traceBackwardInFileWithContext(ctx, filePath, targetVar)
			result.Paths = append(result.Paths, paths...)
			for _, src := range srcs {
				sourceKey := fmt.Sprintf("%s:%s", src.Type, src.Expression)
				if !seenSources[sourceKey] {
					seenSources[sourceKey] = true
					result.Sources = append(result.Sources, src)
				}
			}
		}
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// Parallel processing with worker pool.
	// Worker count is deliberately capped: each worker in TraceBackward holds
	// its own TraceContext (with per-file assignment caches and dedicated
	// parsers). On a large codebase the cumulative memory of many in-flight
	// contexts can easily exceed the MaxMemoryMB limit.
	//   • Default (Workers == 0): use 4 workers — a pragmatic balance between
	//     throughput and memory that empirically keeps usage well below 120 MB
	//     on typical PHP/JS projects.
	//   • Hard ceiling of 8: even when the caller explicitly sets a higher
	//     value, backward-trace contexts are memory-heavy enough that going
	//     beyond 8 offers diminishing returns and risks OOM on large repos.
	//   • Capped to len(filePaths): avoids creating idle goroutines when there
	//     are fewer files than workers.
	numWorkers := t.config.Workers
	if numWorkers <= 0 {
		numWorkers = 4 // Default: 4 workers balances throughput vs. memory
	}
	if numWorkers > 8 {
		numWorkers = 8 // Hard cap: each TraceContext holds parsers + caches
	}
	if numWorkers > len(filePaths) {
		numWorkers = len(filePaths)
	}

	// Create file path channel
	pathChan := make(chan string, len(filePaths))
	for _, fp := range filePaths {
		pathChan <- fp
	}
	close(pathChan)

	// Worker results
	type workerResult struct {
		paths   []types.BackwardPath
		sources []types.SourceInfo
	}
	results := make(chan workerResult, numWorkers)

	// Start workers - each with its own TraceContext
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each worker gets its own context (thread-safe, caches AST within worker)
			ctx := newTraceContext()
			defer ctx.Close()

			localPaths := make([]types.BackwardPath, 0, 16)
			localSources := make([]types.SourceInfo, 0, 8)

			for filePath := range pathChan {
				paths, srcs := t.traceBackwardInFileWithContext(ctx, filePath, targetVar)
				localPaths = append(localPaths, paths...)
				localSources = append(localSources, srcs...)
			}

			results <- workerResult{localPaths, localSources}
		}()
	}

	// Close results when workers done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Merge results
	seenSources := make(map[string]bool)
	for wr := range results {
		result.Paths = append(result.Paths, wr.paths...)
		for _, src := range wr.sources {
			sourceKey := fmt.Sprintf("%s:%s", src.Type, src.Expression)
			if !seenSources[sourceKey] {
				seenSources[sourceKey] = true
				result.Sources = append(result.Sources, src)
			}
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// traceBackwardInFileWithContext processes a single file for backward tracing using a TraceContext.
func (t *Tracer) traceBackwardInFileWithContext(ctx *TraceContext, filePath string, targetVar string) ([]types.BackwardPath, []types.SourceInfo) {
	paths := make([]types.BackwardPath, 0)
	srcs := make([]types.SourceInfo, 0)

	// Lock for reading t.files to get file metadata
	t.mu.RLock()
	fileInfo := t.files[filePath]
	t.mu.RUnlock()

	if fileInfo == nil {
		return paths, srcs
	}

	// Get cached assignments (parses → extracts → immediately discards AST)
	// This is memory-efficient: only assignments are cached, not ASTs
	assignments := ctx.getAssignmentsDirectly(filePath, fileInfo.Language)
	if assignments == nil {
		return paths, srcs
	}

	for _, assign := range assignments {
		// Check if this assignment is to our target variable
		assignTarget := strings.TrimPrefix(assign.Target, "$")
		if assignTarget != targetVar {
			continue
		}

		// Found an assignment to target - trace backward from source
		path := types.BackwardPath{
			Steps:     make([]types.BackwardStep, 0),
			CrossFile: false,
		}

		// Add the assignment as a step
		path.Steps = append(path.Steps, types.BackwardStep{
			StepNumber:  1,
			Expression:  fmt.Sprintf("$%s = %s", targetVar, assign.Source),
			FilePath:    filePath,
			Line:        assign.Line,
			StepType:    "assignment",
			Description: fmt.Sprintf("$%s assigned from %s", targetVar, assign.Source),
		})

		// Check if the source is a superglobal
		sourceInfo := t.identifySource(assign.Source, filePath, assign.Line)
		if sourceInfo != nil {
			path.Source = *sourceInfo
			path.Steps = append([]types.BackwardStep{{
				StepNumber:  0,
				Expression:  sourceInfo.Expression,
				FilePath:    sourceInfo.FilePath,
				Line:        sourceInfo.Line,
				StepType:    "source",
				Description: fmt.Sprintf("Input source: %s (%s)", sourceInfo.Expression, sourceInfo.Type),
			}}, path.Steps...)

			paths = append(paths, path)
			srcs = append(srcs, *sourceInfo)
		} else {
			// The source might be another variable - trace recursively
			if strings.HasPrefix(assign.Source, "$") {
				innerSources := t.traceBackwardRecursiveWithContext(ctx, assign.Source, filePath, make(map[string]bool), 0)
				for _, innerSource := range innerSources {
					innerPath := types.BackwardPath{
						Source:    innerSource,
						Steps:     make([]types.BackwardStep, 0),
						CrossFile: innerSource.FilePath != filePath,
					}
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  0,
						Expression:  innerSource.Expression,
						FilePath:    innerSource.FilePath,
						Line:        innerSource.Line,
						StepType:    "source",
						Description: fmt.Sprintf("Input source: %s", innerSource.Expression),
					})
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  1,
						Expression:  assign.Source,
						FilePath:    filePath,
						Line:        assign.Line,
						StepType:    "intermediate",
						Description: fmt.Sprintf("Via %s", assign.Source),
					})
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  2,
						Expression:  fmt.Sprintf("$%s = %s", targetVar, assign.Source),
						FilePath:    filePath,
						Line:        assign.Line,
						StepType:    "assignment",
						Description: fmt.Sprintf("Assigned to $%s", targetVar),
					})

					paths = append(paths, innerPath)
					srcs = append(srcs, innerSource)
				}
			}
		}
	}

	return paths, srcs
}

// traceBackwardRecursiveWithContext recursively traces backward with caching and early termination.
func (t *Tracer) traceBackwardRecursiveWithContext(ctx *TraceContext, varExpr string, startFile string, visited map[string]bool, depth int) []types.SourceInfo {
	if depth > t.config.MaxDepth {
		return nil
	}

	// Prevent infinite loops
	visitKey := fmt.Sprintf("%s:%s", startFile, varExpr)
	if visited[visitKey] {
		return nil
	}
	visited[visitKey] = true

	var srcs []types.SourceInfo
	varName := strings.TrimPrefix(strings.TrimSpace(varExpr), "$")

	// OPTIMIZATION 1: Search current file FIRST (most common case)
	if found := t.searchFileForVar(ctx, startFile, varName, visited, depth, &srcs); found {
		return srcs // Early termination - found source!
	}

	// OPTIMIZATION 2: Only search other files if not found in current
	// Use read lock, NO map copying
	t.mu.RLock()
	filePaths := make([]string, 0, len(t.files))
	for fp := range t.files {
		if fp != startFile { // Skip already-searched file
			filePaths = append(filePaths, fp)
		}
	}
	t.mu.RUnlock()

	// OPTIMIZATION 3: Search other files with early termination
	for _, filePath := range filePaths {
		if found := t.searchFileForVar(ctx, filePath, varName, visited, depth, &srcs); found {
			return srcs // Early termination - found source!
		}
	}

	return srcs
}

// searchFileForVar searches a single file for variable assignments.
// Returns true if a source was found (for early termination).
func (t *Tracer) searchFileForVar(ctx *TraceContext, filePath string, varName string, visited map[string]bool, depth int, srcs *[]types.SourceInfo) bool {
	// Get file language
	t.mu.RLock()
	fileInfo := t.files[filePath]
	t.mu.RUnlock()
	if fileInfo == nil {
		return false
	}

	// Get cached assignments (parses → extracts → immediately discards AST)
	// This is memory-efficient: only assignments are cached, not ASTs
	assignments := ctx.getAssignmentsDirectly(filePath, fileInfo.Language)

	for _, assign := range assignments {
		if strings.TrimPrefix(assign.Target, "$") != varName {
			continue
		}

		// Check if source is user input
		if sourceInfo := t.identifySource(assign.Source, filePath, assign.Line); sourceInfo != nil {
			*srcs = append(*srcs, *sourceInfo)
			return true // FOUND! Early termination
		}

		// Recurse if source is another variable
		if strings.HasPrefix(assign.Source, "$") {
			innerSources := t.traceBackwardRecursiveWithContext(ctx, assign.Source, filePath, visited, depth+1)
			if len(innerSources) > 0 {
				*srcs = append(*srcs, innerSources...)
				return true // FOUND! Early termination
			}
		}
	}

	return false
}

// identifySource checks if an expression is an input source and returns its info.
// This is only used by backward analysis; forward analysis uses the sources registry.
func (t *Tracer) identifySource(expr string, filePath string, line int) *types.SourceInfo {
	expr = strings.TrimSpace(expr)

	// Check PHP superglobals (using centralized definitions from pkg/sources)
	for sg, sourceType := range sources.SuperglobalToSourceType {
		if strings.Contains(expr, sg) {
			return &types.SourceInfo{
				Type:       types.SourceType(sourceType), // Convert sources.SourceType to types.SourceType
				Expression: expr,
				FilePath:   filePath,
				Line:       line,
			}
		}
	}

	// Check for input/deserialization/network functions using centralized definitions
	// from pkg/sources/php/functions.go (replaces hardcoded inline arrays)
	if sourceType, confidence := phpPatterns.IdentifyExternalDataSource(expr); confidence > 0 {
		return &types.SourceInfo{
			Type:       types.SourceType(sourceType),
			Expression: expr,
			FilePath:   filePath,
			Line:       line,
		}
	}

	// Check property array access and method call patterns using centralized patterns
	if phpPatterns.IsInputPropertyAccess(expr) || phpPatterns.IsInputMethodCall(expr) {
		return &types.SourceInfo{
			Type:       types.SourceUserInput,
			Expression: expr,
			FilePath:   filePath,
			Line:       line,
		}
	}

	return nil
}
