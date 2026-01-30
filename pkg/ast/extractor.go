package ast

import (
	"regexp"
	"strings"
	"sync"

	"github.com/hatlesswizard/inputtracer/pkg/sources/patterns"
	sitter "github.com/smacker/go-tree-sitter"
)

// Assignment represents an assignment operation in code
type Assignment struct {
	LHS       string
	RHS       *sitter.Node
	RHSText   string
	Scope     string
	Line      int
	Column    int
	EndLine   int
	EndColumn int
	Snippet   string
}

// CallArgument represents an argument in a function call
type CallArgument struct {
	Name  string
	Node  *sitter.Node
	Index int
}

// FunctionCall represents a function call in code
type FunctionCall struct {
	Name      string
	Arguments []CallArgument
	Line      int
	Column    int
	EndLine   int
	EndColumn int
	Scope     string
}

// Extractor interface for language-specific AST extraction
type Extractor interface {
	Language() string
	ExtractAssignments(root *sitter.Node, src []byte) []Assignment
	ExtractCalls(root *sitter.Node, src []byte) []FunctionCall
	ExpressionContains(node *sitter.Node, varName string, src []byte) bool
}

// Registry manages AST extractors for all languages
type Registry struct {
	extractors map[string]Extractor
	mu         sync.RWMutex
}

// NewRegistry creates a new AST registry
func NewRegistry() *Registry {
	return &Registry{
		extractors: make(map[string]Extractor),
	}
}

// Register registers an extractor for a language
func (r *Registry) Register(extractor Extractor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.extractors[extractor.Language()] = extractor
}

// GetExtractor returns the extractor for a language
func (r *Registry) GetExtractor(language string) Extractor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.extractors[language]
}

// BaseExtractor provides common functionality for AST extraction
type BaseExtractor struct {
	lang             string
	assignmentTypes  []string
	callTypes        []string
	identifierTypes  []string
}

// NewBaseExtractor creates a new base extractor
func NewBaseExtractor(language string, assignmentTypes, callTypes, identifierTypes []string) *BaseExtractor {
	return &BaseExtractor{
		lang:            language,
		assignmentTypes: assignmentTypes,
		callTypes:       callTypes,
		identifierTypes: identifierTypes,
	}
}

// Language returns the language this extractor handles
func (e *BaseExtractor) Language() string {
	return e.lang
}

// ExtractAssignments extracts all assignments from the AST
func (e *BaseExtractor) ExtractAssignments(root *sitter.Node, src []byte) []Assignment {
	var assignments []Assignment

	e.traverse(root, func(node *sitter.Node) {
		nodeType := node.Type()
		if e.isAssignmentType(nodeType) {
			assign := e.parseAssignment(node, src)
			if assign.LHS != "" {
				assignments = append(assignments, assign)
			}
		}
	})

	return assignments
}

// ExtractCalls extracts all function calls from the AST
func (e *BaseExtractor) ExtractCalls(root *sitter.Node, src []byte) []FunctionCall {
	var calls []FunctionCall

	e.traverse(root, func(node *sitter.Node) {
		nodeType := node.Type()
		if e.isCallType(nodeType) {
			call := e.parseCall(node, src)
			if call.Name != "" {
				calls = append(calls, call)
			}
		}
	})

	return calls
}

// ExpressionContains checks if an expression contains a variable.
// Uses boundary-aware matching to avoid substring false positives
// (e.g. "$order" must not match "$order_id").
func (e *BaseExtractor) ExpressionContains(node *sitter.Node, varName string, src []byte) bool {
	if node == nil {
		return false
	}

	text := string(src[node.StartByte():node.EndByte()])

	// Direct match
	if text == varName {
		return true
	}

	// Check with word boundaries (handles $-prefixed and @-prefixed vars)
	pattern := regexp.MustCompile(patterns.VariableBoundaryPattern(varName))
	return pattern.MatchString(text)
}

// traverse recursively traverses the AST
func (e *BaseExtractor) traverse(node *sitter.Node, callback func(*sitter.Node)) {
	if node == nil {
		return
	}
	callback(node)
	for i := 0; i < int(node.ChildCount()); i++ {
		e.traverse(node.Child(i), callback)
	}
}

// isAssignmentType checks if a node type is an assignment
func (e *BaseExtractor) isAssignmentType(nodeType string) bool {
	for _, t := range e.assignmentTypes {
		if nodeType == t {
			return true
		}
	}
	// Generic fallback
	return strings.Contains(nodeType, "assignment") ||
		strings.Contains(nodeType, "declarator")
}

// isCallType checks if a node type is a function call
func (e *BaseExtractor) isCallType(nodeType string) bool {
	for _, t := range e.callTypes {
		if nodeType == t {
			return true
		}
	}
	// Generic fallback
	return strings.Contains(nodeType, "call")
}

// parseAssignment parses an assignment node
func (e *BaseExtractor) parseAssignment(node *sitter.Node, src []byte) Assignment {
	assign := Assignment{
		Line:      int(node.StartPoint().Row) + 1,
		Column:    int(node.StartPoint().Column),
		EndLine:   int(node.EndPoint().Row) + 1,
		EndColumn: int(node.EndPoint().Column),
		Snippet:   truncateString(string(src[node.StartByte():node.EndByte()]), 100),
	}

	// Extract LHS and RHS based on language patterns
	text := string(src[node.StartByte():node.EndByte()])

	// Try common patterns
	if strings.Contains(text, ":=") {
		parts := strings.SplitN(text, ":=", 2)
		if len(parts) == 2 {
			assign.LHS = strings.TrimSpace(parts[0])
			assign.RHSText = strings.TrimSpace(parts[1])
		}
	} else if strings.Contains(text, "=") && !strings.Contains(text, "==") {
		parts := strings.SplitN(text, "=", 2)
		if len(parts) == 2 {
			// LHS might include type annotation
			lhs := strings.TrimSpace(parts[0])
			// Extract just the variable name (last word)
			lhsParts := strings.Fields(lhs)
			if len(lhsParts) > 0 {
				assign.LHS = lhsParts[len(lhsParts)-1]
			}
			assign.RHSText = strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
		}
	}

	// Find RHS node
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child != nil {
			childType := child.Type()
			// Skip LHS types
			if childType != "identifier" && childType != "variable_name" &&
				childType != "=" && childType != ":=" {
				assign.RHS = child
				break
			}
		}
	}

	return assign
}

// parseCall parses a function call node
func (e *BaseExtractor) parseCall(node *sitter.Node, src []byte) FunctionCall {
	call := FunctionCall{
		Line:      int(node.StartPoint().Row) + 1,
		Column:    int(node.StartPoint().Column),
		EndLine:   int(node.EndPoint().Row) + 1,
		EndColumn: int(node.EndPoint().Column),
		Arguments: make([]CallArgument, 0),
	}

	// Find function name
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		childType := child.Type()

		// Function name
		if childType == "identifier" || childType == "member_expression" ||
			childType == "selector_expression" || childType == "attribute" ||
			childType == "scoped_identifier" {
			if call.Name == "" {
				call.Name = string(src[child.StartByte():child.EndByte()])
			}
		}

		// Arguments
		if strings.Contains(childType, "argument") {
			call.Arguments = e.parseArguments(child, src)
		}
	}

	return call
}

// parseArguments parses function call arguments
func (e *BaseExtractor) parseArguments(node *sitter.Node, src []byte) []CallArgument {
	var args []CallArgument
	argIndex := 0

	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil {
			continue
		}
		childType := child.Type()

		// Skip punctuation
		if childType == "," || childType == "(" || childType == ")" {
			continue
		}

		args = append(args, CallArgument{
			Name:  string(src[child.StartByte():child.EndByte()]),
			Node:  child,
			Index: argIndex,
		})
		argIndex++
	}

	return args
}

// truncateString truncates a string to maxLen
func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = regexp.MustCompile(`\s+`).ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
