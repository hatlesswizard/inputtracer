package tracer

import (
	"strings"

	"github.com/google/uuid"
	"github.com/hatlesswizard/inputtracer/pkg/sources"
	sitter "github.com/smacker/go-tree-sitter"
)

// runInterproceduralAnalysisImpl performs cross-function taint analysis.
// It builds function summaries from all analyzed files, then propagates
// taint through return values transitively up to MaxDepth iterations.
func (t *Tracer) runInterproceduralAnalysisImpl(result *TraceResult) {
	if len(result.TaintedFunctions) == 0 && len(result.TaintedVariables) == 0 {
		return
	}

	maxDepth := t.config.MaxDepth
	if maxDepth <= 0 {
		maxDepth = 3
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

	// Step 3: Create InterproceduralAnalyzer and build function summaries
	analyzer := NewInterproceduralAnalyzer(state, maxDepth)

	for filePath := range filePaths {
		parseResult, err := t.parser.ParseFile(filePath)
		if err != nil || parseResult == nil {
			continue
		}
		buildSummariesFromAST(analyzer, parseResult.Root, parseResult.Source, filePath, parseResult.Language)
	}

	// Step 4: Iterative taint propagation (fixed-point loop)
	for depth := 0; depth < maxDepth; depth++ {
		prevCount := len(state.TaintedFunctions) + len(state.TaintedVariables)

		// For each tainted function, check if its summary shows param-to-return flow
		// and propagate taint to callers that use the return value
		t.propagateReturnTaint(analyzer, state, filePaths)

		// Check for new function calls with newly tainted arguments
		t.propagateCallTaint(state, filePaths)

		newCount := len(state.TaintedFunctions) + len(state.TaintedVariables)
		if newCount == prevCount {
			break // Fixed point reached
		}
	}

	// Step 5: Update result with new findings
	result.TaintedFunctions = state.TaintedFunctions
	result.TaintedVariables = state.TaintedVariables
}

// buildSummariesFromAST walks an AST tree and builds function summaries for all
// function definitions found.
func buildSummariesFromAST(analyzer *InterproceduralAnalyzer, node *sitter.Node, src []byte, filePath string, language string) {
	if node == nil {
		return
	}

	if sources.IsFunctionNodeForLanguage(node.Type(), language) {
		analyzer.BuildFunctionSummary(node, src, filePath, language)
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		buildSummariesFromAST(analyzer, node.Child(i), src, filePath, language)
	}
}

// propagateReturnTaint checks each tainted function: if its summary shows a
// tainted parameter flows to its return value, mark the function as returning
// tainted data, then scan call sites to taint the assigned return values.
func (t *Tracer) propagateReturnTaint(analyzer *InterproceduralAnalyzer, state *FullAnalysisState, filePaths map[string]bool) {
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
		summary := analyzer.GetFunctionSummary(funcName)
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
		parseResult, err := t.parser.ParseFile(filePath)
		if err != nil || parseResult == nil {
			continue
		}
		findReturnTaintCallSites(parseResult.Root, parseResult.Source, filePath, parseResult.Language, state)
	}
}

// findReturnTaintCallSites walks an AST looking for assignments where the RHS
// is a call to a function that returns tainted data.
func findReturnTaintCallSites(node *sitter.Node, src []byte, filePath string, language string, state *FullAnalysisState) {
	if node == nil {
		return
	}

	if isAssignmentNode(node.Type()) {
		checkAssignmentForReturnTaint(node, src, filePath, language, state)
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		findReturnTaintCallSites(node.Child(i), src, filePath, language, state)
	}
}

// checkAssignmentForReturnTaint checks if an assignment's RHS calls a function
// that returns tainted data, and if so taints the LHS variable.
func checkAssignmentForReturnTaint(node *sitter.Node, src []byte, filePath string, language string, state *FullAnalysisState) {
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
		Scope:    "interprocedural",
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
func (t *Tracer) propagateCallTaint(state *FullAnalysisState, filePaths map[string]bool) {
	for filePath := range filePaths {
		parseResult, err := t.parser.ParseFile(filePath)
		if err != nil || parseResult == nil {
			continue
		}
		findNewTaintedCalls(parseResult.Root, parseResult.Source, filePath, parseResult.Language, state)
	}
}

// findNewTaintedCalls walks an AST looking for function calls with tainted arguments.
func findNewTaintedCalls(node *sitter.Node, src []byte, filePath string, language string, state *FullAnalysisState) {
	if node == nil {
		return
	}

	if isCallNode(node.Type()) {
		checkCallForTaintedArgs(node, src, filePath, language, state)
	}

	for i := 0; i < int(node.ChildCount()); i++ {
		findNewTaintedCalls(node.Child(i), src, filePath, language, state)
	}
}

// checkCallForTaintedArgs checks if a function call has tainted arguments.
func checkCallForTaintedArgs(node *sitter.Node, src []byte, filePath string, language string, state *FullAnalysisState) {
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

	// Extract arguments
	var args []string
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if strings.Contains(child.Type(), "argument") {
			for j := 0; j < int(child.ChildCount()); j++ {
				arg := child.Child(j)
				argType := arg.Type()
				if argType != "," && argType != "(" && argType != ")" {
					args = append(args, string(src[arg.StartByte():arg.EndByte()]))
				}
			}
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
	for _, at := range sources.UniversalASTNodeTypes.AssignmentTypes {
		if nodeType == at {
			return true
		}
	}
	return false
}

// isCallNode checks if a node type represents a function call.
func isCallNode(nodeType string) bool {
	for _, ct := range sources.UniversalASTNodeTypes.CallTypes {
		if nodeType == ct {
			return true
		}
	}
	return false
}
