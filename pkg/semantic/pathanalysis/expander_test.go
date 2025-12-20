package pathanalysis

import (
	"testing"
)

func TestNewPathExpander(t *testing.T) {
	pe := NewPathExpander()
	if pe == nil {
		t.Fatal("expected non-nil PathExpander")
	}

	if pe.maxDepth != 10 {
		t.Errorf("expected default maxDepth=10, got %d", pe.maxDepth)
	}

	if pe.maxPathLength != 50 {
		t.Errorf("expected default maxPathLength=50, got %d", pe.maxPathLength)
	}
}

func TestPathExpander_AddCallEdge(t *testing.T) {
	pe := NewPathExpander()

	pe.AddCallEdge("main.go:main", "handler.go:handleRequest")
	pe.AddCallEdge("handler.go:handleRequest", "db.go:query")

	graph := pe.GetCallGraph()

	if len(graph["main.go:main"]) != 1 {
		t.Errorf("expected 1 callee from main, got %d", len(graph["main.go:main"]))
	}

	reverse := pe.GetReverseCallGraph()
	if len(reverse["db.go:query"]) != 1 {
		t.Errorf("expected 1 caller of query, got %d", len(reverse["db.go:query"]))
	}
}

func TestPathExpander_AddFunction(t *testing.T) {
	pe := NewPathExpander()

	pe.AddFunction(&FunctionInfo{
		Name:        "sanitize",
		FilePath:    "utils.go",
		IsSanitizer: true,
	})

	pe.AddFunction(&FunctionInfo{
		Name:        "validate",
		FilePath:    "utils.go",
		IsValidator: true,
	})

	stats := pe.Stats()
	if stats["functions"] != 2 {
		t.Errorf("expected 2 functions, got %d", stats["functions"])
	}
	if stats["sanitizers"] != 1 {
		t.Errorf("expected 1 sanitizer, got %d", stats["sanitizers"])
	}
	if stats["validators"] != 1 {
		t.Errorf("expected 1 validator, got %d", stats["validators"])
	}
}

func TestPathExpander_ExpandPaths_Simple(t *testing.T) {
	pe := NewPathExpander()

	// Create a simple linear path: source -> process -> sink
	pe.AddFunction(&FunctionInfo{Name: "main", FilePath: "main.go"})
	pe.AddFunction(&FunctionInfo{Name: "process", FilePath: "process.go"})
	pe.AddFunction(&FunctionInfo{Name: "sink", FilePath: "sink.go", IsSink: true})

	pe.AddCallEdge("main.go:main", "process.go:process")
	pe.AddCallEdge("process.go:process", "sink.go:sink")

	source := &PathNode{
		Type:         PathNodeSource,
		Name:         "main",
		FunctionName: "main",
		FilePath:     "main.go",
		Line:         10,
	}

	sink := &PathNode{
		Type:         PathNodeSink,
		Name:         "sink",
		FunctionName: "sink",
		FilePath:     "sink.go",
		Line:         5,
	}

	paths := pe.ExpandPaths(source, sink)

	// Note: Path finding depends on call graph structure
	// This test verifies the expander runs without error
	t.Logf("Found %d paths", len(paths))
	for i, p := range paths {
		t.Logf("  Path %d: length=%d, depth=%d, priority=%.2f, feasible=%v",
			i, p.Length, p.Depth, p.Priority, p.Feasible)
	}
}

func TestPathExpander_SanitizerPruning(t *testing.T) {
	pe := NewPathExpander()
	pe.AddSanitizer("htmlspecialchars")

	path := &ExecutionPath{
		Nodes: []*PathNode{
			{Name: "source", Type: PathNodeSource},
			{Name: "htmlspecialchars", Type: PathNodeCall},
			{Name: "echo", Type: PathNodeSink},
		},
		Feasible: true,
	}

	pe.analyzePath(path)

	if !path.HasSanitizer {
		t.Error("path should have sanitizer flag set")
	}

	if path.Feasible {
		t.Error("path with sanitizer should be marked infeasible")
	}

	if path.PruneReason != PruneSanitized {
		t.Errorf("expected PruneSanitized, got %s", path.PruneReason)
	}
}

func TestPathExpander_Priority(t *testing.T) {
	pe := NewPathExpander()
	pe.AddSanitizer("sanitize")
	pe.AddValidator("validate")
	pe.AddAuthCheck("isAuthenticated")

	tests := []struct {
		name     string
		path     *ExecutionPath
		minScore float64
		maxScore float64
	}{
		{
			name: "short_no_guards",
			path: &ExecutionPath{
				Nodes:  make([]*PathNode, 3),
				Length: 3,
				Depth:  1,
			},
			minScore: 80,
			maxScore: 100,
		},
		{
			name: "long_deep",
			path: &ExecutionPath{
				Nodes:  make([]*PathNode, 15),
				Length: 15,
				Depth:  5,
			},
			minScore: 30,
			maxScore: 70,
		},
		{
			name: "with_sanitizer",
			path: &ExecutionPath{
				Nodes: []*PathNode{
					{Name: "source"},
					{Name: "sanitize"},
					{Name: "sink"},
				},
				Length: 3,
			},
			minScore: 0,
			maxScore: 60,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pe.analyzePath(tc.path)

			if tc.path.Priority < tc.minScore || tc.path.Priority > tc.maxScore {
				t.Errorf("expected priority in [%.0f, %.0f], got %.2f",
					tc.minScore, tc.maxScore, tc.path.Priority)
			}
		})
	}
}

func TestPathExpander_FilterByPriority(t *testing.T) {
	pe := NewPathExpander()

	paths := []*ExecutionPath{
		{Priority: 90},
		{Priority: 50},
		{Priority: 30},
		{Priority: 70},
		{Priority: 10},
	}

	filtered := pe.FilterByPriority(paths, 50)
	if len(filtered) != 3 {
		t.Errorf("expected 3 paths with priority >= 50, got %d", len(filtered))
	}

	for _, p := range filtered {
		if p.Priority < 50 {
			t.Errorf("filtered path has priority %.2f < 50", p.Priority)
		}
	}
}

func TestPathExpander_FilterFeasible(t *testing.T) {
	pe := NewPathExpander()

	paths := []*ExecutionPath{
		{Feasible: true},
		{Feasible: false},
		{Feasible: true},
		{Feasible: false},
		{Feasible: true},
	}

	filtered := pe.FilterFeasible(paths)
	if len(filtered) != 3 {
		t.Errorf("expected 3 feasible paths, got %d", len(filtered))
	}

	for _, p := range filtered {
		if !p.Feasible {
			t.Error("filtered path should be feasible")
		}
	}
}

func TestPathExpander_PrunePath(t *testing.T) {
	pe := NewPathExpander()
	pe.SetMaxDepth(5)

	tests := []struct {
		name           string
		path           *ExecutionPath
		shouldPrune    bool
		expectedReason PruneReason
	}{
		{
			name: "normal_path",
			path: &ExecutionPath{
				Length:       10,
				HasSanitizer: false,
			},
			shouldPrune: false,
		},
		{
			name: "too_long",
			path: &ExecutionPath{
				Length: 100,
			},
			shouldPrune:    true,
			expectedReason: PruneMaxDepth,
		},
		{
			name: "sanitized",
			path: &ExecutionPath{
				Length:       5,
				HasSanitizer: true,
			},
			shouldPrune:    true,
			expectedReason: PruneSanitized,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			pruned, reason := pe.PrunePath(tc.path)

			if pruned != tc.shouldPrune {
				t.Errorf("expected pruned=%v, got %v", tc.shouldPrune, pruned)
			}

			if tc.shouldPrune && reason != tc.expectedReason {
				t.Errorf("expected reason %s, got %s", tc.expectedReason, reason)
			}
		})
	}
}

func TestPathExpander_Stats(t *testing.T) {
	pe := NewPathExpander()

	pe.AddFunction(&FunctionInfo{Name: "f1", FilePath: "a.go"})
	pe.AddFunction(&FunctionInfo{Name: "f2", FilePath: "b.go"})
	pe.AddCallEdge("a.go:f1", "b.go:f2")
	pe.AddSanitizer("escape")
	pe.AddValidator("validate")

	stats := pe.Stats()

	if stats["functions"] != 2 {
		t.Errorf("expected 2 functions, got %d", stats["functions"])
	}
	if stats["call_edges"] != 1 {
		t.Errorf("expected 1 call edge, got %d", stats["call_edges"])
	}
	if stats["sanitizers"] != 1 {
		t.Errorf("expected 1 sanitizer, got %d", stats["sanitizers"])
	}
	if stats["validators"] != 1 {
		t.Errorf("expected 1 validator, got %d", stats["validators"])
	}
}

func TestPathNodeTypes(t *testing.T) {
	types := []PathNodeType{
		PathNodeSource,
		PathNodeSink,
		PathNodeCall,
		PathNodeReturn,
		PathNodeAssign,
		PathNodeCondition,
		PathNodeTransform,
	}

	for _, pt := range types {
		if string(pt) == "" {
			t.Error("path node type should not be empty")
		}
	}
}

func TestPruneReasons(t *testing.T) {
	reasons := []PruneReason{
		PruneMaxDepth,
		PruneCycle,
		PruneInfeasible,
		PruneSanitized,
		PruneTypeCoercion,
		PruneDead,
		PruneUnreachable,
		PruneLowPriority,
	}

	for _, pr := range reasons {
		if string(pr) == "" {
			t.Error("prune reason should not be empty")
		}
	}
}

func TestExecutionPath_Conditions(t *testing.T) {
	pe := NewPathExpander()

	path := &ExecutionPath{
		Nodes: []*PathNode{
			{Type: PathNodeSource, Name: "source"},
			{Type: PathNodeCondition, Name: "if1", Condition: "isset($id)"},
			{Type: PathNodeCondition, Name: "if2", Condition: "is_numeric($id)"},
			{Type: PathNodeSink, Name: "query"},
		},
		Feasible: true,
	}

	pe.analyzePath(path)

	if len(path.Conditions) != 2 {
		t.Errorf("expected 2 conditions, got %d", len(path.Conditions))
	}

	expectedConditions := []string{"isset($id)", "is_numeric($id)"}
	for i, expected := range expectedConditions {
		if i < len(path.Conditions) && path.Conditions[i] != expected {
			t.Errorf("expected condition '%s', got '%s'", expected, path.Conditions[i])
		}
	}
}
