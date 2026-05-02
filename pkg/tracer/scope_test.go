package tracer

import (
	"testing"
)

func TestNewScopeManager_StartsAtGlobal(t *testing.T) {
	sm := NewScopeManager()
	scope := sm.CurrentScope()
	if scope == nil {
		t.Fatal("CurrentScope() returned nil")
	}
	if scope.Type != ScopeGlobal {
		t.Errorf("initial scope type = %v, want ScopeGlobal", scope.Type)
	}
	if scope.ID != "global" {
		t.Errorf("initial scope ID = %q, want %q", scope.ID, "global")
	}
}

func TestScopeManager_EnterExitScope(t *testing.T) {
	sm := NewScopeManager()

	inner := sm.EnterScope(ScopeFunction, "myFunc", Location{})
	if inner == nil {
		t.Fatal("EnterScope returned nil")
	}
	if sm.CurrentScope().Name != "myFunc" {
		t.Errorf("after EnterScope, current = %q, want %q", sm.CurrentScope().Name, "myFunc")
	}

	sm.ExitScope()
	if sm.CurrentScope().Type != ScopeGlobal {
		t.Errorf("after ExitScope, scope type = %v, want ScopeGlobal", sm.CurrentScope().Type)
	}
}

func TestScopeManager_ExitScope_AtGlobal_IsNoop(t *testing.T) {
	sm := NewScopeManager()
	sm.ExitScope() // should not panic or move past global
	if sm.CurrentScope().Type != ScopeGlobal {
		t.Errorf("ExitScope at global should stay global, got %v", sm.CurrentScope().Type)
	}
}

func TestScopeManager_LookupVariable_SameScope(t *testing.T) {
	sm := NewScopeManager()
	sm.DeclareVariable("x", false, nil, 0, Location{})

	sv := sm.LookupVariable("x")
	if sv == nil {
		t.Fatal("LookupVariable returned nil for declared variable")
	}
	if sv.Name != "x" {
		t.Errorf("Name = %q, want %q", sv.Name, "x")
	}
}

func TestScopeManager_LookupVariable_OuterScopeVisible(t *testing.T) {
	sm := NewScopeManager()
	sm.DeclareVariable("outerVar", true, &InputSource{ID: "src1"}, 0, Location{})

	sm.EnterScope(ScopeFunction, "fn", Location{})
	sv := sm.LookupVariable("outerVar")
	if sv == nil {
		t.Fatal("inner scope cannot see outer variable")
	}
	if !sv.Tainted {
		t.Error("outer variable should still be tainted when viewed from inner scope")
	}
}

func TestScopeManager_LookupVariable_InnerScopeShadows(t *testing.T) {
	sm := NewScopeManager()
	src1 := &InputSource{ID: "outer-src"}
	sm.DeclareVariable("v", true, src1, 0, Location{})

	sm.EnterScope(ScopeFunction, "fn", Location{})
	sm.DeclareVariable("v", false, nil, 0, Location{})

	sv := sm.LookupVariable("v")
	if sv == nil {
		t.Fatal("LookupVariable returned nil")
	}
	if sv.Tainted {
		t.Error("inner declaration should shadow outer; inner is not tainted")
	}
	if sv.Shadowing == nil {
		t.Error("shadowing variable should reference outer definition")
	}
}

func TestScopeManager_IsTainted(t *testing.T) {
	sm := NewScopeManager()

	if sm.IsTainted("unknown") {
		t.Error("undeclared variable should not be tainted")
	}

	sm.DeclareVariable("safe", false, nil, 0, Location{})
	if sm.IsTainted("safe") {
		t.Error("non-tainted variable should return false")
	}

	sm.DeclareVariable("dirty", true, &InputSource{ID: "s1"}, 0, Location{})
	if !sm.IsTainted("dirty") {
		t.Error("tainted variable should return true")
	}
}

func TestScopeManager_GetTaintSource(t *testing.T) {
	sm := NewScopeManager()
	src := &InputSource{ID: "test-src", Type: "$_GET"}
	sm.DeclareVariable("input", true, src, 0, Location{})

	got := sm.GetTaintSource("input")
	if got == nil {
		t.Fatal("GetTaintSource returned nil for tainted variable")
	}
	if got.ID != "test-src" {
		t.Errorf("source ID = %q, want %q", got.ID, "test-src")
	}

	if sm.GetTaintSource("nothere") != nil {
		t.Error("GetTaintSource should return nil for undeclared variable")
	}
}

func TestScopeManager_MarkTainted(t *testing.T) {
	sm := NewScopeManager()
	sm.DeclareVariable("x", false, nil, 0, Location{})

	if sm.IsTainted("x") {
		t.Fatal("should not be tainted before MarkTainted")
	}

	src := &InputSource{ID: "mark-src"}
	sm.MarkTainted("x", src, 1)
	if !sm.IsTainted("x") {
		t.Error("variable should be tainted after MarkTainted")
	}
	if sm.GetTaintSource("x").ID != "mark-src" {
		t.Errorf("source ID after mark = %q, want %q", sm.GetTaintSource("x").ID, "mark-src")
	}
}

func TestScopeManager_GetAllTaintedInScope(t *testing.T) {
	sm := NewScopeManager()
	sm.DeclareVariable("a", true, &InputSource{ID: "s1"}, 0, Location{})
	sm.DeclareVariable("b", false, nil, 0, Location{})
	sm.DeclareVariable("c", true, &InputSource{ID: "s2"}, 0, Location{})

	tainted := sm.GetAllTaintedInScope()
	if len(tainted) != 2 {
		t.Errorf("GetAllTaintedInScope() len = %d, want 2", len(tainted))
	}
}

func TestScopeManager_GetScopeQualifiedName(t *testing.T) {
	sm := NewScopeManager()
	if sm.GetScopeQualifiedName() != "global" {
		t.Errorf("at global = %q, want %q", sm.GetScopeQualifiedName(), "global")
	}

	sm.EnterScope(ScopeClass, "MyClass", Location{})
	sm.EnterScope(ScopeFunction, "myMethod", Location{})

	got := sm.GetScopeQualifiedName()
	if got != "MyClass.myMethod" {
		t.Errorf("qualified name = %q, want %q", got, "MyClass.myMethod")
	}
}

func TestScopeManager_Reset(t *testing.T) {
	sm := NewScopeManager()
	sm.EnterScope(ScopeFunction, "fn", Location{})
	sm.DeclareVariable("x", true, &InputSource{ID: "s"}, 0, Location{})

	sm.Reset()

	if sm.CurrentScope().Type != ScopeGlobal {
		t.Error("after Reset, scope should be global")
	}
	if sm.LookupVariable("x") != nil {
		t.Error("after Reset, previously declared variable should not exist")
	}
}

func TestScopeManager_Clone(t *testing.T) {
	sm := NewScopeManager()
	src := &InputSource{ID: "orig"}
	sm.DeclareVariable("taintedVar", true, src, 1, Location{})
	sm.DeclareVariable("cleanVar", false, nil, 0, Location{})

	clone := sm.Clone()

	// Clone should have the tainted variable
	if !clone.IsTainted("taintedVar") {
		t.Error("clone should have tainted variable")
	}

	// Clean variable is not cloned (Clone only copies tainted state)
	// Mutations to original should not affect clone
	sm.MarkTainted("cleanVar", &InputSource{ID: "new"}, 2)
	if clone.IsTainted("cleanVar") {
		t.Error("mutating original after clone should not affect clone")
	}
}
