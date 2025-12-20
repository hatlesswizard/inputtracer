package callgraph

import (
	"testing"
)

func TestManager_BasicOperations(t *testing.T) {
	m := NewManager()

	// Add nodes
	m.AddNode(&Node{
		ID:       "main.go:main",
		Name:     "main",
		FilePath: "main.go",
		Type:     NodeTypeEntryPoint,
	})

	m.AddNode(&Node{
		ID:       "handler.go:handleRequest",
		Name:     "handleRequest",
		FilePath: "handler.go",
		Type:     NodeTypeRegular,
	})

	m.AddNode(&Node{
		ID:       "db.go:query",
		Name:     "query",
		FilePath: "db.go",
		Type:     NodeTypeSink, // SQL sink
	})

	// Add edges
	m.AddEdge(&Edge{
		CallerID: "main.go:main",
		CalleeID: "handler.go:handleRequest",
		Line:     10,
	})

	m.AddEdge(&Edge{
		CallerID: "handler.go:handleRequest",
		CalleeID: "db.go:query",
		Line:     25,
	})

	// Test GetNode
	node := m.GetNode("main.go:main")
	if node == nil {
		t.Fatal("expected to find main node")
	}
	if node.Name != "main" {
		t.Errorf("expected name 'main', got '%s'", node.Name)
	}

	// Test GetCallees
	callees := m.GetCallees("main.go:main")
	if len(callees) != 1 {
		t.Fatalf("expected 1 callee, got %d", len(callees))
	}
	if callees[0].Name != "handleRequest" {
		t.Errorf("expected callee 'handleRequest', got '%s'", callees[0].Name)
	}

	// Test GetCallers
	callers := m.GetCallers("db.go:query")
	if len(callers) != 1 {
		t.Fatalf("expected 1 caller, got %d", len(callers))
	}
	if callers[0].Name != "handleRequest" {
		t.Errorf("expected caller 'handleRequest', got '%s'", callers[0].Name)
	}
}

func TestManager_DistanceComputation(t *testing.T) {
	m := NewManager()

	// Create a simple call chain:
	// main (entry) -> process -> validate -> execute -> sqlQuery (sink)
	nodes := []struct {
		id   string
		name string
		typ  NodeType
	}{
		{"a.go:main", "main", NodeTypeEntryPoint},
		{"b.go:process", "process", NodeTypeRegular},
		{"c.go:validate", "validate", NodeTypeRegular},
		{"d.go:execute", "execute", NodeTypeRegular},
		{"e.go:sqlQuery", "sqlQuery", NodeTypeSink},
	}

	for _, n := range nodes {
		m.AddNode(&Node{
			ID:       n.id,
			Name:     n.name,
			FilePath: n.id[:4], // extract file
			Type:     n.typ,
		})
	}

	// Create chain
	m.AddEdge(&Edge{CallerID: "a.go:main", CalleeID: "b.go:process"})
	m.AddEdge(&Edge{CallerID: "b.go:process", CalleeID: "c.go:validate"})
	m.AddEdge(&Edge{CallerID: "c.go:validate", CalleeID: "d.go:execute"})
	m.AddEdge(&Edge{CallerID: "d.go:execute", CalleeID: "e.go:sqlQuery"})

	// Compute distances
	m.ComputeDistanceFromEntryPoints()
	m.ComputeDistanceToSinks()

	// Verify distances from entry
	tests := []struct {
		nodeID           string
		expectedFromEntry int
		expectedToSink   int
	}{
		{"a.go:main", 0, 4},
		{"b.go:process", 1, 3},
		{"c.go:validate", 2, 2},
		{"d.go:execute", 3, 1},
		{"e.go:sqlQuery", 4, 0},
	}

	for _, tc := range tests {
		node := m.GetNode(tc.nodeID)
		if node == nil {
			t.Fatalf("node %s not found", tc.nodeID)
		}

		if node.DistanceFromEntry != tc.expectedFromEntry {
			t.Errorf("%s: expected DistanceFromEntry=%d, got %d",
				tc.nodeID, tc.expectedFromEntry, node.DistanceFromEntry)
		}

		if node.DistanceToSink != tc.expectedToSink {
			t.Errorf("%s: expected DistanceToSink=%d, got %d",
				tc.nodeID, tc.expectedToSink, node.DistanceToSink)
		}
	}
}

func TestManager_ShortestPath(t *testing.T) {
	m := NewManager()

	// Create a graph with multiple paths:
	//   main -> A -> B -> sink
	//   main -> C -> sink (shorter path)
	m.AddNode(&Node{ID: "main", Name: "main", Type: NodeTypeEntryPoint})
	m.AddNode(&Node{ID: "A", Name: "A", Type: NodeTypeRegular})
	m.AddNode(&Node{ID: "B", Name: "B", Type: NodeTypeRegular})
	m.AddNode(&Node{ID: "C", Name: "C", Type: NodeTypeRegular})
	m.AddNode(&Node{ID: "sink", Name: "sink", Type: NodeTypeSink})

	m.AddEdge(&Edge{CallerID: "main", CalleeID: "A"})
	m.AddEdge(&Edge{CallerID: "A", CalleeID: "B"})
	m.AddEdge(&Edge{CallerID: "B", CalleeID: "sink"})
	m.AddEdge(&Edge{CallerID: "main", CalleeID: "C"})
	m.AddEdge(&Edge{CallerID: "C", CalleeID: "sink"})

	// Shortest path from main to sink should be: main -> C -> sink
	path := m.GetShortestPath("main", "sink")
	if len(path) != 3 {
		t.Fatalf("expected path length 3, got %d", len(path))
	}

	if path[0].Name != "main" || path[1].Name != "C" || path[2].Name != "sink" {
		t.Errorf("expected path [main, C, sink], got [%s, %s, %s]",
			path[0].Name, path[1].Name, path[2].Name)
	}

	// Test distance
	dist := m.GetDistance("main", "sink")
	if dist != 2 {
		t.Errorf("expected distance 2, got %d", dist)
	}

	// Test caching
	dist2 := m.GetDistance("main", "sink")
	if dist2 != 2 {
		t.Errorf("cached distance should be 2, got %d", dist2)
	}
}

func TestManager_ReachableSinks(t *testing.T) {
	m := NewManager()

	// Create graph:
	// entry -> A -> sink1
	//       -> B -> sink2
	//       -> C (no sink reachable)
	m.AddNode(&Node{ID: "entry", Name: "entry", Type: NodeTypeEntryPoint})
	m.AddNode(&Node{ID: "A", Name: "A", Type: NodeTypeRegular})
	m.AddNode(&Node{ID: "B", Name: "B", Type: NodeTypeRegular})
	m.AddNode(&Node{ID: "C", Name: "C", Type: NodeTypeRegular})
	m.AddNode(&Node{ID: "sink1", Name: "sink1", Type: NodeTypeSink})
	m.AddNode(&Node{ID: "sink2", Name: "sink2", Type: NodeTypeSink})

	m.AddEdge(&Edge{CallerID: "entry", CalleeID: "A"})
	m.AddEdge(&Edge{CallerID: "entry", CalleeID: "B"})
	m.AddEdge(&Edge{CallerID: "entry", CalleeID: "C"})
	m.AddEdge(&Edge{CallerID: "A", CalleeID: "sink1"})
	m.AddEdge(&Edge{CallerID: "B", CalleeID: "sink2"})

	// Get reachable sinks from entry
	sinks := m.GetReachableSinks("entry")
	if len(sinks) != 2 {
		t.Fatalf("expected 2 reachable sinks, got %d", len(sinks))
	}

	// Get reachable sinks from C (should be none)
	sinksFromC := m.GetReachableSinks("C")
	if len(sinksFromC) != 0 {
		t.Errorf("expected 0 reachable sinks from C, got %d", len(sinksFromC))
	}
}

func TestManager_AllPathsToSinks(t *testing.T) {
	m := NewManager()

	// Create graph with multiple paths to sink:
	// entry -> A -> B -> sink
	//       -> C ------>
	m.AddNode(&Node{ID: "entry", Name: "entry", Type: NodeTypeEntryPoint})
	m.AddNode(&Node{ID: "A", Name: "A", Type: NodeTypeRegular})
	m.AddNode(&Node{ID: "B", Name: "B", Type: NodeTypeRegular})
	m.AddNode(&Node{ID: "C", Name: "C", Type: NodeTypeRegular})
	m.AddNode(&Node{ID: "sink", Name: "sink", Type: NodeTypeSink})

	m.AddEdge(&Edge{CallerID: "entry", CalleeID: "A"})
	m.AddEdge(&Edge{CallerID: "A", CalleeID: "B"})
	m.AddEdge(&Edge{CallerID: "B", CalleeID: "sink"})
	m.AddEdge(&Edge{CallerID: "entry", CalleeID: "C"})
	m.AddEdge(&Edge{CallerID: "C", CalleeID: "sink"})

	paths := m.GetAllPathsToSinks("entry", 10)
	if len(paths) != 2 {
		t.Fatalf("expected 2 paths, got %d", len(paths))
	}

	// Verify path lengths
	pathLengths := make(map[int]bool)
	for _, path := range paths {
		pathLengths[len(path)] = true
	}

	if !pathLengths[3] {
		t.Error("expected a path of length 3 (entry -> C -> sink)")
	}
	if !pathLengths[4] {
		t.Error("expected a path of length 4 (entry -> A -> B -> sink)")
	}
}

func TestManager_PriorityScore(t *testing.T) {
	m := NewManager()

	// entry -> middle -> sink
	m.AddNode(&Node{ID: "entry", Name: "entry", Type: NodeTypeEntryPoint})
	m.AddNode(&Node{ID: "middle", Name: "middle", Type: NodeTypeRegular})
	m.AddNode(&Node{ID: "sink", Name: "sink", Type: NodeTypeSink})
	m.AddNode(&Node{ID: "orphan", Name: "orphan", Type: NodeTypeRegular}) // not connected

	m.AddEdge(&Edge{CallerID: "entry", CalleeID: "middle"})
	m.AddEdge(&Edge{CallerID: "middle", CalleeID: "sink"})

	// Compute distances
	m.ComputeDistanceFromEntryPoints()
	m.ComputeDistanceToSinks()

	// Entry has distance 0 from entry, 2 to sink
	// Middle has distance 1 from entry, 1 to sink (shortest total path)
	// Sink has distance 2 from entry, 0 to sink
	// Orphan is unreachable

	scoreEntry := m.PriorityScore("entry")
	scoreMiddle := m.PriorityScore("middle")
	scoreSink := m.PriorityScore("sink")
	scoreOrphan := m.PriorityScore("orphan")

	// Orphan should have score 0 (unreachable)
	if scoreOrphan != 0 {
		t.Errorf("orphan should have score 0, got %f", scoreOrphan)
	}

	// All connected nodes should have positive scores
	if scoreEntry <= 0 {
		t.Errorf("entry should have positive score, got %f", scoreEntry)
	}
	if scoreMiddle <= 0 {
		t.Errorf("middle should have positive score, got %f", scoreMiddle)
	}
	if scoreSink <= 0 {
		t.Errorf("sink should have positive score, got %f", scoreSink)
	}

	// Middle should have high score (close to both entry and sink)
	// The exact values depend on the scoring formula, but middle should be competitive
	t.Logf("Scores: entry=%f, middle=%f, sink=%f, orphan=%f",
		scoreEntry, scoreMiddle, scoreSink, scoreOrphan)
}

func TestManager_Stats(t *testing.T) {
	m := NewManager()

	m.AddNode(&Node{ID: "entry", Name: "entry", Type: NodeTypeEntryPoint})
	m.AddNode(&Node{ID: "middle", Name: "middle", Type: NodeTypeRegular})
	m.AddNode(&Node{ID: "sink", Name: "sink", Type: NodeTypeSink})

	m.AddEdge(&Edge{CallerID: "entry", CalleeID: "middle"})
	m.AddEdge(&Edge{CallerID: "middle", CalleeID: "sink"})

	stats := m.Stats()

	if stats["total_nodes"] != 3 {
		t.Errorf("expected 3 nodes, got %d", stats["total_nodes"])
	}
	if stats["total_edges"] != 2 {
		t.Errorf("expected 2 edges, got %d", stats["total_edges"])
	}
	if stats["entry_points"] != 1 {
		t.Errorf("expected 1 entry point, got %d", stats["entry_points"])
	}
	if stats["sinks"] != 1 {
		t.Errorf("expected 1 sink, got %d", stats["sinks"])
	}
}

func TestMakeNodeID(t *testing.T) {
	id := MakeNodeID("/path/to/file.go", "myFunc")
	expected := "/path/to/file.go:myFunc"
	if id != expected {
		t.Errorf("expected '%s', got '%s'", expected, id)
	}
}

func TestMakeMethodID(t *testing.T) {
	id := MakeMethodID("/path/to/file.php", "MyClass", "myMethod")
	expected := "/path/to/file.php:MyClass::myMethod"
	if id != expected {
		t.Errorf("expected '%s', got '%s'", expected, id)
	}
}
