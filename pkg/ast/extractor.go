package ast

import (
	"regexp"
	"slices"
	"strings"
	"sync"

	"github.com/hatlesswizard/inputtracer/pkg/sources/patterns"
	sitter "github.com/smacker/go-tree-sitter"
)

// whitespaceRe collapses runs of whitespace; pre-compiled to avoid per-call allocation.
var whitespaceRe = regexp.MustCompile(`\s+`)

// regexCache caches compiled boundary-pattern regexes keyed by variable name pattern.
var regexCache sync.Map // map[string]*regexp.Regexp

func cachedRegex(pattern string) *regexp.Regexp {
	if cached, ok := regexCache.Load(pattern); ok {
		return cached.(*regexp.Regexp)
	}
	compiled := regexp.MustCompile(pattern)
	regexCache.Store(pattern, compiled)
	return compiled
}

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

// Extractor interface for language-specific AST extraction.
// Parameters that were previously *sitter.Node now accept the ast.Node
// interface. Callers that hold a *sitter.Node can pass it directly because
// *sitter.Node satisfies ast.Node (see node.go).
type Extractor interface {
	Language() string
	ExtractAssignments(root Node, src []byte) []Assignment
	ExtractCalls(root Node, src []byte) []FunctionCall
	ExpressionContains(node Node, varName string, src []byte) bool
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
	lang            string
	assignmentTypes []string
	callTypes       []string
	identifierTypes []string
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
func (e *BaseExtractor) ExtractAssignments(root Node, src []byte) []Assignment {
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
func (e *BaseExtractor) ExtractCalls(root Node, src []byte) []FunctionCall {
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
func (e *BaseExtractor) ExpressionContains(node Node, varName string, src []byte) bool {
	// Guard against both a nil interface and a nil *sitter.Node wrapped in an
	// interface (the latter is non-nil at the interface level but still unsafe
	// to dereference).
	if isNilNode(node) {
		return false
	}

	text := string(src[node.StartByte():node.EndByte()])

	// Direct match
	if text == varName {
		return true
	}

	// Check with word boundaries (handles $-prefixed and @-prefixed vars)
	return cachedRegex(patterns.VariableBoundaryPattern(varName)).MatchString(text)
}

// traverse recursively traverses the AST.
// The entry-point accepts Node (the package-local interface) so that
// ExtractAssignments and ExtractCalls can forward their root parameter directly.
// Recursive calls pass *sitter.Node because Child() returns that concrete type.
func (e *BaseExtractor) traverse(node Node, callback func(*sitter.Node)) {
	if node == nil {
		return
	}
	// The callback receives a *sitter.Node so that internal helpers such as
	// parseAssignment and parseCall can call Child/Parent without type-asserting.
	// We assert here once per node rather than at every call-site.
	concrete, ok := node.(*sitter.Node)
	if !ok || concrete == nil {
		return
	}
	callback(concrete)
	for i := 0; i < int(concrete.ChildCount()); i++ {
		e.traverse(concrete.Child(i), callback)
	}
}

// isAssignmentType checks if a node type is an assignment
func (e *BaseExtractor) isAssignmentType(nodeType string) bool {
	if slices.Contains(e.assignmentTypes, nodeType) {
		return true
	}
	// Generic fallback
	return strings.Contains(nodeType, "assignment") ||
		strings.Contains(nodeType, "declarator")
}

// isCallType checks if a node type is a function call
func (e *BaseExtractor) isCallType(nodeType string) bool {
	if slices.Contains(e.callTypes, nodeType) {
		return true
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
		if slices.Contains([]string{
			"identifier", "member_expression",
			"selector_expression", "attribute", "scoped_identifier",
		}, childType) {
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

	IterateChildren(node, func(child *sitter.Node) {
		args = append(args, CallArgument{
			Name:  string(src[child.StartByte():child.EndByte()]),
			Node:  child,
			Index: argIndex,
		})
		argIndex++
	})

	return args
}

// truncateString truncates a string to maxLen
func truncateString(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", "")
	s = whitespaceRe.ReplaceAllString(s, " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
