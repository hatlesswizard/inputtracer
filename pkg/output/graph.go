package output

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/sources"
	"github.com/hatlesswizard/inputtracer/pkg/tracer"
)

// GraphExporter exports flow graphs in various formats
type GraphExporter struct{}

// NewGraphExporter creates a new graph exporter
func NewGraphExporter() *GraphExporter {
	return &GraphExporter{}
}

// ExportDOT exports the flow graph in Graphviz DOT format
func (e *GraphExporter) ExportDOT(graph *tracer.FlowGraph) string {
	var sb strings.Builder

	sb.WriteString("digraph InputFlow {\n")
	sb.WriteString("  rankdir=LR;\n")
	sb.WriteString("  node [shape=box];\n\n")

	// Define node styles
	sb.WriteString("  // Node styles\n")
	sb.WriteString("  node [style=filled];\n\n")

	// Add nodes
	sb.WriteString("  // Nodes\n")
	for _, node := range graph.Nodes {
		color := e.getNodeColor(node.Type)
		label := e.escapeLabel(node.Name)
		if node.Location.Line > 0 {
			label = fmt.Sprintf("%s\\n%s:%d", label, e.truncatePath(node.Location.FilePath), node.Location.Line)
		}
		sb.WriteString(fmt.Sprintf("  \"%s\" [label=\"%s\", fillcolor=\"%s\"];\n", node.ID, label, color))
	}

	sb.WriteString("\n  // Edges\n")
	// Add edges
	for _, edge := range graph.Edges {
		style := e.getEdgeStyle(edge.Type)
		sb.WriteString(fmt.Sprintf("  \"%s\" -> \"%s\" [%s];\n", edge.From, edge.To, style))
	}

	sb.WriteString("}\n")
	return sb.String()
}

// getNodeColor returns the color for a node type using centralized graph styles
func (e *GraphExporter) getNodeColor(nodeType string) string {
	return sources.GetNodeFillColor(sources.GraphNodeType(nodeType))
}

// getEdgeStyle returns DOT style for an edge type using centralized graph styles
func (e *GraphExporter) getEdgeStyle(edgeType string) string {
	return sources.GetDOTEdgeStyle(sources.GraphEdgeType(edgeType))
}

// escapeLabel escapes special characters in DOT labels
func (e *GraphExporter) escapeLabel(s string) string {
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "<", "\\<")
	s = strings.ReplaceAll(s, ">", "\\>")
	return s
}

// truncatePath truncates a file path for display
func (e *GraphExporter) truncatePath(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) <= 2 {
		return path
	}
	return "..." + strings.Join(parts[len(parts)-2:], "/")
}

// ExportMermaid exports the flow graph in Mermaid format
func (e *GraphExporter) ExportMermaid(graph *tracer.FlowGraph) string {
	var sb strings.Builder

	sb.WriteString("graph LR\n")

	// Add nodes with styling
	for _, node := range graph.Nodes {
		nodeShape := e.getMermaidNodeShape(node.Type)
		label := e.escapeLabel(node.Name)
		sb.WriteString(fmt.Sprintf("  %s%s\"%s\"%s\n", node.ID, nodeShape.open, label, nodeShape.close))
	}

	sb.WriteString("\n")

	// Add edges
	for _, edge := range graph.Edges {
		arrowStyle := e.getMermaidArrowStyle(edge.Type)
		sb.WriteString(fmt.Sprintf("  %s %s %s\n", edge.From, arrowStyle, edge.To))
	}

	// Add styling using centralized graph styles
	sb.WriteString("\n")
	sb.WriteString(fmt.Sprintf("  classDef source fill:%s,stroke:%s\n", sources.GetNodeFillColor(sources.GraphNodeSource), sources.GetNodeStrokeColor(sources.GraphNodeSource)))
	sb.WriteString(fmt.Sprintf("  classDef variable fill:%s,stroke:%s\n", sources.GetNodeFillColor(sources.GraphNodeVariable), sources.GetNodeStrokeColor(sources.GraphNodeVariable)))
	sb.WriteString(fmt.Sprintf("  classDef function fill:%s,stroke:%s\n", sources.GetNodeFillColor(sources.GraphNodeFunction), sources.GetNodeStrokeColor(sources.GraphNodeFunction)))

	// Apply classes
	for _, node := range graph.Nodes {
		sb.WriteString(fmt.Sprintf("  class %s %s\n", node.ID, node.Type))
	}

	return sb.String()
}

type mermaidShape struct {
	open  string
	close string
}

func (e *GraphExporter) getMermaidNodeShape(nodeType string) mermaidShape {
	open, close := sources.GetMermaidNodeShape(sources.GraphNodeType(nodeType))
	return mermaidShape{open, close}
}

func (e *GraphExporter) getMermaidArrowStyle(edgeType string) string {
	return sources.GetMermaidArrowStyle(sources.GraphEdgeType(edgeType))
}

// ExportJSON exports the flow graph as JSON
func (e *GraphExporter) ExportJSON(graph *tracer.FlowGraph, pretty bool) (string, error) {
	var data []byte
	var err error

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

// PathFinder finds paths between nodes in the flow graph
type PathFinder struct {
	graph    *tracer.FlowGraph
	adjList  map[string][]string
	maxDepth int
}

// NewPathFinder creates a new path finder
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

// FindAllPaths finds all paths from a source to all reachable nodes
func (pf *PathFinder) FindAllPaths(sourceID string) [][]string {
	var allPaths [][]string
	visited := make(map[string]bool)
	currentPath := []string{sourceID}

	pf.dfs(sourceID, visited, currentPath, &allPaths, 0)
	return allPaths
}

// dfs performs depth-first search to find paths
func (pf *PathFinder) dfs(node string, visited map[string]bool, currentPath []string, allPaths *[][]string, depth int) {
	if depth > pf.maxDepth {
		return
	}

	visited[node] = true

	neighbors, exists := pf.adjList[node]
	if !exists || len(neighbors) == 0 {
		// Leaf node - record path
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

// FindPathsToFunction finds all paths from any source to a specific function
func (pf *PathFinder) FindPathsToFunction(funcID string) [][]string {
	var paths [][]string

	// Find all source nodes
	sourceNodes := make([]string, 0)
	for _, node := range pf.graph.Nodes {
		if node.Type == "source" {
			sourceNodes = append(sourceNodes, node.ID)
		}
	}

	// Find paths from each source
	for _, sourceID := range sourceNodes {
		allPaths := pf.FindAllPaths(sourceID)
		for _, path := range allPaths {
			// Check if path ends at or contains the function
			for _, nodeID := range path {
				if nodeID == funcID {
					paths = append(paths, path)
					break
				}
			}
		}
	}

	return paths
}
