// Package symbolic provides symbolic execution for deep semantic tracing
// This traces object instantiation, constructor execution, method calls, and property population
// Works universally across ALL PHP applications - no framework-specific hints
package symbolic

import (
	"fmt"
	"regexp"
	"strings"
	"sync"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"
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
	ExprTypeUnknown ExpressionType = iota
	ExprTypePropertyAccess    // $obj->property or $obj->property['key']
	ExprTypeMethodCall        // $obj->method('arg') or $obj->method($var)
	ExprTypeStaticCall        // Class::method('arg')
	ExprTypeStaticProperty    // Class::$property
	ExprTypeFunctionCall      // function('arg')
	ExprTypeSuperglobal       // $_GET['key'], $_POST['key'], etc.
	ExprTypeLocalVariable     // $id, $username (simple variable)
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
	IsChained       bool           // true if this is a chained expression
	ChainSteps      []ChainStep    // Steps in the chain
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
	CurrentSources []string      // What sources have flowed into this property
	PopulatedBy    []MethodCall  // Which method calls populated this property
	Assignments    []Assignment  // All assignments to this property
}

// Assignment represents one assignment to a property
type Assignment struct {
	Source      string   // The source expression (e.g., "$_GET", "$array[$key]")
	SourceType  string   // Type of source
	Method      string   // Which method made this assignment
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
	ClassName    string
	MethodName   string
	Arguments    []string
	FilePath     string
	Line         int
	CalledFrom   string // Parent method
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
func (e *ExecutionEngine) TracePropertyAccess(expression string, contextFile string) (*PropertyFlow, error) {
	// Parse the expression to determine its type
	parsed := e.parseExpression(expression)
	if parsed.Type == ExprTypeUnknown {
		return nil, fmt.Errorf("could not parse expression: %s", expression)
	}

	flow := &PropertyFlow{
		Expression: expression,
		Steps:      make([]FlowStep, 0),
		Sources:    make([]UltimateSource, 0),
	}

	// GAP #1 FIX: Handle direct superglobal access
	if parsed.Type == ExprTypeSuperglobal {
		return e.traceSuperglobal(parsed, flow)
	}

	// GAP #2 FIX: Handle local variable tracing
	if parsed.Type == ExprTypeLocalVariable {
		return e.traceLocalVariable(parsed, contextFile, flow)
	}

	// GAP #3 FIX: Handle static method/property calls
	if parsed.Type == ExprTypeStaticCall {
		return e.traceStaticCall(parsed, flow)
	}
	if parsed.Type == ExprTypeStaticProperty {
		return e.traceStaticProperty(parsed, flow)
	}

	// For object-based expressions, find instantiation
	className, instantiationFile, instantiationLine := e.findInstantiation(parsed.VarName, contextFile)
	if className == "" {
		return nil, fmt.Errorf("could not find instantiation of variable %s (searched %d files)", parsed.VarName, len(e.parsedFiles))
	}

	parsed.ClassName = className
	flow.ClassName = className

	// Find the class definition
	classDef, classFile := e.findClassDefinition(className)
	if classDef == nil {
		return nil, fmt.Errorf("could not find class definition for %s", className)
	}

	// GAP #4 FIX: Handle chained expressions like $obj->method()->property
	if parsed.IsChained {
		return e.traceChainedExpression(parsed, classDef, classFile, instantiationFile, instantiationLine, flow)
	}

	switch parsed.Type {
	case ExprTypeMethodCall:
		return e.traceMethodCall(parsed, classDef, classFile, instantiationFile, instantiationLine, flow)
	case ExprTypePropertyAccess:
		return e.tracePropertyAccessExpr(parsed, classDef, classFile, instantiationFile, instantiationLine, flow)
	default:
		return nil, fmt.Errorf("unsupported expression type: %v", parsed.Type)
	}
}

// traceSuperglobal handles direct superglobal access like $_GET['id']
func (e *ExecutionEngine) traceSuperglobal(parsed *ParsedExpression, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.PropertyName = parsed.AccessKey

	// Determine the source type
	var sourceType string
	switch parsed.SuperglobalName {
	case "$_GET":
		sourceType = "http_get"
	case "$_POST":
		sourceType = "http_post"
	case "$_COOKIE":
		sourceType = "http_cookie"
	case "$_REQUEST":
		sourceType = "http_request"
	case "$_SERVER":
		sourceType = "http_header"
	case "$_FILES":
		sourceType = "http_file"
	case "$_ENV":
		sourceType = "env_var"
	case "$_SESSION":
		sourceType = "session"
	default:
		sourceType = "unknown"
	}

	// Add step showing direct superglobal access
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Direct superglobal access: %s['%s']", parsed.SuperglobalName, parsed.AccessKey),
		Code:        fmt.Sprintf("%s['%s']", parsed.SuperglobalName, parsed.AccessKey),
		FilePath:    "",
		Line:        0,
		Type:        "direct_input",
	})

	// Add the source
	flow.Sources = append(flow.Sources, UltimateSource{
		Type:       sourceType,
		Expression: fmt.Sprintf("%s['%s']", parsed.SuperglobalName, parsed.AccessKey),
		FilePath:   "",
		Line:       0,
	})

	return flow, nil
}

// traceLocalVariable traces a local variable to find its source
func (e *ExecutionEngine) traceLocalVariable(parsed *ParsedExpression, contextFile string, flow *PropertyFlow) (*PropertyFlow, error) {
	varName := parsed.VarName

	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Tracing local variable: %s", varName),
		Code:        varName,
		FilePath:    "",
		Line:        0,
		Type:        "local_var",
	})

	// Search all files for assignments to this variable from superglobals
	foundAssignments := e.findVariableAssignments(varName)

	for i, assignment := range foundAssignments {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  i + 2,
			Description: fmt.Sprintf("Assignment found: %s = %s", varName, assignment.source),
			Code:        fmt.Sprintf("%s = %s;", varName, assignment.source),
			FilePath:    assignment.file,
			Line:        assignment.line,
			Type:        "assignment",
		})

		// Check if source is a superglobal
		for _, sg := range superglobals {
			if strings.Contains(assignment.source, sg.pattern) {
				flow.Sources = append(flow.Sources, UltimateSource{
					Type:       sg.sourceType,
					Expression: assignment.source,
					FilePath:   assignment.file,
					Line:       assignment.line,
				})
			}
		}
	}

	if len(foundAssignments) == 0 {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("No assignments found for %s in parsed files", varName),
			Code:        "",
			FilePath:    "",
			Line:        0,
			Type:        "not_found",
		})
	}

	return flow, nil
}

// variableAssignment represents an assignment to a variable
type variableAssignment struct {
	source string
	file   string
	line   int
}

// findVariableAssignments searches for assignments to a variable
func (e *ExecutionEngine) findVariableAssignments(varName string) []variableAssignment {
	var assignments []variableAssignment

	// Remove $ prefix for matching
	varNameClean := strings.TrimPrefix(varName, "$")

	// Search all file contents for assignments
	for file, content := range e.fileContents {
		// Pattern: $varname = something
		assignPattern := getOrCompileRegex(`\$` + regexp.QuoteMeta(varNameClean) + `\s*=\s*([^;]+);`)
		lines := strings.Split(string(content), "\n")

		for lineNum, line := range lines {
			if matches := assignPattern.FindStringSubmatch(line); len(matches) >= 2 {
				assignments = append(assignments, variableAssignment{
					source: strings.TrimSpace(matches[1]),
					file:   file,
					line:   lineNum + 1,
				})
			}
		}
	}

	return assignments
}

// findExternalPropertyAssignments searches all parsed files for external assignments
// to an object's property, like: $varName->propertyName = value;
// This handles dynamic properties assigned outside the class definition
func (e *ExecutionEngine) findExternalPropertyAssignments(varName string, propertyName string) []ExternalAssignment {
	var assignments []ExternalAssignment

	varNameWithoutDollar := strings.TrimPrefix(varName, "$")

	// Search all file contents for external property assignments
	for file, content := range e.fileContents {
		// Pattern: $varname->property = something
		// or $varname->property['key'] = something
		assignPatterns := []*regexp.Regexp{
			// $var->property = value;
			getOrCompileRegex(`\$` + regexp.QuoteMeta(varNameWithoutDollar) + `->` + regexp.QuoteMeta(propertyName) + `\s*=\s*([^;]+);`),
			// $var->property['key'] = value; (array element assignment)
			getOrCompileRegex(`\$` + regexp.QuoteMeta(varNameWithoutDollar) + `->` + regexp.QuoteMeta(propertyName) + `\[['"]?\w+['"]?\]\s*=\s*([^;]+);`),
		}

		lines := strings.Split(string(content), "\n")
		for lineNum, line := range lines {
			for _, pattern := range assignPatterns {
				if matches := pattern.FindStringSubmatch(line); len(matches) >= 2 {
					assignments = append(assignments, ExternalAssignment{
						PropertyName: propertyName,
						Source:       strings.TrimSpace(matches[1]),
						FilePath:     file,
						Line:         lineNum + 1,
					})
				}
			}
		}
	}

	return assignments
}

// traceExternalPropertyAssignment traces an externally assigned property
// This is for properties like $mybb->post_code that are assigned outside the class
func (e *ExecutionEngine) traceExternalPropertyAssignment(parsed *ParsedExpression, externalAssignments []ExternalAssignment, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.PropertyName = parsed.PropertyName
	flow.AccessKey = parsed.AccessKey

	// Show that this is a dynamic property (not defined in class)
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Dynamic property $%s (not defined in class)", parsed.PropertyName),
		Code:        fmt.Sprintf("// $%s is assigned externally, not in class definition", parsed.PropertyName),
		FilePath:    "",
		Line:        0,
		Type:        "dynamic_property",
	})

	// Show all external assignments
	for i, assign := range externalAssignments {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  i + 2,
			Description: fmt.Sprintf("External assignment: %s->%s = %s", parsed.VarName, parsed.PropertyName, assign.Source),
			Code:        fmt.Sprintf("%s->%s = %s;", parsed.VarName, parsed.PropertyName, assign.Source),
			FilePath:    assign.FilePath,
			Line:        assign.Line,
			Type:        "external_assignment",
		})

		// Check if source contains superglobals
		for _, sg := range superglobals {
			if strings.Contains(assign.Source, sg.pattern) {
				flow.Sources = append(flow.Sources, UltimateSource{
					Type:       sg.sourceType,
					Expression: sg.pattern,
					FilePath:   assign.FilePath,
					Line:       assign.Line,
				})
			}
		}

		// If source is a function call, trace into that function
		funcCallPattern := getOrCompileRegex(`^(\w+)\(([^)]*)\)$`)
		if matches := funcCallPattern.FindStringSubmatch(assign.Source); len(matches) >= 2 {
			funcName := matches[1]
			funcArgs := ""
			if len(matches) >= 3 {
				funcArgs = matches[2]
			}

			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  len(flow.Steps) + 1,
				Description: fmt.Sprintf("Calls function: %s(%s)", funcName, funcArgs),
				Code:        fmt.Sprintf("%s(%s)", funcName, funcArgs),
				FilePath:    assign.FilePath,
				Line:        assign.Line,
				Type:        "function_call",
			})

			// Try to find and trace the function
			funcSources := e.traceFunctionForSources(funcName)
			flow.Sources = append(flow.Sources, funcSources...)
		}

		// Check for variable assignments (e.g., $mybb->request_method = $_SERVER['REQUEST_METHOD'])
		if strings.HasPrefix(assign.Source, "$_") {
			// Direct superglobal assignment
			for _, sg := range superglobals {
				if strings.Contains(assign.Source, sg.pattern) {
					flow.Sources = append(flow.Sources, UltimateSource{
						Type:       sg.sourceType,
						Expression: assign.Source,
						FilePath:   assign.FilePath,
						Line:       assign.Line,
					})
				}
			}
		}
	}

	return flow, nil
}

// traceFunctionForSources traces a function to find what superglobals it uses
func (e *ExecutionEngine) traceFunctionForSources(funcName string) []UltimateSource {
	var sources []UltimateSource

	// Search all symbol tables for the function
	for filePath, st := range e.symbolTables {
		if funcDef, ok := st.Functions[funcName]; ok {
			// Check function body for superglobal usage
			if funcDef.BodySource != "" {
				for _, sg := range superglobals {
					if strings.Contains(funcDef.BodySource, sg.pattern) {
						sources = append(sources, UltimateSource{
							Type:       sg.sourceType,
							Expression: sg.pattern,
							FilePath:   filePath,
							Line:       funcDef.Line,
						})
					}
				}
			}
		}
	}

	return sources
}

// MagicPropertyInfo holds information about magic property access patterns
type MagicPropertyInfo struct {
	HasMagicGet        bool   // Class has __get method
	HasDynamicAssign   bool   // Class has $this->$var = $val pattern
	BackingProperty    string // Property used for storage (e.g., "phrases")
	AssignMethodName   string // Method that assigns properties
	SourceType         string // "file_include", "array", etc.
}

// checkMagicPropertyPattern checks if a class uses magic __get or dynamic property assignment
func (e *ExecutionEngine) checkMagicPropertyPattern(classDef *types.ClassDef, classFile string) *MagicPropertyInfo {
	info := &MagicPropertyInfo{}

	// Check for __get magic method
	if method, ok := classDef.Methods["__get"]; ok {
		info.HasMagicGet = true
		// Look for return $this->property[$name] pattern
		backingPattern := getOrCompileRegex(`return\s+\$this->(\w+)\[\$\w+\]`)
		if matches := backingPattern.FindStringSubmatch(method.BodySource); len(matches) >= 2 {
			info.BackingProperty = matches[1]
		}
		return info
	}

	// Check for dynamic property assignment pattern: $this->$key = $val
	// This is used in classes like MyLanguage that load properties dynamically
	dynamicAssignPattern := getOrCompileRegex(`\$this->\$(\w+)\s*=\s*\$(\w+)`)
	for methodName, method := range classDef.Methods {
		if method.BodySource != "" {
			if dynamicAssignPattern.MatchString(method.BodySource) {
				info.HasDynamicAssign = true
				info.AssignMethodName = methodName

				// Check if values come from require/include (file_include)
				if strings.Contains(method.BodySource, "require") || strings.Contains(method.BodySource, "include") {
					info.SourceType = "file_include"
				}

				// Check for foreach pattern: foreach($array as $key => $val) { $this->$key = $val; }
				foreachPattern := getOrCompileRegex(`foreach\s*\(\s*\$(\w+)\s+as\s+\$\w+\s*=>\s*\$\w+`)
				if matches := foreachPattern.FindStringSubmatch(method.BodySource); len(matches) >= 2 {
					info.BackingProperty = matches[1]
				}

				return info
			}
		}
	}

	return nil
}

// traceMagicProperty traces a property accessed via magic method or dynamic assignment
func (e *ExecutionEngine) traceMagicProperty(parsed *ParsedExpression, classDef *types.ClassDef, classFile string, magicInfo *MagicPropertyInfo, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.PropertyName = parsed.PropertyName
	flow.AccessKey = parsed.AccessKey

	if magicInfo.HasMagicGet {
		// Magic __get method
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  1,
			Description: fmt.Sprintf("Dynamic property $%s accessed via __get magic method", parsed.PropertyName),
			Code:        fmt.Sprintf("public function __get($name) { return $this->%s[$name]; }", magicInfo.BackingProperty),
			FilePath:    classFile,
			Line:        0,
			Type:        "magic_get",
		})

		if magicInfo.BackingProperty != "" {
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  2,
				Description: fmt.Sprintf("Returns $this->%s['%s']", magicInfo.BackingProperty, parsed.PropertyName),
				Code:        fmt.Sprintf("return $this->%s['%s'];", magicInfo.BackingProperty, parsed.PropertyName),
				FilePath:    classFile,
				Line:        0,
				Type:        "return",
			})
		}
	} else if magicInfo.HasDynamicAssign {
		// Dynamic property assignment pattern
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  1,
			Description: fmt.Sprintf("Dynamic property $%s assigned via %s() method", parsed.PropertyName, magicInfo.AssignMethodName),
			Code:        fmt.Sprintf("// In %s(): $this->$key = $val;", magicInfo.AssignMethodName),
			FilePath:    classFile,
			Line:        0,
			Type:        "dynamic_assignment",
		})

		if magicInfo.SourceType == "file_include" {
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  2,
				Description: "Property values loaded from included PHP files",
				Code:        fmt.Sprintf("// require/include loads data that populates $%s", parsed.PropertyName),
				FilePath:    classFile,
				Line:        0,
				Type:        "file_include",
			})

			// For language files, the source is file data not user input
			flow.Sources = append(flow.Sources, UltimateSource{
				Type:       "file_data",
				Expression: fmt.Sprintf("$lang->%s (loaded from language file)", parsed.PropertyName),
				FilePath:   classFile,
				Line:       0,
			})
		}

		if magicInfo.BackingProperty != "" {
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  3,
				Description: fmt.Sprintf("Data iterated from $%s array", magicInfo.BackingProperty),
				Code:        fmt.Sprintf("foreach($%s as $key => $val) { $this->$key = $val; }", magicInfo.BackingProperty),
				FilePath:    classFile,
				Line:        0,
				Type:        "loop",
			})
		}
	}

	return flow, nil
}

// traceStaticCall handles static method calls like Class::method()
func (e *ExecutionEngine) traceStaticCall(parsed *ParsedExpression, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.ClassName = parsed.ClassName
	flow.MethodName = parsed.MethodName

	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Static method call: %s::%s()", parsed.ClassName, parsed.MethodName),
		Code:        fmt.Sprintf("%s::%s(%s)", parsed.ClassName, parsed.MethodName, strings.Join(parsed.Arguments, ", ")),
		FilePath:    "",
		Line:        0,
		Type:        "static_call",
	})

	// Find the class definition
	classDef, classFile := e.findClassDefinition(parsed.ClassName)
	if classDef == nil {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Class %s not found", parsed.ClassName),
			Code:        "",
			FilePath:    "",
			Line:        0,
			Type:        "not_found",
		})
		return flow, nil
	}

	// Find the method
	methodDef, ok := classDef.Methods[parsed.MethodName]
	if !ok {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Static method %s not found in class %s", parsed.MethodName, parsed.ClassName),
			Code:        "",
			FilePath:    "",
			Line:        0,
			Type:        "not_found",
		})
		return flow, nil
	}

	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  2,
		Description: fmt.Sprintf("Method %s::%s() defined", parsed.ClassName, parsed.MethodName),
		Code:        fmt.Sprintf("public static function %s() { ... }", parsed.MethodName),
		FilePath:    classFile,
		Line:        methodDef.Line,
		Type:        "method_def",
	})

	// Analyze what the method returns
	returnInfo := e.analyzeMethodReturns(classDef, methodDef, classFile)
	for _, retStmt := range returnInfo.ReturnStatements {
		for _, sg := range superglobals {
			if strings.Contains(retStmt, sg.pattern) {
				flow.Sources = append(flow.Sources, UltimateSource{
					Type:       sg.sourceType,
					Expression: sg.pattern,
					FilePath:   classFile,
					Line:       methodDef.Line,
				})
			}
		}
	}

	return flow, nil
}

// traceStaticProperty handles static property access like Class::$property or Class::CONSTANT
func (e *ExecutionEngine) traceStaticProperty(parsed *ParsedExpression, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.ClassName = parsed.ClassName
	flow.PropertyName = parsed.PropertyName

	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Static property/constant access: %s::%s", parsed.ClassName, parsed.PropertyName),
		Code:        fmt.Sprintf("%s::%s", parsed.ClassName, parsed.PropertyName),
		FilePath:    "",
		Line:        0,
		Type:        "static_property",
	})

	// Find the class definition
	classDef, classFile := e.findClassDefinition(parsed.ClassName)
	if classDef == nil {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Class %s not found", parsed.ClassName),
			Code:        "",
			FilePath:    "",
			Line:        0,
			Type:        "not_found",
		})
		return flow, nil
	}

	// Check for property
	if propDef, ok := classDef.Properties[parsed.PropertyName]; ok {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Static property %s::%s = %s", parsed.ClassName, parsed.PropertyName, propDef.InitialValue),
			Code:        fmt.Sprintf("public static $%s = %s;", parsed.PropertyName, propDef.InitialValue),
			FilePath:    classFile,
			Line:        propDef.Line,
			Type:        "property_def",
		})
	} else {
		// It might be a constant - check for constants in the class body
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Constant %s::%s (value lookup not implemented)", parsed.ClassName, parsed.PropertyName),
			Code:        fmt.Sprintf("const %s = ...;", parsed.PropertyName),
			FilePath:    classFile,
			Line:        0,
			Type:        "constant",
		})
	}

	return flow, nil
}

// traceChainedExpression traces expressions like $obj->method()->property
// GAP #4 FIX: Support chained method calls
func (e *ExecutionEngine) traceChainedExpression(parsed *ParsedExpression, classDef *types.ClassDef, classFile string, instFile string, instLine int, flow *PropertyFlow) (*PropertyFlow, error) {
	stepNum := 1

	// Step 1: Show instantiation
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  stepNum,
		Description: fmt.Sprintf("Variable %s instantiated as new %s()", parsed.VarName, classDef.Name),
		Code:        fmt.Sprintf("%s = new %s();", parsed.VarName, classDef.Name),
		FilePath:    instFile,
		Line:        instLine,
		Type:        "instantiation",
	})
	stepNum++

	// Track the current class as we traverse the chain
	currentClass := classDef
	currentClassFile := classFile

	// Process each step in the chain
	for i, step := range parsed.ChainSteps {
		isLastStep := i == len(parsed.ChainSteps)-1

		if step.Type == ExprTypeMethodCall {
			// Find the method
			methodDef, ok := currentClass.Methods[step.Name]
			if !ok {
				flow.Steps = append(flow.Steps, FlowStep{
					StepNumber:  stepNum,
					Description: fmt.Sprintf("Method %s() not found in class %s", step.Name, currentClass.Name),
					Code:        "",
					FilePath:    currentClassFile,
					Line:        0,
					Type:        "method_not_found",
				})
				break
			}

			// Show method call
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  stepNum,
				Description: fmt.Sprintf("Method call: ->%s(%s)", step.Name, strings.Join(step.Arguments, ", ")),
				Code:        fmt.Sprintf("function %s(%s) { ... }", step.Name, e.formatParams(methodDef.Parameters)),
				FilePath:    currentClassFile,
				Line:        methodDef.Line,
				Type:        "method_call",
			})
			stepNum++

			// Analyze method return type for next step
			returnInfo := e.analyzeMethodReturns(currentClass, methodDef, currentClassFile)

			// If method returns a property, and this is the last step, trace the property
			if isLastStep {
				if returnInfo.ReturnsProperty && returnInfo.PropertyName != "" {
					flow.PropertyName = returnInfo.PropertyName
					flow.MethodName = step.Name

					// Trace sources from that property
					propSteps := e.traceConstructor(currentClass, currentClassFile, returnInfo.PropertyName, step.AccessKey)
					for _, ps := range propSteps {
						ps.StepNumber = stepNum
						flow.Steps = append(flow.Steps, ps)
						stepNum++
					}
					// Extract sources using the existing extractSources function
					flow.Sources = append(flow.Sources, e.extractSources(propSteps)...)
				}
				break
			}

			// Check if method returns $this (fluent interface)
			if returnInfo.ReturnsSelf {
				// Continue with same class
				continue
			}

			// Try to infer return type from method
			returnType := e.inferMethodReturnType(currentClass, methodDef, currentClassFile)
			if returnType != "" {
				newClass, newClassFile := e.findClassDefinition(returnType)
				if newClass != nil {
					currentClass = newClass
					currentClassFile = newClassFile
					continue
				}
			}

			// Can't determine return type - provide partial trace
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  stepNum,
				Description: fmt.Sprintf("Cannot determine return type of %s() - chain tracing stopped", step.Name),
				Code:        "",
				FilePath:    "",
				Line:        0,
				Type:        "return_type_unknown",
			})
			break

		} else if step.Type == ExprTypePropertyAccess {
			// Property access
			propDef, ok := currentClass.Properties[step.Name]
			if !ok {
				// Check for magic methods or external assignments
				flow.Steps = append(flow.Steps, FlowStep{
					StepNumber:  stepNum,
					Description: fmt.Sprintf("Property %s not found in class %s", step.Name, currentClass.Name),
					Code:        "",
					FilePath:    currentClassFile,
					Line:        0,
					Type:        "property_not_found",
				})
				break
			}

			flow.PropertyName = step.Name
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  stepNum,
				Description: fmt.Sprintf("Property access: ->%s", step.Name),
				Code:        fmt.Sprintf("$this->%s = %s;", step.Name, propDef.InitialValue),
				FilePath:    currentClassFile,
				Line:        propDef.Line,
				Type:        "property_access",
			})
			stepNum++

			if isLastStep {
				// Trace the property sources
				propSteps := e.traceConstructor(currentClass, currentClassFile, step.Name, step.AccessKey)
				for _, ps := range propSteps {
					ps.StepNumber = stepNum
					flow.Steps = append(flow.Steps, ps)
					stepNum++
				}
				// Extract sources using the existing extractSources function
				flow.Sources = append(flow.Sources, e.extractSources(propSteps)...)
			}
		}
	}

	return flow, nil
}

// inferMethodReturnType tries to determine what class a method returns
func (e *ExecutionEngine) inferMethodReturnType(classDef *types.ClassDef, methodDef *types.MethodDef, classFile string) string {
	// Check for explicit return type annotation first
	// Pattern: function name(): ReturnType
	if methodDef.ReturnType != "" && methodDef.ReturnType != "void" && methodDef.ReturnType != "self" && methodDef.ReturnType != "static" {
		return methodDef.ReturnType
	}

	// Use the method body source
	body := methodDef.BodySource
	if body == "" {
		return ""
	}

	// Check for return new ClassName()
	newPattern := getOrCompileRegex(`return\s+new\s+(\w+)\(`)
	if matches := newPattern.FindStringSubmatch(body); len(matches) >= 2 {
		return matches[1]
	}

	// Check for @return PHPDoc annotation
	returnDocPattern := getOrCompileRegex(`@return\s+(\w+)`)
	if matches := returnDocPattern.FindStringSubmatch(body); len(matches) >= 2 {
		returnType := matches[1]
		if returnType != "void" && returnType != "self" && returnType != "static" && returnType != "mixed" {
			return returnType
		}
	}

	return ""
}

// parseExpression parses any expression and determines its type
func (e *ExecutionEngine) parseExpression(expr string) *ParsedExpression {
	parsed := &ParsedExpression{
		Type:    ExprTypeUnknown,
		RawExpr: expr,
	}

	expr = strings.TrimSpace(expr)

	// GAP #4 FIX: Try chained expression parsing first
	// This handles expressions like: $obj->method()->property or $obj->method1()->method2('arg')
	if strings.Count(expr, "->") > 1 {
		if chainedParsed := e.parseChainedExpression(expr); chainedParsed != nil {
			return chainedParsed
		}
	}

	// GAP #1 FIX: Try superglobal pattern first: $_GET['key'], $_POST['key'], etc.
	// Pattern: $_SUPERGLOBAL['key'] or $_SUPERGLOBAL["key"]
	superglobalPattern := getOrCompileRegex(`^\$_(GET|POST|COOKIE|REQUEST|SERVER|FILES|ENV|SESSION)\[['"]?(\w+)['"]?\]$`)
	if matches := superglobalPattern.FindStringSubmatch(expr); len(matches) >= 3 {
		parsed.Type = ExprTypeSuperglobal
		parsed.SuperglobalName = "$_" + matches[1]
		parsed.AccessKey = matches[2]
		parsed.IsSuperglobal = true
		parsed.VarName = parsed.SuperglobalName // For compatibility
		return parsed
	}

	// GAP #3 FIX: Try static method call pattern: Class::method('arg') with proper paren handling
	if className, methodName, argsStr, ok := e.extractStaticMethodCall(expr); ok {
		parsed.Type = ExprTypeStaticCall
		parsed.ClassName = className
		parsed.MethodName = methodName
		if argsStr != "" {
			parsed.Arguments = e.parseArguments(argsStr)
			if len(parsed.Arguments) > 0 {
				parsed.AccessKey = strings.Trim(parsed.Arguments[0], "'\"")
			}
		}
		return parsed
	}

	// GAP #3 FIX: Try static property/constant pattern: Class::$property or Class::CONSTANT
	staticPropPattern := getOrCompileRegex(`^(\w+)::\$?(\w+)$`)
	if matches := staticPropPattern.FindStringSubmatch(expr); len(matches) >= 3 {
		parsed.Type = ExprTypeStaticProperty
		parsed.ClassName = matches[1]
		parsed.PropertyName = matches[2]
		return parsed
	}

	// Try method call pattern using smart extraction that handles nested parens
	// This handles: $var->method('arg'), $var->method(func($x)), $var->method("string with (parens)")
	if varName, methodName, argsStr, ok := e.extractMethodCall(expr); ok {
		parsed.Type = ExprTypeMethodCall
		parsed.VarName = varName
		parsed.MethodName = methodName
		if argsStr != "" {
			parsed.Arguments = e.parseArguments(argsStr)
			if len(parsed.Arguments) > 0 {
				parsed.AccessKey = strings.Trim(parsed.Arguments[0], "'\"")
			}
		}
		return parsed
	}

	// Try property access pattern: $var->property or $var->property['key']
	propPattern := getOrCompileRegex(`^\$(\w+)->(\w+)(?:\[['"]?(\w+)['"]?\])?$`)
	if matches := propPattern.FindStringSubmatch(expr); len(matches) >= 3 {
		parsed.Type = ExprTypePropertyAccess
		parsed.VarName = "$" + matches[1]
		parsed.PropertyName = matches[2]
		if len(matches) >= 4 {
			parsed.AccessKey = matches[3]
		}
		return parsed
	}

	// GAP #2 FIX: Try simple local variable pattern: $varname
	// This must come LAST as it's the most generic pattern
	localVarPattern := getOrCompileRegex(`^\$(\w+)$`)
	if matches := localVarPattern.FindStringSubmatch(expr); len(matches) >= 2 {
		parsed.Type = ExprTypeLocalVariable
		parsed.VarName = "$" + matches[1]
		return parsed
	}

	return parsed
}

// parseChainedExpression parses expressions like $obj->method()->property or $obj->method1()->method2('arg')
// GAP #4 FIX: Support chained method calls
func (e *ExecutionEngine) parseChainedExpression(expr string) *ParsedExpression {
	expr = strings.TrimSpace(expr)

	// Must start with a variable
	if !strings.HasPrefix(expr, "$") {
		return nil
	}

	// Find the base variable name
	varNameEnd := strings.Index(expr, "->")
	if varNameEnd == -1 {
		return nil
	}

	basePart := expr[1:varNameEnd] // Remove $ prefix
	if !getOrCompileRegex(`^\w+$`).MatchString(basePart) {
		return nil
	}
	varName := "$" + basePart

	// Parse the chain steps
	remainder := expr[varNameEnd:]
	var steps []ChainStep

	// Patterns for parsing chain steps (property access only - method calls use extractChainMethodCall)
	propWithKeyPattern := getOrCompileRegex(`^->(\w+)\[['"]?(\w+)['"]?\]`)
	propPattern := getOrCompileRegex(`^->(\w+)`)

	for len(remainder) > 0 && strings.HasPrefix(remainder, "->") {
		var step ChainStep
		matched := false

		// Try method call first: ->method(args) with proper paren handling
		if methodName, argsStr, consumed, ok := e.extractChainMethodCall(remainder); ok {
			step.Type = ExprTypeMethodCall
			step.Name = methodName
			if argsStr != "" {
				step.Arguments = e.parseArguments(argsStr)
				if len(step.Arguments) > 0 {
					step.AccessKey = strings.Trim(step.Arguments[0], "'\"")
				}
			}
			steps = append(steps, step)
			remainder = remainder[consumed:]
			matched = true
		}

		// Try property with key: ->property['key']
		if !matched {
			if matches := propWithKeyPattern.FindStringSubmatch(remainder); len(matches) >= 3 {
				step.Type = ExprTypePropertyAccess
				step.Name = matches[1]
				step.AccessKey = matches[2]
				steps = append(steps, step)
				remainder = remainder[len(matches[0]):]
				matched = true
			}
		}

		// Try simple property: ->property
		if !matched {
			if matches := propPattern.FindStringSubmatch(remainder); len(matches) >= 2 {
				step.Type = ExprTypePropertyAccess
				step.Name = matches[1]
				steps = append(steps, step)
				remainder = remainder[len(matches[0]):]
				matched = true
			}
		}

		// No pattern matched, fail
		if !matched {
			return nil
		}
	}

	// Must have at least 2 steps to be a chain
	if len(steps) < 2 {
		return nil
	}

	// We have leftover text that didn't match
	if len(strings.TrimSpace(remainder)) > 0 {
		return nil
	}

	// Build the parsed expression
	// For tracing purposes, we use the first step info
	parsed := &ParsedExpression{
		VarName:    varName,
		IsChained:  true,
		ChainSteps: steps,
	}

	// Set the type based on the first step
	if steps[0].Type == ExprTypeMethodCall {
		parsed.Type = ExprTypeMethodCall
		parsed.MethodName = steps[0].Name
		parsed.Arguments = steps[0].Arguments
		parsed.AccessKey = steps[0].AccessKey
	} else {
		parsed.Type = ExprTypePropertyAccess
		parsed.PropertyName = steps[0].Name
		parsed.AccessKey = steps[0].AccessKey
	}

	return parsed
}

// parseArguments splits method arguments with proper handling of nested parentheses and strings
// This handles complex cases like: "users", "uid='".$mybb->get_input('uid')."'"
func (e *ExecutionEngine) parseArguments(argsStr string) []string {
	var args []string
	var current strings.Builder
	depth := 0
	inString := false
	var stringChar byte = 0

	for i := 0; i < len(argsStr); i++ {
		c := argsStr[i]

		// Handle escape sequences in strings
		if inString && c == '\\' && i+1 < len(argsStr) {
			current.WriteByte(c)
			i++
			current.WriteByte(argsStr[i])
			continue
		}

		// Handle string boundaries
		if (c == '"' || c == '\'') {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar {
				inString = false
				stringChar = 0
			}
			current.WriteByte(c)
			continue
		}

		// Track parenthesis depth (only outside strings)
		if !inString {
			if c == '(' || c == '[' {
				depth++
			} else if c == ')' || c == ']' {
				depth--
			} else if c == ',' && depth == 0 {
				// Argument separator at top level
				arg := strings.TrimSpace(current.String())
				if arg != "" {
					args = append(args, arg)
				}
				current.Reset()
				continue
			}
		}

		current.WriteByte(c)
	}

	// Add the last argument
	if current.Len() > 0 {
		arg := strings.TrimSpace(current.String())
		if arg != "" {
			args = append(args, arg)
		}
	}

	return args
}

// extractMethodCall extracts $var->method(args) with proper handling of nested parentheses
// Returns: varName, methodName, argsStr, success
// This handles complex expressions like:
//   - $db->simple_select("users", "uid='".$mybb->get_input('uid')."'")
//   - $db->write_query("ALTER TABLE posts ADD INDEX (tid)")
//   - $db->escape_string(trim($input))
func (e *ExecutionEngine) extractMethodCall(expr string) (string, string, string, bool) {
	expr = strings.TrimSpace(expr)

	// Must start with $
	if !strings.HasPrefix(expr, "$") {
		return "", "", "", false
	}

	// Find the variable name (ends at ->)
	arrowIdx := strings.Index(expr, "->")
	if arrowIdx == -1 {
		return "", "", "", false
	}

	varName := expr[:arrowIdx]
	// Validate variable name: $word
	if !getOrCompileRegex(`^\$\w+$`).MatchString(varName) {
		return "", "", "", false
	}

	// Find the method name (from after -> to the opening paren)
	remainder := expr[arrowIdx+2:]
	parenIdx := strings.Index(remainder, "(")
	if parenIdx == -1 {
		return "", "", "", false // Not a method call (no parens)
	}

	methodName := remainder[:parenIdx]
	// Validate method name: word characters only
	if !getOrCompileRegex(`^\w+$`).MatchString(methodName) {
		return "", "", "", false
	}

	// Extract arguments by finding the matching closing paren
	argsStart := parenIdx + 1
	afterParen := remainder[argsStart:]

	// Find the matching closing paren, respecting nesting and strings
	argsEnd, ok := findMatchingParen(afterParen)
	if !ok {
		return "", "", "", false
	}

	argsStr := afterParen[:argsEnd]

	// Verify that after the closing paren there's nothing (or just whitespace)
	// This prevents partial matching of longer expressions
	afterArgs := strings.TrimSpace(afterParen[argsEnd+1:])
	if afterArgs != "" {
		// There's more after the method call - this might be a chained call
		// For now, just take what we have
		// In future, we could handle chaining here
	}

	return varName, methodName, argsStr, true
}

// extractStaticMethodCall extracts Class::method(args)
// Returns: className, methodName, argsStr, success
func (e *ExecutionEngine) extractStaticMethodCall(expr string) (string, string, string, bool) {
	expr = strings.TrimSpace(expr)

	// Find the :: separator
	colonIdx := strings.Index(expr, "::")
	if colonIdx == -1 {
		return "", "", "", false
	}

	className := expr[:colonIdx]
	if !getOrCompileRegex(`^\w+$`).MatchString(className) {
		return "", "", "", false
	}

	remainder := expr[colonIdx+2:]

	// Find method name (ends at opening paren)
	parenIdx := strings.Index(remainder, "(")
	if parenIdx == -1 {
		return "", "", "", false // Not a method call
	}

	methodName := remainder[:parenIdx]
	if !getOrCompileRegex(`^\w+$`).MatchString(methodName) {
		return "", "", "", false
	}

	// Find matching closing paren
	afterParen := remainder[parenIdx+1:]
	argsEnd, ok := findMatchingParen(afterParen)
	if !ok {
		return "", "", "", false
	}

	argsStr := afterParen[:argsEnd]

	// Verify nothing after the method call
	afterArgs := strings.TrimSpace(afterParen[argsEnd+1:])
	if afterArgs != "" {
		return "", "", "", false
	}

	return className, methodName, argsStr, true
}

// extractChainMethodCall extracts a method call from a chain step starting with ->
// Input: "->method(args)..."
// Returns: methodName, argsStr, bytesConsumed, success
func (e *ExecutionEngine) extractChainMethodCall(s string) (string, string, int, bool) {
	// Must start with ->
	if !strings.HasPrefix(s, "->") {
		return "", "", 0, false
	}

	remainder := s[2:] // Skip ->

	// Find method name (word characters until '(')
	parenIdx := strings.Index(remainder, "(")
	if parenIdx == -1 {
		return "", "", 0, false
	}

	methodName := remainder[:parenIdx]
	if !getOrCompileRegex(`^\w+$`).MatchString(methodName) {
		return "", "", 0, false
	}

	// Find matching closing paren
	afterParen := remainder[parenIdx+1:]
	argsEnd, ok := findMatchingParen(afterParen)
	if !ok {
		return "", "", 0, false
	}

	argsStr := afterParen[:argsEnd]
	// Total consumed: 2 (for ->) + parenIdx + 1 (for open paren) + argsEnd + 1 (for close paren)
	consumed := 2 + parenIdx + 1 + argsEnd + 1

	return methodName, argsStr, consumed, true
}

// findMatchingParen finds the position of the closing paren that matches the implicit opening paren
// Input is the string AFTER the opening paren
// Returns the index of the matching ')' and success flag
func findMatchingParen(s string) (int, bool) {
	depth := 1
	inString := false
	var stringChar byte = 0

	for i := 0; i < len(s); i++ {
		c := s[i]

		// Handle escape sequences in strings
		if inString && c == '\\' && i+1 < len(s) {
			i++ // Skip the escaped character
			continue
		}

		// Handle string boundaries
		if c == '"' || c == '\'' {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar {
				inString = false
				stringChar = 0
			}
			continue
		}

		// Track parenthesis depth (only outside strings)
		if !inString {
			if c == '(' {
				depth++
			} else if c == ')' {
				depth--
				if depth == 0 {
					return i, true
				}
			}
		}
	}

	return -1, false
}

// traceMethodCall traces a method call expression like $mybb->get_input('timezone')
func (e *ExecutionEngine) traceMethodCall(parsed *ParsedExpression, classDef *types.ClassDef, classFile string, instFile string, instLine int, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.MethodName = parsed.MethodName
	flow.AccessKey = parsed.AccessKey

	// Find the method definition
	methodDef, ok := classDef.Methods[parsed.MethodName]
	if !ok {
		return nil, fmt.Errorf("method %s not found in class %s", parsed.MethodName, parsed.ClassName)
	}

	// Step 1: Show instantiation
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Variable %s instantiated as new %s()", parsed.VarName, parsed.ClassName),
		Code:        fmt.Sprintf("%s = new %s();", parsed.VarName, parsed.ClassName),
		FilePath:    instFile,
		Line:        instLine,
		Type:        "instantiation",
	})

	// Step 2: Show method call
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  2,
		Description: fmt.Sprintf("Method call: %s->%s('%s')", parsed.VarName, parsed.MethodName, parsed.AccessKey),
		Code:        fmt.Sprintf("%s->%s('%s')", parsed.VarName, parsed.MethodName, parsed.AccessKey),
		FilePath:    "",
		Line:        0,
		Type:        "method_call",
	})

	// Step 3: Analyze the method body to find what it returns
	returnInfo := e.analyzeMethodReturns(classDef, methodDef, classFile)

	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  3,
		Description: fmt.Sprintf("Method %s() defined", parsed.MethodName),
		Code:        fmt.Sprintf("function %s(%s) { ... }", parsed.MethodName, e.formatParams(methodDef.Parameters)),
		FilePath:    classFile,
		Line:        methodDef.Line,
		Type:        "method_def",
	})

	// Step 4: Show return analysis
	if returnInfo.ReturnsProperty {
		propName := returnInfo.PropertyName

		if returnInfo.UsesParamAsKey {
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  4,
				Description: fmt.Sprintf("Returns $this->%s[$%s] where $%s = '%s'",
					propName, methodDef.Parameters[returnInfo.ParamIndex].Name,
					methodDef.Parameters[returnInfo.ParamIndex].Name, parsed.AccessKey),
				Code:        fmt.Sprintf("return $this->%s[$%s];", propName, methodDef.Parameters[returnInfo.ParamIndex].Name),
				FilePath:    classFile,
				Line:        methodDef.Line,
				Type:        "return",
			})

			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  5,
				Description: fmt.Sprintf("Resolves to: $this->%s['%s']", propName, parsed.AccessKey),
				Code:        fmt.Sprintf("// %s->%s('%s') == $this->%s['%s']", parsed.VarName, parsed.MethodName, parsed.AccessKey, propName, parsed.AccessKey),
				FilePath:    "",
				Line:        0,
				Type:        "resolution",
			})

			// Now trace the property that the method returns
			flow.PropertyName = propName

			// Find property definition
			if propDef, ok := classDef.Properties[propName]; ok {
				flow.Steps = append(flow.Steps, FlowStep{
					StepNumber:  6,
					Description: fmt.Sprintf("Property $%s starts as %s", propName, propDef.InitialValue),
					Code:        fmt.Sprintf("public $%s = %s;", propName, propDef.InitialValue),
					FilePath:    classFile,
					Line:        propDef.Line,
					Type:        "property_init",
				})
			}

			// Trace constructor to see how property is populated
			if classDef.Constructor != nil {
				e.currentDepth = 0
				constructorFlows := e.traceConstructor(classDef, classFile, propName, parsed.AccessKey)

				// Renumber steps
				for i := range constructorFlows {
					constructorFlows[i].StepNumber = len(flow.Steps) + i + 1
				}
				flow.Steps = append(flow.Steps, constructorFlows...)
			}

		} else {
			// Returns property directly without key
			flow.Steps = append(flow.Steps, FlowStep{
				StepNumber:  4,
				Description: fmt.Sprintf("Returns $this->%s", propName),
				Code:        fmt.Sprintf("return $this->%s;", propName),
				FilePath:    classFile,
				Line:        methodDef.Line,
				Type:        "return",
			})
		}
	}

	// Add return statements analysis
	for _, retStmt := range returnInfo.ReturnStatements {
		// Check for superglobals in return statements
		for _, sg := range superglobals {
			if strings.Contains(retStmt, sg.pattern) {
				flow.Sources = append(flow.Sources, UltimateSource{
					Type:       sg.sourceType,
					Expression: sg.pattern,
					FilePath:   classFile,
					Line:       methodDef.Line,
				})
			}
		}
	}

	// Extract sources from steps
	flow.Sources = append(flow.Sources, e.extractSources(flow.Steps)...)

	return flow, nil
}

// tracePropertyAccessExpr traces a property access expression
func (e *ExecutionEngine) tracePropertyAccessExpr(parsed *ParsedExpression, classDef *types.ClassDef, classFile string, instFile string, instLine int, flow *PropertyFlow) (*PropertyFlow, error) {
	flow.PropertyName = parsed.PropertyName
	flow.AccessKey = parsed.AccessKey

	// Find the property definition
	propDef, found := classDef.Properties[parsed.PropertyName]
	if !found {
		// GAP #6 FIX: Property not found in class definition
		// Check for external property assignments (dynamic properties)
		externalAssignments := e.findExternalPropertyAssignments(parsed.VarName, parsed.PropertyName)
		if len(externalAssignments) > 0 {
			return e.traceExternalPropertyAssignment(parsed, externalAssignments, flow)
		}
		// GAP #6 FIX Phase 2: Check for magic __get method or dynamic property assignment
		magicInfo := e.checkMagicPropertyPattern(classDef, classFile)
		if magicInfo != nil {
			return e.traceMagicProperty(parsed, classDef, classFile, magicInfo, flow)
		}
		return nil, fmt.Errorf("property %s not found in class %s", parsed.PropertyName, parsed.ClassName)
	}

	// Add step for property initialization
	flow.Steps = append(flow.Steps, FlowStep{
		StepNumber:  1,
		Description: fmt.Sprintf("Property $%s starts as %s", parsed.PropertyName, propDef.InitialValue),
		Code:        fmt.Sprintf("public $%s = %s;", parsed.PropertyName, propDef.InitialValue),
		FilePath:    classFile,
		Line:        propDef.Line,
		Type:        "property_init",
	})

	// If variable is instantiated, show that
	if instFile != "" {
		flow.Steps = append(flow.Steps, FlowStep{
			StepNumber:  2,
			Description: fmt.Sprintf("Variable %s instantiated as new %s()", parsed.VarName, parsed.ClassName),
			Code:        fmt.Sprintf("%s = new %s();", parsed.VarName, parsed.ClassName),
			FilePath:    instFile,
			Line:        instLine,
			Type:        "instantiation",
		})
	}

	// Analyze the constructor
	if classDef.Constructor != nil {
		e.currentDepth = 0
		constructorFlows := e.traceConstructor(classDef, classFile, parsed.PropertyName, parsed.AccessKey)
		flow.Steps = append(flow.Steps, constructorFlows...)

		// Extract ultimate sources from constructor
		flow.Sources = e.extractSources(constructorFlows)
	}

	// PHASE 1.2: Trace EXTERNAL method calls made after instantiation
	// This handles cases like: $mybb->parse_cookies() called in init.php:210
	if instFile != "" {
		externalFlows := e.traceExternalCalls(parsed.VarName, instFile, instLine, parsed.PropertyName, parsed.AccessKey, classDef, classFile)
		if len(externalFlows) > 0 {
			// Renumber steps
			for i := range externalFlows {
				externalFlows[i].StepNumber = len(flow.Steps) + i + 1
			}
			flow.Steps = append(flow.Steps, externalFlows...)
			flow.Sources = append(flow.Sources, e.extractSources(externalFlows)...)
		}
	}

	return flow, nil
}

// traceExternalCalls finds and traces method calls made on a variable AFTER its instantiation
// This is critical for cases like: $mybb = new MyBB(); ... $mybb->parse_cookies();
func (e *ExecutionEngine) traceExternalCalls(varName string, instFile string, instLine int, targetProperty string, accessKey string, classDef *types.ClassDef, classFile string) []FlowStep {
	var steps []FlowStep

	// Get the instantiation file content
	content, ok := e.fileContents[instFile]
	if !ok {
		return steps
	}

	// Parse the file to find method calls on this variable
	root, ok := e.parsedFiles[instFile]
	if !ok {
		return steps
	}

	// Find all method calls on this variable after the instantiation line
	methodCalls := e.findExternalMethodCalls(root, content, varName, instLine)

	for _, mc := range methodCalls {
		// Check if this method exists in the class and might populate our target property
		methodDef, ok := classDef.Methods[mc.methodName]
		if !ok {
			continue
		}

		// Check if this method touches the target property
		if methodDef.BodySource != "" && strings.Contains(methodDef.BodySource, "$this->"+targetProperty) {
			steps = append(steps, FlowStep{
				StepNumber:  len(steps) + 20,
				Description: fmt.Sprintf("External call: %s->%s() at line %d", varName, mc.methodName, mc.line),
				Code:        fmt.Sprintf("%s->%s(%s);", varName, mc.methodName, mc.args),
				FilePath:    instFile,
				Line:        mc.line,
				Type:        "external_call",
			})

			// Trace into this method
			e.currentDepth = 0
			methodSteps := e.traceMethod(classDef, methodDef, classFile, targetProperty, accessKey, mc.args)
			steps = append(steps, methodSteps...)
		}
	}

	return steps
}

// externalMethodCall represents a method call found in a file
type externalMethodCall struct {
	methodName string
	args       string
	line       int
}

// findExternalMethodCalls finds all method calls on a variable after a given line
func (e *ExecutionEngine) findExternalMethodCalls(root *sitter.Node, source []byte, varName string, afterLine int) []externalMethodCall {
	var calls []externalMethodCall

	// Find all member_call_expression nodes
	memberCalls := findNodesOfType(root, "member_call_expression")

	varNameWithoutDollar := strings.TrimPrefix(varName, "$")

	for _, call := range memberCalls {
		callLine := int(call.StartPoint().Row) + 1

		// Only look at calls after instantiation
		if callLine <= afterLine {
			continue
		}

		// Check if this is a call on our variable
		// member_call_expression has: object, ->, name, arguments
		if call.ChildCount() < 3 {
			continue
		}

		// Get the object being called on
		obj := call.Child(0)
		if obj == nil {
			continue
		}

		objText := getNodeText(obj, source)

		// Check if it's our variable (handle both $var and $var forms)
		if objText != varName && objText != "$"+varNameWithoutDollar {
			continue
		}

		// Get method name - it's usually the "name" child
		var methodName string
		var argsText string

		for i := 0; i < int(call.ChildCount()); i++ {
			child := call.Child(i)
			if child == nil {
				continue
			}

			switch child.Type() {
			case "name":
				methodName = getNodeText(child, source)
			case "arguments":
				// Extract arguments without parentheses
				argsText = getNodeText(child, source)
				argsText = strings.TrimPrefix(argsText, "(")
				argsText = strings.TrimSuffix(argsText, ")")
			}
		}

		if methodName != "" {
			calls = append(calls, externalMethodCall{
				methodName: methodName,
				args:       argsText,
				line:       callLine,
			})
		}
	}

	return calls
}

// analyzeMethodReturns analyzes what a method returns
func (e *ExecutionEngine) analyzeMethodReturns(classDef *types.ClassDef, method *types.MethodDef, classFile string) *MethodReturnInfo {
	cacheKey := fmt.Sprintf("%s.%s", classDef.Name, method.Name)
	if cached, ok := e.methodReturns[cacheKey]; ok {
		return cached
	}

	info := &MethodReturnInfo{
		ReturnStatements: make([]string, 0),
	}

	if method.BodySource == "" {
		e.methodReturns[cacheKey] = info
		return info
	}

	body := method.BodySource

	// Find all return statements
	returnPattern := getOrCompileRegex(`return\s+([^;]+);`)
	returnMatches := returnPattern.FindAllStringSubmatch(body, -1)

	for _, match := range returnMatches {
		if len(match) >= 2 {
			returnExpr := strings.TrimSpace(match[1])
			info.ReturnStatements = append(info.ReturnStatements, returnExpr)

			// GAP #4 FIX: Check for fluent interface pattern: return $this;
			if returnExpr == "$this" {
				info.ReturnsSelf = true
			}

			// PHASE 2.1: Check if it returns TYPE-CASTED $this->property[$param]
			// Pattern: (int)$this->property[$paramName] or (float)$this->... etc.
			castPropParamPattern := getOrCompileRegex(`\((\w+)\)\s*\$this->(\w+)\[\$(\w+)\]`)
			if propMatch := castPropParamPattern.FindStringSubmatch(returnExpr); len(propMatch) >= 4 {
				info.ReturnsProperty = true
				info.PropertyName = propMatch[2] // property name
				paramName := propMatch[3]        // param used as key

				// Find which parameter index this is
				for i, p := range method.Parameters {
					if p.Name == paramName {
						info.UsesParamAsKey = true
						info.ParamIndex = i
						break
					}
				}
			}

			// Check if it returns $this->property[$param] (without type cast)
			// Pattern: $this->property[$paramName]
			if !info.ReturnsProperty {
				propParamPattern := getOrCompileRegex(`\$this->(\w+)\[\$(\w+)\]`)
				if propMatch := propParamPattern.FindStringSubmatch(returnExpr); len(propMatch) >= 3 {
					info.ReturnsProperty = true
					info.PropertyName = propMatch[1]
					paramName := propMatch[2]

					// Find which parameter index this is
					for i, p := range method.Parameters {
						if p.Name == paramName {
							info.UsesParamAsKey = true
							info.ParamIndex = i
							break
						}
					}
				}
			}

			// PHASE 2.2: Check for null coalescing pattern
			// Pattern: $this->property[$param] ?? $default
			if !info.ReturnsProperty {
				nullCoalescePattern := getOrCompileRegex(`\$this->(\w+)\[\$(\w+)\]\s*\?\?`)
				if propMatch := nullCoalescePattern.FindStringSubmatch(returnExpr); len(propMatch) >= 3 {
					info.ReturnsProperty = true
					info.PropertyName = propMatch[1]
					paramName := propMatch[2]

					for i, p := range method.Parameters {
						if p.Name == paramName {
							info.UsesParamAsKey = true
							info.ParamIndex = i
							break
						}
					}
				}
			}

			// PHASE 2.2: Check for ternary isset pattern
			// Pattern: isset($this->property[$param]) ? $this->property[$param] : default
			if !info.ReturnsProperty {
				ternaryPattern := getOrCompileRegex(`isset\s*\(\s*\$this->(\w+)\[\$(\w+)\]\s*\)\s*\?\s*\$this->(\w+)\[\$(\w+)\]`)
				if propMatch := ternaryPattern.FindStringSubmatch(returnExpr); len(propMatch) >= 5 {
					// Verify both property refs match
					if propMatch[1] == propMatch[3] && propMatch[2] == propMatch[4] {
						info.ReturnsProperty = true
						info.PropertyName = propMatch[1]
						paramName := propMatch[2]

						for i, p := range method.Parameters {
							if p.Name == paramName {
								info.UsesParamAsKey = true
								info.ParamIndex = i
								break
							}
						}
					}
				}
			}

			// Check if it returns $this->property directly
			if !info.ReturnsProperty {
				directPropPattern := getOrCompileRegex(`^\$this->(\w+)$`)
				if propMatch := directPropPattern.FindStringSubmatch(returnExpr); len(propMatch) >= 2 {
					info.ReturnsProperty = true
					info.PropertyName = propMatch[1]
				}
			}

			// Check for superglobals
			for _, sg := range superglobals {
				if strings.Contains(returnExpr, sg.pattern) {
					info.ReturnsUserInput = true
					info.UserInputExpression = returnExpr
					break
				}
			}
		}
	}

	e.methodReturns[cacheKey] = info
	return info
}

// formatParams formats method parameters for display
func (e *ExecutionEngine) formatParams(params []types.ParameterDef) string {
	var parts []string
	for _, p := range params {
		part := "$" + p.Name
		if p.Type != "" {
			part = p.Type + " " + part
		}
		if p.DefaultValue != "" {
			part += " = " + p.DefaultValue
		}
		parts = append(parts, part)
	}
	return strings.Join(parts, ", ")
}

// findInstantiation finds where a variable is instantiated by searching ALL parsed files
// This is fully universal - no framework-specific hints or assumptions
func (e *ExecutionEngine) findInstantiation(varName string, contextFile string) (className, filePath string, line int) {
	// First check the context file (most likely location)
	if root, ok := e.parsedFiles[contextFile]; ok {
		if content, ok := e.fileContents[contextFile]; ok {
			className, line = e.findInstantiationInAST(root, content, varName)
			if className != "" {
				return className, contextFile, line
			}
		}
	}

	// Search ALL files for the instantiation - no assumptions about file names
	for file, root := range e.parsedFiles {
		if file == contextFile {
			continue // Already checked
		}
		if content, ok := e.fileContents[file]; ok {
			className, line = e.findInstantiationInAST(root, content, varName)
			if className != "" {
				return className, file, line
			}
		}
	}

	return "", "", 0
}

// findInstantiationInAST searches an AST for object creation
// Supports:
// - $var = new Class()
// - $GLOBALS['var'] = new Class()
// - $var = $container->get('service') with type hint
func (e *ExecutionEngine) findInstantiationInAST(root *sitter.Node, source []byte, varName string) (className string, line int) {
	// Look for assignment expressions where LHS is varName and RHS is object_creation_expression
	assignments := findNodesOfType(root, "assignment_expression")

	// Strip $ from varName for GLOBALS matching
	varNameWithoutDollar := strings.TrimPrefix(varName, "$")

	for _, assign := range assignments {
		if assign.ChildCount() < 3 {
			continue
		}

		left := assign.Child(0)
		right := assign.Child(2)

		if left == nil || right == nil {
			continue
		}

		leftText := getNodeText(left, source)

		// Check direct assignment: $var = new Class()
		directMatch := leftText == varName

		// Check GLOBALS assignment: $GLOBALS['var'] = new Class()
		globalsMatch := false
		if !directMatch {
			// Pattern: $GLOBALS['varname'] or $GLOBALS["varname"]
			globalsPattern := getOrCompileRegex(`\$GLOBALS\[['"](\w+)['"]\]`)
			if matches := globalsPattern.FindStringSubmatch(leftText); len(matches) >= 2 {
				if matches[1] == varNameWithoutDollar {
					globalsMatch = true
				}
			}
		}

		if !directMatch && !globalsMatch {
			continue
		}

		if right.Type() == "object_creation_expression" {
			// Found it! Extract class name
			nameNode := findChildByType(right, "name")
			if nameNode == nil {
				nameNode = findChildByType(right, "qualified_name")
			}
			if nameNode != nil {
				return getNodeText(nameNode, source), int(assign.StartPoint().Row) + 1
			}
		}

		// Check for DI container pattern: $var = $container->get('service')
		rightText := getNodeText(right, source)
		diPattern := getOrCompileRegex(`\$\w+->get\(['"]([^'"]+)['"]\)`)
		if diPattern.MatchString(rightText) {
			// Found DI container pattern - look for type hint above
			assignLine := int(assign.StartPoint().Row)
			typeHintClass := e.findTypeHintAboveLine(source, assignLine, varNameWithoutDollar)
			if typeHintClass != "" {
				return typeHintClass, assignLine + 1
			}
			// If no type hint, return the service name as a hint
			if matches := diPattern.FindStringSubmatch(rightText); len(matches) >= 2 {
				serviceName := matches[1]
				return fmt.Sprintf("[DI:%s]", serviceName), assignLine + 1
			}
		}
	}

	return "", 0
}

// findTypeHintAboveLine searches for PHPDoc @var type hints above a line
// Pattern: /* @var $varname \namespace\classname */ or /** @var \class $var */
func (e *ExecutionEngine) findTypeHintAboveLine(source []byte, targetLine int, varName string) string {
	lines := strings.Split(string(source), "\n")
	if targetLine <= 0 || targetLine > len(lines) {
		return ""
	}

	// Look up to 5 lines above for a type hint
	startLine := targetLine - 5
	if startLine < 0 {
		startLine = 0
	}

	// Pattern 1: /* @var $varname \namespace\classname */
	// Pattern 2: /** @var \namespace\classname $varname */
	typeHintPatterns := []*regexp.Regexp{
		// /* @var $request \phpbb\request\request_interface */
		getOrCompileRegex(`@var\s+\$` + regexp.QuoteMeta(varName) + `\s+\\?([\w\\]+)`),
		// /* @var \phpbb\request\request_interface $request */
		getOrCompileRegex(`@var\s+\\?([\w\\]+)\s+\$` + regexp.QuoteMeta(varName)),
	}

	for i := targetLine - 1; i >= startLine; i-- {
		line := lines[i]
		for _, pattern := range typeHintPatterns {
			if matches := pattern.FindStringSubmatch(line); len(matches) >= 2 {
				// Extract class name from fully qualified name
				fqn := matches[1]
				parts := strings.Split(fqn, "\\")
				return parts[len(parts)-1] // Return just the class name
			}
		}
	}

	return ""
}

// findClassDefinition finds a class definition across all symbol tables
// Handles interfaces by stripping _interface suffix and looking for implementing class
func (e *ExecutionEngine) findClassDefinition(className string) (*types.ClassDef, string) {
	// First try exact match
	for filePath, st := range e.symbolTables {
		if classDef, ok := st.Classes[className]; ok {
			return classDef, filePath
		}
	}

	// Try case-insensitive match
	lowerClassName := strings.ToLower(className)
	for filePath, st := range e.symbolTables {
		for name, classDef := range st.Classes {
			if strings.ToLower(name) == lowerClassName {
				return classDef, filePath
			}
		}
	}

	// If name ends with _interface, try stripping it
	if strings.HasSuffix(lowerClassName, "_interface") {
		baseName := strings.TrimSuffix(className, "_interface")
		baseName = strings.TrimSuffix(baseName, "_Interface")
		for filePath, st := range e.symbolTables {
			for name, classDef := range st.Classes {
				if strings.EqualFold(name, baseName) {
					return classDef, filePath
				}
			}
		}
	}

	// Try to find a class that implements this interface
	for filePath, st := range e.symbolTables {
		for _, classDef := range st.Classes {
			for _, iface := range classDef.Implements {
				if strings.EqualFold(iface, className) {
					return classDef, filePath
				}
			}
		}
	}

	return nil, ""
}

// traceConstructor traces through a constructor to find property population
func (e *ExecutionEngine) traceConstructor(classDef *types.ClassDef, classFile string, targetProperty string, accessKey string) []FlowStep {
	var steps []FlowStep

	if classDef.Constructor == nil {
		return steps
	}

	constructor := classDef.Constructor

	steps = append(steps, FlowStep{
		StepNumber:  len(steps) + 2,
		Description: "Constructor runs",
		Code:        fmt.Sprintf("function __construct() { ... }"),
		FilePath:    classFile,
		Line:        constructor.Line,
		Type:        "constructor_call",
	})

	// Parse the constructor body
	if constructor.BodySource == "" {
		return steps
	}

	// Look for method calls that might populate the property
	// Parse: $this->methodName($arg)
	methodCallPattern := getOrCompileRegex(`\$this->(\w+)\(([^)]*)\)`)
	methodCalls := methodCallPattern.FindAllStringSubmatch(constructor.BodySource, -1)

	for _, call := range methodCalls {
		methodName := call[1]
		methodArgs := call[2]

		// Check if this method populates our target property
		if methodDef, ok := classDef.Methods[methodName]; ok {
			// Trace into the method FIRST to see if it affects target property
			methodSteps := e.traceMethod(classDef, methodDef, classFile, targetProperty, accessKey, methodArgs)

			// Only add method call step if method actually affects the target property
			if len(methodSteps) > 0 {
				steps = append(steps, FlowStep{
					StepNumber:  len(steps) + 2,
					Description: fmt.Sprintf("Calls $this->%s(%s)", methodName, methodArgs),
					Code:        fmt.Sprintf("$this->%s(%s);", methodName, methodArgs),
					FilePath:    classFile,
					Line:        e.findLineInBody(constructor.BodySource, constructor.BodyStart, methodName),
					Type:        "method_call",
				})
				steps = append(steps, methodSteps...)
			}
		}
	}

	// Also trace the constructor body directly for direct assignments
	// This handles cases like: if($_SERVER['REQUEST_METHOD'] == "POST") { $this->request_method = "post"; }
	constructorAsMethod := &types.MethodDef{
		Name:       "__construct",
		Line:       constructor.Line,
		BodySource: constructor.BodySource,
		BodyStart:  constructor.BodyStart,
	}
	// Reset depth to allow analyzing the constructor body directly
	// This is not a recursive call, just analyzing the current body
	savedDepth := e.currentDepth
	e.currentDepth = 0
	directSteps := e.traceMethod(classDef, constructorAsMethod, classFile, targetProperty, accessKey, "")
	e.currentDepth = savedDepth
	steps = append(steps, directSteps...)

	return steps
}

// traceMethod traces through a method to find property assignments
func (e *ExecutionEngine) traceMethod(classDef *types.ClassDef, method *types.MethodDef, classFile string, targetProperty string, accessKey string, callArgs string) []FlowStep {
	var steps []FlowStep

	e.currentDepth++
	if e.currentDepth > e.maxDepth {
		return steps
	}

	if method.BodySource == "" {
		return steps
	}

	body := method.BodySource

	// PHASE 1.1: Look for foreach loops iterating over SUPERGLOBALS directly
	// Pattern: foreach($_SUPERGLOBAL as $key => $val)
	// This handles methods like parse_cookies() that don't take parameters
	superglobalForeachPattern := getOrCompileRegex(`foreach\s*\(\s*(\$_\w+)\s+as\s+\$(\w+)\s*=>\s*\$(\w+)\s*\)`)
	superglobalMatches := superglobalForeachPattern.FindAllStringSubmatch(body, -1)

	for _, match := range superglobalMatches {
		superglobalName := match[1] // e.g., "$_COOKIE"
		keyVar := match[2]          // e.g., "key"
		valVar := match[3]          // e.g., "val"

		// Check if this superglobal is a known user input source
		var sourceType string
		for _, sg := range superglobals {
			if superglobalName == sg.pattern {
				sourceType = sg.sourceType
				break
			}
		}

		if sourceType != "" {
			// Look for property assignment inside the loop FIRST
			// Pattern: $this->property[$key] = $val
			propAssignPattern := getOrCompileRegex(`\$this->(\w+)\[\$` + keyVar + `\]\s*=\s*\$` + valVar)
			propMatches := propAssignPattern.FindAllStringSubmatch(body, -1)

			for _, propMatch := range propMatches {
				assignedProperty := propMatch[1]
				// Only add steps if this assigns to our target property
				if assignedProperty == targetProperty {
					// Add the loop step only if we're assigning to target property
					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Inside %s() - loops through %s superglobal", method.Name, superglobalName),
						Code:        fmt.Sprintf("foreach(%s as $%s => $%s)", superglobalName, keyVar, valVar),
						FilePath:    classFile,
						Line:        e.findLineInBody(body, method.BodyStart, "foreach"),
						Type:        "loop",
					})

					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Assigns $this->%s[$%s] = $%s from %s", targetProperty, keyVar, valVar, superglobalName),
						Code:        fmt.Sprintf("$this->%s[$%s] = $%s;", targetProperty, keyVar, valVar),
						FilePath:    classFile,
						Line:        e.findLineInBody(body, method.BodyStart, "$this->"+targetProperty),
						Type:        "assignment",
					})

					// The ultimate source is the superglobal
					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Result: $%s['%s'] now contains %s['%s']", targetProperty, accessKey, superglobalName, accessKey),
						Code:        fmt.Sprintf("// $this->%s['%s'] = %s['%s']", targetProperty, accessKey, superglobalName, accessKey),
						FilePath:    classFile,
						Line:        0,
						Type:        "result",
					})
				}
			}
		}
	}

	// Look for foreach loops iterating over the method parameter
	// Pattern: foreach($array as $key => $val)
	foreachPattern := getOrCompileRegex(`foreach\s*\(\s*\$(\w+)\s+as\s+\$(\w+)\s*=>\s*\$(\w+)\s*\)`)
	foreachMatches := foreachPattern.FindAllStringSubmatch(body, -1)

	for _, match := range foreachMatches {
		arrayVar := match[1]
		keyVar := match[2]
		valVar := match[3]

		// Skip if this is a superglobal (already handled above)
		if strings.HasPrefix(arrayVar, "_") {
			continue
		}

		// Check if the array variable matches the parameter
		if len(method.Parameters) > 0 && method.Parameters[0].Name == arrayVar {
			// Look for property assignment inside the loop FIRST
			// Pattern: $this->property[$key] = $val
			propAssignPattern := getOrCompileRegex(`\$this->(\w+)\[\$` + keyVar + `\]\s*=\s*\$` + valVar)
			propMatches := propAssignPattern.FindAllStringSubmatch(body, -1)

			for _, propMatch := range propMatches {
				assignedProperty := propMatch[1]
				// Only add steps if this assigns to our target property
				if assignedProperty == targetProperty {
					// Add the loop step only if we're assigning to target property
					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Inside %s() - loops through %s parameter", method.Name, callArgs),
						Code:        fmt.Sprintf("foreach(%s as $%s => $%s)", callArgs, keyVar, valVar),
						FilePath:    classFile,
						Line:        e.findLineInBody(body, method.BodyStart, "foreach"),
						Type:        "loop",
					})

					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Assigns $this->%s[$%s] = $%s from %s", targetProperty, keyVar, valVar, callArgs),
						Code:        fmt.Sprintf("$this->%s[$%s] = $%s;", targetProperty, keyVar, valVar),
						FilePath:    classFile,
						Line:        e.findLineInBody(body, method.BodyStart, "$this->"+targetProperty),
						Type:        "assignment",
					})

					// The ultimate source is the call argument
					steps = append(steps, FlowStep{
						StepNumber:  len(steps) + 10,
						Description: fmt.Sprintf("Result: $%s['%s'] now contains %s['%s']", targetProperty, accessKey, callArgs, accessKey),
						Code:        fmt.Sprintf("// $this->%s['%s'] = %s['%s']", targetProperty, accessKey, callArgs, accessKey),
						FilePath:    classFile,
						Line:        0,
						Type:        "result",
					})
				}
			}
		}
	}

	// Also look for direct assignments
	// Pattern: $this->property = $something
	directAssignPattern := getOrCompileRegex(`\$this->` + regexp.QuoteMeta(targetProperty) + `\s*=\s*([^;]+)`)
	directMatches := directAssignPattern.FindAllStringSubmatch(body, -1)

	for _, match := range directMatches {
		source := strings.TrimSpace(match[1])
		steps = append(steps, FlowStep{
			StepNumber:  len(steps) + 10,
			Description: fmt.Sprintf("Assigns $this->%s = %s", targetProperty, source),
			Code:        fmt.Sprintf("$this->%s = %s;", targetProperty, source),
			FilePath:    classFile,
			Line:        e.findLineInBody(body, method.BodyStart, "$this->"+targetProperty),
			Type:        "assignment",
		})
	}

	// NEW: Look for conditional assignments based on superglobals
	// Pattern: if($_SUPERGLOBAL['key']... { $this->property = value }
	// This handles cases like: if($_SERVER['REQUEST_METHOD'] == "POST") { $this->request_method = "post"; }
	for _, sg := range superglobals {
		if strings.Contains(body, sg.pattern) && strings.Contains(body, "$this->"+targetProperty) {
			// Check if superglobal is used in a condition and property is assigned nearby
			// Pattern: if($_SUPERGLOBAL[anything])
			condPattern := getOrCompileRegex(`if\s*\(\s*` + regexp.QuoteMeta(sg.pattern) + `\[['"]?(\w+)['"]?\]`)
			if condMatches := condPattern.FindStringSubmatch(body); len(condMatches) >= 2 {
				superglobalKey := condMatches[1]
				steps = append(steps, FlowStep{
					StepNumber:  len(steps) + 10,
					Description: fmt.Sprintf("Conditional on %s['%s']", sg.pattern, superglobalKey),
					Code:        fmt.Sprintf("if(%s['%s'] == ...) { $this->%s = ...; }", sg.pattern, superglobalKey, targetProperty),
					FilePath:    classFile,
					Line:        e.findLineInBody(body, method.BodyStart, sg.pattern),
					Type:        "conditional",
				})
				steps = append(steps, FlowStep{
					StepNumber:  len(steps) + 10,
					Description: fmt.Sprintf("Property $%s is controlled by %s['%s']", targetProperty, sg.pattern, superglobalKey),
					Code:        fmt.Sprintf("// $this->%s value depends on %s['%s']", targetProperty, sg.pattern, superglobalKey),
					FilePath:    classFile,
					Line:        0,
					Type:        "taint",
				})
			}
		}
	}

	return steps
}

// findLineInBody finds the approximate line number for a pattern in the body
func (e *ExecutionEngine) findLineInBody(body string, startLine int, pattern string) int {
	lines := strings.Split(body, "\n")
	for i, line := range lines {
		if strings.Contains(line, pattern) {
			return startLine + i
		}
	}
	return startLine
}

// Superglobals list for source detection
var superglobals = []struct {
	pattern    string
	sourceType string
}{
	{"$_GET", "http_get"},
	{"$_POST", "http_post"},
	{"$_COOKIE", "http_cookie"},
	{"$_REQUEST", "http_request"},
	{"$_SERVER", "http_header"},
	{"$_FILES", "http_file"},
	{"$_ENV", "env_var"},
}

// extractSources extracts ultimate sources from the flow steps
func (e *ExecutionEngine) extractSources(steps []FlowStep) []UltimateSource {
	var sources []UltimateSource

	for _, step := range steps {
		for _, sg := range superglobals {
			if strings.Contains(step.Code, sg.pattern) || strings.Contains(step.Description, sg.pattern) {
				sources = append(sources, UltimateSource{
					Type:       sg.sourceType,
					Expression: sg.pattern,
					FilePath:   step.FilePath,
					Line:       step.Line,
				})
			}
		}
	}

	return sources
}

// GenerateFlowReport generates a human-readable flow report
func (flow *PropertyFlow) GenerateFlowReport() string {
	var sb strings.Builder

	sb.WriteString("=== Input Flow Trace ===\n\n")
	sb.WriteString(fmt.Sprintf("Expression: %s\n", flow.Expression))
	sb.WriteString(fmt.Sprintf("Class: %s\n", flow.ClassName))
	if flow.MethodName != "" {
		sb.WriteString(fmt.Sprintf("Method: %s()\n", flow.MethodName))
	}
	if flow.PropertyName != "" {
		sb.WriteString(fmt.Sprintf("Property: %s\n", flow.PropertyName))
	}
	if flow.AccessKey != "" {
		sb.WriteString(fmt.Sprintf("Access Key: '%s'\n", flow.AccessKey))
	}
	sb.WriteString("\n--- Flow Steps ---\n\n")

	for _, step := range flow.Steps {
		sb.WriteString(fmt.Sprintf("Step %d: %s\n", step.StepNumber, step.Description))
		sb.WriteString(fmt.Sprintf("   Code: %s\n", step.Code))
		if step.FilePath != "" && step.Line > 0 {
			sb.WriteString(fmt.Sprintf("   Location: %s:%d\n", step.FilePath, step.Line))
		}
		sb.WriteString("\n")
	}

	if len(flow.Sources) > 0 {
		sb.WriteString("--- Ultimate Sources ---\n\n")
		seen := make(map[string]bool)
		for _, src := range flow.Sources {
			key := fmt.Sprintf("%s:%d", src.Expression, src.Line)
			if seen[key] {
				continue
			}
			seen[key] = true
			sb.WriteString(fmt.Sprintf("  %s (%s)\n", src.Expression, src.Type))
			if src.FilePath != "" && src.Line > 0 {
				sb.WriteString(fmt.Sprintf("    at %s:%d\n", src.FilePath, src.Line))
			}
		}
	}

	return sb.String()
}

// GenerateMermaidDiagram generates a Mermaid flowchart for the flow
func (flow *PropertyFlow) GenerateMermaidDiagram() string {
	var sb strings.Builder

	sb.WriteString("flowchart TD\n")
	sb.WriteString("    classDef source fill:#ff6b6b,stroke:#c0392b,color:white\n")
	sb.WriteString("    classDef property fill:#4ecdc4,stroke:#16a085\n")
	sb.WriteString("    classDef method fill:#45b7d1,stroke:#2980b9\n")
	sb.WriteString("    classDef result fill:#95e1a3,stroke:#27ae60\n\n")

	// Add nodes for each step
	prevNode := ""
	for i, step := range flow.Steps {
		nodeID := fmt.Sprintf("step%d", i)
		label := strings.ReplaceAll(step.Description, "\"", "'")
		label = strings.ReplaceAll(label, "$", "\\$")

		nodeClass := ""
		switch step.Type {
		case "property_init":
			nodeClass = ":::property"
		case "method_call", "constructor_call", "method_def":
			nodeClass = ":::method"
		case "assignment":
			nodeClass = ":::source"
		case "result", "resolution":
			nodeClass = ":::result"
		}

		sb.WriteString(fmt.Sprintf("    %s[\"%s\"]%s\n", nodeID, label, nodeClass))

		if prevNode != "" {
			sb.WriteString(fmt.Sprintf("    %s --> %s\n", prevNode, nodeID))
		}
		prevNode = nodeID
	}

	// Add source nodes
	seen := make(map[string]bool)
	for i, src := range flow.Sources {
		if seen[src.Expression] {
			continue
		}
		seen[src.Expression] = true
		srcID := fmt.Sprintf("source%d", i)
		sb.WriteString(fmt.Sprintf("    %s((%s)):::source\n", srcID, src.Expression))
		sb.WriteString(fmt.Sprintf("    %s -.->|\"originates from\"| %s\n", srcID, prevNode))
	}

	return sb.String()
}

// Helper functions

func findNodesOfType(root *sitter.Node, nodeType string) []*sitter.Node {
	var nodes []*sitter.Node
	traverseTree(root, func(node *sitter.Node) bool {
		if node.Type() == nodeType {
			nodes = append(nodes, node)
		}
		return true
	})
	return nodes
}

func traverseTree(node *sitter.Node, callback func(*sitter.Node) bool) {
	if node == nil {
		return
	}
	if !callback(node) {
		return
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		traverseTree(node.Child(i), callback)
	}
}

func findChildByType(node *sitter.Node, nodeType string) *sitter.Node {
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

func getNodeText(node *sitter.Node, source []byte) string {
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

// CreateParser creates a new PHP parser
func CreateParser() *sitter.Parser {
	parser := sitter.NewParser()
	parser.SetLanguage(php.GetLanguage())
	return parser
}
