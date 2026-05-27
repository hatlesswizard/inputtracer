package tracer

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/hatlesswizard/inputtracer/pkg/sources/constants"
)

// FlowNodeType represents the type of a node in the data flow graph.
// Re-exported from pkg/sources/constants.
type FlowNodeType = constants.FlowNodeType

// FlowEdgeType represents how data flows between two nodes.
// Re-exported from pkg/sources/constants.
type FlowEdgeType = constants.FlowEdgeType

// FlowNode type constants
const (
	FlowNodeSource   FlowNodeType = constants.NodeSource
	FlowNodeVariable FlowNodeType = constants.NodeVariable
	FlowNodeFunction FlowNodeType = constants.NodeFunction
	FlowNodeParam    FlowNodeType = constants.NodeParam
	FlowNodeCarrier  FlowNodeType = constants.NodeCarrier
	FlowNodeProperty FlowNodeType = constants.NodeProperty
	FlowNodeReturn   FlowNodeType = constants.NodeReturn
)

// FlowEdge type constants
const (
	FlowEdgeAssignment FlowEdgeType = constants.EdgeAssignment
	FlowEdgeCall       FlowEdgeType = constants.EdgeCall
	FlowEdgeReturn     FlowEdgeType = constants.EdgeReturn
	FlowEdgeTaint      FlowEdgeType = constants.EdgeDataFlow
	FlowEdgeParameter  FlowEdgeType = constants.EdgeParameter
)

// InputLabel categorizes the type of user input
// Re-exported from pkg/sources/constants for backward compatibility
type InputLabel = constants.InputLabel

// Re-export InputLabel constants for backward compatibility
const (
	LabelHTTPGet     = constants.LabelHTTPGet
	LabelHTTPPost    = constants.LabelHTTPPost
	LabelHTTPCookie  = constants.LabelHTTPCookie
	LabelHTTPHeader  = constants.LabelHTTPHeader
	LabelHTTPBody    = constants.LabelHTTPBody
	LabelCLI         = constants.LabelCLI
	LabelEnvironment = constants.LabelEnvironment
	LabelFile        = constants.LabelFile
	LabelDatabase    = constants.LabelDatabase
	LabelNetwork     = constants.LabelNetwork
	LabelUserInput   = constants.LabelUserInput
)

// Location represents a precise location in source code
type Location struct {
	FilePath  string `json:"file_path"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndLine   int    `json:"end_line"`
	EndColumn int    `json:"end_column"`
	Snippet   string `json:"snippet,omitempty"`
}

// InputSource represents where user input enters the code
type InputSource struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"` // e.g., "$_GET", "req.body", "argv"
	Key      string       `json:"key"`  // e.g., "username" in $_GET['username']
	Location Location     `json:"location"`
	Labels   []InputLabel `json:"labels"`
	Language string       `json:"language"`
}

// TaintedVariable represents a variable that holds user input at some point
type TaintedVariable struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Scope    string       `json:"scope"`  // Function/class scope
	Source   *InputSource `json:"source"` // Original input source
	Location Location     `json:"location"`
	Depth    int          `json:"depth"` // How many assignments from original source
	Language string       `json:"language"`
}

// TaintedParam represents a function parameter that receives user input
type TaintedParam struct {
	Index  int              `json:"index"`
	Name   string           `json:"name"`
	Source *InputSource     `json:"source"`
	Path   *PropagationPath `json:"path,omitempty"`
}

// TaintedFunction represents a function that receives user input
type TaintedFunction struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	FilePath        string            `json:"file_path"`
	Line            int               `json:"line"`
	Language        string            `json:"language"`
	TaintedParams   []TaintedParam    `json:"tainted_params"`
	ReceivesThrough []PropagationPath `json:"receives_through,omitempty"`
}

// PropagationStepType defines the type of propagation step
// Re-exported from pkg/sources/constants for backward compatibility
type PropagationStepType = constants.PropagationStepType

// Re-export PropagationStepType constants for backward compatibility
const (
	StepAssignment            = constants.StepAssignment
	StepParameterPass         = constants.StepParameterPass
	StepReturn                = constants.StepReturn
	StepInterproceduralReturn = constants.StepInterproceduralReturn
	StepConcatenation         = constants.StepConcatenation
	StepArrayAccess           = constants.StepArrayAccess
	StepObjectAccess          = constants.StepObjectAccess
	StepDestructure           = constants.StepDestructure
)

// PropagationStep is one step in the propagation chain
type PropagationStep struct {
	Type     PropagationStepType `json:"type"`
	Variable string              `json:"variable"`
	Function string              `json:"function,omitempty"` // If crossing function boundary
	Location Location            `json:"location"`
}

// PropagationPath shows how input flows from source to destination
type PropagationPath struct {
	Source      *InputSource      `json:"source"`
	Steps       []PropagationStep `json:"steps"`
	Destination Location          `json:"destination"`
}

// FlowNode is a node in the flow graph
type FlowNode struct {
	ID       string       `json:"id"`
	Type     FlowNodeType `json:"type"`
	Name     string       `json:"name"`
	Location Location     `json:"location"`
}

// FlowEdge connects two nodes showing data flow
type FlowEdge struct {
	From     string       `json:"from"` // Node ID
	To       string       `json:"to"`   // Node ID
	Type     FlowEdgeType `json:"type"`
	Location Location     `json:"location"`
}

// FlowGraph represents the complete input flow graph
type FlowGraph struct {
	Nodes []FlowNode `json:"nodes"`
	Edges []FlowEdge `json:"edges"`
}

// TraceStats contains analysis statistics
type TraceStats struct {
	FilesAnalyzed     int            `json:"files_analyzed"`
	SourcesFound      int            `json:"sources_found"`
	TaintedVarsFound  int            `json:"tainted_variables_found"`
	TaintedFuncsFound int            `json:"tainted_functions_found"`
	PropagationPaths  int            `json:"propagation_paths"`
	AnalysisDuration  time.Duration  `json:"analysis_duration_ns"`
	DurationMs        int64          `json:"analysis_duration_ms"`
	ByLanguage        map[string]int `json:"files_by_language"`
}

// TraceResult is the complete result of tracing a codebase
type TraceResult struct {
	// All discovered input sources
	Sources []*InputSource `json:"sources"`

	// All variables that hold user input at some point
	TaintedVariables []*TaintedVariable `json:"tainted_variables"`

	// All functions that receive user input (directly or transitively)
	TaintedFunctions []*TaintedFunction `json:"tainted_functions"`

	// Complete flow graph
	FlowGraph *FlowGraph `json:"flow_graph"`

	// Statistics
	Stats TraceStats `json:"stats"`

	// Errors encountered during analysis (parse errors, permission errors, etc.)
	Errors []error `json:"errors,omitempty"`
}

// MarshalJSON implements json.Marshaler so that the []error Errors field is
// serialized as a JSON array of strings (by calling .Error() on each entry)
// rather than as an array of empty objects (which is what encoding/json
// produces for interface values by default).
func (r *TraceResult) MarshalJSON() ([]byte, error) {
	// Use an alias to avoid infinite recursion.
	type alias TraceResult

	// Convert []error → []string for JSON output.
	errStrings := make([]string, len(r.Errors))
	for i, e := range r.Errors {
		if e != nil {
			errStrings[i] = e.Error()
		}
	}

	// Intermediate struct with string errors.
	v := struct {
		alias
		Errors []string `json:"errors,omitempty"`
	}{
		alias:  alias(*r),
		Errors: errStrings,
	}
	// Clear the embedded Errors so it doesn't conflict.
	v.alias.Errors = nil

	return json.Marshal(v)
}

// ToJSON converts the trace result to JSON
func (r *TraceResult) ToJSON() (string, error) {
	data, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ParameterInfo contains information about a function parameter
type ParameterInfo struct {
	Index int    `json:"index"`
	Name  string `json:"name"`
	Type  string `json:"type,omitempty"`
}

// FunctionSummary captures how a function propagates input
type FunctionSummary struct {
	Name            string          `json:"name"`
	FilePath        string          `json:"file_path"`
	Language        string          `json:"language"`
	StartLine       int             `json:"start_line"`
	EndLine         int             `json:"end_line"`
	Parameters      []ParameterInfo `json:"parameters"`
	ParamsToReturn  []int           `json:"params_to_return"` // Indices of params that flow to return
	ParamsToParams  map[int][]int   `json:"params_to_params"` // Param N flows to param M in nested calls
	IsSource        bool            `json:"is_source"`        // Function itself returns user input
	CalledFunctions []string        `json:"called_functions"`
}

// ScopeType represents the type of scope
// Re-exported from pkg/sources/constants for backward compatibility
type ScopeType = constants.ScopeType

// Re-export ScopeType constants for backward compatibility
const (
	ScopeGlobal   = constants.ScopeGlobal
	ScopeFile     = constants.ScopeFile
	ScopeModule   = constants.ScopeModule
	ScopeClass    = constants.ScopeClass
	ScopeFunction = constants.ScopeFunction
	ScopeBlock    = constants.ScopeBlock
)

// Scope represents a variable scope in the code
type Scope struct {
	ID        string                      `json:"id"`
	Type      ScopeType                   `json:"type"`
	Name      string                      `json:"name"`
	Parent    *Scope                      `json:"-"` // Avoid circular JSON
	ParentID  string                      `json:"parent_id,omitempty"`
	Children  []*Scope                    `json:"-"` // Child scopes
	Variables map[string]*TaintedVariable `json:"-"`
	StartLine int                         `json:"start_line"`
	EndLine   int                         `json:"end_line"`
	StartLoc  Location                    `json:"start_location,omitempty"`
}

// ScopeState manages the scope stack during AST traversal.
// It is a pure scope-management concern: push/pop scopes, look up variables
// within the lexical hierarchy. It knows nothing about taint tracking.
type ScopeState struct {
	CurrentScope *Scope
	ScopeStack   []*Scope
}

// newScopeState creates a ScopeState rooted at a fresh global scope.
func newScopeState() ScopeState {
	globalScope := &Scope{
		ID:        "global",
		Type:      ScopeGlobal,
		Name:      "global",
		Variables: make(map[string]*TaintedVariable),
	}
	return ScopeState{
		CurrentScope: globalScope,
		ScopeStack:   []*Scope{globalScope},
	}
}

// EnterScope pushes a new named scope onto the stack.
func (ss *ScopeState) EnterScope(scopeType ScopeType, name string, startLine, endLine int) *Scope {
	newScope := &Scope{
		ID:        name + "_" + string(scopeType),
		Type:      scopeType,
		Name:      name,
		Parent:    ss.CurrentScope,
		Variables: make(map[string]*TaintedVariable),
		StartLine: startLine,
		EndLine:   endLine,
	}
	if ss.CurrentScope != nil {
		newScope.ParentID = ss.CurrentScope.ID
	}
	ss.ScopeStack = append(ss.ScopeStack, newScope)
	ss.CurrentScope = newScope
	return newScope
}

// ExitScope pops the current scope and restores the parent.
func (ss *ScopeState) ExitScope() *Scope {
	if len(ss.ScopeStack) <= 1 {
		return ss.CurrentScope
	}
	ss.ScopeStack = ss.ScopeStack[:len(ss.ScopeStack)-1]
	ss.CurrentScope = ss.ScopeStack[len(ss.ScopeStack)-1]
	return ss.CurrentScope
}

// LookupVariable searches for a variable from the current scope upward.
func (ss *ScopeState) LookupVariable(name string) (*TaintedVariable, bool) {
	scope := ss.CurrentScope
	for scope != nil {
		if v, ok := scope.Variables[name]; ok {
			return v, true
		}
		scope = scope.Parent
	}
	return nil, false
}

// AnalysisState maintains the current state during analysis of a single file.
//
// It embeds ScopeState for scope management (push/pop scopes, hierarchical
// variable lookup) and adds taint-tracking fields (TaintedValues,
// FunctionSummaries, VisitedFunctions) as a separate concern. Keeping them in
// one struct is intentional for this package: both concerns are needed together
// on every file-analysis pass and separating them into two arguments everywhere
// would add noise without benefit. The split is expressed structurally via the
// embedded ScopeState so that the two concerns remain conceptually distinct.
//
// See also: pkg/semantic/types.AnalysisState, which is an unrelated type used
// by the deep semantic analysis layer. The two types serve different levels of
// abstraction and should not be merged (see M26 note in types.go).
type AnalysisState struct {
	ScopeState                                    // scope management (push/pop, lookup)
	TaintedValues     map[string]*TaintedVariable // variable name -> tainted info (taint tracking)
	FunctionSummaries map[string]*FunctionSummary
	VisitedFunctions  map[string]bool
}

// NewAnalysisState creates a new analysis state with global scope
func NewAnalysisState() *AnalysisState {
	return &AnalysisState{
		ScopeState:        newScopeState(),
		TaintedValues:     make(map[string]*TaintedVariable),
		FunctionSummaries: make(map[string]*FunctionSummary),
		VisitedFunctions:  make(map[string]bool),
	}
}

// LookupVariable looks up a variable in current and parent scopes, then falls
// back to the flat TaintedValues map for file-global tainted variables.
func (s *AnalysisState) LookupVariable(name string) (*TaintedVariable, bool) {
	// Scope-aware hierarchical lookup first.
	if v, ok := s.ScopeState.LookupVariable(name); ok {
		return v, true
	}
	// Fall back to flat global taint map.
	if v, ok := s.TaintedValues[name]; ok {
		return v, true
	}
	return nil, false
}

// SetTainted marks a variable as tainted in the current scope and in the flat
// TaintedValues map so that findTaintInfo can reach it via O(1) lookup.
func (s *AnalysisState) SetTainted(name string, tainted *TaintedVariable) {
	if s.CurrentScope != nil {
		s.CurrentScope.Variables[name] = tainted
	}
	s.TaintedValues[name] = tainted
}

// EnterScope creates and enters a new scope (delegates to ScopeState).
func (s *AnalysisState) EnterScope(scopeType ScopeType, name string, startLine, endLine int) *Scope {
	return s.ScopeState.EnterScope(scopeType, name, startLine, endLine)
}

// ExitScope exits the current scope and returns to parent (delegates to ScopeState).
func (s *AnalysisState) ExitScope() *Scope {
	return s.ScopeState.ExitScope()
}

// Additional fields for full analysis state with O(1) lookups
type FullAnalysisState struct {
	*AnalysisState

	// Maps for O(1) deduplication
	sourcesMap      map[string]*InputSource     // sourceID -> source
	taintedVarsMap  map[string]*TaintedVariable // varKey -> variable
	taintedFuncsMap map[string]*TaintedFunction // funcKey -> function

	// Slices for output (computed on demand)
	Sources          []*InputSource
	TaintedVariables []*TaintedVariable
	TaintedFunctions []*TaintedFunction

	PropagationPaths map[string][]*PropagationPath // source ID -> paths
	ReturnsTainted   map[string]*InputSource       // function name -> source
}

// NewFullAnalysisState creates a complete analysis state with optimized maps
func NewFullAnalysisState() *FullAnalysisState {
	return &FullAnalysisState{
		AnalysisState:    NewAnalysisState(),
		sourcesMap:       make(map[string]*InputSource, 128),
		taintedVarsMap:   make(map[string]*TaintedVariable, 256),
		taintedFuncsMap:  make(map[string]*TaintedFunction, 128),
		Sources:          make([]*InputSource, 0, 128),
		TaintedVariables: make([]*TaintedVariable, 0, 256),
		TaintedFunctions: make([]*TaintedFunction, 0, 128),
		PropagationPaths: make(map[string][]*PropagationPath, 64),
		ReturnsTainted:   make(map[string]*InputSource, 64),
	}
}

// AddSource adds a new input source with O(1) deduplication
func (s *FullAnalysisState) AddSource(source *InputSource) {
	if _, exists := s.sourcesMap[source.ID]; !exists {
		s.sourcesMap[source.ID] = source
		s.Sources = append(s.Sources, source)
	}
}

// AddTaintedVariable adds a tainted variable with O(1) deduplication.
// taintedVarsMap is the single source of truth; AnalysisState.TaintedValues is
// not written here to avoid the dual-map divergence.
func (s *FullAnalysisState) AddTaintedVariable(tv *TaintedVariable) {
	key := tv.Name + ":" + tv.Scope + ":" + tv.Location.FilePath
	if existing, exists := s.taintedVarsMap[key]; exists {
		// Update depth if this path is shorter
		if tv.Depth < existing.Depth {
			s.taintedVarsMap[key] = tv
		}
	} else {
		s.taintedVarsMap[key] = tv
		s.TaintedVariables = append(s.TaintedVariables, tv)
		// Note: deliberately NOT calling s.SetTainted — taintedVarsMap is the
		// authoritative store; keeping AnalysisState.TaintedValues in sync
		// causes dual-map divergence when depths are updated above.
	}
}

// IsTainted reports whether the variable identified by (name, scope, filePath)
// is tracked in the FullAnalysisState's dedup map.
func (s *FullAnalysisState) IsTainted(name, scope, filePath string) (*TaintedVariable, bool) {
	key := name + ":" + scope + ":" + filePath
	tv, ok := s.taintedVarsMap[key]
	return tv, ok
}

// AddTaintedFunction adds a tainted function with O(1) deduplication
func (s *FullAnalysisState) AddTaintedFunction(tf *TaintedFunction) {
	key := tf.Name + ":" + tf.FilePath
	if existing, exists := s.taintedFuncsMap[key]; exists {
		// Merge tainted params (deduplicated)
		paramSet := make(map[string]TaintedParam)
		for _, p := range existing.TaintedParams {
			paramKey := strconv.Itoa(p.Index) + ":" + p.Name
			paramSet[paramKey] = p
		}
		for _, p := range tf.TaintedParams {
			paramKey := strconv.Itoa(p.Index) + ":" + p.Name
			paramSet[paramKey] = p
		}
		merged := make([]TaintedParam, 0, len(paramSet))
		for _, p := range paramSet {
			merged = append(merged, p)
		}
		existing.TaintedParams = merged
	} else {
		s.taintedFuncsMap[key] = tf
		s.TaintedFunctions = append(s.TaintedFunctions, tf)
	}
}

// AddPropagationStep adds a propagation step for a source
func (s *FullAnalysisState) AddPropagationStep(source *InputSource, step PropagationStep) {
	if source == nil {
		return
	}
	paths, exists := s.PropagationPaths[source.ID]
	if !exists || len(paths) == 0 {
		// Create new path
		path := &PropagationPath{
			Source: source,
			Steps:  []PropagationStep{step},
		}
		s.PropagationPaths[source.ID] = []*PropagationPath{path}
	} else {
		// Add to existing path
		paths[len(paths)-1].Steps = append(paths[len(paths)-1].Steps, step)
	}
}

// AddReturnsTaintedFunction marks a function as returning tainted data
func (s *FullAnalysisState) AddReturnsTaintedFunction(funcName string, source *InputSource) {
	s.ReturnsTainted[funcName] = source
}

// GetTaintedVariables returns all tainted variables
func (s *FullAnalysisState) GetTaintedVariables() []*TaintedVariable {
	return s.TaintedVariables
}

// BuildFlowGraph builds a flow graph from the analysis state
func (s *FullAnalysisState) BuildFlowGraph() *FlowGraph {
	graph := &FlowGraph{
		Nodes: make([]FlowNode, 0),
		Edges: make([]FlowEdge, 0),
	}

	nodeIDMap := make(map[string]string) // unique key -> node ID

	// Add source nodes
	for _, source := range s.Sources {
		node := FlowNode{
			ID:       source.ID,
			Type:     FlowNodeSource,
			Name:     source.Type,
			Location: source.Location,
		}
		graph.Nodes = append(graph.Nodes, node)
		nodeIDMap["source:"+source.ID] = source.ID
	}

	// Add variable nodes
	for _, tv := range s.TaintedVariables {
		nodeKey := "var:" + tv.Name + ":" + tv.Location.FilePath
		if _, exists := nodeIDMap[nodeKey]; !exists {
			node := FlowNode{
				ID:       tv.ID,
				Type:     FlowNodeVariable,
				Name:     tv.Name,
				Location: tv.Location,
			}
			graph.Nodes = append(graph.Nodes, node)
			nodeIDMap[nodeKey] = tv.ID

			// Add edge from source to variable
			if tv.Source != nil {
				edge := FlowEdge{
					From:     tv.Source.ID,
					To:       tv.ID,
					Type:     FlowEdgeTaint,
					Location: tv.Location,
				}
				graph.Edges = append(graph.Edges, edge)
			}
		}
	}

	// Add function nodes
	for _, tf := range s.TaintedFunctions {
		nodeKey := "func:" + tf.Name + ":" + tf.FilePath
		if _, exists := nodeIDMap[nodeKey]; !exists {
			node := FlowNode{
				ID:   tf.ID,
				Type: FlowNodeFunction,
				Name: tf.Name,
				Location: Location{
					FilePath: tf.FilePath,
					Line:     tf.Line,
				},
			}
			graph.Nodes = append(graph.Nodes, node)
			nodeIDMap[nodeKey] = tf.ID

			// Add edges from sources to function
			for _, param := range tf.TaintedParams {
				if param.Source != nil {
					edge := FlowEdge{
						From: param.Source.ID,
						To:   tf.ID,
						Type: FlowEdgeCall,
						Location: Location{
							FilePath: tf.FilePath,
							Line:     tf.Line,
						},
					}
					graph.Edges = append(graph.Edges, edge)
				}
			}
		}
	}

	return graph
}
