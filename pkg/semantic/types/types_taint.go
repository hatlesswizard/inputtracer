package types

// ============================================================================
// Assignment and Call Tracking
// ============================================================================

// Assignment represents a variable assignment
type Assignment struct {
	Target      string `json:"target"`      // Variable being assigned to
	TargetType  string `json:"target_type"` // "variable", "property", "array_element"
	Source      string `json:"source"`      // Expression being assigned
	SourceType  string `json:"source_type"` // Type of source expression
	Line        int    `json:"line"`
	Column      int    `json:"column"`
	FilePath    string `json:"file_path"`
	Scope       string `json:"scope"`
	IsTainted   bool   `json:"is_tainted"`
	TaintSource string `json:"taint_source,omitempty"`

	// For compound assignments
	Operator string `json:"operator,omitempty"` // =, +=, .=, etc.

	// For array/object access
	Keys []string `json:"keys,omitempty"` // Access path: ["input", "thumbnail"]
}

// CallSite represents a function/method call
type CallSite struct {
	FunctionName string    `json:"function_name"`
	ClassName    string    `json:"class_name,omitempty"`
	MethodName   string    `json:"method_name,omitempty"`
	Arguments    []CallArg `json:"arguments"`
	Line         int       `json:"line"`
	Column       int       `json:"column"`
	FilePath     string    `json:"file_path"`
	Scope        string    `json:"scope"`

	// Result assignment
	ResultVar string `json:"result_var,omitempty"`

	// Call type
	IsStatic      bool `json:"is_static"`
	IsConstructor bool `json:"is_constructor"`

	// Taint info
	HasTaintedArgs    bool  `json:"has_tainted_args"`
	TaintedArgIndices []int `json:"tainted_arg_indices,omitempty"`
}

// CallArg represents a function call argument
type CallArg struct {
	Index       int         `json:"index"`
	Value       string      `json:"value"`
	Type        string      `json:"type,omitempty"`
	IsTainted   bool        `json:"is_tainted"`
	TaintSource string      `json:"taint_source,omitempty"`
	TaintChain  *TaintChain `json:"taint_chain,omitempty"`
}

// ============================================================================
// Taint Chain Types
// ============================================================================

// TaintChain tracks the complete propagation path of tainted data,
// enabling precise tracking of how data flows from source to usage.
type TaintChain struct {
	// Original source information
	OriginalSource string     `json:"original_source"` // e.g., "$_GET['id']"
	OriginalType   SourceType `json:"original_type"`   // e.g., "http_get"
	OriginalFile   string     `json:"original_file"`
	OriginalLine   int        `json:"original_line"`

	// Chain of transformations/assignments
	Steps []TaintStep `json:"steps"`

	// Current state
	CurrentExpression string `json:"current_expression"` // What the taint looks like now
	Depth             int    `json:"depth"`              // How many hops from source
}

// TaintStep represents one step in the taint propagation chain
type TaintStep struct {
	StepType    string `json:"step_type"`  // "assignment", "parameter", "return", "property", "method_call"
	Expression  string `json:"expression"` // The code at this step
	FilePath    string `json:"file_path"`
	Line        int    `json:"line"`
	Description string `json:"description"` // Human-readable description
}

// NewTaintChain creates a new taint chain from an original source
func NewTaintChain(source, sourceType, file string, line int) *TaintChain {
	return &TaintChain{
		OriginalSource:    source,
		OriginalType:      SourceType(sourceType),
		OriginalFile:      file,
		OriginalLine:      line,
		Steps:             make([]TaintStep, 0),
		CurrentExpression: source,
		Depth:             0,
	}
}

// AddStep adds a propagation step to the taint chain
func (tc *TaintChain) AddStep(stepType, expression, file string, line int, description string) {
	tc.Steps = append(tc.Steps, TaintStep{
		StepType:    stepType,
		Expression:  expression,
		FilePath:    file,
		Line:        line,
		Description: description,
	})
	tc.CurrentExpression = expression
	tc.Depth++
}

// Clone creates a copy of the taint chain for branching flows
func (tc *TaintChain) Clone() *TaintChain {
	if tc == nil {
		return nil
	}
	clone := &TaintChain{
		OriginalSource:    tc.OriginalSource,
		OriginalType:      tc.OriginalType,
		OriginalFile:      tc.OriginalFile,
		OriginalLine:      tc.OriginalLine,
		Steps:             make([]TaintStep, len(tc.Steps)),
		CurrentExpression: tc.CurrentExpression,
		Depth:             tc.Depth,
	}
	copy(clone.Steps, tc.Steps)
	return clone
}

// ============================================================================
// Analysis State
// ============================================================================

// AnalysisState holds the current state during analysis
type AnalysisState struct {
	// Symbol tables by file
	SymbolTables map[string]*SymbolTable `json:"symbol_tables"`

	// All discovered sources
	Sources []FlowNode `json:"sources"`

	// All discovered carriers
	Carriers []FlowNode `json:"carriers"`

	// Tainted variables by scope
	TaintedVars map[string]map[string]*TaintInfo `json:"tainted_vars"` // scope -> name -> info

	// Object instances being tracked
	ObjectInstances map[string]*ObjectInstance `json:"object_instances"`

	// Call graph
	CallGraph map[string][]string `json:"call_graph"`

	// File dependencies
	FileDependencies map[string][]string `json:"file_dependencies"`

	// Current context
	CurrentFile   string `json:"current_file"`
	CurrentClass  string `json:"current_class"`
	CurrentMethod string `json:"current_method"`
	CurrentScope  string `json:"current_scope"`

	// Analysis depth tracking
	Depth    int `json:"depth"`
	MaxDepth int `json:"max_depth"`

	// Visited tracking (prevent infinite loops)
	Visited map[string]bool `json:"-"`
}

// NewAnalysisState creates a new analysis state
func NewAnalysisState(maxDepth int) *AnalysisState {
	return &AnalysisState{
		SymbolTables:     make(map[string]*SymbolTable),
		Sources:          make([]FlowNode, 0),
		Carriers:         make([]FlowNode, 0),
		TaintedVars:      make(map[string]map[string]*TaintInfo),
		ObjectInstances:  make(map[string]*ObjectInstance),
		CallGraph:        make(map[string][]string),
		FileDependencies: make(map[string][]string),
		MaxDepth:         maxDepth,
		Visited:          make(map[string]bool),
	}
}

// TaintInfo holds simplified taint information for a variable during analysis.
// Use TaintChain when full propagation history is needed.
type TaintInfo struct {
	Source     *FlowNode  `json:"source"`
	SourceType SourceType `json:"source_type"`
	SourceKey  string     `json:"source_key"`
	Depth      int        `json:"depth"`
	Path       []string   `json:"path"` // How taint reached this var
}

// ObjectInstance represents a tracked object instance
type ObjectInstance struct {
	VariableName string                `json:"variable_name"`
	ClassName    string                `json:"class_name"`
	CreatedAt    Location              `json:"created_at"`
	Properties   map[string]*TaintInfo `json:"properties"`
	Framework    string                `json:"framework,omitempty"`
}

// Location represents a code location
type Location struct {
	FilePath  string `json:"file_path"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndLine   int    `json:"end_line,omitempty"`
	EndColumn int    `json:"end_column,omitempty"`
}
