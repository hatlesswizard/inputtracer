// Package analyzer defines the interface for language-specific analyzers
package analyzer

import (
	"fmt"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// LanguageAnalyzer defines the interface that all language analyzers must implement
type LanguageAnalyzer interface {
	// Language returns the language this analyzer handles
	Language() string

	// SupportedExtensions returns file extensions this analyzer handles
	SupportedExtensions() []string

	// BuildSymbolTable parses a file and builds its symbol table
	BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error)

	// ResolveImports resolves import/include statements to actual file paths
	ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error)

	// ExtractClasses extracts class definitions from the AST
	ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error)

	// ExtractFunctions extracts function definitions from the AST
	ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error)

	// ExtractAssignments extracts all variable assignments from the AST
	ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error)

	// ExtractCalls extracts all function/method calls from the AST
	ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error)

	// FindInputSources finds all user input sources in the AST
	FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error)

	// AnalyzeMethodBody analyzes a method body for data flow
	AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*MethodFlowAnalysis, error)

	// DetectFrameworks detects which frameworks are used in the code
	DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error)

	// GetFrameworkPatterns returns known framework patterns for this language
	GetFrameworkPatterns() []*types.FrameworkPattern

	// TraceExpression traces a specific expression back to its sources
	TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error)
}

// MethodFlowAnalysis represents the result of analyzing a method body
type MethodFlowAnalysis struct {
	// Which parameters flow to return value
	ParamsToReturn []int

	// Which parameters flow to which properties
	ParamsToProperties map[int][]string

	// Which parameters flow to which method calls (param index -> call sites)
	ParamsToCallArgs map[int][]*types.CallSite

	// Variables that become tainted within the method
	TaintedVariables map[string]*types.TaintInfo

	// All assignments in the method
	Assignments []*types.Assignment

	// All calls in the method
	Calls []*types.CallSite

	// Return statements
	Returns []ReturnInfo

	// Does this method return user input?
	ReturnsInput bool

	// Does this method modify properties with input?
	ModifiesProperties bool
}

// ReturnInfo represents a return statement
type ReturnInfo struct {
	Line       int
	Expression string
	IsTainted  bool
	TaintSource string
}

// BaseAnalyzer provides common functionality for all analyzers
type BaseAnalyzer struct {
	language          string
	extensions        []string
	frameworkPatterns []*types.FrameworkPattern
}

// NewBaseAnalyzer creates a new base analyzer
func NewBaseAnalyzer(language string, extensions []string) *BaseAnalyzer {
	return &BaseAnalyzer{
		language:          language,
		extensions:        extensions,
		frameworkPatterns: make([]*types.FrameworkPattern, 0),
	}
}

// Language returns the language
func (b *BaseAnalyzer) Language() string {
	return b.language
}

// SupportedExtensions returns supported extensions
func (b *BaseAnalyzer) SupportedExtensions() []string {
	return b.extensions
}

// GetFrameworkPatterns returns framework patterns
func (b *BaseAnalyzer) GetFrameworkPatterns() []*types.FrameworkPattern {
	return b.frameworkPatterns
}

// AddFrameworkPattern adds a framework pattern
func (b *BaseAnalyzer) AddFrameworkPattern(pattern *types.FrameworkPattern) {
	b.frameworkPatterns = append(b.frameworkPatterns, pattern)
}

// ============================================================================
// AST Helper Functions
// ============================================================================

// GetNodeText extracts the text content of a node
func GetNodeText(node *sitter.Node, source []byte) string {
	if node == nil {
		return ""
	}
	start := node.StartByte()
	end := node.EndByte()
	if start >= uint32(len(source)) || end > uint32(len(source)) {
		return ""
	}
	return string(source[start:end])
}

// FindChildByType finds the first child with a specific type
func FindChildByType(node *sitter.Node, nodeType string) *sitter.Node {
	if node == nil {
		return nil
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == nodeType {
			return child
		}
	}
	return nil
}

// FindChildrenByType finds all children with a specific type
func FindChildrenByType(node *sitter.Node, nodeType string) []*sitter.Node {
	var children []*sitter.Node
	if node == nil {
		return children
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == nodeType {
			children = append(children, child)
		}
	}
	return children
}

// FindChildByFieldName finds a child by its field name
func FindChildByFieldName(node *sitter.Node, fieldName string) *sitter.Node {
	if node == nil {
		return nil
	}
	return node.ChildByFieldName(fieldName)
}

// TraverseTree traverses the AST and calls the callback for each node
func TraverseTree(node *sitter.Node, callback func(*sitter.Node) bool) {
	if node == nil {
		return
	}
	if !callback(node) {
		return
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		TraverseTree(node.Child(i), callback)
	}
}

// FindNodesOfType finds all nodes of a specific type in the tree
func FindNodesOfType(root *sitter.Node, nodeType string) []*sitter.Node {
	var nodes []*sitter.Node
	TraverseTree(root, func(node *sitter.Node) bool {
		if node.Type() == nodeType {
			nodes = append(nodes, node)
		}
		return true
	})
	return nodes
}

// FindNodesOfTypes finds all nodes matching any of the given types
func FindNodesOfTypes(root *sitter.Node, nodeTypes []string) []*sitter.Node {
	typeSet := make(map[string]bool)
	for _, t := range nodeTypes {
		typeSet[t] = true
	}

	var nodes []*sitter.Node
	TraverseTree(root, func(node *sitter.Node) bool {
		if typeSet[node.Type()] {
			nodes = append(nodes, node)
		}
		return true
	})
	return nodes
}

// GetAncestorOfType finds the first ancestor of a specific type
func GetAncestorOfType(node *sitter.Node, nodeType string) *sitter.Node {
	if node == nil {
		return nil
	}
	parent := node.Parent()
	for parent != nil {
		if parent.Type() == nodeType {
			return parent
		}
		parent = parent.Parent()
	}
	return nil
}

// GetEnclosingFunction finds the enclosing function/method definition
func GetEnclosingFunction(node *sitter.Node, functionTypes []string) *sitter.Node {
	typeSet := make(map[string]bool)
	for _, t := range functionTypes {
		typeSet[t] = true
	}

	parent := node.Parent()
	for parent != nil {
		if typeSet[parent.Type()] {
			return parent
		}
		parent = parent.Parent()
	}
	return nil
}

// GetEnclosingClass finds the enclosing class definition
func GetEnclosingClass(node *sitter.Node, classTypes []string) *sitter.Node {
	typeSet := make(map[string]bool)
	for _, t := range classTypes {
		typeSet[t] = true
	}

	parent := node.Parent()
	for parent != nil {
		if typeSet[parent.Type()] {
			return parent
		}
		parent = parent.Parent()
	}
	return nil
}

// NodeLocation creates a Location from a node
func NodeLocation(node *sitter.Node, filePath string) types.Location {
	if node == nil {
		return types.Location{FilePath: filePath}
	}
	return types.Location{
		FilePath:  filePath,
		Line:      int(node.StartPoint().Row) + 1,
		Column:    int(node.StartPoint().Column),
		EndLine:   int(node.EndPoint().Row) + 1,
		EndColumn: int(node.EndPoint().Column),
	}
}

// CreateFlowNode creates a FlowNode from an AST node
func CreateFlowNode(node *sitter.Node, source []byte, filePath, language string, nodeType types.FlowNodeType) *types.FlowNode {
	return &types.FlowNode{
		ID:       GenerateNodeID(filePath, node),
		Type:     nodeType,
		Language: language,
		FilePath: filePath,
		Line:     int(node.StartPoint().Row) + 1,
		Column:   int(node.StartPoint().Column),
		EndLine:  int(node.EndPoint().Row) + 1,
		EndColumn: int(node.EndPoint().Column),
		Snippet:  GetNodeText(node, source),
		Metadata: make(map[string]interface{}),
	}
}

// GenerateNodeID generates a unique ID for a node
func GenerateNodeID(filePath string, node *sitter.Node) string {
	if node == nil {
		return ""
	}
	return fmt.Sprintf("%s:%d:%d", filePath, node.StartPoint().Row+1, node.StartPoint().Column)
}

// ============================================================================
// Registry
// ============================================================================

// Registry holds all registered language analyzers
type Registry struct {
	analyzers map[string]LanguageAnalyzer
}

// NewRegistry creates a new analyzer registry
func NewRegistry() *Registry {
	return &Registry{
		analyzers: make(map[string]LanguageAnalyzer),
	}
}

// Register registers an analyzer for a language
func (r *Registry) Register(analyzer LanguageAnalyzer) {
	r.analyzers[analyzer.Language()] = analyzer
}

// Get returns the analyzer for a language
func (r *Registry) Get(language string) LanguageAnalyzer {
	return r.analyzers[language]
}

// GetByExtension returns the analyzer for a file extension
func (r *Registry) GetByExtension(ext string) LanguageAnalyzer {
	for _, analyzer := range r.analyzers {
		for _, supportedExt := range analyzer.SupportedExtensions() {
			if supportedExt == ext {
				return analyzer
			}
		}
	}
	return nil
}

// Languages returns all registered languages
func (r *Registry) Languages() []string {
	languages := make([]string, 0, len(r.analyzers))
	for lang := range r.analyzers {
		languages = append(languages, lang)
	}
	return languages
}

// DefaultRegistry is the global analyzer registry
var DefaultRegistry = NewRegistry()
