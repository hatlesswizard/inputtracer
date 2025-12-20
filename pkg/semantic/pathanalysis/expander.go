// Package pathanalysis provides inter-procedural path expansion and pruning
// for taint analysis, inspired by ATLANTIS's approach to vulnerability discovery.
//
// Key features:
// - Expands paths through function calls to discover all execution paths
// - Prunes infeasible paths based on condition analysis
// - Prioritizes paths more likely to be exploitable
// - Handles recursive calls and cycles gracefully
package pathanalysis

import (
	"container/heap"
	"sync"
)

// PathNodeType represents the type of a path node
type PathNodeType string

const (
	PathNodeSource   PathNodeType = "source"   // Input source
	PathNodeSink     PathNodeType = "sink"     // Security-sensitive sink
	PathNodeCall     PathNodeType = "call"     // Function call
	PathNodeReturn   PathNodeType = "return"   // Return from function
	PathNodeAssign   PathNodeType = "assign"   // Variable assignment
	PathNodeCondition PathNodeType = "condition" // Conditional branch
	PathNodeTransform PathNodeType = "transform" // Data transformation
)

// PruneReason explains why a path was pruned
type PruneReason string

const (
	PruneNone          PruneReason = ""
	PruneMaxDepth      PruneReason = "max_depth_exceeded"
	PruneCycle         PruneReason = "cycle_detected"
	PruneInfeasible    PruneReason = "infeasible_condition"
	PruneSanitized     PruneReason = "data_sanitized"
	PruneTypeCoercion  PruneReason = "type_coercion"
	PruneDead          PruneReason = "dead_code"
	PruneUnreachable   PruneReason = "unreachable"
	PruneLowPriority   PruneReason = "low_priority"
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

// ExecutionPath represents a complete path from source to sink
type ExecutionPath struct {
	ID          string       `json:"id"`
	Nodes       []*PathNode  `json:"nodes"`

	// Source and sink info
	SourceNode  *PathNode    `json:"source_node"`
	SinkNode    *PathNode    `json:"sink_node"`

	// Path characteristics
	Depth       int          `json:"depth"`        // Function call depth
	Length      int          `json:"length"`       // Number of nodes
	HasSanitizer bool        `json:"has_sanitizer"`
	HasValidator bool        `json:"has_validator"`
	HasAuthCheck bool        `json:"has_auth_check"`

	// Scoring
	Priority    float64      `json:"priority"`     // Higher = more likely exploitable
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

	// Sanitizer/validator tracking
	sanitizers    map[string]bool
	validators    map[string]bool
	authChecks    map[string]bool

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
	IsSanitizer  bool
	IsValidator  bool
	IsSink       bool
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
		sanitizers:    make(map[string]bool),
		validators:    make(map[string]bool),
		authChecks:    make(map[string]bool),
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

	if info.IsSanitizer {
		pe.sanitizers[info.Name] = true
	}
	if info.IsValidator {
		pe.validators[info.Name] = true
	}
}

// AddSanitizer marks a function as a sanitizer
func (pe *PathExpander) AddSanitizer(funcName string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.sanitizers[funcName] = true
}

// AddValidator marks a function as a validator
func (pe *PathExpander) AddValidator(funcName string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.validators[funcName] = true
}

// AddAuthCheck marks a function as an authentication check
func (pe *PathExpander) AddAuthCheck(funcName string) {
	pe.mu.Lock()
	defer pe.mu.Unlock()
	pe.authChecks[funcName] = true
}

// ExpandPaths finds all paths from source to sink, expanding through calls
func (pe *PathExpander) ExpandPaths(source, sink *PathNode) []*ExecutionPath {
	pe.mu.Lock()
	pe.pathsExplored = 0
	pe.pathsPruned = 0
	pe.pathsFound = 0
	pe.mu.Unlock()

	var paths []*ExecutionPath
	visited := make(map[string]bool)
	currentPath := make([]*PathNode, 0)

	pe.expandDFS(source, sink, currentPath, visited, 0, &paths)

	// Sort by priority
	pe.sortByPriority(paths)

	return paths
}

// expandDFS performs depth-first expansion of paths
func (pe *PathExpander) expandDFS(
	current *PathNode,
	sink *PathNode,
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

	// Check if we reached the sink
	if pe.isMatch(current, sink) {
		path := pe.createPath(currentPath)
		pe.analyzePath(path)

		// Apply pruning
		if pe.enablePruning && !path.Feasible {
			pe.mu.Lock()
			pe.pathsPruned++
			pe.mu.Unlock()
			return
		}

		*paths = append(*paths, path)
		pe.mu.Lock()
		pe.pathsFound++
		pe.mu.Unlock()
		return
	}

	// Get next nodes
	nextNodes := pe.getNextNodes(current, sink, depth)

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

		pe.expandDFS(next, sink, currentPath, visited, newDepth, paths)
	}
}

// getNextNodes returns possible next nodes from current position
func (pe *PathExpander) getNextNodes(current *PathNode, sink *PathNode, depth int) []*PathNode {
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

	// If at a sink-containing function, add sink node
	if current.FilePath == sink.FilePath && current.FunctionName == sink.FunctionName {
		if current.Line < sink.Line {
			nodes = append(nodes, sink)
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
		path.SinkNode = pathNodes[len(pathNodes)-1]
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

// analyzePath analyzes path for sanitizers, validators, and feasibility
func (pe *PathExpander) analyzePath(path *ExecutionPath) {
	pe.mu.RLock()
	defer pe.mu.RUnlock()

	for _, node := range path.Nodes {
		if node == nil {
			continue
		}

		// Check for sanitizers
		if pe.sanitizers[node.Name] || pe.sanitizers[node.FunctionName] {
			path.HasSanitizer = true
		}

		// Check for validators
		if pe.validators[node.Name] || pe.validators[node.FunctionName] {
			path.HasValidator = true
		}

		// Check for auth checks
		if pe.authChecks[node.Name] || pe.authChecks[node.FunctionName] {
			path.HasAuthCheck = true
		}

		// Collect conditions
		if node.Type == PathNodeCondition && node.Condition != "" {
			path.Conditions = append(path.Conditions, node.Condition)
		}
	}

	// Calculate priority score
	path.Priority = pe.calculatePriority(path)

	// Determine feasibility (simple heuristic)
	if path.HasSanitizer && pe.enablePruning {
		path.Feasible = false
		path.PruneReason = PruneSanitized
	}
}

// calculatePriority calculates a priority score for a path
func (pe *PathExpander) calculatePriority(path *ExecutionPath) float64 {
	score := 100.0

	// Shorter paths are generally more exploitable
	if path.Length > 0 {
		score -= float64(path.Length) * 2
	}

	// Deeper call stacks are more complex
	if path.Depth > 0 {
		score -= float64(path.Depth) * 5
	}

	// Sanitizers reduce exploitability significantly
	if path.HasSanitizer {
		score -= 50
	}

	// Validators reduce exploitability somewhat
	if path.HasValidator {
		score -= 30
	}

	// Auth checks may prevent exploitation
	if path.HasAuthCheck {
		score -= 20
	}

	// More conditions means more constraints
	score -= float64(len(path.Conditions)) * 3

	if score < 0 {
		score = 0
	}

	return score
}

// sortByPriority sorts paths by priority (highest first)
func (pe *PathExpander) sortByPriority(paths []*ExecutionPath) {
	// Use heap for efficient sorting
	h := &pathHeap{paths: paths}
	heap.Init(h)

	sorted := make([]*ExecutionPath, 0, len(paths))
	for h.Len() > 0 {
		sorted = append(sorted, heap.Pop(h).(*ExecutionPath))
	}

	// Copy back
	copy(paths, sorted)
}

// pathHeap implements heap.Interface for priority sorting
type pathHeap struct {
	paths []*ExecutionPath
}

func (h *pathHeap) Len() int           { return len(h.paths) }
func (h *pathHeap) Less(i, j int) bool { return h.paths[i].Priority > h.paths[j].Priority } // Higher priority first
func (h *pathHeap) Swap(i, j int)      { h.paths[i], h.paths[j] = h.paths[j], h.paths[i] }
func (h *pathHeap) Push(x interface{}) { h.paths = append(h.paths, x.(*ExecutionPath)) }
func (h *pathHeap) Pop() interface{} {
	old := h.paths
	n := len(old)
	x := old[n-1]
	h.paths = old[0 : n-1]
	return x
}

// PrunePath checks if a path should be pruned
func (pe *PathExpander) PrunePath(path *ExecutionPath) (bool, PruneReason) {
	if path.Length > pe.maxPathLength {
		return true, PruneMaxDepth
	}

	if path.HasSanitizer {
		return true, PruneSanitized
	}

	// Add more pruning rules as needed

	return false, PruneNone
}

// FilterByPriority filters paths to keep only high-priority ones
func (pe *PathExpander) FilterByPriority(paths []*ExecutionPath, minPriority float64) []*ExecutionPath {
	var filtered []*ExecutionPath
	for _, p := range paths {
		if p.Priority >= minPriority {
			filtered = append(filtered, p)
		}
	}
	return filtered
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
		"sanitizers":     len(pe.sanitizers),
		"validators":     len(pe.validators),
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
