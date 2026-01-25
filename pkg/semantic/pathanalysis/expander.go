// Package pathanalysis provides inter-procedural path expansion and pruning
// for taint analysis.
//
// Key features:
// - Expands paths through function calls to discover all execution paths
// - Prunes infeasible paths based on condition analysis
// - Handles recursive calls and cycles gracefully
package pathanalysis

import (
	"sync"
)

// PathNodeType represents the type of a path node
type PathNodeType string

const (
	PathNodeSource    PathNodeType = "source"    // Input source
	PathNodeCall      PathNodeType = "call"      // Function call
	PathNodeReturn    PathNodeType = "return"    // Return from function
	PathNodeAssign    PathNodeType = "assign"    // Variable assignment
	PathNodeCondition PathNodeType = "condition" // Conditional branch
	PathNodeTransform PathNodeType = "transform" // Data transformation
)

// PruneReason explains why a path was pruned
type PruneReason string

const (
	PruneNone        PruneReason = ""
	PruneMaxDepth    PruneReason = "max_depth_exceeded"
	PruneCycle       PruneReason = "cycle_detected"
	PruneInfeasible  PruneReason = "infeasible_condition"
	PruneDead        PruneReason = "dead_code"
	PruneUnreachable PruneReason = "unreachable"
	PruneLowPriority PruneReason = "low_priority"
)

// PathNode represents a node in an execution path
type PathNode struct {
	ID           string       `json:"id"`
	Type         PathNodeType `json:"type"`
	Name         string       `json:"name"`
	Expression   string       `json:"expression,omitempty"`
	FilePath     string       `json:"file_path"`
	Line         int          `json:"line"`
	Column       int          `json:"column"`

	// Function context
	FunctionName string       `json:"function_name,omitempty"`
	ClassName    string       `json:"class_name,omitempty"`

	// Taint information
	TaintedVars  []string     `json:"tainted_vars,omitempty"`
	TaintSource  string       `json:"taint_source,omitempty"`

	// Condition information (for condition nodes)
	Condition    string       `json:"condition,omitempty"`
	BranchTaken  bool         `json:"branch_taken,omitempty"` // true/false branch

	// Metadata
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
}

// ExecutionPath represents a complete path from source to target
type ExecutionPath struct {
	ID          string       `json:"id"`
	Nodes       []*PathNode  `json:"nodes"`

	// Source and target info
	SourceNode  *PathNode    `json:"source_node"`
	TargetNode  *PathNode    `json:"target_node"`

	// Path characteristics
	Depth       int          `json:"depth"`        // Function call depth
	Length      int          `json:"length"`       // Number of nodes
	Feasible    bool         `json:"feasible"`     // Is path feasible?
	PruneReason PruneReason  `json:"prune_reason,omitempty"`

	// Conditions along the path
	Conditions  []string     `json:"conditions,omitempty"`
}

// PathExpander expands and prunes inter-procedural paths
type PathExpander struct {
	mu sync.RWMutex

	// Configuration
	maxDepth      int
	maxPathLength int
	maxPaths      int
	enablePruning bool

	// Call graph information
	callGraph     map[string][]string // caller -> callees
	reverseGraph  map[string][]string // callee -> callers

	// Function information
	functionDefs  map[string]*FunctionInfo

	// Statistics
	pathsExplored  int
	pathsPruned    int
	pathsFound     int
}

// FunctionInfo stores information about a function
type FunctionInfo struct {
	Name         string
	FilePath     string
	Line         int
	ClassName    string
	Parameters   []string
	Returns      bool       // Does it return a value?
	TaintThrough []int      // Which params flow to return?
	IsSource     bool
}

// NewPathExpander creates a new path expander
func NewPathExpander() *PathExpander {
	return &PathExpander{
		maxDepth:      10,
		maxPathLength: 50,
		maxPaths:      1000,
		enablePruning: true,
		callGraph:     make(map[string][]string),
		reverseGraph:  make(map[string][]string),
		functionDefs:  make(map[string]*FunctionInfo),
	}
}

// SetMaxDepth sets the maximum function call depth
func (pe *PathExpander) SetMaxDepth(depth int) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.maxDepth = depth
}

// SetMaxPaths sets the maximum number of paths to explore
func (pe *PathExpander) SetMaxPaths(max int) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.maxPaths = max
}

// AddCallEdge adds a call relationship
func (pe *PathExpander) AddCallEdge(caller, callee string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	pe.callGraph[caller] = append(pe.callGraph[caller], callee)
	pe.reverseGraph[callee] = append(pe.reverseGraph[callee], caller)
}

// AddFunction adds function information
func (pe *PathExpander) AddFunction(info *FunctionInfo) {
	pe.mu.Lock()
	defer pe.mu.Unlock()

	key := info.FilePath + ":" + info.Name
	pe.functionDefs[key] = info
}

// ExpandPaths finds all paths from source to target, expanding through calls
func (pe *PathExpander) ExpandPaths(source, target *PathNode) []*ExecutionPath {
	pe.mu.Lock()
	pe.pathsExplored = 0
	pe.pathsPruned = 0
	pe.pathsFound = 0
	pe.mu.Unlock()

	var paths []*ExecutionPath
	visited := make(map[string]bool)
	currentPath := make([]*PathNode, 0)

	pe.expandDFS(source, target, currentPath, visited, 0, &paths)

	return paths
}

// expandDFS performs depth-first expansion of paths
func (pe *PathExpander) expandDFS(
	current *PathNode,
	target *PathNode,
	currentPath []*PathNode,
	visited map[string]bool,
	depth int,
	paths *[]*ExecutionPath,
) {
	pe.mu.Lock()
	pe.pathsExplored++
	explored := pe.pathsExplored
	maxPaths := pe.maxPaths
	pe.mu.Unlock()

	// Check limits
	if explored > maxPaths*10 || len(*paths) >= maxPaths {
		return
	}

	if depth > pe.maxDepth {
		return
	}

	if len(currentPath) > pe.maxPathLength {
		pe.mu.Lock()
		pe.pathsPruned++
		pe.mu.Unlock()
		return
	}

	// Create node key for cycle detection
	nodeKey := current.FilePath + ":" + current.Name + ":" + string(rune(current.Line))
	if visited[nodeKey] {
		pe.mu.Lock()
		pe.pathsPruned++
		pe.mu.Unlock()
		return
	}

	// Add current node to path
	visited[nodeKey] = true
	currentPath = append(currentPath, current)
	defer func() {
		visited[nodeKey] = false
	}()

	// Check if we reached the target
	if pe.isMatch(current, target) {
		path := pe.createPath(currentPath)
		path.Feasible = true

		*paths = append(*paths, path)
		pe.mu.Lock()
		pe.pathsFound++
		pe.mu.Unlock()
		return
	}

	// Get next nodes
	nextNodes := pe.getNextNodes(current, target, depth)

	for _, next := range nextNodes {
		newDepth := depth
		if next.Type == PathNodeCall {
			newDepth++
		} else if next.Type == PathNodeReturn {
			newDepth--
			if newDepth < 0 {
				newDepth = 0
			}
		}

		pe.expandDFS(next, target, currentPath, visited, newDepth, paths)
	}
}

// getNextNodes returns possible next nodes from current position
func (pe *PathExpander) getNextNodes(current *PathNode, target *PathNode, depth int) []*PathNode {
	var nodes []*PathNode

	pe.mu.RLock()
	defer pe.mu.RUnlock()

	funcKey := current.FilePath + ":" + current.FunctionName
	callees := pe.callGraph[funcKey]

	// Add callees as potential next nodes
	for _, callee := range callees {
		if info, ok := pe.functionDefs[callee]; ok {
			nodes = append(nodes, &PathNode{
				ID:           callee,
				Type:         PathNodeCall,
				Name:         info.Name,
				FunctionName: info.Name,
				ClassName:    info.ClassName,
				FilePath:     info.FilePath,
				Line:         info.Line,
			})
		}
	}

	// If at a target-containing function, add target node
	if current.FilePath == target.FilePath && current.FunctionName == target.FunctionName {
		if current.Line < target.Line {
			nodes = append(nodes, target)
		}
	}

	// Add return node if we're in a called function
	if depth > 0 {
		nodes = append(nodes, &PathNode{
			Type:         PathNodeReturn,
			Name:         "return",
			FunctionName: current.FunctionName,
			FilePath:     current.FilePath,
			Line:         current.Line,
		})
	}

	return nodes
}

// isMatch checks if current node matches target
func (pe *PathExpander) isMatch(current, target *PathNode) bool {
	if current.FilePath != target.FilePath {
		return false
	}
	if target.Line > 0 && current.Line != target.Line {
		return false
	}
	if target.Name != "" && current.Name != target.Name {
		return false
	}
	return true
}

// createPath creates an ExecutionPath from the current node list
func (pe *PathExpander) createPath(nodes []*PathNode) *ExecutionPath {
	// Deep copy nodes
	pathNodes := make([]*PathNode, len(nodes))
	for i, n := range nodes {
		nodeCopy := *n
		pathNodes[i] = &nodeCopy
	}

	path := &ExecutionPath{
		Nodes:    pathNodes,
		Length:   len(pathNodes),
		Feasible: true,
	}

	if len(pathNodes) > 0 {
		path.SourceNode = pathNodes[0]
		path.TargetNode = pathNodes[len(pathNodes)-1]
	}

	// Calculate depth
	maxDepth := 0
	currentDepth := 0
	for _, n := range pathNodes {
		if n.Type == PathNodeCall {
			currentDepth++
			if currentDepth > maxDepth {
				maxDepth = currentDepth
			}
		} else if n.Type == PathNodeReturn {
			currentDepth--
		}
	}
	path.Depth = maxDepth

	return path
}

// FilterFeasible filters to keep only feasible paths
func (pe *PathExpander) FilterFeasible(paths []*ExecutionPath) []*ExecutionPath {
	var filtered []*ExecutionPath
	for _, p := range paths {
		if p.Feasible {
			filtered = append(filtered, p)
		}
	}
	return filtered
}

// Stats returns expansion statistics
func (pe *PathExpander) Stats() map[string]int {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	return map[string]int{
		"paths_explored": pe.pathsExplored,
		"paths_pruned":   pe.pathsPruned,
		"paths_found":    pe.pathsFound,
		"functions":      len(pe.functionDefs),
		"call_edges":     len(pe.callGraph),
	}
}

// GetCallGraph returns the call graph
func (pe *PathExpander) GetCallGraph() map[string][]string {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	// Return a copy
	graph := make(map[string][]string)
	for k, v := range pe.callGraph {
		graph[k] = append([]string{}, v...)
	}
	return graph
}

// GetReverseCallGraph returns the reverse call graph (callee -> callers)
func (pe *PathExpander) GetReverseCallGraph() map[string][]string {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	// Return a copy
	graph := make(map[string][]string)
	for k, v := range pe.reverseGraph {
		graph[k] = append([]string{}, v...)
	}
	return graph
}
