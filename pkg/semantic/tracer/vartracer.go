// Package tracer provides variable tracing across codebases
package tracer

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/discovery"
)

// VariableTracer traces a variable across all definitions in a codebase
type VariableTracer struct {
	parser     *sitter.Parser
	carrierMap *discovery.CarrierMap
	codebase   string
}

// VariableDefinition represents one definition of a variable
type VariableDefinition struct {
	File         string   `json:"file"`
	Line         int      `json:"line"`
	FunctionName string   `json:"function_name,omitempty"`
	ClassName    string   `json:"class_name,omitempty"`
	InitialValue string   `json:"initial_value"`
	CodeSnippet  string   `json:"code_snippet"`
}

// VariableTraceResult contains the complete trace for a variable in one file
type VariableTraceResult struct {
	File           string              `json:"file"`
	Line           int                 `json:"line"`
	FunctionName   string              `json:"function_name,omitempty"`
	ClassName      string              `json:"class_name,omitempty"`
	HasUserInput   bool                `json:"has_user_input"`
	InputSources   []string            `json:"input_sources,omitempty"`
	FlowPath       []string            `json:"flow_path,omitempty"`
	MatchedCarrier string              `json:"matched_carrier,omitempty"`
	Assignments    []Assignment        `json:"assignments,omitempty"`
	ParameterInfo  *ParameterTaintInfo `json:"parameter_info,omitempty"`
}

// Assignment represents one assignment to the variable
type Assignment struct {
	Line        int      `json:"line"`
	Expression  string   `json:"expression"`
	HasInput    bool     `json:"has_input"`
	Sources     []string `json:"sources,omitempty"`
}

// ParameterTaintInfo tracks taint propagation through function parameters
type ParameterTaintInfo struct {
	FunctionName   string         `json:"function_name"`
	ParameterName  string         `json:"parameter_name"`
	ParameterIndex int            `json:"parameter_index"`
	IsTainted      bool           `json:"is_tainted"`
	Sources        []string       `json:"sources,omitempty"`
	CallSites      []CallSiteInfo `json:"call_sites,omitempty"`
}

// CallSiteInfo represents a call site where a function is invoked
type CallSiteInfo struct {
	File     string   `json:"file"`
	Line     int      `json:"line"`
	Argument string   `json:"argument"`
	HasInput bool     `json:"has_input"`
	Sources  []string `json:"sources,omitempty"`
}

// TraceReport is the complete report for a variable
type TraceReport struct {
	Variable           string                `json:"variable"`
	Codebase           string                `json:"codebase"`
	TotalDefinitions   int                   `json:"total_definitions"`
	WithUserInput      int                   `json:"with_user_input"`
	WithoutUserInput   int                   `json:"without_user_input"`
	Definitions        []VariableTraceResult `json:"definitions"`
}

// NewVariableTracer creates a new variable tracer
func NewVariableTracer(codebase string, carrierMap *discovery.CarrierMap) *VariableTracer {
	parser := sitter.NewParser()
	parser.SetLanguage(php.GetLanguage())

	return &VariableTracer{
		parser:     parser,
		carrierMap: carrierMap,
		codebase:   codebase,
	}
}

// TraceVariable traces a variable across the entire codebase
func (t *VariableTracer) TraceVariable(varName string) (*TraceReport, error) {
	report := &TraceReport{
		Variable:    varName,
		Codebase:    t.codebase,
		Definitions: make([]VariableTraceResult, 0),
	}

	// Normalize variable name (ensure it starts with $)
	if !strings.HasPrefix(varName, "$") {
		varName = "$" + varName
	}

	// Find all definitions of this variable
	definitions, err := t.findAllDefinitions(varName)
	if err != nil {
		return nil, err
	}

	report.TotalDefinitions = len(definitions)

	// Trace each definition
	for _, def := range definitions {
		result := t.traceDefinition(varName, def)
		report.Definitions = append(report.Definitions, result)

		if result.HasUserInput {
			report.WithUserInput++
		} else {
			report.WithoutUserInput++
		}
	}

	return report, nil
}

// findAllDefinitions finds all places where the variable is initially assigned
func (t *VariableTracer) findAllDefinitions(varName string) ([]VariableDefinition, error) {
	var definitions []VariableDefinition

	// Pattern to match variable assignment: $varName =
	// We look for the first assignment in each function/scope
	assignPattern := regexp.MustCompile(regexp.QuoteMeta(varName) + `\s*=\s*`)

	err := filepath.Walk(t.codebase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".php") {
			return nil
		}

		// Skip vendor, cache directories
		if strings.Contains(path, "/vendor/") ||
			strings.Contains(path, "/cache/") ||
			strings.Contains(path, "/.git/") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Quick check if file contains the variable
		if !assignPattern.Match(content) {
			return nil
		}

		// Parse the file and find definitions
		fileDefs := t.findDefinitionsInFile(path, content, varName)
		definitions = append(definitions, fileDefs...)

		return nil
	})

	return definitions, err
}

// findDefinitionsInFile finds variable definitions in a single file
func (t *VariableTracer) findDefinitionsInFile(filePath string, content []byte, varName string) []VariableDefinition {
	var definitions []VariableDefinition

	tree, err := t.parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return definitions
	}
	defer tree.Close()

	lines := strings.Split(string(content), "\n")

	// Track which functions/scopes we've seen definitions in
	seenScopes := make(map[string]bool)

	// Walk the AST to find assignments
	t.walkForDefinitions(tree.RootNode(), content, lines, filePath, varName, "", "", &definitions, seenScopes)

	return definitions
}

// walkForDefinitions walks AST to find variable definitions
func (t *VariableTracer) walkForDefinitions(node *sitter.Node, content []byte, lines []string, filePath, varName, currentClass, currentFunc string, definitions *[]VariableDefinition, seenScopes map[string]bool) {
	if node == nil {
		return
	}

	nodeType := node.Type()

	// Track class context
	if nodeType == "class_declaration" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			currentClass = nameNode.Content(content)
		}
	}

	// Track function/method context
	if nodeType == "method_declaration" || nodeType == "function_definition" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil {
			currentFunc = nameNode.Content(content)
		}
	}

	// Check for assignment expression
	if nodeType == "assignment_expression" {
		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if leftNode != nil && rightNode != nil {
			leftContent := leftNode.Content(content)

			// Check if this is an assignment to our variable (exact match or concatenation)
			if leftContent == varName || strings.HasPrefix(leftContent, varName+"[") {
				scopeKey := filePath + ":" + currentClass + ":" + currentFunc

				// Only record the first definition per scope
				if !seenScopes[scopeKey] {
					seenScopes[scopeKey] = true

					line := int(node.StartPoint().Row) + 1
					snippet := ""
					if line > 0 && line <= len(lines) {
						snippet = strings.TrimSpace(lines[line-1])
					}

					def := VariableDefinition{
						File:         t.relativePath(filePath),
						Line:         line,
						FunctionName: currentFunc,
						ClassName:    currentClass,
						InitialValue: rightNode.Content(content),
						CodeSnippet:  snippet,
					}
					*definitions = append(*definitions, def)
				}
			}
		}
	}

	// Recurse into children
	for i := 0; i < int(node.ChildCount()); i++ {
		t.walkForDefinitions(node.Child(i), content, lines, filePath, varName, currentClass, currentFunc, definitions, seenScopes)
	}
}

// traceDefinition traces a single definition to see if user input flows into it
func (t *VariableTracer) traceDefinition(varName string, def VariableDefinition) VariableTraceResult {
	result := VariableTraceResult{
		File:         def.File,
		Line:         def.Line,
		FunctionName: def.FunctionName,
		ClassName:    def.ClassName,
		Assignments:  make([]Assignment, 0),
		InputSources: make([]string, 0),
		FlowPath:     make([]string, 0),
	}

	// Read the full file
	fullPath := filepath.Join(t.codebase, def.File)
	content, err := os.ReadFile(fullPath)
	if err != nil {
		return result
	}

	tree, err := t.parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return result
	}
	defer tree.Close()

	lines := strings.Split(string(content), "\n")

	// First, find all tainted variables in this file (variables assigned from user input)
	taintedVars := t.findTaintedVariables(tree.RootNode(), content)

	// If we're inside a function, check if parameters receive tainted input from call sites
	// and add them to taintedVars so subsequent checks find them
	if def.FunctionName != "" {
		scopeNode := t.findScopeNode(tree.RootNode(), content, def.FunctionName, def.ClassName)
		if scopeNode != nil {
			params := t.extractFunctionParameters(scopeNode, content)
			for paramIdx, paramName := range params {
				paramTaint := t.traceParameterSources(def.FunctionName, paramName, paramIdx)
				if paramTaint.IsTainted {
					// Add parameter to tainted vars so it's detected in expressions
					taintedVars[paramName] = paramTaint.Sources
				}
			}
		}
	}

	// Find all assignments to this variable in the same scope
	assignments := t.findAssignmentsInScope(tree.RootNode(), content, lines, varName, def.FunctionName, def.ClassName)

	// Check each assignment for user input (direct or via tainted variables)
	for i := range assignments {
		// Check direct input sources
		sources := t.findInputSources(assignments[i].Expression)

		// Also check if any tainted variable is used in the expression
		for taintedVar, taintSources := range taintedVars {
			if strings.Contains(assignments[i].Expression, taintedVar) {
				sources = append(sources, taintSources...)
				// Add the tainted variable itself as a flow step
				sources = append(sources, taintedVar+" (tainted)")
			}
		}

		if len(sources) > 0 {
			assignments[i].HasInput = true
			assignments[i].Sources = uniqueStrings(sources)
		}
	}

	// Inter-procedural analysis: Check if any assignment uses a function parameter
	if def.FunctionName != "" {
		scopeNode := t.findScopeNode(tree.RootNode(), content, def.FunctionName, def.ClassName)
		if scopeNode != nil {
			params := t.extractFunctionParameters(scopeNode, content)

			for i := range assignments {
				// Skip if already has input
				if assignments[i].HasInput {
					continue
				}

				// Check if expression contains a parameter
				for paramIdx, paramName := range params {
					if strings.Contains(assignments[i].Expression, paramName) {
						// This assignment uses a parameter - trace call sites
						paramTaint := t.traceParameterSources(def.FunctionName, paramName, paramIdx)

						if paramTaint.IsTainted {
							assignments[i].HasInput = true
							assignments[i].Sources = append(assignments[i].Sources, paramTaint.Sources...)
							assignments[i].Sources = append(assignments[i].Sources, paramName+" (param)")
							assignments[i].Sources = uniqueStrings(assignments[i].Sources)

							// Store parameter info in result
							result.ParameterInfo = &paramTaint
						}
						break
					}
				}
			}
		}
	}

	result.Assignments = assignments

	// Check each assignment for user input
	sourcesSeen := make(map[string]bool)
	for _, assign := range assignments {
		if assign.HasInput {
			result.HasUserInput = true
			for _, src := range assign.Sources {
				if !sourcesSeen[src] {
					sourcesSeen[src] = true
					result.InputSources = append(result.InputSources, src)
				}
			}
		}
	}

	// Build flow path
	if result.HasUserInput {
		result.FlowPath = t.buildFlowPath(varName, assignments)

		// Check if any source matches a carrier
		for _, src := range result.InputSources {
			if carrier := t.matchCarrier(src); carrier != "" {
				result.MatchedCarrier = carrier
				break
			}
		}
	}

	return result
}

// findTaintedVariables finds all variables in a file that are assigned from user input
func (t *VariableTracer) findTaintedVariables(root *sitter.Node, content []byte) map[string][]string {
	tainted := make(map[string][]string)

	t.walkForTaintedVars(root, content, tainted)

	return tainted
}

// walkForTaintedVars walks AST to find variables assigned from user input
func (t *VariableTracer) walkForTaintedVars(node *sitter.Node, content []byte, tainted map[string][]string) {
	if node == nil {
		return
	}

	if node.Type() == "assignment_expression" {
		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if leftNode != nil && rightNode != nil {
			leftContent := leftNode.Content(content)

			// Check if the right side IS a direct user input source (not just contains one)
			sources := t.findDirectInputSources(rightNode, content)
			if len(sources) > 0 {
				// This variable is tainted
				// Normalize: $view['conditions']['x'] -> $view['conditions']
				baseVar := extractBaseVariable(leftContent)
				if baseVar != "" {
					tainted[baseVar] = append(tainted[baseVar], sources...)
				}
			}
		}
	}

	// Recurse
	for i := 0; i < int(node.ChildCount()); i++ {
		t.walkForTaintedVars(node.Child(i), content, tainted)
	}
}

// findDirectInputSources checks if an AST node IS a direct user input source
// This is different from findInputSources which does string matching.
// Key principle: User input as a function ARGUMENT does NOT taint the return value.
func (t *VariableTracer) findDirectInputSources(node *sitter.Node, content []byte) []string {
	if node == nil {
		return nil
	}

	nodeType := node.Type()
	nodeContent := node.Content(content)

	// Case 1: Subscript expression on a superglobal: $_GET['x'], $_POST['y']
	if nodeType == "subscript_expression" {
		// In PHP tree-sitter, the base of subscript is the first child, not a named field
		var objNode *sitter.Node
		if node.ChildCount() > 0 {
			objNode = node.Child(0)
		}
		if objNode != nil {
			objContent := objNode.Content(content)
			for _, sg := range []string{"$_GET", "$_POST", "$_COOKIE", "$_REQUEST", "$_SERVER", "$_FILES", "$_ENV"} {
				if objContent == sg {
					return []string{sg}
				}
			}
			// Check for carrier property access: $mybb->input['x']
			if objNode.Type() == "member_access_expression" {
				sources := t.checkMemberAccessForInput(objNode, content)
				if len(sources) > 0 {
					return sources
				}
			}
		}
	}

	// Case 2: Member access that IS an input source: $mybb->input, $request->query
	if nodeType == "member_access_expression" {
		sources := t.checkMemberAccessForInput(node, content)
		if len(sources) > 0 {
			return sources
		}
	}

	// Case 3: Method call that IS an input function: $mybb->get_input('x'), $request->input('x')
	if nodeType == "member_call_expression" {
		sources := t.checkMethodCallForInput(node, content)
		if len(sources) > 0 {
			return sources
		}
		// If it's a method call but NOT an input function, return empty
		// (user input as argument does NOT taint return value)
		return nil
	}

	// Case 4: Regular function call - user input as argument does NOT taint return
	if nodeType == "function_call_expression" {
		// Check if the function itself is an input source (rare, but possible)
		funcNode := node.ChildByFieldName("function")
		if funcNode != nil {
			funcName := funcNode.Content(content)
			// Known input functions at global scope
			if funcName == "file_get_contents" {
				// Check if reading from php://input
				argsNode := node.ChildByFieldName("arguments")
				if argsNode != nil && strings.Contains(argsNode.Content(content), "php://input") {
					return []string{"php://input"}
				}
			}
		}
		// For other function calls, input as argument does NOT taint return
		return nil
	}

	// Case 5: Binary expression (concatenation, etc.) - check both sides
	if nodeType == "binary_expression" {
		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")
		var sources []string
		if leftNode != nil {
			sources = append(sources, t.findDirectInputSources(leftNode, content)...)
		}
		if rightNode != nil {
			sources = append(sources, t.findDirectInputSources(rightNode, content)...)
		}
		return sources
	}

	// Case 6: Parenthesized expression - unwrap
	if nodeType == "parenthesized_expression" {
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			if child.Type() != "(" && child.Type() != ")" {
				return t.findDirectInputSources(child, content)
			}
		}
	}

	// Case 7: Encapsed string (double-quoted with variables)
	if nodeType == "encapsed_string" {
		var sources []string
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			sources = append(sources, t.findDirectInputSources(child, content)...)
		}
		return sources
	}

	// Case 8: Variable - check if it's a direct superglobal reference
	if nodeType == "variable_name" {
		for _, sg := range []string{"$_GET", "$_POST", "$_COOKIE", "$_REQUEST", "$_SERVER", "$_FILES", "$_ENV"} {
			if nodeContent == sg {
				return []string{sg}
			}
		}
	}

	return nil
}

// checkMemberAccessForInput checks if a member access is an input source
func (t *VariableTracer) checkMemberAccessForInput(node *sitter.Node, content []byte) []string {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	propName := nameNode.Content(content)

	// Check against carrier map
	if t.carrierMap != nil {
		for _, carrier := range t.carrierMap.Carriers {
			if carrier.PropertyName == propName {
				return carrier.SourceTypes
			}
		}
	}

	// Common patterns without carrier map
	if propName == "input" || propName == "cookies" || propName == "query" || propName == "request" {
		return []string{"$" + propName}
	}

	return nil
}

// checkMethodCallForInput checks if a method call returns user input
func (t *VariableTracer) checkMethodCallForInput(node *sitter.Node, content []byte) []string {
	nameNode := node.ChildByFieldName("name")
	if nameNode == nil {
		return nil
	}
	methodName := nameNode.Content(content)

	// Known input-returning methods
	inputMethods := map[string][]string{
		"get_input": {"$_GET", "$_POST"},
		"input":     {"$_GET", "$_POST"},
		"get":       {"$_GET"},
		"post":      {"$_POST"},
		"cookie":    {"$_COOKIE"},
		"server":    {"$_SERVER"},
		"file":      {"$_FILES"},
	}

	if sources, ok := inputMethods[methodName]; ok {
		return sources
	}

	return nil
}

// extractJustVariable extracts just the variable name without any array indices
// e.g., $view['conditions']['username'] -> $view
func extractJustVariable(varExpr string) string {
	if !strings.HasPrefix(varExpr, "$") {
		return ""
	}

	// Find the first '[' or end of string
	idx := strings.Index(varExpr, "[")
	if idx == -1 {
		return varExpr
	}
	return varExpr[:idx]
}

// extractBaseVariable extracts the base variable from an array access
// e.g., $view['conditions']['username'] -> $view['conditions']
func extractBaseVariable(varExpr string) string {
	// Handle $var['key']['key2'] -> $var['key']
	// or $var['key'] -> $var['key']
	// or $var -> $var

	if !strings.HasPrefix(varExpr, "$") {
		return ""
	}

	// Find position up to second-to-last bracket
	brackets := 0
	lastBracket := -1
	for i, c := range varExpr {
		if c == '[' {
			brackets++
			if brackets == 1 {
				lastBracket = i
			}
		}
	}

	if brackets >= 1 && lastBracket > 0 {
		// Return up to and including first bracket group
		depth := 0
		for i := lastBracket; i < len(varExpr); i++ {
			if varExpr[i] == '[' {
				depth++
			} else if varExpr[i] == ']' {
				depth--
				if depth == 0 {
					return varExpr[:i+1]
				}
			}
		}
	}

	return varExpr
}

// uniqueStrings returns unique strings from a slice
func uniqueStrings(slice []string) []string {
	seen := make(map[string]bool)
	result := make([]string, 0)
	for _, s := range slice {
		if !seen[s] {
			seen[s] = true
			result = append(result, s)
		}
	}
	return result
}

// findAssignmentsInScope finds all assignments to variable in a specific scope
func (t *VariableTracer) findAssignmentsInScope(root *sitter.Node, content []byte, lines []string, varName, funcName, className string) []Assignment {
	var assignments []Assignment

	// Find the function/method node that matches our scope
	scopeNode := t.findScopeNode(root, content, funcName, className)
	if scopeNode == nil {
		scopeNode = root // Global scope
	}

	// Walk the scope to find all assignments
	t.walkForAssignments(scopeNode, content, lines, varName, &assignments)

	return assignments
}

// findScopeNode finds the AST node for a specific function/class scope
func (t *VariableTracer) findScopeNode(node *sitter.Node, content []byte, funcName, className string) *sitter.Node {
	if node == nil {
		return nil
	}

	nodeType := node.Type()

	// If looking for a method in a class
	if className != "" && nodeType == "class_declaration" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil && nameNode.Content(content) == className {
			// Found the class, now find the method
			if funcName != "" {
				return t.findMethodInClass(node, content, funcName)
			}
			return node
		}
	}

	// If looking for a function (not in a class)
	if className == "" && funcName != "" && nodeType == "function_definition" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil && nameNode.Content(content) == funcName {
			return node
		}
	}

	// Recurse
	for i := 0; i < int(node.ChildCount()); i++ {
		if found := t.findScopeNode(node.Child(i), content, funcName, className); found != nil {
			return found
		}
	}

	return nil
}

// findMethodInClass finds a method node within a class
func (t *VariableTracer) findMethodInClass(classNode *sitter.Node, content []byte, methodName string) *sitter.Node {
	for i := 0; i < int(classNode.ChildCount()); i++ {
		child := classNode.Child(i)
		if child.Type() == "declaration_list" {
			for j := 0; j < int(child.ChildCount()); j++ {
				member := child.Child(j)
				if member.Type() == "method_declaration" {
					nameNode := member.ChildByFieldName("name")
					if nameNode != nil && nameNode.Content(content) == methodName {
						return member
					}
				}
			}
		}
	}
	return nil
}

// extractFunctionParameters extracts parameter names from a function/method node
func (t *VariableTracer) extractFunctionParameters(funcNode *sitter.Node, content []byte) []string {
	var params []string

	if funcNode == nil {
		return params
	}

	// Find the formal_parameters or parameters node
	paramsNode := funcNode.ChildByFieldName("parameters")
	if paramsNode == nil {
		return params
	}

	// Walk through children to find simple_parameter nodes
	for i := 0; i < int(paramsNode.ChildCount()); i++ {
		child := paramsNode.Child(i)
		if child.Type() == "simple_parameter" || child.Type() == "property_promotion_parameter" {
			// Get the variable name from the parameter
			nameNode := child.ChildByFieldName("name")
			if nameNode != nil {
				params = append(params, nameNode.Content(content))
			}
		}
	}

	return params
}

// findFunctionCallSites finds all call sites of a function across the codebase
func (t *VariableTracer) findFunctionCallSites(funcName string) []CallSiteInfo {
	var callSites []CallSiteInfo

	// Create a pattern to quickly filter files
	callPattern := regexp.MustCompile(regexp.QuoteMeta(funcName) + `\s*\(`)

	filepath.Walk(t.codebase, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() || !strings.HasSuffix(strings.ToLower(path), ".php") {
			return nil
		}

		// Skip vendor, cache directories
		if strings.Contains(path, "/vendor/") ||
			strings.Contains(path, "/cache/") ||
			strings.Contains(path, "/.git/") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Quick check if file contains the function call
		if !callPattern.Match(content) {
			return nil
		}

		// Parse and find call sites
		tree, err := t.parser.ParseCtx(context.Background(), nil, content)
		if err != nil {
			return nil
		}
		defer tree.Close()

		relPath := t.relativePath(path)
		t.walkForCallSites(tree.RootNode(), content, relPath, funcName, &callSites)

		return nil
	})

	return callSites
}

// walkForCallSites walks AST to find function call sites
func (t *VariableTracer) walkForCallSites(node *sitter.Node, content []byte, filePath, funcName string, callSites *[]CallSiteInfo) {
	if node == nil {
		return
	}

	nodeType := node.Type()

	// Check for function_call_expression: funcName($arg)
	if nodeType == "function_call_expression" {
		funcNode := node.ChildByFieldName("function")
		if funcNode != nil && funcNode.Content(content) == funcName {
			t.extractCallSite(node, content, filePath, callSites)
		}
	}

	// Check for member_call_expression: $obj->funcName($arg)
	if nodeType == "member_call_expression" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil && nameNode.Content(content) == funcName {
			t.extractCallSite(node, content, filePath, callSites)
		}
	}

	// Check for scoped_call_expression: self::funcName($arg) or ClassName::funcName($arg)
	if nodeType == "scoped_call_expression" {
		nameNode := node.ChildByFieldName("name")
		if nameNode != nil && nameNode.Content(content) == funcName {
			t.extractCallSite(node, content, filePath, callSites)
		}
	}

	// Recurse
	for i := 0; i < int(node.ChildCount()); i++ {
		t.walkForCallSites(node.Child(i), content, filePath, funcName, callSites)
	}
}

// extractCallSite extracts call site information from a call expression node
func (t *VariableTracer) extractCallSite(callNode *sitter.Node, content []byte, filePath string, callSites *[]CallSiteInfo) {
	line := int(callNode.StartPoint().Row) + 1

	// Find arguments node
	argsNode := callNode.ChildByFieldName("arguments")
	if argsNode == nil {
		return
	}

	// Find tainted variables in this file to check if arguments are tainted
	tree, err := t.parser.ParseCtx(context.Background(), nil, content)
	var fileTaintedVars map[string][]string
	if err == nil {
		fileTaintedVars = t.findTaintedVariables(tree.RootNode(), content)
		tree.Close()
	}

	// Extract each argument
	argIndex := 0
	for i := 0; i < int(argsNode.ChildCount()); i++ {
		child := argsNode.Child(i)
		// Skip commas, parentheses
		if child.Type() == "argument" || (child.Type() != "," && child.Type() != "(" && child.Type() != ")") {
			argContent := child.Content(content)

			// Skip empty arguments
			if strings.TrimSpace(argContent) == "" || argContent == "(" || argContent == ")" || argContent == "," {
				continue
			}

			// Check if this argument contains user input (direct patterns)
			sources := t.findInputSources(argContent)

			// Also check if argument is a tainted variable from this file
			if fileTaintedVars != nil {
				// Extract base variable from argument (e.g., "$admin_view" from "$admin_view['x']")
				argBaseVar := extractJustVariable(argContent)

				for taintedVar, taintSources := range fileTaintedVars {
					// Extract base from tainted var too
					taintedBaseVar := extractJustVariable(taintedVar)

					// Match if:
					// 1. Argument equals tainted var: $admin_view == $admin_view
					// 2. Argument is a sub-access of tainted: $admin_view['x']['y'] starts with $admin_view['x']
					// 3. Tainted is a sub-access of argument: $admin_view['x'] when passing $admin_view
					// 4. Same base variable: $admin_view['x'] and $admin_view['y'] both taint $admin_view
					if argContent == taintedVar ||
						strings.HasPrefix(argContent, taintedVar) ||
						strings.HasPrefix(taintedVar, argContent) ||
						(argBaseVar != "" && taintedBaseVar != "" && argBaseVar == taintedBaseVar) {
						sources = append(sources, taintSources...)
						sources = append(sources, taintedVar+" (tainted)")
					}
				}
			}

			hasInput := len(sources) > 0
			sources = uniqueStrings(sources)

			callSite := CallSiteInfo{
				File:     filePath,
				Line:     line,
				Argument: argContent,
				HasInput: hasInput,
				Sources:  sources,
			}

			// Store with argument index
			if argIndex < len(*callSites) || true {
				*callSites = append(*callSites, callSite)
			}
			argIndex++
		}
	}
}

// traceParameterSources traces whether a function parameter receives tainted input at call sites
func (t *VariableTracer) traceParameterSources(funcName, paramName string, paramIndex int) ParameterTaintInfo {
	result := ParameterTaintInfo{
		FunctionName:   funcName,
		ParameterName:  paramName,
		ParameterIndex: paramIndex,
		CallSites:      make([]CallSiteInfo, 0),
		Sources:        make([]string, 0),
	}

	// Find all call sites
	allCallSites := t.findFunctionCallSites(funcName)

	// Group call sites by file:line and extract the argument at paramIndex
	callSiteMap := make(map[string][]CallSiteInfo)
	for _, cs := range allCallSites {
		key := cs.File + ":" + itoa(cs.Line)
		callSiteMap[key] = append(callSiteMap[key], cs)
	}

	// For each call site location, get the argument at the right index
	sourcesSeen := make(map[string]bool)
	for _, callSiteArgs := range callSiteMap {
		if paramIndex < len(callSiteArgs) {
			arg := callSiteArgs[paramIndex]
			result.CallSites = append(result.CallSites, arg)

			if arg.HasInput {
				result.IsTainted = true
				for _, src := range arg.Sources {
					if !sourcesSeen[src] {
						sourcesSeen[src] = true
						result.Sources = append(result.Sources, src)
					}
				}
			}
		}
	}

	return result
}

// walkForAssignments finds all assignments to a variable
func (t *VariableTracer) walkForAssignments(node *sitter.Node, content []byte, lines []string, varName string, assignments *[]Assignment) {
	if node == nil {
		return
	}

	if node.Type() == "assignment_expression" {
		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if leftNode != nil && rightNode != nil {
			leftContent := leftNode.Content(content)

			// Check if assigning to our variable (exact or concatenation .=)
			if leftContent == varName || strings.HasPrefix(leftContent, varName) {
				line := int(node.StartPoint().Row) + 1
				rightContent := rightNode.Content(content)

				assign := Assignment{
					Line:       line,
					Expression: rightContent,
				}

				// Check if the right side contains user input
				sources := t.findInputSources(rightContent)
				if len(sources) > 0 {
					assign.HasInput = true
					assign.Sources = sources
				}

				*assignments = append(*assignments, assign)
			}
		}
	}

	// Also check for compound assignment (.=)
	if node.Type() == "augmented_assignment_expression" {
		leftNode := node.ChildByFieldName("left")
		rightNode := node.ChildByFieldName("right")

		if leftNode != nil && rightNode != nil {
			leftContent := leftNode.Content(content)

			if leftContent == varName {
				line := int(node.StartPoint().Row) + 1
				rightContent := rightNode.Content(content)

				assign := Assignment{
					Line:       line,
					Expression: rightContent,
				}

				sources := t.findInputSources(rightContent)
				if len(sources) > 0 {
					assign.HasInput = true
					assign.Sources = sources
				}

				*assignments = append(*assignments, assign)
			}
		}
	}

	// Recurse
	for i := 0; i < int(node.ChildCount()); i++ {
		t.walkForAssignments(node.Child(i), content, lines, varName, assignments)
	}
}

// findInputSources checks an expression for user input sources
func (t *VariableTracer) findInputSources(expr string) []string {
	var sources []string
	seen := make(map[string]bool)

	// Check for direct superglobals
	superglobals := []string{"$_GET", "$_POST", "$_COOKIE", "$_REQUEST", "$_SERVER", "$_FILES", "$_ENV", "$_SESSION"}
	for _, sg := range superglobals {
		if strings.Contains(expr, sg) {
			if !seen[sg] {
				seen[sg] = true
				sources = append(sources, sg)
			}
		}
	}

	// Check for known carriers from carrier map
	if t.carrierMap != nil {
		for _, carrier := range t.carrierMap.Carriers {
			var pattern string
			if carrier.PropertyName != "" {
				// Match $var->property or $var->property[
				pattern = `\$\w+->\s*` + regexp.QuoteMeta(carrier.PropertyName)
			} else if carrier.MethodName != "" {
				pattern = `\$\w+->\s*` + regexp.QuoteMeta(carrier.MethodName) + `\s*\(`
			}

			if pattern != "" {
				re := regexp.MustCompile(pattern)
				if re.MatchString(expr) {
					carrierName := carrier.ClassName + "->" + carrier.PropertyName
					if carrier.MethodName != "" {
						carrierName = carrier.ClassName + "->" + carrier.MethodName + "()"
					}
					if !seen[carrierName] {
						seen[carrierName] = true
						sources = append(sources, carrierName)
						// Also add the underlying source types
						for _, st := range carrier.SourceTypes {
							if !seen[st] {
								seen[st] = true
								sources = append(sources, st)
							}
						}
					}
				}
			}
		}
	}

	// Check for common input patterns (without carrier map)
	inputPatterns := []struct {
		pattern string
		name    string
	}{
		{`\$mybb->input\[`, "$mybb->input"},
		{`\$mybb->cookies\[`, "$mybb->cookies"},
		{`\$mybb->get_input\(`, "$mybb->get_input()"},
		{`\$request->input\(`, "$request->input()"},
		{`\$request->get\(`, "$request->get()"},
		{`\$request->post\(`, "$request->post()"},
		{`\$request->query\[`, "$request->query"},
		{`\$_REQUEST\[`, "$_REQUEST"},
	}

	for _, p := range inputPatterns {
		re := regexp.MustCompile(p.pattern)
		if re.MatchString(expr) {
			if !seen[p.name] {
				seen[p.name] = true
				sources = append(sources, p.name)
			}
		}
	}

	return sources
}

// matchCarrier checks if a source matches a known carrier
func (t *VariableTracer) matchCarrier(source string) string {
	if t.carrierMap == nil {
		return ""
	}

	for _, carrier := range t.carrierMap.Carriers {
		carrierName := carrier.ClassName + "->" + carrier.PropertyName
		if carrier.MethodName != "" {
			carrierName = carrier.ClassName + "->" + carrier.MethodName + "()"
		}
		if strings.Contains(source, carrier.ClassName) {
			return carrierName
		}
	}
	return ""
}

// buildFlowPath creates a human-readable flow path
func (t *VariableTracer) buildFlowPath(varName string, assignments []Assignment) []string {
	var path []string

	for _, assign := range assignments {
		if assign.HasInput {
			for _, src := range assign.Sources {
				path = append(path, src+" → "+varName)
			}
		}
	}

	return path
}

// relativePath converts absolute path to relative
func (t *VariableTracer) relativePath(absPath string) string {
	rel, err := filepath.Rel(t.codebase, absPath)
	if err != nil {
		return absPath
	}
	return rel
}

// Summary returns a human-readable summary
func (r *TraceReport) Summary() string {
	var sb strings.Builder

	sb.WriteString("Variable Trace Report\n")
	sb.WriteString("=====================\n")
	sb.WriteString("Variable: " + r.Variable + "\n")
	sb.WriteString("Codebase: " + r.Codebase + "\n\n")
	sb.WriteString("Statistics:\n")
	sb.WriteString("  - Total Definitions: " + itoa(r.TotalDefinitions) + "\n")
	sb.WriteString("  - With User Input: " + itoa(r.WithUserInput) + "\n")
	sb.WriteString("  - Without User Input: " + itoa(r.WithoutUserInput) + "\n\n")

	sb.WriteString("Definitions:\n")
	for _, def := range r.Definitions {
		status := "✗ NO USER INPUT"
		if def.HasUserInput {
			status = "✓ HAS USER INPUT"
		}

		sb.WriteString("\n  " + def.File + ":" + itoa(def.Line) + " [" + status + "]\n")
		if def.FunctionName != "" {
			if def.ClassName != "" {
				sb.WriteString("    Scope: " + def.ClassName + "::" + def.FunctionName + "()\n")
			} else {
				sb.WriteString("    Scope: " + def.FunctionName + "()\n")
			}
		}

		if def.HasUserInput {
			sb.WriteString("    Sources: " + strings.Join(def.InputSources, ", ") + "\n")

			// Show parameter tracing info if available
			if def.ParameterInfo != nil && def.ParameterInfo.IsTainted {
				sb.WriteString("    Parameter: " + def.ParameterInfo.ParameterName + " receives user input from:\n")
				for _, cs := range def.ParameterInfo.CallSites {
					if cs.HasInput {
						sb.WriteString("      - " + cs.File + ":" + itoa(cs.Line) + ": " + cs.Argument + "\n")
					}
				}
			}

			if len(def.FlowPath) > 0 {
				sb.WriteString("    Flow: " + strings.Join(def.FlowPath, " | ") + "\n")
			}
		}
	}

	return sb.String()
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}
