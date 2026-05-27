package tracer

import (
	"testing"
)

func TestNewScopeState_StartsAtGlobal(t *testing.T) {
	ss := newScopeState()
	scope := ss.CurrentScope
	if scope == nil {
		t.Fatal("CurrentScope returned nil")
	}
	if scope.Type != ScopeGlobal {
		t.Errorf("initial scope type = %v, want ScopeGlobal", scope.Type)
	}
	if scope.ID != "global" {
		t.Errorf("initial scope ID = %q, want %q", scope.ID, "global")
	}
}

func TestScopeState_EnterExitScope(t *testing.T) {
	ss := newScopeState()

	inner := ss.EnterScope(ScopeFunction, "myFunc", 1, 10)
	if inner == nil {
		t.Fatal("EnterScope returned nil")
	}
	if ss.CurrentScope.Name != "myFunc" {
		t.Errorf("after EnterScope, current = %q, want %q", ss.CurrentScope.Name, "myFunc")
	}

	ss.ExitScope()
	if ss.CurrentScope.Type != ScopeGlobal {
		t.Errorf("after ExitScope, scope type = %v, want ScopeGlobal", ss.CurrentScope.Type)
	}
}

func TestScopeState_ExitScope_AtGlobal_IsNoop(t *testing.T) {
	ss := newScopeState()
	ss.ExitScope() // should not panic or move past global
	if ss.CurrentScope.Type != ScopeGlobal {
		t.Errorf("ExitScope at global should stay global, got %v", ss.CurrentScope.Type)
	}
}

func TestScopeState_LookupVariable_SameScope(t *testing.T) {
	ss := newScopeState()
	tv := &TaintedVariable{Name: "x"}
	ss.CurrentScope.Variables["x"] = tv

	got, ok := ss.LookupVariable("x")
	if !ok {
		t.Fatal("LookupVariable returned not found for declared variable")
	}
	if got.Name != "x" {
		t.Errorf("Name = %q, want %q", got.Name, "x")
	}
}

func TestScopeState_LookupVariable_OuterScopeVisible(t *testing.T) {
	ss := newScopeState()
	ss.CurrentScope.Variables["outerVar"] = &TaintedVariable{
		Name:   "outerVar",
		Source: &InputSource{ID: "src1"},
	}

	ss.EnterScope(ScopeFunction, "fn", 1, 10)
	got, ok := ss.LookupVariable("outerVar")
	if !ok {
		t.Fatal("inner scope cannot see outer variable")
	}
	if got.Source == nil {
		t.Error("outer variable should still have Source when viewed from inner scope")
	}
}

func TestScopeState_LookupVariable_InnerScopeHides(t *testing.T) {
	ss := newScopeState()
	ss.CurrentScope.Variables["v"] = &TaintedVariable{
		Name:   "v",
		Source: &InputSource{ID: "outer-src"},
	}

	ss.EnterScope(ScopeFunction, "fn", 1, 10)
	// Declare a clean "v" in inner scope (no source = not tainted)
	ss.CurrentScope.Variables["v"] = &TaintedVariable{Name: "v"}

	got, ok := ss.LookupVariable("v")
	if !ok {
		t.Fatal("LookupVariable returned not found")
	}
	if got.Source != nil {
		t.Error("inner declaration should shadow outer; inner has no source (not tainted)")
	}
}

func TestAnalysisState_SetAndLookupTainted(t *testing.T) {
	as := NewAnalysisState()

	src := &InputSource{ID: "test-src", Type: "$_GET"}
	tv := &TaintedVariable{Name: "input", Source: src}
	as.SetTainted("input", tv)

	got, ok := as.LookupVariable("input")
	if !ok {
		t.Fatal("LookupVariable returned not found after SetTainted")
	}
	if got.Source.ID != "test-src" {
		t.Errorf("source ID = %q, want %q", got.Source.ID, "test-src")
	}
}

func TestAnalysisState_LookupFallsBackToTaintedValues(t *testing.T) {
	as := NewAnalysisState()

	// Set directly in TaintedValues map (global fallback)
	tv := &TaintedVariable{Name: "global_var", Source: &InputSource{ID: "g1"}}
	as.TaintedValues["global_var"] = tv

	// Enter a new scope (variable not in scope hierarchy)
	as.EnterScope(ScopeFunction, "fn", 1, 10)

	got, ok := as.LookupVariable("global_var")
	if !ok {
		t.Fatal("LookupVariable should fall back to TaintedValues map")
	}
	if got.Source == nil {
		t.Error("looked-up variable should have a Source (tainted)")
	}
}

func TestAnalysisState_EnterExitScopeDelegates(t *testing.T) {
	as := NewAnalysisState()

	as.EnterScope(ScopeClass, "MyClass", 1, 100)
	if as.CurrentScope.Name != "MyClass" {
		t.Errorf("after EnterScope, current = %q, want %q", as.CurrentScope.Name, "MyClass")
	}

	as.ExitScope()
	if as.CurrentScope.Type != ScopeGlobal {
		t.Errorf("after ExitScope, scope type = %v, want ScopeGlobal", as.CurrentScope.Type)
	}
}

func TestAnalysisState_NewAnalysisState_Initialized(t *testing.T) {
	as := NewAnalysisState()
	if as.TaintedValues == nil {
		t.Error("TaintedValues should be initialized")
	}
	if as.FunctionSummaries == nil {
		t.Error("FunctionSummaries should be initialized")
	}
	if as.VisitedFunctions == nil {
		t.Error("VisitedFunctions should be initialized")
	}
	if as.CurrentScope == nil {
		t.Error("CurrentScope should be initialized")
	}
}

func TestScopeState_NestedScopes(t *testing.T) {
	ss := newScopeState()

	// global → class → method
	ss.EnterScope(ScopeClass, "MyClass", 1, 100)
	ss.EnterScope(ScopeFunction, "myMethod", 5, 20)

	if len(ss.ScopeStack) != 3 {
		t.Errorf("ScopeStack len = %d, want 3 (global + class + method)", len(ss.ScopeStack))
	}

	ss.ExitScope()
	if ss.CurrentScope.Name != "MyClass" {
		t.Errorf("after ExitScope (method), current = %q, want MyClass", ss.CurrentScope.Name)
	}

	ss.ExitScope()
	if ss.CurrentScope.Type != ScopeGlobal {
		t.Errorf("after ExitScope (class), should be back at global, got %v", ss.CurrentScope.Type)
	}
}

func TestScopeState_ParentLink(t *testing.T) {
	ss := newScopeState()
	global := ss.CurrentScope

	inner := ss.EnterScope(ScopeFunction, "fn", 1, 10)
	if inner.Parent != global {
		t.Error("inner scope Parent should point to global scope")
	}
}
