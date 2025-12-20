// Package callgraph provides sophisticated call graph management with
// distance computation, inspired by ATLANTIS's approach to vulnerability analysis.
// Unlike simple caller->callee mappings, this package provides:
// - Bidirectional graph traversal (caller->callee and callee->caller)
// - Distance computation from entry points (HTTP handlers, main, etc.)
// - Sink reachability analysis
// - Shortest path computation between functions
// - LRU caching for expensive computations
package callgraph

import (
	"container/list"
	"math"
	"sync"
)

// NodeType represents the type of a call graph node
type NodeType int

const (
	NodeTypeRegular NodeType = iota
	NodeTypeEntryPoint       // HTTP handlers, main functions, CLI entry points
	NodeTypeSink             // Security-sensitive functions (SQL, exec, etc.)
	NodeTypeSource           // Input source functions
)

// Node represents a function in the call graph
type Node struct {
	ID           string   `json:"id"`            // Unique identifier (usually "filepath:funcname")
	Name         string   `json:"name"`          // Function name
	FilePath     string   `json:"file_path"`
	Line         int      `json:"line"`
	Type         NodeType `json:"type"`
	Language     string   `json:"language"`
	ClassName    string   `json:"class_name,omitempty"`    // For methods
	IsPublic     bool     `json:"is_public"`               // Visibility
	Signature    string   `json:"signature,omitempty"`     // Full function signature

	// Distance metrics (computed lazily)
	DistanceFromEntry int `json:"distance_from_entry"` // -1 if unreachable
	DistanceToSink    int `json:"distance_to_sink"`    // -1 if no path to sink
}

// Edge represents a call relationship
type Edge struct {
	CallerID string `json:"caller_id"`
	CalleeID string `json:"callee_id"`
	Line     int    `json:"line"`     // Line of the call site
	Column   int    `json:"column"`
	FilePath string `json:"file_path"`

	// Call metadata
	ArgumentCount int      `json:"argument_count"`
	IsConditional bool     `json:"is_conditional"` // Inside if/switch
	InLoop        bool     `json:"in_loop"`        // Inside loop
	BranchDepth   int      `json:"branch_depth"`   // Nesting level
}

// Manager provides sophisticated call graph operations
type Manager struct {
	mu sync.RWMutex

	// Core graph data
	nodes map[string]*Node // nodeID -> Node

	// Bidirectional edges for efficient traversal
	outEdges map[string][]*Edge // caller -> edges to callees
	inEdges  map[string][]*Edge // callee -> edges from callers

	// Categorized nodes for quick lookup
	entryPoints map[string]*Node
	sinks       map[string]*Node
	sources     map[string]*Node

	// Precomputed distances (lazily populated)
	distanceCache     map[string]map[string]int // from -> to -> distance
	distanceCacheLRU  *list.List
	distanceCacheMap  map[string]*list.Element
	maxDistanceCache  int

	// Configuration
	maxPathLength int
}

// NewManager creates a new call graph manager
func NewManager() *Manager {
	return &Manager{
		nodes:            make(map[string]*Node),
		outEdges:         make(map[string][]*Edge),
		inEdges:          make(map[string][]*Edge),
		entryPoints:      make(map[string]*Node),
		sinks:            make(map[string]*Node),
		sources:          make(map[string]*Node),
		distanceCache:    make(map[string]map[string]int),
		distanceCacheLRU: list.New(),
		distanceCacheMap: make(map[string]*list.Element),
		maxDistanceCache: 10000,
		maxPathLength:    50, // Prevent infinite loops in complex graphs
	}
}

// AddNode adds a function node to the graph
func (m *Manager) AddNode(node *Node) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if node.DistanceFromEntry == 0 {
		node.DistanceFromEntry = -1 // Mark as not computed
	}
	if node.DistanceToSink == 0 {
		node.DistanceToSink = -1
	}

	m.nodes[node.ID] = node

	// Categorize by type
	switch node.Type {
	case NodeTypeEntryPoint:
		m.entryPoints[node.ID] = node
		node.DistanceFromEntry = 0 // Entry points have distance 0
	case NodeTypeSink:
		m.sinks[node.ID] = node
		node.DistanceToSink = 0 // Sinks have distance 0 to themselves
	case NodeTypeSource:
		m.sources[node.ID] = node
	}
}

// AddEdge adds a call edge to the graph
func (m *Manager) AddEdge(edge *Edge) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Add to outgoing edges (caller -> callee)
	m.outEdges[edge.CallerID] = append(m.outEdges[edge.CallerID], edge)

	// Add to incoming edges (callee <- caller)
	m.inEdges[edge.CalleeID] = append(m.inEdges[edge.CalleeID], edge)

	// Invalidate relevant distance cache entries
	m.invalidateCacheFor(edge.CallerID)
	m.invalidateCacheFor(edge.CalleeID)
}

// GetNode retrieves a node by ID
func (m *Manager) GetNode(id string) *Node {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.nodes[id]
}

// GetCallees returns all functions called by the given function
func (m *Manager) GetCallees(callerID string) []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	edges := m.outEdges[callerID]
	result := make([]*Node, 0, len(edges))
	seen := make(map[string]bool)

	for _, edge := range edges {
		if !seen[edge.CalleeID] {
			if node := m.nodes[edge.CalleeID]; node != nil {
				result = append(result, node)
				seen[edge.CalleeID] = true
			}
		}
	}
	return result
}

// GetCallers returns all functions that call the given function
func (m *Manager) GetCallers(calleeID string) []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	edges := m.inEdges[calleeID]
	result := make([]*Node, 0, len(edges))
	seen := make(map[string]bool)

	for _, edge := range edges {
		if !seen[edge.CallerID] {
			if node := m.nodes[edge.CallerID]; node != nil {
				result = append(result, node)
				seen[edge.CallerID] = true
			}
		}
	}
	return result
}

// ComputeDistanceFromEntryPoints uses BFS to compute distances from all entry points
// This is the ATLANTIS-inspired approach for prioritizing analysis
func (m *Manager) ComputeDistanceFromEntryPoints() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset all distances
	for _, node := range m.nodes {
		if node.Type == NodeTypeEntryPoint {
			node.DistanceFromEntry = 0
		} else {
			node.DistanceFromEntry = -1
		}
	}

	// BFS from all entry points simultaneously (multi-source BFS)
	queue := list.New()
	visited := make(map[string]bool)

	// Initialize queue with entry points
	for id, node := range m.entryPoints {
		queue.PushBack(id)
		visited[id] = true
		node.DistanceFromEntry = 0
	}

	// BFS traversal
	for queue.Len() > 0 {
		elem := queue.Front()
		queue.Remove(elem)
		currentID := elem.Value.(string)
		currentNode := m.nodes[currentID]
		if currentNode == nil {
			continue
		}

		currentDist := currentNode.DistanceFromEntry

		// Visit all callees
		for _, edge := range m.outEdges[currentID] {
			calleeID := edge.CalleeID
			if !visited[calleeID] {
				visited[calleeID] = true
				if callee := m.nodes[calleeID]; callee != nil {
					callee.DistanceFromEntry = currentDist + 1
					queue.PushBack(calleeID)
				}
			}
		}
	}
}

// ComputeDistanceToSinks uses reverse BFS to compute distances to nearest sink
// This helps prioritize functions that are closer to security-sensitive operations
func (m *Manager) ComputeDistanceToSinks() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Reset all sink distances
	for _, node := range m.nodes {
		if node.Type == NodeTypeSink {
			node.DistanceToSink = 0
		} else {
			node.DistanceToSink = -1
		}
	}

	// Reverse BFS from all sinks (using incoming edges to traverse backwards)
	queue := list.New()
	visited := make(map[string]bool)

	// Initialize queue with sinks
	for id, node := range m.sinks {
		queue.PushBack(id)
		visited[id] = true
		node.DistanceToSink = 0
	}

	// BFS traversal (going backwards via callers)
	for queue.Len() > 0 {
		elem := queue.Front()
		queue.Remove(elem)
		currentID := elem.Value.(string)
		currentNode := m.nodes[currentID]
		if currentNode == nil {
			continue
		}

		currentDist := currentNode.DistanceToSink

		// Visit all callers (reverse direction)
		for _, edge := range m.inEdges[currentID] {
			callerID := edge.CallerID
			if !visited[callerID] {
				visited[callerID] = true
				if caller := m.nodes[callerID]; caller != nil {
					caller.DistanceToSink = currentDist + 1
					queue.PushBack(callerID)
				}
			}
		}
	}
}

// GetShortestPath finds the shortest call path between two functions
// Returns nil if no path exists
func (m *Manager) GetShortestPath(fromID, toID string) []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if fromID == toID {
		if node := m.nodes[fromID]; node != nil {
			return []*Node{node}
		}
		return nil
	}

	// BFS to find shortest path
	queue := list.New()
	visited := make(map[string]bool)
	parent := make(map[string]string) // child -> parent for path reconstruction

	queue.PushBack(fromID)
	visited[fromID] = true

	found := false
	for queue.Len() > 0 && !found {
		elem := queue.Front()
		queue.Remove(elem)
		currentID := elem.Value.(string)

		for _, edge := range m.outEdges[currentID] {
			calleeID := edge.CalleeID
			if !visited[calleeID] {
				visited[calleeID] = true
				parent[calleeID] = currentID
				if calleeID == toID {
					found = true
					break
				}
				queue.PushBack(calleeID)
			}
		}
	}

	if !found {
		return nil
	}

	// Reconstruct path
	path := make([]*Node, 0)
	current := toID
	for current != "" {
		if node := m.nodes[current]; node != nil {
			path = append([]*Node{node}, path...)
		}
		current = parent[current]
	}

	return path
}

// GetDistance returns the shortest distance between two nodes
// Uses caching for performance
func (m *Manager) GetDistance(fromID, toID string) int {
	if fromID == toID {
		return 0
	}

	// Check cache first
	m.mu.RLock()
	if fromCache, ok := m.distanceCache[fromID]; ok {
		if dist, ok := fromCache[toID]; ok {
			m.mu.RUnlock()
			return dist
		}
	}
	m.mu.RUnlock()

	// Compute via BFS
	path := m.GetShortestPath(fromID, toID)
	dist := -1
	if path != nil {
		dist = len(path) - 1
	}

	// Cache the result
	m.mu.Lock()
	m.cacheDistance(fromID, toID, dist)
	m.mu.Unlock()

	return dist
}

// cacheDistance stores a distance in the LRU cache (must hold write lock)
func (m *Manager) cacheDistance(fromID, toID string, distance int) {
	key := fromID + ":" + toID

	// Check if already in cache
	if elem, ok := m.distanceCacheMap[key]; ok {
		m.distanceCacheLRU.MoveToFront(elem)
		return
	}

	// Evict if necessary
	for m.distanceCacheLRU.Len() >= m.maxDistanceCache {
		oldest := m.distanceCacheLRU.Back()
		if oldest != nil {
			oldKey := oldest.Value.(string)
			m.distanceCacheLRU.Remove(oldest)
			delete(m.distanceCacheMap, oldKey)
			// Parse and delete from distanceCache
			for i := 0; i < len(oldKey); i++ {
				if oldKey[i] == ':' {
					delete(m.distanceCache[oldKey[:i]], oldKey[i+1:])
					break
				}
			}
		}
	}

	// Add to cache
	if m.distanceCache[fromID] == nil {
		m.distanceCache[fromID] = make(map[string]int)
	}
	m.distanceCache[fromID][toID] = distance

	elem := m.distanceCacheLRU.PushFront(key)
	m.distanceCacheMap[key] = elem
}

// invalidateCacheFor removes cache entries involving the given node
func (m *Manager) invalidateCacheFor(nodeID string) {
	// Remove entries where this node is the source
	if fromCache, ok := m.distanceCache[nodeID]; ok {
		for toID := range fromCache {
			key := nodeID + ":" + toID
			if elem, ok := m.distanceCacheMap[key]; ok {
				m.distanceCacheLRU.Remove(elem)
				delete(m.distanceCacheMap, key)
			}
		}
		delete(m.distanceCache, nodeID)
	}

	// We don't remove entries where this node is the destination
	// as that would require scanning all entries (expensive)
	// Those will be evicted naturally via LRU
}

// GetEntryPoints returns all entry point nodes
func (m *Manager) GetEntryPoints() []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Node, 0, len(m.entryPoints))
	for _, node := range m.entryPoints {
		result = append(result, node)
	}
	return result
}

// GetSinks returns all sink nodes
func (m *Manager) GetSinks() []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Node, 0, len(m.sinks))
	for _, node := range m.sinks {
		result = append(result, node)
	}
	return result
}

// GetReachableSinks returns all sinks reachable from the given node
func (m *Manager) GetReachableSinks(fromID string) []*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]*Node, 0)
	visited := make(map[string]bool)

	// DFS to find all reachable nodes
	var dfs func(id string, depth int)
	dfs = func(id string, depth int) {
		if depth > m.maxPathLength || visited[id] {
			return
		}
		visited[id] = true

		// Check if this is a sink
		if sink, ok := m.sinks[id]; ok {
			result = append(result, sink)
		}

		// Continue to callees
		for _, edge := range m.outEdges[id] {
			dfs(edge.CalleeID, depth+1)
		}
	}

	dfs(fromID, 0)
	return result
}

// GetAllPathsToSinks returns all paths from a node to any sink
// Limited by maxPathLength to prevent explosion
func (m *Manager) GetAllPathsToSinks(fromID string, maxPaths int) [][]*Node {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var paths [][]*Node
	currentPath := make([]*Node, 0)
	visited := make(map[string]bool)

	var dfs func(id string)
	dfs = func(id string) {
		if len(paths) >= maxPaths || len(currentPath) > m.maxPathLength || visited[id] {
			return
		}

		visited[id] = true
		node := m.nodes[id]
		if node == nil {
			visited[id] = false
			return
		}

		currentPath = append(currentPath, node)

		// Check if we reached a sink
		if node.Type == NodeTypeSink {
			// Save a copy of the path
			pathCopy := make([]*Node, len(currentPath))
			copy(pathCopy, currentPath)
			paths = append(paths, pathCopy)
		} else {
			// Continue to callees
			for _, edge := range m.outEdges[id] {
				dfs(edge.CalleeID)
			}
		}

		currentPath = currentPath[:len(currentPath)-1]
		visited[id] = false
	}

	dfs(fromID)
	return paths
}

// PriorityScore computes a priority score for analyzing a function
// Higher score = higher priority for vulnerability analysis
// Based on ATLANTIS's approach of prioritizing functions closer to both entry and sink
func (m *Manager) PriorityScore(nodeID string) float64 {
	m.mu.RLock()
	node := m.nodes[nodeID]
	m.mu.RUnlock()

	if node == nil {
		return 0
	}

	// Ensure distances are computed
	if node.DistanceFromEntry == -1 {
		m.ComputeDistanceFromEntryPoints()
	}
	if node.DistanceToSink == -1 {
		m.ComputeDistanceToSinks()
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	distEntry := node.DistanceFromEntry
	distSink := node.DistanceToSink

	// Unreachable from entry or can't reach sink = low priority
	if distEntry < 0 || distSink < 0 {
		return 0
	}

	// Score formula: prioritize functions that are:
	// 1. Reachable from entry points (lower distance = better)
	// 2. Can reach sinks (lower distance = better)
	// 3. Have shorter total path (entry -> function -> sink)

	totalPath := distEntry + distSink
	if totalPath == 0 {
		return 100.0 // Entry point that is also a sink (rare but maximum priority)
	}

	// Use inverse of path length, scaled
	// Also boost functions that are entry points or closer to sinks
	score := 100.0 / (1.0 + float64(totalPath))

	// Bonus for being close to sink (more likely to be exploitable)
	if distSink <= 2 {
		score *= 1.5
	}

	// Bonus for being entry point
	if distEntry == 0 {
		score *= 1.3
	}

	return math.Round(score*100) / 100
}

// Stats returns statistics about the call graph
func (m *Manager) Stats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	edgeCount := 0
	for _, edges := range m.outEdges {
		edgeCount += len(edges)
	}

	reachableFromEntry := 0
	canReachSink := 0
	for _, node := range m.nodes {
		if node.DistanceFromEntry >= 0 {
			reachableFromEntry++
		}
		if node.DistanceToSink >= 0 {
			canReachSink++
		}
	}

	return map[string]int{
		"total_nodes":          len(m.nodes),
		"total_edges":          edgeCount,
		"entry_points":         len(m.entryPoints),
		"sinks":                len(m.sinks),
		"sources":              len(m.sources),
		"reachable_from_entry": reachableFromEntry,
		"can_reach_sink":       canReachSink,
		"cache_size":           m.distanceCacheLRU.Len(),
	}
}

// MakeNodeID creates a standard node ID from file path and function name
func MakeNodeID(filePath, funcName string) string {
	return filePath + ":" + funcName
}

// MakeMethodID creates a node ID for a class method
func MakeMethodID(filePath, className, methodName string) string {
	return filePath + ":" + className + "::" + methodName
}
