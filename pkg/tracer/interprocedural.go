package tracer

import (
	"slices"
	"strings"
	"sync"

	"github.com/google/uuid"
	"github.com/hatlesswizard/inputtracer/pkg/ast"
	"github.com/hatlesswizard/inputtracer/pkg/parser"
	sitter "github.com/smacker/go-tree-sitter"
)

const (
	// defaultInterproceduralDepth is the fallback maximum cross-function analysis
	// depth when the caller does not provide a positive MaxDepth value.
	// Three iterations covers the vast majority of real-world call chains while
	// keeping analysis time predictable.
	defaultInterproceduralDepth = 3

	// scopeInterprocedural is the scope label assigned to tainted variables that
	// are discovered during inter-procedural (cross-function) taint propagation.
	scopeInterprocedural = "interprocedural"
)

// callGraph stores directed edges from callers to the functions they call.
// It is the sole owner of the edges map and provides thread-safe access via its
// own RW-mutex so that InterproceduralAnalyzer does not need to manage that
// locking concern.
type callGraph struct {
	edges map[string][]string // caller -> []callee
	mu    sync.RWMutex
}

func newCallGraph() *callGraph {
	return &callGraph{edges: make(map[string][]string)}
}

// addEdges merges a batch of caller→callees entries into the graph.
// Batching the write lets callers accumulate edges without the mutex and
// then commit them in a single lock acquisition.
func (cg *callGraph) addEdges(batch map[string][]string) {
	cg.mu.Lock()
	defer cg.mu.Unlock()
	for caller, callees := range batch {
		cg.edges[caller] = append(cg.edges[caller], callees...)
	}
}

// snapshot returns a deep copy of the edge map, safe for callers to read
// without holding any lock.
func (cg *callGraph) snapshot() map[string][]string {
	cg.mu.RLock()
	defer cg.mu.RUnlock()
	out := make(map[string][]string, len(cg.edges))
	for k, v := range cg.edges {
		out[k] = append([]string{}, v...)
	}
	return out
}

// InterproceduralAnalyzer coordinates cross-function taint analysis.
// Its single responsibility is orchestrating three subordinate concerns:
//  1. Building per-function summaries (BuildFunctionSummary and helpers).
//  2. Maintaining the call graph (via the embedded *callGraph).
//  3. Propagating taint across function boundaries (PropagateInterproceduralTaint,
//     RunAnalysis, propagateReturnTaint, propagateCallTaint).
type InterproceduralAnalyzer struct {
	state        *FullAnalysisState
	summaries    map[string]*FunctionSummary // function name -> summary
	graph        *callGraph
	parser       *parser.Service
	maxDepth     int
	currentDepth int
	mu           sync.RWMutex
}

// NewInterproceduralAnalyzer creates a new inter-procedural analyzer
func NewInterproceduralAnalyzer(state *FullAnalysisState, maxDepth int, parserSvc *parser.Service) *InterproceduralAnalyzer {
	return &InterproceduralAnalyzer{
		state:     state,
		summaries: make(map[string]*FunctionSummary),
		graph:     newCallGraph(),
		parser:    parserSvc,
		maxDepth:  maxDepth,
	}
}

// BuildFunctionSummary builds a summary for a function definition
func (ipa *InterproceduralAnalyzer) BuildFunctionSummary(node *sitter.Node, src []byte, filePath string, language string) *FunctionSummary {
	funcName := ipa.extractFunctionName(node, src)
	if funcName == "" {
		return nil
	}

	ipa.mu.Lock()
	defer ipa.mu.Unlock()

	// Check if already summarized
	if existing, exists := ipa.summaries[funcName]; exists {
		return existing
	}

	summary := &FunctionSummary{
		Name:            funcName,
		FilePath:        filePath,
		Language:        language,
		ParamsToReturn:  make([]int, 0),
		ParamsToParams:  make(map[int][]int),
		CalledFunctions: make([]string, 0),
	}

	// Extract parameters
	params := ipa.extractParameters(node, src)
	summary.Parameters = params

	// Analyze function body for:
	// 1. Which params flow to return
	// 2. Which params flow to nested function calls
	// 3. Which functions are called
	//
	// localCallGraph accumulates call-graph edges discovered during traversal.
	// Passing it as a parameter avoids acquiring ipa.graph's mutex inside
	// traverseForFlow while ipa.mu is already held (consistent lock order:
	// ipa.mu is always acquired before ipa.graph.mu, never the reverse).
	localCallGraph := make(map[string][]string)
	ipa.analyzeFlowWithinFunction(node, src, summary, localCallGraph)

	// Merge local edges into the call graph.  addEdges acquires ipa.graph.mu
	// internally; acquiring it while ipa.mu is held is safe because the lock
	// ordering (ipa.mu → ipa.graph.mu) is consistent throughout this file.
	ipa.graph.addEdges(localCallGraph)

	ipa.summaries[funcName] = summary
	return summary
}

// extractFunctionName extracts the function name from a definition node
func (ipa *InterproceduralAnalyzer) extractFunctionName(node *sitter.Node, src []byte) string {
	nodeType := node.Type()

	// Look for name/identifier child
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()
		if childType == "identifier" || childType == "name" ||
			childType == "property_identifier" {
			return string(src[child.StartByte():child.EndByte()])
		}
		// For Go/Rust: function_declarator
		if strings.Contains(childType, "declarator") {
			return ipa.extractFunctionName(child, src)
		}
	}

	// Check if the node itself contains the name (e.g., method definitions)
	if strings.Contains(nodeType, "definition") || strings.Contains(nodeType, "declaration") {
		// Try to extract from the text
		text := string(src[node.StartByte():node.EndByte()])
		// Find first identifier-like pattern
		for _, part := range strings.Fields(text) {
			part = strings.TrimSuffix(part, "(")
			if isValidIdentifier(part) && part != "function" && part != "def" &&
				part != "fn" && part != "func" && part != "void" &&
				part != "public" && part != "private" {
				return part
			}
		}
	}

	return ""
}

// extractParameters extracts parameter information from a function definition
func (ipa *InterproceduralAnalyzer) extractParameters(node *sitter.Node, src []byte) []ParameterInfo {
	var params []ParameterInfo

	// Find parameters/formal_parameters node
	var paramsNode *sitter.Node
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()
		if strings.Contains(childType, "parameter") ||
			childType == "arguments" ||
			childType == "formal_parameters" {
			paramsNode = child
			break
		}
	}

	if paramsNode == nil {
		return params
	}

	// Extract each parameter, skipping punctuation tokens
	ast.IterateChildren(paramsNode, func(param *sitter.Node) {
		name := ipa.extractParameterName(param, src)
		if name != "" {
			params = append(params, ParameterInfo{
				Index: len(params),
				Name:  name,
			})
		}
	})

	return params
}

// extractParameterName extracts the name from a parameter node
func (ipa *InterproceduralAnalyzer) extractParameterName(node *sitter.Node, src []byte) string {
	nodeType := node.Type()

	// Direct identifier
	if nodeType == "identifier" || nodeType == "variable_name" {
		return string(src[node.StartByte():node.EndByte()])
	}

	// Look for identifier in children (for typed params like "int x")
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" || child.Type() == "variable_name" ||
			child.Type() == "name" {
			return string(src[child.StartByte():child.EndByte()])
		}
	}

	// Fallback: try to extract from text
	text := string(src[node.StartByte():node.EndByte()])
	parts := strings.Fields(text)
	for i := len(parts) - 1; i >= 0; i-- {
		if isValidIdentifier(parts[i]) {
			return parts[i]
		}
	}

	return ""
}

// analyzeFlowWithinFunction analyzes data flow within a function.
// localCallGraph is an accumulator for call-graph edges; it must not be nil.
func (ipa *InterproceduralAnalyzer) analyzeFlowWithinFunction(node *sitter.Node, src []byte, summary *FunctionSummary, localCallGraph map[string][]string) {
	// Find function body
	var body *sitter.Node
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()
		if strings.Contains(childType, "body") ||
			childType == "block" ||
			childType == "compound_statement" {
			body = child
			break
		}
	}

	if body == nil {
		return
	}

	// Traverse body looking for:
	// 1. Return statements
	// 2. Function calls
	// 3. Assignments involving parameters
	ipa.traverseForFlow(body, src, summary, localCallGraph)
}

// traverseForFlow traverses AST looking for flow patterns.
// localCallGraph accumulates call-graph edges; the caller (BuildFunctionSummary)
// commits them to ipa.graph in a single batched write, so no lock is acquired here.
func (ipa *InterproceduralAnalyzer) traverseForFlow(node *sitter.Node, src []byte, summary *FunctionSummary, localCallGraph map[string][]string) {
	if node == nil {
		return
	}

	nodeType := node.Type()
	nodeText := string(src[node.StartByte():node.EndByte()])

	// Check for return statements
	if strings.Contains(nodeType, "return") {
		// Check if any parameter is used in return value
		for i, param := range summary.Parameters {
			if strings.Contains(nodeText, param.Name) {
				// This parameter flows to return
				if !slices.Contains(summary.ParamsToReturn, i) {
					summary.ParamsToReturn = append(summary.ParamsToReturn, i)
				}
			}
		}
	}

	// Check for function calls
	if strings.Contains(nodeType, "call") {
		callName := ipa.extractCallName(node, src)
		if callName != "" && !slices.Contains(summary.CalledFunctions, callName) {
			summary.CalledFunctions = append(summary.CalledFunctions, callName)
		}

		// Accumulate into the local (unsynchronised) copy; the caller
		// commits it to ipa.graph in one batched write.
		localCallGraph[summary.Name] = append(localCallGraph[summary.Name], callName)
	}

	// Recurse into children
	for i := 0; i < int(node.ChildCount()); i++ {
		ipa.traverseForFlow(node.Child(i), src, summary, localCallGraph)
	}
}

// extractCallName extracts the function name from a call expression
func (ipa *InterproceduralAnalyzer) extractCallName(node *sitter.Node, src []byte) string {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()
		if childType == "identifier" || childType == "member_expression" ||
			childType == "selector_expression" || childType == "attribute" {
			text := string(src[child.StartByte():child.EndByte()])
			// Remove trailing parenthesis
			if idx := strings.Index(text, "("); idx > 0 {
				text = text[:idx]
			}
			return text
		}
	}
	return ""
}

// PropagateInterproceduralTaint propagates taint across function boundaries.
// visited is a caller-owned map used to prevent re-entrant processing of the
// same call site; the caller must allocate it (make(map[string]bool)) once per
// top-level propagation request and pass the same map on recursive invocations.
// Keeping visited outside the struct eliminates the data race that arose when
// ipa.visited was read and written without holding ipa.mu.
func (ipa *InterproceduralAnalyzer) PropagateInterproceduralTaint(callNode *sitter.Node, src []byte, filePath string, callerState *FullAnalysisState, visited map[string]bool) {
	if ipa.currentDepth >= ipa.maxDepth {
		return
	}

	funcName := ipa.extractCallName(callNode, src)
	if funcName == "" {
		return
	}

	// Check for recursion
	callKey := funcName + ":" + filePath
	if visited[callKey] {
		return
	}
	visited[callKey] = true
	defer func() { delete(visited, callKey) }()

	ipa.currentDepth++
	defer func() { ipa.currentDepth-- }()

	// Get function summary
	ipa.mu.RLock()
	summary, exists := ipa.summaries[funcName]
	ipa.mu.RUnlock()

	if !exists {
		return
	}

	// Check if any tainted argument maps to a return value
	args := ipa.extractCallArguments(callNode, src)

	for i, arg := range args {
		// Check if argument is tainted
		for _, tv := range callerState.TaintedVariables {
			if strings.Contains(arg, tv.Name) {
				// If this param flows to return, mark call result as tainted
				if slices.Contains(summary.ParamsToReturn, i) {
					// The call result is tainted
					// Find assignment target if this is part of an assignment
					if target := ipa.findAssignmentTarget(callNode, src); target != "" {
						newTV := &TaintedVariable{
							Name:     target,
							Scope:    scopeInterprocedural,
							Source:   tv.Source,
							Location: nodeToLocation(callNode, src, filePath),
							Depth:    tv.Depth + 1,
						}
						callerState.AddTaintedVariable(newTV)

						// Add propagation step
						step := PropagationStep{
							Type:     StepInterproceduralReturn,
							Variable: target,
							Function: funcName,
							Location: newTV.Location,
						}
						callerState.AddPropagationStep(tv.Source, step)
					}
				}

				// Mark function as tainted
				tf := &TaintedFunction{
					Name:     funcName,
					FilePath: summary.FilePath,
					Line:     int(callNode.StartPoint().Row) + 1,
					Language: summary.Language,
					TaintedParams: []TaintedParam{
						{
							Index:  i,
							Name:   summary.GetParamName(i),
							Source: tv.Source,
						},
					},
				}
				callerState.AddTaintedFunction(tf)
			}
		}
	}
}

// extractCallArguments extracts argument strings from a call
func (ipa *InterproceduralAnalyzer) extractCallArguments(node *sitter.Node, src []byte) []string {
	var args []string

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if strings.Contains(child.Type(), "argument") {
			ast.IterateChildren(child, func(arg *sitter.Node) {
				args = append(args, string(src[arg.StartByte():arg.EndByte()]))
			})
			break
		}
	}

	return args
}

// findAssignmentTarget finds assignment target if call is part of assignment
func (ipa *InterproceduralAnalyzer) findAssignmentTarget(node *sitter.Node, src []byte) string {
	parent := node.Parent()
	for parent != nil {
		parentType := parent.Type()
		if strings.Contains(parentType, "assignment") ||
			strings.Contains(parentType, "declarator") {
			// Find left-hand side
			for i := 0; i < int(parent.ChildCount()); i++ {
				child := parent.Child(i)
				childType := child.Type()
				if childType == "identifier" || childType == "variable_name" {
					return string(src[child.StartByte():child.EndByte()])
				}
			}
		}
		parent = parent.Parent()
	}
	return ""
}

// GetCallGraph returns a snapshot of the call graph (caller → []callee).
// The returned map is a deep copy; callers may read it without any lock.
func (ipa *InterproceduralAnalyzer) GetCallGraph() map[string][]string {
	return ipa.graph.snapshot()
}

// GetFunctionSummary returns a function summary by name
func (ipa *InterproceduralAnalyzer) GetFunctionSummary(name string) *FunctionSummary {
	ipa.mu.RLock()
	defer ipa.mu.RUnlock()
	return ipa.summaries[name]
}

// GetAllSummaries returns all function summaries
func (ipa *InterproceduralAnalyzer) GetAllSummaries() map[string]*FunctionSummary {
	ipa.mu.RLock()
	defer ipa.mu.RUnlock()

	// Return a copy
	summaries := make(map[string]*FunctionSummary)
	for k, v := range ipa.summaries {
		summaries[k] = v
	}
	return summaries
}

// GetParamName returns parameter name by index
func (fs *FunctionSummary) GetParamName(index int) string {
	if index >= 0 && index < len(fs.Parameters) {
		return fs.Parameters[index].Name
	}
	return ""
}

// Helper functions
func isValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}
	for i, c := range s {
		if i == 0 {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' || c == '$') {
				return false
			}
		} else {
			if !((c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') || c == '_' || c == '$') {
				return false
			}
		}
	}
	return true
}

// RunAnalysis performs cross-function taint analysis against the provided result,
// collecting all unique file paths, building function summaries, and iteratively
// propagating taint until a fixed point is reached.
func (ipa *InterproceduralAnalyzer) RunAnalysis(result *TraceResult) {
	if len(result.TaintedFunctions) == 0 && len(result.TaintedVariables) == 0 {
		return
	}

	// Step 1: Collect all unique file paths from results
	filePaths := make(map[string]bool)
	for _, src := range result.Sources {
		if src.Location.FilePath != "" {
			filePaths[src.Location.FilePath] = true
		}
	}
	for _, tv := range result.TaintedVariables {
		if tv.Location.FilePath != "" {
			filePaths[tv.Location.FilePath] = true
		}
	}
	for _, tf := range result.TaintedFunctions {
		if tf.FilePath != "" {
			filePaths[tf.FilePath] = true
		}
	}

	// Step 2: Build FullAnalysisState from existing results
	state := NewFullAnalysisState()
	for _, s := range result.Sources {
		state.AddSource(s)
	}
	for _, v := range result.TaintedVariables {
		state.AddTaintedVariable(v)
	}
	for _, f := range result.TaintedFunctions {
		state.AddTaintedFunction(f)
	}

	// Step 3: Build function summaries from parsed ASTs
	for filePath := range filePaths {
		parseResult, err := ipa.parser.ParseFile(filePath)
		if err != nil || parseResult == nil {
			continue
		}
		buildSummariesFromAST(ipa, parseResult.Root, parseResult.Source, filePath, parseResult.Language)
	}

	// Step 4: Iterative taint propagation (fixed-point loop)
	for depth := 0; depth < ipa.maxDepth; depth++ {
		prevCount := len(state.TaintedFunctions) + len(state.TaintedVariables)

		ipa.propagateReturnTaint(state, filePaths)
		ipa.propagateCallTaint(state, filePaths)

		newCount := len(state.TaintedFunctions) + len(state.TaintedVariables)
		if newCount == prevCount {
			break // Fixed point reached
		}
	}

	// Step 5: Update result with new findings
	result.TaintedFunctions = state.TaintedFunctions
	result.TaintedVariables = state.TaintedVariables
}

// astWalkState bundles per-file context shared by AST walker helpers,
// replacing the recurring (src, filePath, language, state) parameter group.
type astWalkState struct {
	src      []byte
	filePath string
	language string
	state    *FullAnalysisState
}

// buildSummariesFromAST walks an AST tree and builds function summaries for all
// function definitions found.
func buildSummariesFromAST(ipa *InterproceduralAnalyzer, node *sitter.Node, src []byte, filePath string, language string) {
	if node == nil {
		return
	}

	if ast.IsFunctionNodeForLanguage(node.Type(), language) {
		ipa.BuildFunctionSummary(node, src, filePath, language)
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		buildSummariesFromAST(ipa, node.Child(i), src, filePath, language)
	}
}

// propagateReturnTaint checks each tainted function: if its summary shows a
// tainted parameter flows to its return value, mark the function as returning
// tainted data, then scan call sites to taint the assigned return values.
func (ipa *InterproceduralAnalyzer) propagateReturnTaint(state *FullAnalysisState, filePaths map[string]bool) {
	// Build map of function name -> which param indices are tainted
	taintedParamsByFunc := make(map[string]map[int]*InputSource)
	for _, tf := range state.TaintedFunctions {
		if _, ok := taintedParamsByFunc[tf.Name]; !ok {
			taintedParamsByFunc[tf.Name] = make(map[int]*InputSource)
		}
		for _, tp := range tf.TaintedParams {
			taintedParamsByFunc[tf.Name][tp.Index] = tp.Source
		}
	}

	// For each function with tainted params, check if any tainted param flows to return
	for funcName, paramSources := range taintedParamsByFunc {
		summary := ipa.GetFunctionSummary(funcName)
		if summary == nil {
			continue
		}
		for _, returnParamIdx := range summary.ParamsToReturn {
			source, ok := paramSources[returnParamIdx]
			if !ok || source == nil {
				continue
			}
			state.AddReturnsTaintedFunction(funcName, source)
		}
	}

	if len(state.ReturnsTainted) == 0 {
		return
	}

	// Scan all files for call sites of functions that return tainted data
	for filePath := range filePaths {
		parseResult, err := ipa.parser.ParseFile(filePath)
		if err != nil || parseResult == nil {
			continue
		}
		ws := &astWalkState{src: parseResult.Source, filePath: filePath, language: parseResult.Language, state: state}
		findReturnTaintCallSites(parseResult.Root, ws)
	}
}

// findReturnTaintCallSites walks an AST looking for assignments where the RHS
// is a call to a function that returns tainted data.
func findReturnTaintCallSites(node *sitter.Node, ws *astWalkState) {
	if node == nil {
		return
	}

	if isAssignmentNode(node.Type()) {
		checkAssignmentForReturnTaint(node, ws)
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		findReturnTaintCallSites(node.Child(i), ws)
	}
}

// checkAssignmentForReturnTaint checks if an assignment's RHS calls a function
// that returns tainted data, and if so taints the LHS variable.
func checkAssignmentForReturnTaint(node *sitter.Node, ws *astWalkState) {
	src, filePath, language, state := ws.src, ws.filePath, ws.language, ws.state
	var lhsName string
	var rhsNode *sitter.Node

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()

		if lhsName == "" {
			if childType == "identifier" || childType == "variable_name" || childType == "name" {
				lhsName = string(src[child.StartByte():child.EndByte()])
			}
		} else if childType != "=" && childType != ":=" {
			rhsNode = child
			break
		}
	}

	if lhsName == "" || rhsNode == nil {
		return
	}

	callName := findCallInNode(rhsNode, src)
	if callName == "" {
		return
	}

	source, ok := state.ReturnsTainted[callName]
	if !ok || source == nil {
		return
	}

	tv := &TaintedVariable{
		ID:       uuid.New().String(),
		Name:     lhsName,
		Scope:    scopeInterprocedural,
		Source:   source,
		Location: nodeToLocation(node, src, filePath),
		Depth:    2,
		Language: language,
	}
	state.AddTaintedVariable(tv)
}

// findCallInNode finds a function call name within a node (recursively).
func findCallInNode(node *sitter.Node, src []byte) string {
	if node == nil {
		return ""
	}

	if isCallNode(node.Type()) {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			childType := child.Type()
			if childType == "identifier" || childType == "member_expression" ||
				childType == "selector_expression" || childType == "attribute" ||
				childType == "name" {
				text := string(src[child.StartByte():child.EndByte()])
				if idx := strings.Index(text, "("); idx > 0 {
					text = text[:idx]
				}
				return text
			}
		}
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		if name := findCallInNode(node.Child(i), src); name != "" {
			return name
		}
	}
	return ""
}

// propagateCallTaint re-scans files for function calls where arguments became
// newly tainted and creates new TaintedFunction entries.
func (ipa *InterproceduralAnalyzer) propagateCallTaint(state *FullAnalysisState, filePaths map[string]bool) {
	for filePath := range filePaths {
		parseResult, err := ipa.parser.ParseFile(filePath)
		if err != nil || parseResult == nil {
			continue
		}
		ws := &astWalkState{src: parseResult.Source, filePath: filePath, language: parseResult.Language, state: state}
		findNewTaintedCalls(parseResult.Root, ws)
	}
}

// findNewTaintedCalls walks an AST looking for function calls with tainted arguments.
func findNewTaintedCalls(node *sitter.Node, ws *astWalkState) {
	if node == nil {
		return
	}

	if isCallNode(node.Type()) {
		checkCallForTaintedArgs(node, ws)
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		findNewTaintedCalls(node.Child(i), ws)
	}
}

// checkCallForTaintedArgs checks if a function call has tainted arguments.
func checkCallForTaintedArgs(node *sitter.Node, ws *astWalkState) {
	src, filePath, language, state := ws.src, ws.filePath, ws.language, ws.state
	var funcName string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()
		if childType == "identifier" || childType == "member_expression" ||
			childType == "selector_expression" || childType == "attribute" ||
			childType == "name" {
			funcName = string(src[child.StartByte():child.EndByte()])
			if idx := strings.Index(funcName, "("); idx > 0 {
				funcName = funcName[:idx]
			}
			break
		}
	}
	if funcName == "" {
		return
	}

	// Extract arguments, skipping punctuation tokens
	var args []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if strings.Contains(child.Type(), "argument") {
			ast.IterateChildren(child, func(arg *sitter.Node) {
				args = append(args, string(src[arg.StartByte():arg.EndByte()]))
			})
			break
		}
	}

	var taintedParams []TaintedParam
	for i, argText := range args {
		for _, tv := range state.TaintedVariables {
			if strings.Contains(argText, tv.Name) {
				taintedParams = append(taintedParams, TaintedParam{
					Index:  i,
					Name:   argText,
					Source: tv.Source,
				})
				break
			}
		}
	}

	if len(taintedParams) > 0 {
		tf := &TaintedFunction{
			ID:            uuid.New().String(),
			Name:          funcName,
			FilePath:      filePath,
			Line:          int(node.StartPoint().Row) + 1,
			Language:      language,
			TaintedParams: taintedParams,
		}
		state.AddTaintedFunction(tf)
	}
}

// isAssignmentNode checks if a node type represents an assignment.
func isAssignmentNode(nodeType string) bool {
	return ast.IsAssignmentNode(nodeType)
}

// isCallNode checks if a node type represents a function call.
func isCallNode(nodeType string) bool {
	return ast.IsCallNode(nodeType)
}

