// Package types defines universal data structures for semantic input tracing
// across all supported programming languages.
package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// ============================================================================
// Core Flow Types
// ============================================================================

// FlowNodeType represents the type of a node in the data flow graph
type FlowNodeType string

const (
	NodeSource   FlowNodeType = "source"   // Original input source (e.g., $_GET, req.body)
	NodeCarrier  FlowNodeType = "carrier"  // Object/variable that holds user data
	NodeVariable FlowNodeType = "variable" // Regular variable in the flow
	NodeFunction FlowNodeType = "function" // Function/method call
	NodeProperty FlowNodeType = "property" // Object property access
	NodeParam    FlowNodeType = "param"    // Function parameter
	NodeReturn   FlowNodeType = "return"   // Return value
)

// FlowEdgeType represents how data flows between nodes
type FlowEdgeType string

const (
	EdgeAssignment   FlowEdgeType = "assignment"    // $x = $y
	EdgeParameter    FlowEdgeType = "parameter"     // func($x)
	EdgeReturn       FlowEdgeType = "return"        // return $x
	EdgeProperty     FlowEdgeType = "property"      // $obj->prop = $x
	EdgeArraySet     FlowEdgeType = "array_set"     // $arr['key'] = $x
	EdgeArrayGet     FlowEdgeType = "array_get"     // $x = $arr['key']
	EdgeMethodCall   FlowEdgeType = "method_call"   // $obj->method($x)
	EdgeConstructor  FlowEdgeType = "constructor"   // new Class($x)
	EdgeFramework    FlowEdgeType = "framework"     // Framework-specific flow
	EdgeConcatenate  FlowEdgeType = "concatenate"   // $x . $y
	EdgeDestructure  FlowEdgeType = "destructure"   // const {a, b} = obj
	EdgeIteration    FlowEdgeType = "iteration"     // foreach/for loop
	EdgeConditional  FlowEdgeType = "conditional"   // if/else branch
	EdgeCall         FlowEdgeType = "call"          // Function call
	EdgeDataFlow     FlowEdgeType = "data_flow"     // Generic data flow
)

// SourceType represents the type of input source
// Re-exported from pkg/sources/common for backward compatibility
type SourceType = common.SourceType

// Re-export source type constants for backward compatibility
const (
	SourceHTTPGet     = common.SourceHTTPGet
	SourceHTTPPost    = common.SourceHTTPPost
	SourceHTTPBody    = common.SourceHTTPBody
	SourceHTTPJSON    = common.SourceHTTPJSON
	SourceHTTPHeader  = common.SourceHTTPHeader
	SourceHTTPCookie  = common.SourceHTTPCookie
	SourceHTTPPath    = common.SourceHTTPPath
	SourceHTTPFile    = common.SourceHTTPFile
	SourceHTTPRequest = common.SourceHTTPRequest
	SourceSession     = common.SourceSession
	SourceCLIArg      = common.SourceCLIArg
	SourceEnvVar      = common.SourceEnvVar
	SourceStdin       = common.SourceStdin
	SourceFile        = common.SourceFile
	SourceDatabase    = common.SourceDatabase
	SourceNetwork     = common.SourceNetwork
	SourceUserInput   = common.SourceUserInput
	SourceUnknown     = common.SourceUnknown
)

// FlowNode represents a node in the data flow graph
type FlowNode struct {
	ID         string       `json:"id"`
	Type       FlowNodeType `json:"type"`
	Language   string       `json:"language"`

	// Location information
	FilePath   string `json:"file_path"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`
	EndLine    int    `json:"end_line,omitempty"`
	EndColumn  int    `json:"end_column,omitempty"`

	// Semantic information
	Name       string `json:"name"`                  // Variable/function/property name
	ClassName  string `json:"class_name,omitempty"`  // If part of a class
	MethodName string `json:"method_name,omitempty"` // If inside a method
	Scope      string `json:"scope,omitempty"`       // Scope identifier

	// Type information
	TypeInfo   *TypeInfo `json:"type_info,omitempty"`

	// Source information (if this is a source node)
	SourceType SourceType `json:"source_type,omitempty"`
	SourceKey  string     `json:"source_key,omitempty"` // Parameter name

	// Carrier information
	CarrierType string `json:"carrier_type,omitempty"` // "array", "object_property", etc.

	// Code snippet
	Snippet    string `json:"snippet"`

	// Metadata
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
}

// FlowEdge represents a directed edge in the data flow graph
type FlowEdge struct {
	ID          string       `json:"id"`
	From        string       `json:"from"`        // Source node ID
	To          string       `json:"to"`          // Target node ID
	Type        FlowEdgeType `json:"type"`

	// Location where flow occurs
	FilePath    string `json:"file_path"`
	Line        int    `json:"line"`

	// Human-readable description
	Description string `json:"description"`

	// Code causing the flow
	Code        string `json:"code,omitempty"`

	// Additional context
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// TypeInfo holds type information for a node
type TypeInfo struct {
	Name       string   `json:"name"`
	Kind       string   `json:"kind"` // "class", "interface", "primitive", "array", "map"
	Package    string   `json:"package,omitempty"`
	Generics   []string `json:"generics,omitempty"`
	IsNullable bool     `json:"is_nullable,omitempty"`
}

// ============================================================================
// Flow Map (Analysis Result)
// ============================================================================

// FlowMap represents the complete data flow analysis result
// Memory-optimized with internal deduplication maps
type FlowMap struct {
	// Target expression being traced
	Target FlowTarget `json:"target"`

	// Ultimate sources (where data originally comes from)
	Sources []FlowNode `json:"sources"`

	// Complete paths from sources to target
	Paths []FlowPath `json:"paths"`

	// All intermediate carriers
	Carriers []FlowNode `json:"carriers"`

	// All nodes in the flow graph
	AllNodes []FlowNode `json:"all_nodes"`

	// All edges in the flow graph
	AllEdges []FlowEdge `json:"all_edges"`

	// Usage locations (where the data is used)
	Usages []FlowNode `json:"usages"`

	// Carrier chain information
	CarrierChain *CarrierChain `json:"carrier_chain,omitempty"`

	// Call graph relevant to this flow
	CallGraph map[string][]string `json:"call_graph,omitempty"`

	// Analysis metadata
	Metadata FlowMapMetadata `json:"metadata"`

	// Internal deduplication maps (not serialized)
	nodeIndex map[string]bool `json:"-"` // nodeID -> exists
	edgeIndex map[string]bool `json:"-"` // edgeKey -> exists

	// Configurable limits (not serialized)
	maxNodes int `json:"-"`
	maxEdges int `json:"-"`
}

// Default limits for flow graph size
const (
	// DefaultMaxFlowNodes limits total nodes to prevent unbounded memory growth in large codebases
	DefaultMaxFlowNodes = 10000

	// DefaultMaxFlowEdges limits total edges to prevent unbounded memory growth in large codebases
	DefaultMaxFlowEdges = 20000
)

// NewFlowMap creates an optimized FlowMap with default limits and deduplication support
// MEMORY FIX: Reduced pre-allocation sizes for multi-threaded usage
func NewFlowMap() *FlowMap {
	return NewFlowMapWithLimits(DefaultMaxFlowNodes, DefaultMaxFlowEdges)
}

// NewFlowMapWithLimits creates a FlowMap with custom node/edge limits.
// Use maxNodes=0 or maxEdges=0 to use the default limits.
func NewFlowMapWithLimits(maxNodes, maxEdges int) *FlowMap {
	if maxNodes <= 0 {
		maxNodes = DefaultMaxFlowNodes
	}
	if maxEdges <= 0 {
		maxEdges = DefaultMaxFlowEdges
	}
	return &FlowMap{
		Sources:   make([]FlowNode, 0, 16),
		Paths:     make([]FlowPath, 0, 8),
		Carriers:  make([]FlowNode, 0, 8),
		AllNodes:  make([]FlowNode, 0, 64),
		AllEdges:  make([]FlowEdge, 0, 128),
		Usages:    make([]FlowNode, 0, 16),
		CallGraph: make(map[string][]string),
		nodeIndex: make(map[string]bool, 64),
		edgeIndex: make(map[string]bool, 128),
		maxNodes:  maxNodes,
		maxEdges:  maxEdges,
	}
}

// AddNode adds a node with O(1) deduplication
// MEMORY FIX: Limits total nodes to prevent unbounded growth
func (fm *FlowMap) AddNode(node FlowNode) bool {
	if fm.nodeIndex == nil {
		fm.nodeIndex = make(map[string]bool, 256)
	}
	if fm.nodeIndex[node.ID] {
		return false // Already exists
	}
	// Use instance limit, falling back to default if not set
	limit := fm.maxNodes
	if limit == 0 {
		limit = DefaultMaxFlowNodes
	}
	if len(fm.AllNodes) >= limit {
		return false // At capacity
	}
	fm.nodeIndex[node.ID] = true
	fm.AllNodes = append(fm.AllNodes, node)
	return true
}

// AddEdge adds an edge with O(1) deduplication
// MEMORY FIX: Limits total edges to prevent unbounded growth
func (fm *FlowMap) AddEdge(edge FlowEdge) bool {
	if fm.edgeIndex == nil {
		fm.edgeIndex = make(map[string]bool, 512)
	}
	edgeKey := edge.From + "->" + edge.To + ":" + string(edge.Type)
	if fm.edgeIndex[edgeKey] {
		return false // Already exists
	}
	// Use instance limit, falling back to default if not set
	limit := fm.maxEdges
	if limit == 0 {
		limit = DefaultMaxFlowEdges
	}
	if len(fm.AllEdges) >= limit {
		return false // At capacity
	}
	fm.edgeIndex[edgeKey] = true
	fm.AllEdges = append(fm.AllEdges, edge)
	return true
}

// AddSource adds a source node with deduplication
func (fm *FlowMap) AddSource(source FlowNode) bool {
	if fm.AddNode(source) {
		fm.Sources = append(fm.Sources, source)
		return true
	}
	return false
}

// AddCarrier adds a carrier node with deduplication
func (fm *FlowMap) AddCarrier(carrier FlowNode) bool {
	if fm.AddNode(carrier) {
		fm.Carriers = append(fm.Carriers, carrier)
		return true
	}
	return false
}

// AddUsage adds a usage node with deduplication
func (fm *FlowMap) AddUsage(usage FlowNode) bool {
	if fm.AddNode(usage) {
		fm.Usages = append(fm.Usages, usage)
		return true
	}
	return false
}

// HasNode checks if a node ID exists in O(1)
func (fm *FlowMap) HasNode(nodeID string) bool {
	if fm.nodeIndex == nil {
		return false
	}
	return fm.nodeIndex[nodeID]
}

// HasEdge checks if an edge exists in O(1)
func (fm *FlowMap) HasEdge(from, to string, edgeType FlowEdgeType) bool {
	if fm.edgeIndex == nil {
		return false
	}
	edgeKey := from + "->" + to + ":" + string(edgeType)
	return fm.edgeIndex[edgeKey]
}

// FlowTarget specifies what expression to trace
type FlowTarget struct {
	FilePath   string `json:"file_path"`
	Line       int    `json:"line"`
	Column     int    `json:"column,omitempty"`
	Expression string `json:"expression"`
}

// FlowPath represents a complete path from source to target
type FlowPath struct {
	ID          string     `json:"id"`
	Description string     `json:"description"`
	Steps       []FlowStep `json:"steps"`
	Source      *FlowNode  `json:"source"`
	Target      *FlowNode  `json:"target"`
}

// FlowStep represents one step in a flow path
type FlowStep struct {
	Node        FlowNode  `json:"node"`
	Edge        *FlowEdge `json:"edge,omitempty"` // Edge to next step
	Description string    `json:"description"`
	StepNumber  int       `json:"step_number"`
}

// CarrierChain describes how a carrier object propagates input
type CarrierChain struct {
	ClassName        string   `json:"class_name"`
	PropertyName     string   `json:"property_name"`
	Initialization   string   `json:"initialization"`
	PopulationMethod string   `json:"population_method,omitempty"`
	PopulationCalls  []string `json:"population_calls,omitempty"`
	Framework        string   `json:"framework,omitempty"`
}

// FlowMapMetadata contains analysis metadata
type FlowMapMetadata struct {
	AnalyzedAt       time.Time `json:"analyzed_at"`
	Duration         string    `json:"duration"`
	FilesAnalyzed    int       `json:"files_analyzed"`
	Language         string    `json:"language"`
	Framework        string    `json:"framework,omitempty"`
	TracerVersion    string    `json:"tracer_version"`
	Confidence       float64   `json:"confidence"` // 0.0 to 1.0
}

// ============================================================================
// Symbol Table Types
// ============================================================================

// SymbolTable holds all symbols discovered in a file
type SymbolTable struct {
	FilePath   string                  `json:"file_path"`
	Language   string                  `json:"language"`
	Imports    []ImportInfo            `json:"imports"`
	Classes    map[string]*ClassDef    `json:"classes"`
	Functions  map[string]*FunctionDef `json:"functions"`
	Variables  map[string]*VariableDef `json:"variables"`
	Constants  map[string]*ConstantDef `json:"constants"`
	Namespace  string                  `json:"namespace,omitempty"`

	// File-level metadata
	Framework  string                 `json:"framework,omitempty"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
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
// MEMORY FIX: Call this after analysis is complete to free large string memory
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
	Path       string `json:"path"`        // Import path/module name
	Alias      string `json:"alias,omitempty"`
	Names      []string `json:"names,omitempty"` // Specific imports (from X import a, b)
	IsRelative bool   `json:"is_relative"`
	Line       int    `json:"line"`
	Type       string `json:"type"` // "import", "require", "include", "use"
}

// ClassDef represents a class definition
type ClassDef struct {
	Name        string                  `json:"name"`
	FilePath    string                  `json:"file_path"`
	Line        int                     `json:"line"`
	EndLine     int                     `json:"end_line"`

	// Inheritance
	Extends     string                  `json:"extends,omitempty"`
	Implements  []string                `json:"implements,omitempty"`
	Traits      []string                `json:"traits,omitempty"` // PHP traits

	// Members
	Properties  map[string]*PropertyDef `json:"properties"`
	Methods     map[string]*MethodDef   `json:"methods"`
	Constructor *MethodDef              `json:"constructor,omitempty"`

	// For framework detection
	IsCarrier   bool                    `json:"is_carrier"`
	CarrierInfo *CarrierInfo            `json:"carrier_info,omitempty"`

	// Visibility
	Visibility  string                  `json:"visibility"` // public, private, protected
	IsAbstract  bool                    `json:"is_abstract"`
	IsFinal     bool                    `json:"is_final"`

	// Namespace/package
	Namespace   string                  `json:"namespace,omitempty"`
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

// ReleaseBodySources releases all method body sources to free memory
// MEMORY FIX: Call this after symbol table analysis is complete
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
	Name        string          `json:"name"`
	Parameters  []ParameterDef  `json:"parameters"`
	ReturnType  string          `json:"return_type,omitempty"`
	Line        int             `json:"line"`
	EndLine     int             `json:"end_line"`
	Visibility  string          `json:"visibility"`
	IsStatic    bool            `json:"is_static"`
	IsAbstract  bool            `json:"is_abstract"`
	IsAsync     bool            `json:"is_async"`

	// Body information
	BodyStart   int             `json:"body_start"`
	BodyEnd     int             `json:"body_end"`
	BodySource  string          `json:"body_source,omitempty"` // Actual source code

	// Flow analysis results
	ParamsToReturn []int          `json:"params_to_return,omitempty"` // Which params flow to return
	ParamsToProps  map[int]string `json:"params_to_props,omitempty"`  // Param -> property flows
	CallsInternal  []string       `json:"calls_internal,omitempty"`   // Internal method calls
	CallsExternal  []string       `json:"calls_external,omitempty"`   // External function calls
	ReturnsInput   bool           `json:"returns_input"`              // Does it return user input?

	// Annotations/decorators
	Annotations    []AnnotationDef `json:"annotations,omitempty"`
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
	ReceivesInput bool         `json:"receives_input"`
	InputSource   string       `json:"input_source,omitempty"`
	TaintChain    *TaintChain  `json:"taint_chain,omitempty"` // Taint chain if param is tainted (GAP 5)
}

// AnnotationDef represents a decorator/annotation
type AnnotationDef struct {
	Name       string                 `json:"name"`
	Arguments  map[string]interface{} `json:"arguments,omitempty"`
	Line       int                    `json:"line"`
}

// FunctionDef represents a standalone function definition
type FunctionDef struct {
	Name        string          `json:"name"`
	FilePath    string          `json:"file_path"`
	Parameters  []ParameterDef  `json:"parameters"`
	ReturnType  string          `json:"return_type,omitempty"`
	Line        int             `json:"line"`
	EndLine     int             `json:"end_line"`
	IsExported  bool            `json:"is_exported"`
	IsAsync     bool            `json:"is_async"`

	// Body information
	BodyStart   int             `json:"body_start"`
	BodyEnd     int             `json:"body_end"`
	BodySource  string          `json:"body_source,omitempty"`

	// Flow analysis results
	ParamsToReturn []int          `json:"params_to_return,omitempty"`
	ReturnsInput   bool           `json:"returns_input"`
	CallsExternal  []string       `json:"calls_external,omitempty"`

	// Enhanced taint tracking (GAP 5)
	ReturnTaintChain *TaintChain            `json:"return_taint_chain,omitempty"` // Taint chain for return value
	ParamTaintChains map[int]*TaintChain    `json:"param_taint_chains,omitempty"` // Taint chains by param index
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
	IsTainted    bool         `json:"is_tainted"`
	TaintSource  string       `json:"taint_source,omitempty"`
	TaintDepth   int          `json:"taint_depth,omitempty"`
	TaintChain   *TaintChain  `json:"taint_chain,omitempty"` // Full taint chain (GAP 5)
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

// ============================================================================
// Assignment and Call Tracking
// ============================================================================

// Assignment represents a variable assignment
type Assignment struct {
	Target      string   `json:"target"`       // Variable being assigned to
	TargetType  string   `json:"target_type"`  // "variable", "property", "array_element"
	Source      string   `json:"source"`       // Expression being assigned
	SourceType  string   `json:"source_type"`  // Type of source expression
	Line        int      `json:"line"`
	Column      int      `json:"column"`
	FilePath    string   `json:"file_path"`
	Scope       string   `json:"scope"`
	IsTainted   bool     `json:"is_tainted"`
	TaintSource string   `json:"taint_source,omitempty"`

	// For compound assignments
	Operator    string   `json:"operator,omitempty"` // =, +=, .=, etc.

	// For array/object access
	Keys        []string `json:"keys,omitempty"` // Access path: ["input", "thumbnail"]
}

// CallSite represents a function/method call
type CallSite struct {
	FunctionName string       `json:"function_name"`
	ClassName    string       `json:"class_name,omitempty"`
	MethodName   string       `json:"method_name,omitempty"`
	Arguments    []CallArg    `json:"arguments"`
	Line         int          `json:"line"`
	Column       int          `json:"column"`
	FilePath     string       `json:"file_path"`
	Scope        string       `json:"scope"`

	// Result assignment
	ResultVar    string       `json:"result_var,omitempty"`

	// Call type
	IsStatic     bool         `json:"is_static"`
	IsConstructor bool        `json:"is_constructor"`

	// Taint info
	HasTaintedArgs bool       `json:"has_tainted_args"`
	TaintedArgIndices []int   `json:"tainted_arg_indices,omitempty"`
}

// CallArg represents a function call argument
type CallArg struct {
	Index       int    `json:"index"`
	Value       string `json:"value"`
	Type        string `json:"type,omitempty"`
	IsTainted   bool   `json:"is_tainted"`
	TaintSource string `json:"taint_source,omitempty"`

	// Enhanced taint tracking (GAP 5)
	TaintChain  *TaintChain `json:"taint_chain,omitempty"` // Full taint propagation chain
}

// TaintChain tracks the complete propagation path of tainted data
// This enables precise tracking of how data flows from source to usage
type TaintChain struct {
	// Original source information
	OriginalSource   string     `json:"original_source"`    // e.g., "$_GET['id']"
	OriginalType     SourceType `json:"original_type"`      // e.g., "http_get"
	OriginalFile     string     `json:"original_file"`
	OriginalLine     int        `json:"original_line"`

	// Chain of transformations/assignments
	Steps []TaintStep `json:"steps"`

	// Current state
	CurrentExpression string `json:"current_expression"` // What the taint looks like now
	Depth             int    `json:"depth"`              // How many hops from source
}

// TaintStep represents one step in the taint propagation chain
type TaintStep struct {
	StepType    string `json:"step_type"`    // "assignment", "parameter", "return", "property", "method_call"
	Expression  string `json:"expression"`   // The code at this step
	FilePath    string `json:"file_path"`
	Line        int    `json:"line"`
	Description string `json:"description"`  // Human-readable description
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
// Backward Taint Analysis Types (GAP 2)
// ============================================================================

// BackwardTraceResult represents the result of backward taint analysis
// This traces from a target expression back to its input sources
type BackwardTraceResult struct {
	// Target expression being traced
	TargetExpression string       `json:"target_expression"`
	TargetFile       string       `json:"target_file"`
	TargetLine       int          `json:"target_line"`

	// All paths from sources to this target
	Paths []BackwardPath `json:"paths"`

	// Summary of all sources found
	Sources []SourceInfo `json:"sources"`

	// Analysis metadata
	AnalyzedFiles int           `json:"analyzed_files"`
	Duration      time.Duration `json:"duration"`
}

// BackwardPath represents one path from a source to the target
type BackwardPath struct {
	// Source information
	Source SourceInfo `json:"source"`

	// Steps from source to target (in forward order for readability)
	Steps []BackwardStep `json:"steps"`

	// Confidence score (0.0 to 1.0)
	Confidence float64 `json:"confidence"`

	// Whether path crosses file boundaries
	CrossFile bool `json:"cross_file"`
}

// BackwardStep represents one step in a backward trace path
type BackwardStep struct {
	StepNumber  int          `json:"step_number"`
	Expression  string       `json:"expression"`      // The code at this step
	FilePath    string       `json:"file_path"`
	Line        int          `json:"line"`
	StepType    string       `json:"step_type"`       // "source", "assignment", "parameter", "return", "property"
	Description string       `json:"description"`
}

// BatchTraceResult represents the result of batch backward taint analysis
// This traces multiple target expressions in a SINGLE pass through the codebase
// CRITICAL for performance: reduces file reads from N*files to just files
type BatchTraceResult struct {
	// Whether ANY variable traces back to user input
	HasUserInput bool `json:"has_user_input"`

	// Results for each variable traced
	PerVariable map[string]*BackwardTraceResult `json:"per_variable"`

	// Analysis metadata
	TotalDuration  time.Duration `json:"total_duration"`
	AnalyzedFiles  int           `json:"analyzed_files"`
	VariablesFound int           `json:"variables_found"`
}

// SourceInfo provides details about a discovered input source
type SourceInfo struct {
	Type        SourceType `json:"type"`          // http_get, http_post, etc.
	Expression  string     `json:"expression"`    // e.g., "$_GET['id']"
	FilePath    string     `json:"file_path"`
	Line        int        `json:"line"`
	Confidence  float64    `json:"confidence"`
}

// ============================================================================
// Framework Knowledge Types
// ============================================================================

// FrameworkPattern defines a known framework input pattern
type FrameworkPattern struct {
	ID          string        `json:"id"`
	Framework   string        `json:"framework"`
	Language    string        `json:"language"`
	Name        string        `json:"name"`
	Description string        `json:"description"`

	// Pattern matching
	ClassPattern    string    `json:"class_pattern,omitempty"`    // Regex for class names
	MethodPattern   string    `json:"method_pattern,omitempty"`   // Regex for method names
	PropertyPattern string    `json:"property_pattern,omitempty"` // Regex for property names
	AccessPattern   string    `json:"access_pattern,omitempty"`   // How data is accessed

	// Source mapping
	SourceType      SourceType `json:"source_type"`
	SourceKey       string     `json:"source_key,omitempty"` // How to extract the key

	// Flow information
	CarrierClass    string    `json:"carrier_class,omitempty"`
	CarrierProperty string    `json:"carrier_property,omitempty"`
	PopulatedBy     string    `json:"populated_by,omitempty"`     // Method that populates
	PopulatedFrom   []string  `json:"populated_from,omitempty"`   // Original sources

	// Confidence
	Confidence      float64   `json:"confidence"` // 0.0 to 1.0
}

// FrameworkPatternData is a simple struct for importing patterns from pkg/sources
// This avoids import cycles by using a plain data structure
type FrameworkPatternData struct {
	ID              string
	Framework       string
	Language        string
	Name            string
	Description     string
	ClassPattern    string
	MethodPattern   string
	PropertyPattern string
	AccessPattern   string
	SourceType      string
	SourceKey       string
	CarrierClass    string
	CarrierProperty string
	PopulatedBy     string
	PopulatedFrom   []string
	Confidence      float64
}

// FromData creates a FrameworkPattern from FrameworkPatternData
func (d *FrameworkPatternData) ToFrameworkPattern() *FrameworkPattern {
	return &FrameworkPattern{
		ID:              d.ID,
		Framework:       d.Framework,
		Language:        d.Language,
		Name:            d.Name,
		Description:     d.Description,
		ClassPattern:    d.ClassPattern,
		MethodPattern:   d.MethodPattern,
		PropertyPattern: d.PropertyPattern,
		AccessPattern:   d.AccessPattern,
		SourceType:      SourceType(d.SourceType),
		SourceKey:       d.SourceKey,
		CarrierClass:    d.CarrierClass,
		CarrierProperty: d.CarrierProperty,
		PopulatedBy:     d.PopulatedBy,
		PopulatedFrom:   d.PopulatedFrom,
		Confidence:      d.Confidence,
	}
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
	Depth         int `json:"depth"`
	MaxDepth      int `json:"max_depth"`

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

// TaintInfo holds taint information for a variable
type TaintInfo struct {
	Source     *FlowNode `json:"source"`
	SourceType SourceType `json:"source_type"`
	SourceKey  string    `json:"source_key"`
	Depth      int       `json:"depth"`
	Path       []string  `json:"path"` // How taint reached this var
}

// ObjectInstance represents a tracked object instance
type ObjectInstance struct {
	VariableName string                 `json:"variable_name"`
	ClassName    string                 `json:"class_name"`
	CreatedAt    Location               `json:"created_at"`
	Properties   map[string]*TaintInfo  `json:"properties"`
	Framework    string                 `json:"framework,omitempty"`
}

// Location represents a code location
type Location struct {
	FilePath  string `json:"file_path"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndLine   int    `json:"end_line,omitempty"`
	EndColumn int    `json:"end_column,omitempty"`
}

// ============================================================================
// Output Helpers
// ============================================================================

// ToJSON converts a FlowMap to JSON string
func (fm *FlowMap) ToJSON() (string, error) {
	data, err := json.MarshalIndent(fm, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ToMermaid generates a Mermaid diagram for the flow
func (fm *FlowMap) ToMermaid() string {
	var sb strings.Builder
	sb.WriteString("graph TD\n")

	// Add nodes
	nodeIDs := make(map[string]string)
	for i, node := range fm.AllNodes {
		nodeID := fmt.Sprintf("N%d", i)
		nodeIDs[node.ID] = nodeID

		label := node.Name
		if node.Snippet != "" && len(node.Snippet) < 50 {
			label = node.Snippet
		}
		label = strings.ReplaceAll(label, "\"", "'")

		style := ""
		switch node.Type {
		case NodeSource:
			style = ":::source"
		case NodeCarrier:
			style = ":::carrier"
		}

		sb.WriteString(fmt.Sprintf("    %s[\"%s\"]%s\n", nodeID, label, style))
	}

	// Add edges
	for _, edge := range fm.AllEdges {
		fromID := nodeIDs[edge.From]
		toID := nodeIDs[edge.To]
		if fromID != "" && toID != "" {
			label := string(edge.Type)
			sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n", fromID, label, toID))
		}
	}

	// Add styles
	sb.WriteString("\n    classDef source fill:#ff6b6b,color:white\n")
	sb.WriteString("    classDef carrier fill:#4ecdc4,color:white\n")

	return sb.String()
}

// ToDOT generates a DOT graph for the flow
func (fm *FlowMap) ToDOT() string {
	var sb strings.Builder
	sb.WriteString("digraph FlowGraph {\n")
	sb.WriteString("    rankdir=TB;\n")
	sb.WriteString("    node [shape=box];\n\n")

	// Add nodes
	for _, node := range fm.AllNodes {
		label := node.Name
		if node.Snippet != "" && len(node.Snippet) < 50 {
			label = node.Snippet
		}
		label = strings.ReplaceAll(label, "\"", "\\\"")

		color := "white"
		switch node.Type {
		case NodeSource:
			color = "#ff6b6b"
		case NodeCarrier:
			color = "#4ecdc4"
		}

		sb.WriteString(fmt.Sprintf("    \"%s\" [label=\"%s\" fillcolor=\"%s\" style=filled];\n",
			node.ID, label, color))
	}

	sb.WriteString("\n")

	// Add edges
	for _, edge := range fm.AllEdges {
		label := string(edge.Type)
		sb.WriteString(fmt.Sprintf("    \"%s\" -> \"%s\" [label=\"%s\"];\n",
			edge.From, edge.To, label))
	}

	sb.WriteString("}\n")
	return sb.String()
}

// Summary returns a human-readable summary of the flow
func (fm *FlowMap) Summary() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("TARGET: %s @ %s:%d\n\n",
		fm.Target.Expression, fm.Target.FilePath, fm.Target.Line))

	sb.WriteString("ULTIMATE SOURCES:\n")
	for _, src := range fm.Sources {
		key := src.SourceKey
		if key == "" {
			key = "*"
		}
		sb.WriteString(fmt.Sprintf("  - %s[%s]\n", src.SourceType, key))
	}

	sb.WriteString("\nFLOW PATHS:\n")
	for i, path := range fm.Paths {
		sb.WriteString(fmt.Sprintf("  Path %d: %s\n", i+1, path.Description))
		for _, step := range path.Steps {
			sb.WriteString(fmt.Sprintf("    %d. %s @ %s:%d\n",
				step.StepNumber, step.Description, step.Node.FilePath, step.Node.Line))
		}
	}

	if fm.CarrierChain != nil {
		sb.WriteString("\nCARRIER CHAIN:\n")
		sb.WriteString(fmt.Sprintf("  Class: %s\n", fm.CarrierChain.ClassName))
		sb.WriteString(fmt.Sprintf("  Property: %s\n", fm.CarrierChain.PropertyName))
		sb.WriteString(fmt.Sprintf("  Initialization: %s\n", fm.CarrierChain.Initialization))
		if fm.CarrierChain.PopulationMethod != "" {
			sb.WriteString(fmt.Sprintf("  Population Method: %s\n", fm.CarrierChain.PopulationMethod))
		}
	}

	sb.WriteString("\nUSAGE LOCATIONS:\n")
	for _, usage := range fm.Usages {
		sb.WriteString(fmt.Sprintf("  - %s:%d - %s\n",
			usage.FilePath, usage.Line, usage.Snippet))
	}

	return sb.String()
}
