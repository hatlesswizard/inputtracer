// Package sources - graph_styles.go provides centralized graph visualization styles
// All graph colors, shapes, and edge styles should be defined here
package sources

// GraphNodeType represents node types in flow graphs
type GraphNodeType string

const (
	GraphNodeSource    GraphNodeType = "source"
	GraphNodeVariable  GraphNodeType = "variable"
	GraphNodeFunction  GraphNodeType = "function"
	GraphNodeParameter GraphNodeType = "parameter"
	GraphNodeCarrier   GraphNodeType = "carrier"
	GraphNodeProperty  GraphNodeType = "property"
	GraphNodeReturn    GraphNodeType = "return"
)

// GraphEdgeType represents edge types in flow graphs
type GraphEdgeType string

const (
	GraphEdgeAssignment GraphEdgeType = "assignment"
	GraphEdgeCall       GraphEdgeType = "call"
	GraphEdgeReturn     GraphEdgeType = "return"
	GraphEdgeTaint      GraphEdgeType = "taint"
	GraphEdgeParameter  GraphEdgeType = "parameter"
	GraphEdgeProperty   GraphEdgeType = "property"
)

// NodeStyle defines visual styling for graph nodes
type NodeStyle struct {
	FillColor   string // Hex color for fill
	StrokeColor string // Hex color for stroke/border
	TextColor   string // Hex color for text
	Shape       string // DOT shape name
}

// EdgeStyle defines visual styling for graph edges
type EdgeStyle struct {
	LineStyle  string // "solid", "dashed", "dotted", "bold"
	Color      string // Hex color
	ArrowStyle string // Mermaid arrow style
	Label      string // Optional label
}

// MermaidShape defines Mermaid node shape delimiters
type MermaidShape struct {
	Open  string
	Close string
}

// NodeStyles maps node types to their visual styles
// Replaces hardcoded switch in graph.go getNodeColor()
var NodeStyles = map[GraphNodeType]NodeStyle{
	GraphNodeSource:    {FillColor: "#ff6b6b", StrokeColor: "#333", TextColor: "white", Shape: "ellipse"},
	GraphNodeVariable:  {FillColor: "#4ecdc4", StrokeColor: "#333", TextColor: "white", Shape: "box"},
	GraphNodeFunction:  {FillColor: "#45b7d1", StrokeColor: "#333", TextColor: "white", Shape: "box"},
	GraphNodeParameter: {FillColor: "#96ceb4", StrokeColor: "#333", TextColor: "white", Shape: "box"},
	GraphNodeCarrier:   {FillColor: "#4ecdc4", StrokeColor: "#333", TextColor: "white", Shape: "box"},
	GraphNodeProperty:  {FillColor: "#f9f9f9", StrokeColor: "#333", TextColor: "black", Shape: "box"},
	GraphNodeReturn:    {FillColor: "#f9f9f9", StrokeColor: "#333", TextColor: "black", Shape: "box"},
}

// DefaultNodeStyle is returned when node type is unknown
var DefaultNodeStyle = NodeStyle{
	FillColor:   "#f9f9f9",
	StrokeColor: "#333",
	TextColor:   "black",
	Shape:       "box",
}

// EdgeStyles maps edge types to their visual styles
// Replaces hardcoded switch in graph.go getEdgeStyle()
var EdgeStyles = map[GraphEdgeType]EdgeStyle{
	GraphEdgeAssignment: {LineStyle: "solid", Color: "black", ArrowStyle: "-->"},
	GraphEdgeCall:       {LineStyle: "dashed", Color: "blue", ArrowStyle: "-.->|call|"},
	GraphEdgeReturn:     {LineStyle: "dotted", Color: "green", ArrowStyle: "==>|return|"},
	GraphEdgeTaint:      {LineStyle: "bold", Color: "red", ArrowStyle: "-->|taint|"},
	GraphEdgeParameter:  {LineStyle: "dashed", Color: "purple", ArrowStyle: "-.->"},
	GraphEdgeProperty:   {LineStyle: "solid", Color: "gray", ArrowStyle: "-->"},
}

// DefaultEdgeStyle is returned when edge type is unknown
var DefaultEdgeStyle = EdgeStyle{
	LineStyle:  "solid",
	Color:      "gray",
	ArrowStyle: "-->",
}

// MermaidNodeShapes maps node types to Mermaid shape delimiters
// Replaces hardcoded switch in graph.go getMermaidNodeShape()
var MermaidNodeShapes = map[GraphNodeType]MermaidShape{
	GraphNodeSource:   {Open: "((", Close: "))"},
	GraphNodeFunction: {Open: "[/", Close: "/]"},
}

// DefaultMermaidShape is returned when node type is unknown
var DefaultMermaidShape = MermaidShape{Open: "[", Close: "]"}

// MermaidClassDefs defines Mermaid class styling
// Replaces hardcoded strings in graph.go ExportMermaid()
var MermaidClassDefs = map[GraphNodeType]string{
	GraphNodeSource:   "classDef source fill:#ff6b6b,stroke:#333",
	GraphNodeVariable: "classDef variable fill:#4ecdc4,stroke:#333",
	GraphNodeFunction: "classDef function fill:#45b7d1,stroke:#333",
}

// GetNodeStyle returns the style for a node type
func GetNodeStyle(nodeType GraphNodeType) NodeStyle {
	if style, ok := NodeStyles[nodeType]; ok {
		return style
	}
	return DefaultNodeStyle
}

// GetNodeStyleByString returns the style for a node type string
func GetNodeStyleByString(nodeType string) NodeStyle {
	return GetNodeStyle(GraphNodeType(nodeType))
}

// GetEdgeStyle returns the style for an edge type
func GetEdgeStyle(edgeType GraphEdgeType) EdgeStyle {
	if style, ok := EdgeStyles[edgeType]; ok {
		return style
	}
	return DefaultEdgeStyle
}

// GetEdgeStyleByString returns the style for an edge type string
func GetEdgeStyleByString(edgeType string) EdgeStyle {
	return GetEdgeStyle(GraphEdgeType(edgeType))
}

// GetMermaidNodeShape returns Mermaid shape delimiters for a node type
func GetMermaidNodeShape(nodeType GraphNodeType) (open, close string) {
	if shape, ok := MermaidNodeShapes[nodeType]; ok {
		return shape.Open, shape.Close
	}
	return DefaultMermaidShape.Open, DefaultMermaidShape.Close
}

// GetMermaidNodeShapeByString returns Mermaid shape delimiters for a node type string
func GetMermaidNodeShapeByString(nodeType string) (open, close string) {
	return GetMermaidNodeShape(GraphNodeType(nodeType))
}

// GetDOTEdgeAttributes returns DOT format attributes for an edge type
func GetDOTEdgeAttributes(edgeType GraphEdgeType) string {
	style := GetEdgeStyle(edgeType)
	return "style=" + style.LineStyle + ", color=" + style.Color
}

// GetDOTEdgeAttributesByString returns DOT format attributes for an edge type string
func GetDOTEdgeAttributesByString(edgeType string) string {
	return GetDOTEdgeAttributes(GraphEdgeType(edgeType))
}

// GetMermaidArrowStyle returns the Mermaid arrow style for an edge type
func GetMermaidArrowStyle(edgeType GraphEdgeType) string {
	style := GetEdgeStyle(edgeType)
	return style.ArrowStyle
}

// GetMermaidArrowStyleByString returns the Mermaid arrow style for an edge type string
func GetMermaidArrowStyleByString(edgeType string) string {
	return GetMermaidArrowStyle(GraphEdgeType(edgeType))
}

// GetMermaidClassDefs returns all Mermaid class definitions as a string
func GetMermaidClassDefs() string {
	result := ""
	for _, def := range MermaidClassDefs {
		result += "  " + def + "\n"
	}
	return result
}

// GetNodeFillColor returns the fill color for a node type
func GetNodeFillColor(nodeType GraphNodeType) string {
	return GetNodeStyle(nodeType).FillColor
}

// GetNodeStrokeColor returns the stroke color for a node type
func GetNodeStrokeColor(nodeType GraphNodeType) string {
	return GetNodeStyle(nodeType).StrokeColor
}

// GetDOTEdgeStyle returns DOT format style string for an edge type
func GetDOTEdgeStyle(edgeType GraphEdgeType) string {
	return GetDOTEdgeAttributes(edgeType)
}
