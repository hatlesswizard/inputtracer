// Package output - styles.go provides centralized graph visualization styles.
// All graph colors, shapes, and edge styles for DOT/Mermaid output are defined here.
package output

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

// mermaidShapeDelimiters defines Mermaid node shape delimiters
type mermaidShapeDelimiters struct {
	Open  string
	Close string
}

// NodeStyles maps node types to their visual styles
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

// mermaidNodeShapes maps node types to Mermaid shape delimiters
var mermaidNodeShapes = map[GraphNodeType]mermaidShapeDelimiters{
	GraphNodeSource:   {Open: "((", Close: "))"},
	GraphNodeFunction: {Open: "[/", Close: "/]"},
}

// defaultMermaidShape is returned when node type is unknown
var defaultMermaidShape = mermaidShapeDelimiters{Open: "[", Close: "]"}

// getNodeStyle returns the style for a node type
func getNodeStyle(nodeType GraphNodeType) NodeStyle {
	if style, ok := NodeStyles[nodeType]; ok {
		return style
	}
	return DefaultNodeStyle
}

// getNodeFillColor returns the fill color for a node type
func getNodeFillColor(nodeType GraphNodeType) string {
	return getNodeStyle(nodeType).FillColor
}

// getNodeStrokeColor returns the stroke color for a node type
func getNodeStrokeColor(nodeType GraphNodeType) string {
	return getNodeStyle(nodeType).StrokeColor
}

// getEdgeStyle returns the style for an edge type
func getEdgeStyle(edgeType GraphEdgeType) EdgeStyle {
	if style, ok := EdgeStyles[edgeType]; ok {
		return style
	}
	return DefaultEdgeStyle
}

// getDOTEdgeAttributes returns DOT format attributes for an edge type
func getDOTEdgeAttributes(edgeType GraphEdgeType) string {
	style := getEdgeStyle(edgeType)
	return "style=" + style.LineStyle + ", color=" + style.Color
}

// getMermaidNodeShapeDelimiters returns Mermaid shape delimiters for a node type
func getMermaidNodeShapeDelimiters(nodeType GraphNodeType) (open, close string) {
	if shape, ok := mermaidNodeShapes[nodeType]; ok {
		return shape.Open, shape.Close
	}
	return defaultMermaidShape.Open, defaultMermaidShape.Close
}

// getMermaidArrowStyle returns the Mermaid arrow style for an edge type
func getMermaidArrowStyle(edgeType GraphEdgeType) string {
	return getEdgeStyle(edgeType).ArrowStyle
}
