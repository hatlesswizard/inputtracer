package tracer

import (
	"encoding/json"
	"time"
)

// InputLabel categorizes the type of user input
type InputLabel string

const (
	LabelHTTPGet     InputLabel = "http_get"
	LabelHTTPPost    InputLabel = "http_post"
	LabelHTTPCookie  InputLabel = "http_cookie"
	LabelHTTPHeader  InputLabel = "http_header"
	LabelHTTPBody    InputLabel = "http_body"
	LabelCLI         InputLabel = "cli"
	LabelEnvironment InputLabel = "environment"
	LabelFile        InputLabel = "file"
	LabelDatabase    InputLabel = "database"
	LabelNetwork     InputLabel = "network"
	LabelUserInput   InputLabel = "user_input"
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
	Type     string       `json:"type"`      // e.g., "$_GET", "req.body", "argv"
	Key      string       `json:"key"`       // e.g., "username" in $_GET['username']
	Location Location     `json:"location"`
	Labels   []InputLabel `json:"labels"`
	Language string       `json:"language"`
}

// TaintedVariable represents a variable that holds user input at some point
type TaintedVariable struct {
	ID       string       `json:"id"`
	Name     string       `json:"name"`
	Scope    string       `json:"scope"`    // Function/class scope
	Source   *InputSource `json:"source"`   // Original input source
	Location Location     `json:"location"`
	Depth    int          `json:"depth"`    // How many assignments from original source
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
type PropagationStepType string

const (
	StepAssignment    PropagationStepType = "assignment"
	StepParameterPass PropagationStepType = "parameter_pass"
	StepReturn        PropagationStepType = "return"
	StepConcatenation PropagationStepType = "concatenation"
	StepArrayAccess   PropagationStepType = "array_access"
	StepObjectAccess  PropagationStepType = "object_access"
	StepDestructure   PropagationStepType = "destructure"
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
	ID       string   `json:"id"`
	Type     string   `json:"type"` // "source", "variable", "function", "parameter"
	Name     string   `json:"name"`
	Location Location `json:"location"`
}

// FlowEdge connects two nodes showing data flow
type FlowEdge struct {
	From     string   `json:"from"`     // Node ID
	To       string   `json:"to"`       // Node ID
	Type     string   `json:"type"`     // "assignment", "call", "return"
	Location Location `json:"location"`
}

// FlowGraph represents the complete input flow graph
type FlowGraph struct {
	Nodes []FlowNode `json:"nodes"`
	Edges []FlowEdge `json:"edges"`
}

// TraceStats contains analysis statistics
type TraceStats struct {
	FilesAnalyzed     int           `json:"files_analyzed"`
	SourcesFound      int           `json:"sources_found"`
	TaintedVarsFound  int           `json:"tainted_variables_found"`
	TaintedFuncsFound int           `json:"tainted_functions_found"`
	PropagationPaths  int           `json:"propagation_paths"`
	AnalysisDuration  time.Duration `json:"analysis_duration_ns"`
	DurationMs        int64         `json:"analysis_duration_ms"`
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

	// Errors encountered during analysis
	Errors []string `json:"errors,omitempty"`
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
	ParamsToReturn  []int           `json:"params_to_return"`  // Indices of params that flow to return
	ParamsToParams  map[int][]int   `json:"params_to_params"`  // Param N flows to param M in nested calls
	IsSource        bool            `json:"is_source"`         // Function itself returns user input
	CalledFunctions []string        `json:"called_functions"`
}

// Scope represents a variable scope in the code
type Scope struct {
	ID        string                       `json:"id"`
	Type      ScopeType                    `json:"type"`
	Name      string                       `json:"name"`
	Parent    *Scope                       `json:"-"` // Avoid circular JSON
	ParentID  string                       `json:"parent_id,omitempty"`
	Children  []*Scope                     `json:"-"` // Child scopes
	Variables map[string]*TaintedVariable  `json:"-"`
	StartLine int                          `json:"start_line"`
	EndLine   int                          `json:"end_line"`
	StartLoc  Location                     `json:"start_location,omitempty"`
}

// Note: ScopeType is defined in scope.go

// AnalysisState maintains the current state during analysis
type AnalysisState struct {
	CurrentScope     *Scope
	ScopeStack       []*Scope
	TaintedValues    map[string]*TaintedVariable // variable name -> tainted info
	FunctionSummaries map[string]*FunctionSummary
	VisitedFunctions map[string]bool
}

// NewAnalysisState creates a new analysis state with global scope
func NewAnalysisState() *AnalysisState {
	globalScope := &Scope{
		ID:        "global",
		Type:      ScopeGlobal,
		Name:      "global",
		Variables: make(map[string]*TaintedVariable),
	}
	return &AnalysisState{
		CurrentScope:      globalScope,
		ScopeStack:        []*Scope{globalScope},
		TaintedValues:     make(map[string]*TaintedVariable),
		FunctionSummaries: make(map[string]*FunctionSummary),
		VisitedFunctions:  make(map[string]bool),
	}
}

// EnterScope creates and enters a new scope
func (s *AnalysisState) EnterScope(scopeType ScopeType, name string, startLine, endLine int) *Scope {
	newScope := &Scope{
		ID:        name + "_" + string(scopeType),
		Type:      scopeType,
		Name:      name,
		Parent:    s.CurrentScope,
		Variables: make(map[string]*TaintedVariable),
		StartLine: startLine,
		EndLine:   endLine,
	}
	if s.CurrentScope != nil {
		newScope.ParentID = s.CurrentScope.ID
	}
	s.ScopeStack = append(s.ScopeStack, newScope)
	s.CurrentScope = newScope
	return newScope
}

// ExitScope exits the current scope and returns to parent
func (s *AnalysisState) ExitScope() *Scope {
	if len(s.ScopeStack) <= 1 {
		return s.CurrentScope
	}
	s.ScopeStack = s.ScopeStack[:len(s.ScopeStack)-1]
	s.CurrentScope = s.ScopeStack[len(s.ScopeStack)-1]
	return s.CurrentScope
}

// LookupVariable looks up a variable in current and parent scopes
func (s *AnalysisState) LookupVariable(name string) (*TaintedVariable, bool) {
	// Check current scope and walk up
	scope := s.CurrentScope
	for scope != nil {
		if v, ok := scope.Variables[name]; ok {
			return v, true
		}
		scope = scope.Parent
	}
	// Also check global tainted values
	if v, ok := s.TaintedValues[name]; ok {
		return v, true
	}
	return nil, false
}

// SetTainted marks a variable as tainted in current scope
func (s *AnalysisState) SetTainted(name string, tainted *TaintedVariable) {
	if s.CurrentScope != nil {
		s.CurrentScope.Variables[name] = tainted
	}
	s.TaintedValues[name] = tainted
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

	// Track if slices need rebuilding
	slicesStale bool
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
		slicesStale:      true,
	}
}

// AddSource adds a new input source with O(1) deduplication
func (s *FullAnalysisState) AddSource(source *InputSource) {
	if _, exists := s.sourcesMap[source.ID]; !exists {
		s.sourcesMap[source.ID] = source
		s.Sources = append(s.Sources, source)
	}
}

// AddTaintedVariable adds a tainted variable with O(1) deduplication
func (s *FullAnalysisState) AddTaintedVariable(tv *TaintedVariable) {
	key := tv.Name + ":" + tv.Scope + ":" + tv.Location.FilePath
	if existing, exists := s.taintedVarsMap[key]; exists {
		// Update depth if this path is shorter
		if tv.Depth < existing.Depth {
			s.taintedVarsMap[key] = tv
			s.slicesStale = true
		}
	} else {
		s.taintedVarsMap[key] = tv
		s.TaintedVariables = append(s.TaintedVariables, tv)
		s.SetTainted(tv.Name, tv)
	}
}

// AddTaintedFunction adds a tainted function with O(1) deduplication
func (s *FullAnalysisState) AddTaintedFunction(tf *TaintedFunction) {
	key := tf.Name + ":" + tf.FilePath
	if existing, exists := s.taintedFuncsMap[key]; exists {
		// Merge tainted params (deduplicated)
		paramSet := make(map[string]TaintedParam)
		for _, p := range existing.TaintedParams {
			paramKey := string(rune(p.Index)) + ":" + p.Name
			paramSet[paramKey] = p
		}
		for _, p := range tf.TaintedParams {
			paramKey := string(rune(p.Index)) + ":" + p.Name
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
			Type:     "source",
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
				Type:     "variable",
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
					Type:     "taint",
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
				Type: "function",
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
						Type: "call",
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
