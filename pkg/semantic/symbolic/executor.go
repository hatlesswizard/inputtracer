// Package symbolic provides symbolic execution for deep semantic tracing
// This traces object instantiation, constructor execution, method calls, and property population
// Works universally across ALL PHP applications - no framework-specific hints
package symbolic

import (
	"regexp"
	"sync"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// Regex cache for avoiding repeated compilation of the same patterns
var regexCache sync.Map // pattern string -> *regexp.Regexp

// getOrCompileRegex returns a cached compiled regex, compiling it if not already cached
func getOrCompileRegex(pattern string) *regexp.Regexp {
	if cached, ok := regexCache.Load(pattern); ok {
		return cached.(*regexp.Regexp)
	}
	compiled := regexp.MustCompile(pattern)
	regexCache.Store(pattern, compiled)
	return compiled
}

// ExpressionType represents the type of expression being traced
type ExpressionType int

const (
	ExprTypeUnknown        ExpressionType = iota
	ExprTypePropertyAccess                // $obj->property or $obj->property['key']
	ExprTypeMethodCall                    // $obj->method('arg') or $obj->method($var)
	ExprTypeStaticCall                    // Class::method('arg')
	ExprTypeStaticProperty                // Class::$property
	ExprTypeFunctionCall                  // function('arg')
	ExprTypeSuperglobal                   // $_GET['key'], $_POST['key'], etc.
	ExprTypeLocalVariable                 // $id, $username (simple variable)
)

// ParsedExpression holds the parsed components of an expression
type ParsedExpression struct {
	Type            ExpressionType
	RawExpr         string
	VarName         string   // $mybb
	ClassName       string   // MyBB (resolved)
	PropertyName    string   // input
	MethodName      string   // get_input
	AccessKey       string   // 'thumbnail' or 'timezone'
	Arguments       []string // method arguments
	SuperglobalName string   // $_GET, $_POST, etc. (for ExprTypeSuperglobal)
	IsSuperglobal   bool     // true if this is a superglobal access

	// Chained expression support
	IsChained  bool        // true if this is a chained expression
	ChainSteps []ChainStep // Steps in the chain
}

// ChainStep represents one step in a chained expression
type ChainStep struct {
	Type      ExpressionType // PropertyAccess or MethodCall
	Name      string         // method or property name
	Arguments []string       // method arguments if method call
	AccessKey string         // array access key if any
}

// ExecutionEngine performs symbolic execution to trace data flow through objects
// Memory-optimized with LRU file caching to prevent unbounded memory growth
type ExecutionEngine struct {
	// Global symbol tables from all parsed files
	symbolTables map[string]*types.SymbolTable

	// Object instances: variable name -> class name
	instances map[string]*ObjectInstance

	// Property states: "ClassName.propertyName" -> PropertyState
	properties map[string]*PropertyState

	// Method call chain for tracing
	callStack []MethodCall

	// Traced flows
	flows []*PropertyFlow

	// Maximum call depth to prevent infinite recursion
	maxDepth int

	// Current analysis depth
	currentDepth int

	// MEMORY OPTIMIZATION: LRU file cache instead of unbounded maps
	// Keeps only recently-used files in memory, evicts LRU entries
	fileCache *LRUFileCache

	// Legacy maps for backward compatibility (deprecated - use fileCache)
	parsedFiles  map[string]*sitter.Node
	fileContents map[string][]byte

	// Method return analysis cache: "ClassName.methodName" -> what it returns
	methodReturns map[string]*MethodReturnInfo
}

// MethodReturnInfo captures what a method returns
type MethodReturnInfo struct {
	ReturnsProperty     bool     // returns $this->property
	PropertyName        string   // which property
	UsesParamAsKey      bool     // returns $this->property[$param]
	ParamIndex          int      // which parameter is used as key
	ReturnsParam        bool     // returns a parameter directly
	ReturnStatements    []string // all return statement code
	ReturnsUserInput    bool     // directly returns user input
	UserInputExpression string   // e.g., "$_GET['key']"
	ReturnsSelf         bool     // returns $this (fluent interface)
}

// ObjectInstance represents an instantiated object
type ObjectInstance struct {
	VariableName string
	ClassName    string
	FilePath     string
	Line         int
	Properties   map[string]*PropertyState
}

// PropertyState tracks the state of a class property
type PropertyState struct {
	ClassName      string
	PropertyName   string
	InitialValue   string
	CurrentSources []string     // What sources have flowed into this property
	PopulatedBy    []MethodCall // Which method calls populated this property
	Assignments    []Assignment // All assignments to this property
}

// Assignment represents one assignment to a property
type Assignment struct {
	Source      string // The source expression (e.g., "$_GET", "$array[$key]")
	SourceType  string // Type of source
	Method      string // Which method made this assignment
	Line        int
	FilePath    string
	IsUserInput bool     // Whether this comes from user input
	TaintChain  []string // Chain of taints
}

// ExternalAssignment represents a property assigned outside the class definition
// This handles dynamic properties like: $mybb->post_code = generate_post_check();
type ExternalAssignment struct {
	PropertyName string // The property being assigned
	Source       string // The value assigned (e.g., "generate_post_check()")
	FilePath     string
	Line         int
}

// MethodCall represents a method invocation
type MethodCall struct {
	ClassName  string
	MethodName string
	Arguments  []string
	FilePath   string
	Line       int
	CalledFrom string // Parent method
}

// PropertyFlow represents the complete flow analysis for a property access
type PropertyFlow struct {
	// The expression being traced (e.g., "$mybb->input['thumbnail']" or "$mybb->get_input('timezone')")
	Expression string

	// The class and property/method
	ClassName    string
	PropertyName string
	MethodName   string
	AccessKey    string // e.g., "thumbnail" for array access or method argument

	// The complete trace
	Steps []FlowStep

	// Ultimate sources
	Sources []UltimateSource
}

// FlowStep represents one step in the flow trace
type FlowStep struct {
	StepNumber  int
	Description string
	Code        string
	FilePath    string
	Line        int
	Type        string // "property_init", "constructor_call", "method_call", "assignment", "loop", "return"
}

// UltimateSource represents the original user input source
type UltimateSource struct {
	Type       string // "http_get", "http_post", "http_cookie", etc.
	Expression string // e.g., "$_GET['thumbnail']"
	FilePath   string
	Line       int
}

// NewExecutionEngine creates a new symbolic execution engine
// Uses an LRU file cache to limit memory usage
func NewExecutionEngine() *ExecutionEngine {
	return &ExecutionEngine{
		symbolTables:  make(map[string]*types.SymbolTable),
		instances:     make(map[string]*ObjectInstance),
		properties:    make(map[string]*PropertyState),
		callStack:     make([]MethodCall, 0, 32), // Pre-allocate reasonable capacity
		flows:         make([]*PropertyFlow, 0, 64),
		maxDepth:      10,
		fileCache:     NewLRUFileCache(100), // Keep max 100 files in memory
		parsedFiles:   make(map[string]*sitter.Node),
		fileContents:  make(map[string][]byte),
		methodReturns: make(map[string]*MethodReturnInfo),
	}
}

// NewExecutionEngineWithCacheSize creates an engine with custom cache size
func NewExecutionEngineWithCacheSize(cacheSize int) *ExecutionEngine {
	e := NewExecutionEngine()
	e.fileCache = NewLRUFileCache(cacheSize)
	return e
}

// AddSymbolTable adds a symbol table from a parsed file
func (e *ExecutionEngine) AddSymbolTable(filePath string, st *types.SymbolTable) {
	e.symbolTables[filePath] = st
}

// AddParsedFile adds a parsed file AST
// DEPRECATED: Use SetFilePath and let the LRU cache handle loading
func (e *ExecutionEngine) AddParsedFile(filePath string, root *sitter.Node, content []byte) {
	// Store in legacy maps for backward compatibility
	e.parsedFiles[filePath] = root
	e.fileContents[filePath] = content
}

// GetFileContent retrieves file content using LRU cache (lazy loading)
func (e *ExecutionEngine) GetFileContent(filePath string) ([]byte, error) {
	// Try LRU cache first
	if e.fileCache != nil {
		return e.fileCache.GetContent(filePath)
	}
	// Fall back to legacy map
	if content, ok := e.fileContents[filePath]; ok {
		return content, nil
	}
	return nil, nil
}

// GetParsedFile retrieves parsed AST using LRU cache (lazy loading)
func (e *ExecutionEngine) GetParsedFile(filePath string) (*sitter.Node, error) {
	// Try LRU cache first
	if e.fileCache != nil {
		return e.fileCache.GetParsedFile(filePath)
	}
	// Fall back to legacy map
	if root, ok := e.parsedFiles[filePath]; ok {
		return root, nil
	}
	return nil, nil
}

// ClearFileCache releases all cached files to free memory
func (e *ExecutionEngine) ClearFileCache() {
	if e.fileCache != nil {
		e.fileCache.Clear()
	}
	// Also clear legacy maps
	e.parsedFiles = make(map[string]*sitter.Node)
	e.fileContents = make(map[string][]byte)
}

// FileCacheStats returns cache statistics for monitoring
func (e *ExecutionEngine) FileCacheStats() (hits, misses, memUsage int64) {
	if e.fileCache != nil {
		return e.fileCache.Stats()
	}
	return 0, 0, 0
}

// TracePropertyAccess traces any expression - property access OR method call
// This is the main entry point for symbolic tracing

type variableAssignment struct {
	source string
	file   string
	line   int
}

type MagicPropertyInfo struct {
	HasMagicGet      bool   // Class has __get method
	HasDynamicAssign bool   // Class has $this->$var = $val pattern
	BackingProperty  string // Property used for storage (e.g., "phrases")
	AssignMethodName string // Method that assigns properties
	SourceType       string // "file_include", "array", etc.
}

type externalMethodCall struct {
	methodName string
	args       string
	line       int
}
