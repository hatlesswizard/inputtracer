package output

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/tracer"
)

// ExportDOT exports the flow graph in Graphviz DOT format.
// The output is structured in four phases: header, nodes, edges, footer.
func ExportDOT(graph *tracer.FlowGraph) string {
	var sb strings.Builder
	writeDOTHeader(&sb)
	writeDOTNodes(&sb, graph)
	writeDOTEdges(&sb, graph)
	writeDOTFooter(&sb)
	return sb.String()
}

// writeDOTHeader writes the DOT graph declaration and global attributes.
func writeDOTHeader(sb *strings.Builder) {
	sb.WriteString("digraph InputFlow {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box];\n\n")
	sb.WriteString("  // Node styles\n")
	sb.WriteString("  node [style=filled];\n\n")
}

// writeDOTNodes writes a DOT node statement for every node in the graph.
func writeDOTNodes(sb *strings.Builder, graph *tracer.FlowGraph) {
	sb.WriteString("  // Nodes\n")
	for _, node := range graph.Nodes {
		color := dotNodeColor(string(node.Type))
		label := dotEscapeLabel(node.Name)
		if node.Location.Line > 0 {
			label = fmt.Sprintf("%s\\n%s:%d", label, dotTruncatePath(node.Location.FilePath), node.Location.Line)
		}
		fmt.Fprintf(sb, "  \"%s\" [label=\"%s\", fillcolor=\"%s\"];\n", node.ID, label, color)
	}
}

// writeDOTEdges writes a DOT edge statement for every edge in the graph.
func writeDOTEdges(sb *strings.Builder, graph *tracer.FlowGraph) {
	sb.WriteString("\n  // Edges\n")
	for _, edge := range graph.Edges {
		style := dotEdgeStyle(string(edge.Type))
		fmt.Fprintf(sb, "  \"%s\" -> \"%s\" [%s];\n", edge.From, edge.To, style)
	}
}

// writeDOTFooter writes the closing brace of the DOT graph.
func writeDOTFooter(sb *strings.Builder) {
	sb.WriteString("}\n")
}

// dotNodeColor returns the fill colour for a DOT node based on its type.
func dotNodeColor(nodeType string) string {
	return getNodeFillColor(GraphNodeType(nodeType))
}

// dotEdgeStyle returns the DOT attribute string for an edge based on its type.
func dotEdgeStyle(edgeType string) string {
	return getDOTEdgeAttributes(GraphEdgeType(edgeType))
}

// dotEscapeLabel escapes special characters in a DOT label string.
func dotEscapeLabel(s string) string {
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "<", "\\<")
	s = strings.ReplaceAll(s, ">", "\\>")
	return s
}

// dotTruncatePath shortens a file path for display by keeping only the last
// two path components, prefixed with "...".
func dotTruncatePath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return path
	}
	return "..." + strings.Join(parts[len(parts)-2:], "/")
}

// ExportMermaid exports the flow graph in Mermaid format.
func ExportMermaid(graph *tracer.FlowGraph) string {
	var sb strings.Builder

	sb.WriteString("graph LR\n")

	// Add nodes with styling
	for _, node := range graph.Nodes {
		open, close := getMermaidNodeShapeDelimiters(GraphNodeType(string(node.Type)))
		label := dotEscapeLabel(node.Name)
		fmt.Fprintf(&sb, "  %s%s\"%s\"%s\n", node.ID, open, label, close)
	}

	sb.WriteString("\n")

	// Add edges
	for _, edge := range graph.Edges {
		arrowStyle := getMermaidArrowStyle(GraphEdgeType(string(edge.Type)))
		fmt.Fprintf(&sb, "  %s %s %s\n", edge.From, arrowStyle, edge.To)
	}

	// Add class definitions using centralised graph styles
	sb.WriteString("\n")
	fmt.Fprintf(&sb, "  classDef source fill:%s,stroke:%s\n", getNodeFillColor(GraphNodeSource), getNodeStrokeColor(GraphNodeSource))
	fmt.Fprintf(&sb, "  classDef variable fill:%s,stroke:%s\n", getNodeFillColor(GraphNodeVariable), getNodeStrokeColor(GraphNodeVariable))
	fmt.Fprintf(&sb, "  classDef function fill:%s,stroke:%s\n", getNodeFillColor(GraphNodeFunction), getNodeStrokeColor(GraphNodeFunction))

	// Apply classes
	for _, node := range graph.Nodes {
		fmt.Fprintf(&sb, "  class %s %s\n", node.ID, node.Type)
	}

	return sb.String()
}

// ExportJSON exports the flow graph as JSON.
func ExportJSON(graph *tracer.FlowGraph, pretty bool) (string, error) {
	var (
		data []byte
		err  error
	)
	if pretty {
		data, err = json.MarshalIndent(graph, "", "  ")
	} else {
		data, err = json.Marshal(graph)
	}
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// PathFinder finds paths between nodes in the flow graph.
type PathFinder struct {
	graph    *tracer.FlowGraph
	adjList  map[string][]string
	maxDepth int
}

// NewPathFinder creates a new PathFinder for the given graph.
func NewPathFinder(graph *tracer.FlowGraph, maxDepth int) *PathFinder {
	pf := &PathFinder{
		graph:    graph,
		adjList:  make(map[string][]string),
		maxDepth: maxDepth,
	}

	// Build adjacency list
	for _, edge := range graph.Edges {
		pf.adjList[edge.From] = append(pf.adjList[edge.From], edge.To)
	}

	return pf
}

// FindAllPaths finds all paths from a source node to all reachable leaf nodes.
func (pf *PathFinder) FindAllPaths(sourceID string) [][]string {
	var allPaths [][]string
	visited := make(map[string]bool)
	currentPath := []string{sourceID}

	pf.dfs(sourceID, visited, currentPath, &allPaths, 0)
	return allPaths
}

// dfs performs a depth-first search to enumerate paths.
func (pf *PathFinder) dfs(node string, visited map[string]bool, currentPath []string, allPaths *[][]string, depth int) {
	if depth > pf.maxDepth {
		return
	}

	visited[node] = true

	neighbors, exists := pf.adjList[node]
	if !exists || len(neighbors) == 0 {
		// Leaf node — record a copy of the current path.
		pathCopy := make([]string, len(currentPath))
		copy(pathCopy, currentPath)
		*allPaths = append(*allPaths, pathCopy)
	}

	for _, neighbor := range neighbors {
		if !visited[neighbor] {
			currentPath = append(currentPath, neighbor)
			pf.dfs(neighbor, visited, currentPath, allPaths, depth+1)
			currentPath = currentPath[:len(currentPath)-1]
		}
	}

	visited[node] = false
}

// FindPathsToFunction finds all paths from any source node to a specific function node.
func (pf *PathFinder) FindPathsToFunction(funcID string) [][]string {
	var paths [][]string

	// Collect all source nodes.
	sourceNodes := make([]string, 0)
	for _, node := range pf.graph.Nodes {
		if node.Type == "source" {
			sourceNodes = append(sourceNodes, node.ID)
		}
	}

	// Find paths from each source that pass through the target function.
	for _, sourceID := range sourceNodes {
		allPaths := pf.FindAllPaths(sourceID)
		for _, path := range allPaths {
			if slices.Contains(path, funcID) {
				paths = append(paths, path)
			}
		}
	}

	return paths
}
