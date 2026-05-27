package types

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ============================================================================
// Core Flow Node/Edge Types
// ============================================================================

// FlowNode represents a node in the data flow graph
type FlowNode struct {
	ID       string       `json:"id"`
	Type     FlowNodeType `json:"type"`
	Language string       `json:"language"`

	// Location information
	FilePath  string `json:"file_path"`
	Line      int    `json:"line"`
	Column    int    `json:"column"`
	EndLine   int    `json:"end_line,omitempty"`
	EndColumn int    `json:"end_column,omitempty"`

	// Semantic information
	Name       string `json:"name"`                  // Variable/function/property name
	ClassName  string `json:"class_name,omitempty"`  // If part of a class
	MethodName string `json:"method_name,omitempty"` // If inside a method
	Scope      string `json:"scope,omitempty"`       // Scope identifier

	// Type information
	TypeInfo *TypeInfo `json:"type_info,omitempty"`

	// Source information (if this is a source node)
	SourceType SourceType `json:"source_type,omitempty"`
	SourceKey  string     `json:"source_key,omitempty"` // Parameter name

	// Carrier information
	CarrierType string `json:"carrier_type,omitempty"` // "array", "object_property", etc.

	// Code snippet
	Snippet string `json:"snippet"`

	// Metadata
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// FlowEdge represents a directed edge in the data flow graph
type FlowEdge struct {
	ID   string       `json:"id"`
	From string       `json:"from"` // Source node ID
	To   string       `json:"to"`   // Target node ID
	Type FlowEdgeType `json:"type"`

	// Location where flow occurs
	FilePath string `json:"file_path"`
	Line     int    `json:"line"`

	// Human-readable description
	Description string `json:"description"`

	// Code causing the flow
	Code string `json:"code,omitempty"`

	// Additional context
	Metadata map[string]interface{} `json:"metadata,omitempty"`
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
func (fm *FlowMap) AddNode(node FlowNode) bool {
	if fm.nodeIndex == nil {
		fm.nodeIndex = make(map[string]bool, 256)
	}
	if fm.nodeIndex[node.ID] {
		return false // Already exists
	}
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
func (fm *FlowMap) AddEdge(edge FlowEdge) bool {
	if fm.edgeIndex == nil {
		fm.edgeIndex = make(map[string]bool, 512)
	}
	edgeKey := edge.From + "->" + edge.To + ":" + string(edge.Type)
	if fm.edgeIndex[edgeKey] {
		return false // Already exists
	}
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
	AnalyzedAt    time.Time `json:"analyzed_at"`
	Duration      string    `json:"duration"`
	FilesAnalyzed int       `json:"files_analyzed"`
	Language      string    `json:"language"`
	Framework     string    `json:"framework,omitempty"`
	TracerVersion string    `json:"tracer_version"`
}
