package tracer

import (
	"slices"
	"testing"
)

func TestNewInterproceduralAnalyzer(t *testing.T) {
	state := NewFullAnalysisState()
	ipa := NewInterproceduralAnalyzer(state, 5, nil)
	if ipa == nil {
		t.Fatal("NewInterproceduralAnalyzer returned nil")
	}
	if ipa.maxDepth != 5 {
		t.Errorf("maxDepth = %d, want 5", ipa.maxDepth)
	}
	if ipa.currentDepth != 0 {
		t.Errorf("currentDepth = %d, want 0", ipa.currentDepth)
	}
}

func TestInterproceduralAnalyzer_GetCallGraph_StartsEmpty(t *testing.T) {
	ipa := NewInterproceduralAnalyzer(NewFullAnalysisState(), 3, nil)
	graph := ipa.GetCallGraph()
	if len(graph) != 0 {
		t.Errorf("initial call graph len = %d, want 0", len(graph))
	}
}

func TestInterproceduralAnalyzer_GetCallGraph_ReturnsCopy(t *testing.T) {
	ipa := NewInterproceduralAnalyzer(NewFullAnalysisState(), 3, nil)
	ipa.graph.mu.Lock()
	ipa.graph.edges["foo"] = []string{"bar"}
	ipa.graph.mu.Unlock()

	graph := ipa.GetCallGraph()
	graph["foo"] = append(graph["foo"], "baz") // mutate the returned copy

	// Original should be unaffected
	ipa.graph.mu.RLock()
	if len(ipa.graph.edges["foo"]) != 1 {
		t.Error("mutating GetCallGraph result should not affect internal state")
	}
	ipa.graph.mu.RUnlock()
}

func TestInterproceduralAnalyzer_GetFunctionSummary_UnknownReturnsNil(t *testing.T) {
	ipa := NewInterproceduralAnalyzer(NewFullAnalysisState(), 3, nil)
	if ipa.GetFunctionSummary("noSuchFunc") != nil {
		t.Error("GetFunctionSummary for unknown function should return nil")
	}
}

func TestInterproceduralAnalyzer_GetAllSummaries_ReturnsCopy(t *testing.T) {
	ipa := NewInterproceduralAnalyzer(NewFullAnalysisState(), 3, nil)
	ipa.mu.Lock()
	ipa.summaries["myFunc"] = &FunctionSummary{Name: "myFunc"}
	ipa.mu.Unlock()

	summaries := ipa.GetAllSummaries()
	if len(summaries) != 1 {
		t.Fatalf("GetAllSummaries len = %d, want 1", len(summaries))
	}

	delete(summaries, "myFunc") // mutate the copy
	if ipa.GetFunctionSummary("myFunc") == nil {
		t.Error("mutating GetAllSummaries result should not affect internal state")
	}
}

func TestFunctionSummary_GetParamName(t *testing.T) {
	fs := &FunctionSummary{
		Parameters: []ParameterInfo{
			{Index: 0, Name: "first"},
			{Index: 1, Name: "second"},
		},
	}

	tests := []struct {
		index int
		want  string
	}{
		{0, "first"},
		{1, "second"},
		{2, ""},  // out of range
		{-1, ""}, // negative
	}

	for _, tt := range tests {
		got := fs.GetParamName(tt.index)
		if got != tt.want {
			t.Errorf("GetParamName(%d) = %q, want %q", tt.index, got, tt.want)
		}
	}
}

func TestContainsInt(t *testing.T) {
	tests := []struct {
		slice []int
		val   int
		want  bool
	}{
		{[]int{1, 2, 3}, 2, true},
		{[]int{1, 2, 3}, 5, false},
		{[]int{}, 0, false},
		{[]int{0}, 0, true},
	}

	for _, tt := range tests {
		got := slices.Contains(tt.slice, tt.val)
		if got != tt.want {
			t.Errorf("slices.Contains(%v, %d) = %v, want %v", tt.slice, tt.val, got, tt.want)
		}
	}
}

func TestContainsString(t *testing.T) {
	tests := []struct {
		slice []string
		val   string
		want  bool
	}{
		{[]string{"a", "b", "c"}, "b", true},
		{[]string{"a", "b", "c"}, "d", false},
		{[]string{}, "x", false},
		{[]string{""}, "", true},
	}

	for _, tt := range tests {
		got := slices.Contains(tt.slice, tt.val)
		if got != tt.want {
			t.Errorf("slices.Contains(%v, %q) = %v, want %v", tt.slice, tt.val, got, tt.want)
		}
	}
}

func TestIsValidIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"foo", true},
		{"_bar", true},
		{"$var", true},
		{"myVar123", true},
		{"", false},
		{"123start", false},
		{"has-dash", false},
		{"has space", false},
	}

	for _, tt := range tests {
		got := isValidIdentifier(tt.input)
		if got != tt.want {
			t.Errorf("isValidIdentifier(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
