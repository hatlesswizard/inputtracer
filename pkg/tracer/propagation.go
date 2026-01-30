package tracer

import (
	"regexp"
	"strings"
	"sync"

	"github.com/hatlesswizard/inputtracer/pkg/sources"
	"github.com/hatlesswizard/inputtracer/pkg/sources/patterns"
	sitter "github.com/smacker/go-tree-sitter"
)

// regexCache caches compiled regex patterns for O(1) reuse
var regexCache sync.Map // map[string]*regexp.Regexp

// getOrCompileRegex gets a cached regex or compiles and caches it
func getOrCompileRegex(pattern string) *regexp.Regexp {
	if cached, ok := regexCache.Load(pattern); ok {
		return cached.(*regexp.Regexp)
	}
	compiled := regexp.MustCompile(pattern)
	regexCache.Store(pattern, compiled)
	return compiled
}

// TaintPropagator handles taint propagation through code
type TaintPropagator struct {
	state    *FullAnalysisState
	language string
}

// NewTaintPropagator creates a new taint propagator
func NewTaintPropagator(state *FullAnalysisState, language string) *TaintPropagator {
	return &TaintPropagator{
		state:    state,
		language: language,
	}
}

// PropagateFromAssignment propagates taint from an assignment expression
func (tp *TaintPropagator) PropagateFromAssignment(node *sitter.Node, src []byte, filePath string) {
	// Find assignment target and value
	target, value := tp.extractAssignmentParts(node, src)
	if target == "" || value == "" {
		return
	}

	// Check if value is tainted
	if taintInfo := tp.checkTainted(value, node, src); taintInfo != nil {
		// Create new tainted variable
		tv := &TaintedVariable{
			Name:     target,
			Scope:    tp.getCurrentScope(node),
			Source:   taintInfo.Source,
			Location: nodeToLocation(node, src, filePath),
			Depth:    taintInfo.Depth + 1,
		}

		// Add to state
		tp.state.AddTaintedVariable(tv)

		// Create propagation step
		step := PropagationStep{
			Type:     "assignment",
			Variable: target,
			Location: tv.Location,
		}
		tp.state.AddPropagationStep(taintInfo.Source, step)
	}
}

// PropagateFromFunctionCall propagates taint through function calls
func (tp *TaintPropagator) PropagateFromFunctionCall(node *sitter.Node, src []byte, filePath string) {
	funcName := tp.extractFunctionName(node, src)
	if funcName == "" {
		return
	}

	args := tp.extractArguments(node, src)

	// Check each argument for taint
	for i, arg := range args {
		if taintInfo := tp.checkTainted(arg.Text, node, src); taintInfo != nil {
			// Record that this function receives tainted input
			tf := &TaintedFunction{
				Name:     funcName,
				FilePath: filePath,
				Line:     int(node.StartPoint().Row) + 1,
				Language: tp.language,
				TaintedParams: []TaintedParam{
					{
						Index:  i,
						Name:   arg.Name,
						Source: taintInfo.Source,
					},
				},
			}

			tp.state.AddTaintedFunction(tf)

			// Create propagation step
			step := PropagationStep{
				Type:     "parameter_pass",
				Variable: arg.Text,
				Function: funcName,
				Location: nodeToLocation(node, src, filePath),
			}
			tp.state.AddPropagationStep(taintInfo.Source, step)
		}
	}
}

// PropagateFromReturn propagates taint from return statements
func (tp *TaintPropagator) PropagateFromReturn(node *sitter.Node, src []byte, filePath string) {
	// Extract return value
	returnValue := tp.extractReturnValue(node, src)
	if returnValue == "" {
		return
	}

	// Check if return value is tainted
	if taintInfo := tp.checkTainted(returnValue, node, src); taintInfo != nil {
		// Find containing function
		funcNode := tp.findContainingFunction(node)
		if funcNode == nil {
			return
		}

		funcName := tp.extractFunctionNameFromDef(funcNode, src)

		// Mark function as returning tainted data
		tp.state.AddReturnsTaintedFunction(funcName, taintInfo.Source)

		// Create propagation step
		step := PropagationStep{
			Type:     "return",
			Variable: returnValue,
			Function: funcName,
			Location: nodeToLocation(node, src, filePath),
		}
		tp.state.AddPropagationStep(taintInfo.Source, step)
	}
}

// TaintInfo contains information about a tainted value
type TaintInfo struct {
	Source *InputSource
	Depth  int
}

// checkTainted checks if a value is tainted
func (tp *TaintPropagator) checkTainted(value string, node *sitter.Node, src []byte) *TaintInfo {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	// Check if it's a known tainted variable
	for _, tv := range tp.state.TaintedVariables {
		if tp.matchesVariable(value, tv.Name) {
			return &TaintInfo{
				Source: tv.Source,
				Depth:  tv.Depth,
			}
		}
	}

	// Check if it's a direct input source (handled elsewhere)
	return nil
}

// matchesVariable checks if a value references a tainted variable.
// Handles exact matches, property access (var.prop, var['prop']),
// and expression containment with boundary-aware matching.
func (tp *TaintPropagator) matchesVariable(value, varName string) bool {
	// Exact match
	if value == varName {
		return true
	}

	// Check for property access (var.prop or var['prop'])
	if strings.HasPrefix(value, varName+".") ||
		strings.HasPrefix(value, varName+"[") {
		return true
	}

	// Check if variable is used in expression
	// Use boundary matching that handles $-prefixed and @-prefixed vars
	pattern := patterns.VariableBoundaryPattern(varName)
	return getOrCompileRegex(pattern).MatchString(value)
}

// Argument represents a function argument
type Argument struct {
	Name  string
	Text  string
	Index int
}

// extractAssignmentParts extracts target and value from an assignment
func (tp *TaintPropagator) extractAssignmentParts(node *sitter.Node, src []byte) (target, value string) {
	nodeType := node.Type()

	switch tp.language {
	case "php":
		return tp.extractPHPAssignment(node, src)
	case "javascript", "typescript", "tsx":
		return tp.extractJSAssignment(node, src)
	case "python":
		return tp.extractPythonAssignment(node, src)
	case "go":
		return tp.extractGoAssignment(node, src)
	case "java":
		return tp.extractJavaAssignment(node, src)
	case "c", "cpp":
		return tp.extractCAssignment(node, src)
	case "c_sharp":
		return tp.extractCSharpAssignment(node, src)
	case "ruby":
		return tp.extractRubyAssignment(node, src)
	case "rust":
		return tp.extractRustAssignment(node, src)
	}

	// Generic fallback - look for = operator
	text := string(src[node.StartByte():node.EndByte()])
	if strings.Contains(nodeType, "assignment") || strings.Contains(text, "=") {
		parts := strings.SplitN(text, "=", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}

	return "", ""
}

// Language-specific assignment extractors
func (tp *TaintPropagator) extractPHPAssignment(node *sitter.Node, src []byte) (string, string) {
	// PHP: $var = value;
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "variable_name" || child.Type() == "member_access_expression" {
			target := string(src[child.StartByte():child.EndByte()])
			// Find value (sibling after =)
			for j := i + 1; j < int(node.ChildCount()); j++ {
				sibling := node.Child(j)
				if sibling.Type() != "=" {
					value := string(src[sibling.StartByte():sibling.EndByte()])
					return target, value
				}
			}
		}
	}
	return "", ""
}

func (tp *TaintPropagator) extractJSAssignment(node *sitter.Node, src []byte) (string, string) {
	// JS: let/const/var name = value; or name = value;
	nodeType := node.Type()

	if nodeType == "variable_declarator" || nodeType == "assignment_expression" {
		var target, value string
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			childType := child.Type()
			if childType == "identifier" || childType == "member_expression" {
				if target == "" {
					target = string(src[child.StartByte():child.EndByte()])
				}
			} else if childType != "=" && childType != "type_annotation" && target != "" {
				value = string(src[child.StartByte():child.EndByte()])
				break
			}
		}
		return target, value
	}
	return "", ""
}

func (tp *TaintPropagator) extractPythonAssignment(node *sitter.Node, src []byte) (string, string) {
	// Python: name = value
	if node.Type() == "assignment" || node.Type() == "augmented_assignment" {
		var target, value string
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			childType := child.Type()
			if childType == "identifier" || childType == "attribute" || childType == "subscript" {
				if target == "" {
					target = string(src[child.StartByte():child.EndByte()])
				}
			} else if childType != "=" && childType != "type" && target != "" {
				value = string(src[child.StartByte():child.EndByte()])
				break
			}
		}
		return target, value
	}
	return "", ""
}

func (tp *TaintPropagator) extractGoAssignment(node *sitter.Node, src []byte) (string, string) {
	// Go: name := value or name = value
	nodeType := node.Type()
	if nodeType == "short_var_declaration" || nodeType == "assignment_statement" {
		text := string(src[node.StartByte():node.EndByte()])
		// Handle := and =
		if strings.Contains(text, ":=") {
			parts := strings.SplitN(text, ":=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(text, "=") {
			parts := strings.SplitN(text, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			}
		}
	}
	return "", ""
}

func (tp *TaintPropagator) extractJavaAssignment(node *sitter.Node, src []byte) (string, string) {
	// Java: Type name = value; or name = value;
	text := string(src[node.StartByte():node.EndByte()])
	if strings.Contains(text, "=") {
		parts := strings.SplitN(text, "=", 2)
		if len(parts) == 2 {
			// Left side might be "Type name" or just "name"
			leftParts := strings.Fields(strings.TrimSpace(parts[0]))
			target := leftParts[len(leftParts)-1]
			return target, strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
		}
	}
	return "", ""
}

func (tp *TaintPropagator) extractCAssignment(node *sitter.Node, src []byte) (string, string) {
	// C/C++: type name = value; or name = value;
	return tp.extractJavaAssignment(node, src) // Similar pattern
}

func (tp *TaintPropagator) extractCSharpAssignment(node *sitter.Node, src []byte) (string, string) {
	// C#: Type name = value; or name = value;
	return tp.extractJavaAssignment(node, src) // Similar pattern
}

func (tp *TaintPropagator) extractRubyAssignment(node *sitter.Node, src []byte) (string, string) {
	// Ruby: name = value
	text := string(src[node.StartByte():node.EndByte()])
	if strings.Contains(text, "=") && !strings.Contains(text, "==") {
		parts := strings.SplitN(text, "=", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}
	return "", ""
}

func (tp *TaintPropagator) extractRustAssignment(node *sitter.Node, src []byte) (string, string) {
	// Rust: let name = value; or let mut name = value;
	text := string(src[node.StartByte():node.EndByte()])
	if strings.Contains(text, "=") {
		// Remove "let" and "mut" keywords
		text = strings.TrimPrefix(strings.TrimSpace(text), "let")
		text = strings.TrimPrefix(strings.TrimSpace(text), "mut")
		parts := strings.SplitN(text, "=", 2)
		if len(parts) == 2 {
			// Handle type annotations
			target := strings.TrimSpace(parts[0])
			if colonIdx := strings.Index(target, ":"); colonIdx > 0 {
				target = strings.TrimSpace(target[:colonIdx])
			}
			return target, strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
		}
	}
	return "", ""
}

// extractFunctionName extracts the function name from a call expression
func (tp *TaintPropagator) extractFunctionName(node *sitter.Node, src []byte) string {
	// Look for identifier or member expression in first child
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()
		if childType == "identifier" || childType == "member_expression" ||
			childType == "selector_expression" || childType == "attribute" ||
			childType == "scoped_identifier" || childType == "call" {
			text := string(src[child.StartByte():child.EndByte()])
			// Remove parentheses if present
			if idx := strings.Index(text, "("); idx > 0 {
				text = text[:idx]
			}
			return text
		}
	}
	return ""
}

// extractArguments extracts arguments from a function call
func (tp *TaintPropagator) extractArguments(node *sitter.Node, src []byte) []Argument {
	var args []Argument

	// Find arguments node
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()
		if childType == "arguments" || childType == "argument_list" ||
			strings.Contains(childType, "argument") {
			// Extract each argument
			for j := 0; j < int(child.ChildCount()); j++ {
				arg := child.Child(j)
				argType := arg.Type()
				// Skip punctuation
				if argType == "," || argType == "(" || argType == ")" {
					continue
				}
				text := string(src[arg.StartByte():arg.EndByte()])
				args = append(args, Argument{
					Text:  text,
					Index: len(args),
				})
			}
			break
		}
	}

	return args
}

// extractReturnValue extracts the value from a return statement
func (tp *TaintPropagator) extractReturnValue(node *sitter.Node, src []byte) string {
	text := string(src[node.StartByte():node.EndByte()])

	// Remove "return" keyword
	text = strings.TrimPrefix(strings.TrimSpace(text), "return")
	text = strings.TrimSuffix(strings.TrimSpace(text), ";")

	return strings.TrimSpace(text)
}

// findContainingFunction finds the function containing a node using centralized AST patterns
func (tp *TaintPropagator) findContainingFunction(node *sitter.Node) *sitter.Node {
	current := node.Parent()
	for current != nil {
		nodeType := current.Type()
		if sources.IsFunctionNode(nodeType) || strings.Contains(nodeType, "function") ||
			strings.Contains(nodeType, "method") {
			return current
		}
		current = current.Parent()
	}
	return nil
}

// extractFunctionNameFromDef extracts function name from a function definition
func (tp *TaintPropagator) extractFunctionNameFromDef(node *sitter.Node, src []byte) string {
	// Look for identifier child
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "identifier" || child.Type() == "name" ||
			child.Type() == "function_declarator" {
			return string(src[child.StartByte():child.EndByte()])
		}
	}
	return ""
}

// getCurrentScope returns the current scope identifier using centralized AST patterns
func (tp *TaintPropagator) getCurrentScope(node *sitter.Node) string {
	// Walk up to find containing scope
	parts := []string{}
	current := node.Parent()
	for current != nil {
		nodeType := current.Type()
		if sources.IsScopeNode(nodeType) {
			// Try to get name
			name := tp.extractFunctionNameFromDef(current, nil)
			if name != "" {
				parts = append([]string{name}, parts...)
			}
		}
		current = current.Parent()
	}

	if len(parts) == 0 {
		return "global"
	}
	return strings.Join(parts, ".")
}

// nodeToLocation converts a tree-sitter node to a Location
func nodeToLocation(node *sitter.Node, src []byte, filePath string) Location {
	text := ""
	if src != nil {
		text = string(src[node.StartByte():node.EndByte()])
		// Truncate for snippet
		if len(text) > 100 {
			text = text[:97] + "..."
		}
	}

	return Location{
		FilePath:  filePath,
		Line:      int(node.StartPoint().Row) + 1,
		Column:    int(node.StartPoint().Column),
		EndLine:   int(node.EndPoint().Row) + 1,
		EndColumn: int(node.EndPoint().Column),
		Snippet:   text,
	}
}
