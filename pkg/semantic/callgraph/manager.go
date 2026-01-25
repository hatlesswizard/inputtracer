// Package callgraph provides sophisticated call graph management with
// distance computation for input flow analysis.
// Unlike simple caller->callee mappings, this package provides:
// - Bidirectional graph traversal (caller->callee and callee->caller)
// - Distance computation from entry points (HTTP handlers, main, etc.)
// - Shortest path computation between functions
// - LRU caching for expensive computations
package callgraph

import (
	"container/list"
	"sync"
)

// NodeType represents the type of a call graph node
type NodeType int

const (
	NodeTypeRegular NodeType = iota
	NodeTypeEntryPoint       // HTTP handlers, main functions, CLI entry points
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

	m.nodes[node.ID] = node

	// Categorize by type
	switch node.Type {
	case NodeTypeEntryPoint:
		m.entryPoints[node.ID] = node
		node.DistanceFromEntry = 0 // Entry points have distance 0
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

// Stats returns statistics about the call graph
func (m *Manager) Stats() map[string]int {
	m.mu.RLock()
	defer m.mu.RUnlock()

	edgeCount := 0
	for _, edges := range m.outEdges {
		edgeCount += len(edges)
	}

	reachableFromEntry := 0
	for _, node := range m.nodes {
		if node.DistanceFromEntry >= 0 {
			reachableFromEntry++
		}
	}

	return map[string]int{
		"total_nodes":          len(m.nodes),
		"total_edges":          edgeCount,
		"entry_points":         len(m.entryPoints),
		"sources":              len(m.sources),
		"reachable_from_entry": reachableFromEntry,
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
