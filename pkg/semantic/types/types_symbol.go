package types

// ============================================================================
// Symbol Table Types
// ============================================================================

// SymbolTable holds all symbols discovered in a file
type SymbolTable struct {
	FilePath  string                  `json:"file_path"`
	Language  string                  `json:"language"`
	Imports   []ImportInfo            `json:"imports"`
	Classes   map[string]*ClassDef    `json:"classes"`
	Functions map[string]*FunctionDef `json:"functions"`
	Variables map[string]*VariableDef `json:"variables"`
	Constants map[string]*ConstantDef `json:"constants"`
	Namespace string                  `json:"namespace,omitempty"`

	// File-level metadata
	Framework string                 `json:"framework,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NewSymbolTable creates a new empty symbol table
func NewSymbolTable(filePath, language string) *SymbolTable {
	return &SymbolTable{
		FilePath:  filePath,
		Language:  language,
		Imports:   make([]ImportInfo, 0),
		Classes:   make(map[string]*ClassDef),
		Functions: make(map[string]*FunctionDef),
		Variables: make(map[string]*VariableDef),
		Constants: make(map[string]*ConstantDef),
		Metadata:  make(map[string]interface{}),
	}
}

// ReleaseBodySources releases all body sources from classes and functions
// to free large string memory after analysis is complete.
func (st *SymbolTable) ReleaseBodySources() {
	for _, class := range st.Classes {
		class.ReleaseBodySources()
	}
	for _, fn := range st.Functions {
		fn.BodySource = ""
	}
}

// ImportInfo represents an import/include/require statement
type ImportInfo struct {
	Path       string   `json:"path"` // Import path/module name
	Alias      string   `json:"alias,omitempty"`
	Names      []string `json:"names,omitempty"` // Specific imports (from X import a, b)
	IsRelative bool     `json:"is_relative"`
	Line       int      `json:"line"`
	Type       string   `json:"type"` // "import", "require", "include", "use"
}

// ClassDef represents a class definition
type ClassDef struct {
	Name     string `json:"name"`
	FilePath string `json:"file_path"`
	Line     int    `json:"line"`
	EndLine  int    `json:"end_line"`

	// Inheritance
	Extends    string   `json:"extends,omitempty"`
	Implements []string `json:"implements,omitempty"`
	Traits     []string `json:"traits,omitempty"` // PHP traits

	// Members
	Properties  map[string]*PropertyDef `json:"properties"`
	Methods     map[string]*MethodDef   `json:"methods"`
	Constructor *MethodDef              `json:"constructor,omitempty"`

	// For framework detection
	IsCarrier   bool         `json:"is_carrier"`
	CarrierInfo *CarrierInfo `json:"carrier_info,omitempty"`

	// Visibility
	Visibility string `json:"visibility"` // public, private, protected
	IsAbstract bool   `json:"is_abstract"`
	IsFinal    bool   `json:"is_final"`

	// Namespace/package
	Namespace string `json:"namespace,omitempty"`
}

// NewClassDef creates a new class definition
func NewClassDef(name, filePath string, line int) *ClassDef {
	return &ClassDef{
		Name:       name,
		FilePath:   filePath,
		Line:       line,
		Properties: make(map[string]*PropertyDef),
		Methods:    make(map[string]*MethodDef),
		Implements: make([]string, 0),
		Traits:     make([]string, 0),
	}
}

// ReleaseBodySources releases all method body sources to free memory.
func (cd *ClassDef) ReleaseBodySources() {
	for _, method := range cd.Methods {
		method.BodySource = ""
	}
	if cd.Constructor != nil {
		cd.Constructor.BodySource = ""
	}
}

// PropertyDef represents a class property/field
type PropertyDef struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	Visibility   string `json:"visibility"` // public, private, protected
	InitialValue string `json:"initial_value,omitempty"`
	Line         int    `json:"line"`
	IsStatic     bool   `json:"is_static"`
	IsReadonly   bool   `json:"is_readonly"`

	// Flow analysis results
	ReceivesInput bool     `json:"receives_input"`
	InputSources  []string `json:"input_sources,omitempty"`
	TaintDepth    int      `json:"taint_depth,omitempty"`
}

// MethodDef represents a method/function definition
type MethodDef struct {
	Name       string         `json:"name"`
	Parameters []ParameterDef `json:"parameters"`
	ReturnType string         `json:"return_type,omitempty"`
	Line       int            `json:"line"`
	EndLine    int            `json:"end_line"`
	Visibility string         `json:"visibility"`
	IsStatic   bool           `json:"is_static"`
	IsAbstract bool           `json:"is_abstract"`
	IsAsync    bool           `json:"is_async"`

	// Body information
	BodyStart  int    `json:"body_start"`
	BodyEnd    int    `json:"body_end"`
	BodySource string `json:"body_source,omitempty"` // Actual source code

	// Flow analysis results
	ParamsToReturn []int          `json:"params_to_return,omitempty"` // Which params flow to return
	ParamsToProps  map[int]string `json:"params_to_props,omitempty"`  // Param -> property flows
	CallsInternal  []string       `json:"calls_internal,omitempty"`   // Internal method calls
	CallsExternal  []string       `json:"calls_external,omitempty"`   // External function calls
	ReturnsInput   bool           `json:"returns_input"`              // Does it return user input?

	// Annotations/decorators
	Annotations []AnnotationDef `json:"annotations,omitempty"`
}

// ParameterDef represents a function/method parameter
type ParameterDef struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	DefaultValue string `json:"default_value,omitempty"`
	Index        int    `json:"index"`
	IsVariadic   bool   `json:"is_variadic"`
	IsReference  bool   `json:"is_reference"` // PHP &$param

	// Flow analysis
	ReceivesInput bool        `json:"receives_input"`
	InputSource   string      `json:"input_source,omitempty"`
	TaintChain    *TaintChain `json:"taint_chain,omitempty"`
}

// AnnotationDef represents a decorator/annotation
type AnnotationDef struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments,omitempty"`
	Line      int                    `json:"line"`
}

// FunctionDef represents a standalone function definition
type FunctionDef struct {
	Name       string         `json:"name"`
	FilePath   string         `json:"file_path"`
	Parameters []ParameterDef `json:"parameters"`
	ReturnType string         `json:"return_type,omitempty"`
	Line       int            `json:"line"`
	EndLine    int            `json:"end_line"`
	IsExported bool           `json:"is_exported"`
	IsAsync    bool           `json:"is_async"`

	// Body information
	BodyStart  int    `json:"body_start"`
	BodyEnd    int    `json:"body_end"`
	BodySource string `json:"body_source,omitempty"`

	// Flow analysis results
	ParamsToReturn   []int               `json:"params_to_return,omitempty"`
	ReturnsInput     bool                `json:"returns_input"`
	CallsExternal    []string            `json:"calls_external,omitempty"`
	ReturnTaintChain *TaintChain         `json:"return_taint_chain,omitempty"`
	ParamTaintChains map[int]*TaintChain `json:"param_taint_chains,omitempty"`
}

// VariableDef represents a variable definition
type VariableDef struct {
	Name         string `json:"name"`
	Type         string `json:"type,omitempty"`
	InitialValue string `json:"initial_value,omitempty"`
	Line         int    `json:"line"`
	Scope        string `json:"scope"`
	IsGlobal     bool   `json:"is_global"`
	IsConstant   bool   `json:"is_constant"`

	// Flow analysis
	IsTainted   bool        `json:"is_tainted"`
	TaintSource string      `json:"taint_source,omitempty"`
	TaintDepth  int         `json:"taint_depth,omitempty"`
	TaintChain  *TaintChain `json:"taint_chain,omitempty"`
}

// ConstantDef represents a constant definition
type ConstantDef struct {
	Name  string `json:"name"`
	Value string `json:"value"`
	Type  string `json:"type,omitempty"`
	Line  int    `json:"line"`
}

// CarrierInfo describes how a class carries user input
type CarrierInfo struct {
	PropertyName      string   `json:"property_name"`
	SourceTypes       []string `json:"source_types"`       // Which source types it carries
	PopulationMethod  string   `json:"population_method"`  // Method that populates it
	PopulationPattern string   `json:"population_pattern"` // Pattern used
	AccessPattern     string   `json:"access_pattern"`     // How to access: "array", "method", "property"
}
