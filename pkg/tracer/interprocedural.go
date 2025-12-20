package tracer

import (
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
)

// InterproceduralAnalyzer handles cross-function taint analysis
type InterproceduralAnalyzer struct {
	state        *FullAnalysisState
	summaries    map[string]*FunctionSummary // function name -> summary
	callGraph    map[string][]string         // caller -> callees
	maxDepth     int
	currentDepth int
	visited      map[string]bool // Prevent infinite recursion
	mu           sync.RWMutex
}

// NewInterproceduralAnalyzer creates a new inter-procedural analyzer
func NewInterproceduralAnalyzer(state *FullAnalysisState, maxDepth int) *InterproceduralAnalyzer {
	return &InterproceduralAnalyzer{
		state:     state,
		summaries: make(map[string]*FunctionSummary),
		callGraph: make(map[string][]string),
		maxDepth:  maxDepth,
		visited:   make(map[string]bool),
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
	ipa.analyzeFlowWithinFunction(node, src, summary)

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

	// Extract each parameter
	for i := 0; i < int(paramsNode.ChildCount()); i++ {
		param := paramsNode.Child(i)
		paramType := param.Type()

		// Skip punctuation
		if paramType == "," || paramType == "(" || paramType == ")" {
			continue
		}

		// Get parameter name
		name := ipa.extractParameterName(param, src)
		if name != "" {
			params = append(params, ParameterInfo{
				Index: len(params),
				Name:  name,
			})
		}
	}

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

// analyzeFlowWithinFunction analyzes data flow within a function
func (ipa *InterproceduralAnalyzer) analyzeFlowWithinFunction(node *sitter.Node, src []byte, summary *FunctionSummary) {
	// Track which parameters flow where
	paramFlows := make(map[string]bool) // param name -> has been used

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
	ipa.traverseForFlow(body, src, summary, paramFlows)
}

// traverseForFlow traverses AST looking for flow patterns
func (ipa *InterproceduralAnalyzer) traverseForFlow(node *sitter.Node, src []byte, summary *FunctionSummary, paramFlows map[string]bool) {
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
				if !containsInt(summary.ParamsToReturn, i) {
					summary.ParamsToReturn = append(summary.ParamsToReturn, i)
				}
			}
		}
	}

	// Check for function calls
	if strings.Contains(nodeType, "call") {
		callName := ipa.extractCallName(node, src)
		if callName != "" && !containsString(summary.CalledFunctions, callName) {
			summary.CalledFunctions = append(summary.CalledFunctions, callName)
		}

		// Add to call graph
		ipa.mu.Lock()
		ipa.callGraph[summary.Name] = append(ipa.callGraph[summary.Name], callName)
		ipa.mu.Unlock()

		// Check if parameters flow to this call
		for _, param := range summary.Parameters {
			if strings.Contains(nodeText, param.Name) {
				// Track param -> call param flow (simplified)
				paramFlows[param.Name] = true
			}
		}
	}

	// Recurse into children
	for i := 0; i < int(node.ChildCount()); i++ {
		ipa.traverseForFlow(node.Child(i), src, summary, paramFlows)
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

// PropagateInterproceduralTaint propagates taint across function boundaries
func (ipa *InterproceduralAnalyzer) PropagateInterproceduralTaint(callNode *sitter.Node, src []byte, filePath string, callerState *FullAnalysisState) {
	if ipa.currentDepth >= ipa.maxDepth {
		return
	}

	funcName := ipa.extractCallName(callNode, src)
	if funcName == "" {
		return
	}

	// Check for recursion
	callKey := funcName + ":" + filePath
	if ipa.visited[callKey] {
		return
	}
	ipa.visited[callKey] = true
	defer func() { delete(ipa.visited, callKey) }()

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
				if containsInt(summary.ParamsToReturn, i) {
					// The call result is tainted
					// Find assignment target if this is part of an assignment
					if target := ipa.findAssignmentTarget(callNode, src); target != "" {
						newTV := &TaintedVariable{
							Name:     target,
							Scope:    "interprocedural",
							Source:   tv.Source,
							Location: nodeToLocation(callNode, src, filePath),
							Depth:    tv.Depth + 1,
						}
						callerState.AddTaintedVariable(newTV)

						// Add propagation step
						step := PropagationStep{
							Type:     "interprocedural_return",
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
		childType := child.Type()
		if strings.Contains(childType, "argument") {
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

// GetCallGraph returns the call graph
func (ipa *InterproceduralAnalyzer) GetCallGraph() map[string][]string {
	ipa.mu.RLock()
	defer ipa.mu.RUnlock()

	// Return a copy
	graph := make(map[string][]string)
	for k, v := range ipa.callGraph {
		graph[k] = append([]string{}, v...)
	}
	return graph
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

func containsInt(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

func containsString(slice []string, val string) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}
