package tracer

import (
	"regexp"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/ast"
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
	"github.com/hatlesswizard/inputtracer/pkg/sources/patterns"
	sitter "github.com/smacker/go-tree-sitter"
)

// maxSnippetLength is the maximum number of runes retained for a source-code
// snippet stored in a Location. Snippets longer than this are truncated and
// suffixed with "..." (three characters), so the usable content length is
// maxSnippetLength - 3.
const maxSnippetLength = 100

// getOrCompileRegex gets a cached regex or compiles and caches it.
// It delegates to the shared cache in pkg/sources/common so all packages
// reuse the same compiled *regexp.Regexp instances. Internal patterns are
// always syntactically valid, so compilation errors are treated as fatal.
func getOrCompileRegex(pattern string) *regexp.Regexp {
	re, err := common.GetOrCompileRegex(pattern)
	if err != nil {
		panic("inputtracer: invalid regex pattern: " + pattern + ": " + err.Error())
	}
	return re
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
func (prop *TaintPropagator) PropagateFromAssignment(node *sitter.Node, src []byte, filePath string) {
	// Find assignment target and value
	target, value := prop.extractAssignmentParts(node, src)
	if target == "" || value == "" {
		return
	}

	// Check if value is tainted
	if taintInfo := prop.findTaintInfo(value); taintInfo != nil {
		// Create new tainted variable
		tv := &TaintedVariable{
			Name:     target,
			Scope:    prop.getCurrentScope(node, src),
			Source:   taintInfo.Source,
			Location: nodeToLocation(node, src, filePath),
			Depth:    taintInfo.Depth + 1,
		}

		// Add to state
		prop.state.AddTaintedVariable(tv)

		// Create propagation step
		step := PropagationStep{
			Type:     StepAssignment,
			Variable: target,
			Location: tv.Location,
		}
		prop.state.AddPropagationStep(taintInfo.Source, step)
	}
}

// PropagateFromFunctionCall propagates taint through function calls
func (prop *TaintPropagator) PropagateFromFunctionCall(node *sitter.Node, src []byte, filePath string) {
	funcName := prop.extractFunctionName(node, src)
	if funcName == "" {
		return
	}

	args := prop.extractArguments(node, src)

	// Check each argument for taint
	for i, arg := range args {
		if taintInfo := prop.findTaintInfo(arg.Text); taintInfo != nil {
			// Record that this function receives tainted input
			tf := &TaintedFunction{
				Name:     funcName,
				FilePath: filePath,
				Line:     int(node.StartPoint().Row) + 1,
				Language: prop.language,
				TaintedParams: []TaintedParam{
					{
						Index:  i,
						Name:   arg.Name,
						Source: taintInfo.Source,
					},
				},
			}

			prop.state.AddTaintedFunction(tf)

			// Create propagation step
			step := PropagationStep{
				Type:     StepParameterPass,
				Variable: arg.Text,
				Function: funcName,
				Location: nodeToLocation(node, src, filePath),
			}
			prop.state.AddPropagationStep(taintInfo.Source, step)
		}
	}
}

// PropagateFromReturn propagates taint from return statements
func (prop *TaintPropagator) PropagateFromReturn(node *sitter.Node, src []byte, filePath string) {
	// Extract return value
	returnValue := prop.extractReturnValue(node, src)
	if returnValue == "" {
		return
	}

	// Check if return value is tainted
	if taintInfo := prop.findTaintInfo(returnValue); taintInfo != nil {
		// Find containing function
		funcNode := prop.findContainingFunction(node)
		if funcNode == nil {
			return
		}

		funcName := prop.extractFunctionNameFromDef(funcNode, src)

		// Mark function as returning tainted data
		prop.state.AddReturnsTaintedFunction(funcName, taintInfo.Source)

		// Create propagation step
		step := PropagationStep{
			Type:     StepReturn,
			Variable: returnValue,
			Function: funcName,
			Location: nodeToLocation(node, src, filePath),
		}
		prop.state.AddPropagationStep(taintInfo.Source, step)
	}
}

// TaintInfo contains information about a tainted value
type TaintInfo struct {
	Source *InputSource
	Depth  int
}

// findTaintInfo returns taint information for a value expression, or nil if not tainted.
// It first attempts an O(1) map lookup using the base variable name extracted from the
// expression, falling back to boundary-aware regex matching only when the expression
// contains operators (property access, subscript, etc.) that prevent a direct lookup.
func (prop *TaintPropagator) findTaintInfo(value string) *TaintInfo {
	value = strings.TrimSpace(value)
	if value == "" {
		return nil
	}

	// Extract the base variable name for the O(1) map lookup.
	// For simple identifiers (e.g. "$user", "req") this is the whole value;
	// for property/subscript access (e.g. "$user['id']", "req.body") it is the
	// portion before the first "." or "[".
	baseName := value
	if idx := strings.IndexAny(value, ".["); idx > 0 {
		baseName = value[:idx]
	}

	// O(1) lookup in the name-keyed map maintained by AnalysisState.SetTainted.
	if tv, ok := prop.state.TaintedValues[baseName]; ok {
		return &TaintInfo{
			Source: tv.Source,
			Depth:  tv.Depth,
		}
	}

	// Follow alias chain: if baseName is an alias for a tainted variable,
	// resolve through up to 3 levels to find the original tainted source.
	// Aliases are stored without sigil prefix ($/@), so strip it for lookup.
	if prop.state.Aliases != nil {
		strippedBase := stripVarSigil(baseName)
		resolvedName := strippedBase
		visited := make(map[string]bool)
		const maxAliasDepth = 3
		for i := 0; i < maxAliasDepth; i++ {
			aliased, ok := prop.state.Aliases[resolvedName]
			if !ok || visited[resolvedName] {
				break
			}
			visited[resolvedName] = true
			resolvedName = aliased
			// Try lookup with resolved name as-is (for non-sigil languages)
			if tv, ok := prop.state.TaintedValues[resolvedName]; ok {
				return &TaintInfo{
					Source: tv.Source,
					Depth:  tv.Depth,
				}
			}
			// Try lookup with $ prefix (for PHP variables stored with sigil)
			if tv, ok := prop.state.TaintedValues["$"+resolvedName]; ok {
				return &TaintInfo{
					Source: tv.Source,
					Depth:  tv.Depth,
				}
			}
		}
	}

	// Boundary-aware fallback for expressions where the base name did not match
	// directly (e.g. complex concatenations or language-specific sigil differences).
	for _, tv := range prop.state.TaintedVariables {
		if tv.Name != baseName && prop.matchesVariable(value, tv.Name) {
			return &TaintInfo{
				Source: tv.Source,
				Depth:  tv.Depth,
			}
		}
	}

	// Not a tainted value; direct input sources are handled elsewhere.
	return nil
}

// matchesVariable checks if a value references a tainted variable.
// Handles exact matches, property access (var.prop, var['prop']),
// and expression containment with boundary-aware matching.
func (prop *TaintPropagator) matchesVariable(value, varName string) bool {
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

// extractAssignmentParts extracts target and value from an assignment.
// Language-specific extraction logic lives in pkg/ast.ExtractAssignmentParts.
func (prop *TaintPropagator) extractAssignmentParts(node *sitter.Node, src []byte) (target, value string) {
	return ast.ExtractAssignmentParts(node, src, prop.language)
}

// extractFunctionName extracts the function name from a call expression
func (prop *TaintPropagator) extractFunctionName(node *sitter.Node, src []byte) string {
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
func (prop *TaintPropagator) extractArguments(node *sitter.Node, src []byte) []Argument {
	var args []Argument

	// Find arguments node
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		childType := child.Type()
		if childType == "arguments" || childType == "argument_list" ||
			strings.Contains(childType, "argument") {
			// Extract each argument, skipping punctuation tokens
			ast.IterateChildren(child, func(arg *sitter.Node) {
				args = append(args, Argument{
					Text:  string(src[arg.StartByte():arg.EndByte()]),
					Index: len(args),
				})
			})
			break
		}
	}

	return args
}

// extractReturnValue extracts the value from a return statement
func (prop *TaintPropagator) extractReturnValue(node *sitter.Node, src []byte) string {
	text := string(src[node.StartByte():node.EndByte()])

	// Remove "return" keyword
	text = strings.TrimPrefix(strings.TrimSpace(text), "return")
	text = strings.TrimSuffix(strings.TrimSpace(text), ";")

	return strings.TrimSpace(text)
}

// findContainingFunction finds the function containing a node using centralized AST patterns
func (prop *TaintPropagator) findContainingFunction(node *sitter.Node) *sitter.Node {
	current := node.Parent()
	for current != nil {
		nodeType := current.Type()
		if ast.IsFunctionNode(nodeType) || strings.Contains(nodeType, "function") ||
			strings.Contains(nodeType, "method") {
			return current
		}
		current = current.Parent()
	}
	return nil
}

// extractFunctionNameFromDef extracts function name from a function definition
func (prop *TaintPropagator) extractFunctionNameFromDef(node *sitter.Node, src []byte) string {
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

// getCurrentScope returns the current scope identifier using centralized AST patterns.
// src must be the source bytes of the file being analysed; it is forwarded to
// extractFunctionNameFromDef so that node text can be read correctly.
func (prop *TaintPropagator) getCurrentScope(node *sitter.Node, src []byte) string {
	// Walk up to find containing scope
	parts := []string{}
	current := node.Parent()
	for current != nil {
		nodeType := current.Type()
		if ast.IsScopeNode(nodeType) {
			// Try to get name
			name := prop.extractFunctionNameFromDef(current, src)
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
		if len(text) > maxSnippetLength {
			text = text[:maxSnippetLength-3] + "..."
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
