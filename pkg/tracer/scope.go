package tracer

import (
	"strings"
	"sync"

	"github.com/hatlesswizard/inputtracer/pkg/sources/constants"
)

// ScopeType represents the type of scope
// Re-exported from pkg/sources/constants for backward compatibility
type ScopeType = constants.ScopeType

// Re-export ScopeType constants for backward compatibility
const (
	ScopeGlobal   = constants.ScopeGlobal
	ScopeFile     = constants.ScopeFile
	ScopeModule   = constants.ScopeModule
	ScopeClass    = constants.ScopeClass
	ScopeFunction = constants.ScopeFunction
	ScopeBlock    = constants.ScopeBlock
)

// ScopeManager manages variable scopes during analysis
type ScopeManager struct {
	scopes    []*Scope
	current   *Scope
	variables map[string][]*ScopedVariable // variable name -> definitions in different scopes
	mu        sync.RWMutex
}

// ScopedVariable represents a variable within a specific scope
type ScopedVariable struct {
	Name      string
	Scope     *Scope
	Tainted   bool
	Source    *InputSource
	Depth     int
	Location  Location
	Shadowing *ScopedVariable // Previous definition being shadowed
}

// NewScopeManager creates a new scope manager
func NewScopeManager() *ScopeManager {
	global := &Scope{
		ID:       "global",
		Type:     ScopeGlobal,
		Name:     "global",
		Parent:   nil,
		Children: make([]*Scope, 0),
	}

	return &ScopeManager{
		scopes:    []*Scope{global},
		current:   global,
		variables: make(map[string][]*ScopedVariable),
	}
}

// EnterScope creates and enters a new scope
func (sm *ScopeManager) EnterScope(scopeType ScopeType, name string, location Location) *Scope {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	scope := &Scope{
		ID:       sm.generateScopeID(name),
		Type:     scopeType,
		Name:     name,
		Parent:   sm.current,
		Children: make([]*Scope, 0),
		StartLoc: location,
	}

	sm.current.Children = append(sm.current.Children, scope)
	sm.scopes = append(sm.scopes, scope)
	sm.current = scope

	return scope
}

// ExitScope exits the current scope
func (sm *ScopeManager) ExitScope() *Scope {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if sm.current.Parent != nil {
		sm.current = sm.current.Parent
	}
	return sm.current
}

// CurrentScope returns the current scope
func (sm *ScopeManager) CurrentScope() *Scope {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.current
}

// DeclareVariable declares a variable in the current scope
func (sm *ScopeManager) DeclareVariable(name string, tainted bool, source *InputSource, depth int, location Location) *ScopedVariable {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// Check for shadowing in current scope chain
	var shadowing *ScopedVariable
	if existing := sm.lookupVariableInScope(name, sm.current); existing != nil {
		shadowing = existing
	}

	sv := &ScopedVariable{
		Name:      name,
		Scope:     sm.current,
		Tainted:   tainted,
		Source:    source,
		Depth:     depth,
		Location:  location,
		Shadowing: shadowing,
	}

	sm.variables[name] = append(sm.variables[name], sv)
	return sv
}

// LookupVariable looks up a variable by name, respecting scope rules
func (sm *ScopeManager) LookupVariable(name string) *ScopedVariable {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return sm.lookupVariableInScope(name, sm.current)
}

// lookupVariableInScope looks up a variable in a specific scope chain
func (sm *ScopeManager) lookupVariableInScope(name string, scope *Scope) *ScopedVariable {
	defs, exists := sm.variables[name]
	if !exists || len(defs) == 0 {
		return nil
	}

	// Search from current scope upward
	currentScope := scope
	for currentScope != nil {
		// Find most recent definition in this scope
		for i := len(defs) - 1; i >= 0; i-- {
			if defs[i].Scope == currentScope {
				return defs[i]
			}
		}
		currentScope = currentScope.Parent
	}

	return nil
}

// IsTainted checks if a variable is tainted in the current scope
func (sm *ScopeManager) IsTainted(name string) bool {
	sv := sm.LookupVariable(name)
	return sv != nil && sv.Tainted
}

// GetTaintSource returns the taint source for a variable
func (sm *ScopeManager) GetTaintSource(name string) *InputSource {
	sv := sm.LookupVariable(name)
	if sv != nil && sv.Tainted {
		return sv.Source
	}
	return nil
}

// MarkTainted marks a variable as tainted
func (sm *ScopeManager) MarkTainted(name string, source *InputSource, depth int) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	sv := sm.lookupVariableInScope(name, sm.current)
	if sv != nil {
		sv.Tainted = true
		sv.Source = source
		sv.Depth = depth
	}
}

// GetAllTaintedInScope returns all tainted variables visible in current scope
func (sm *ScopeManager) GetAllTaintedInScope() []*ScopedVariable {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var tainted []*ScopedVariable
	seen := make(map[string]bool)

	// Walk from current scope upward
	currentScope := sm.current
	for currentScope != nil {
		for name, defs := range sm.variables {
			if seen[name] {
				continue // Already found in closer scope
			}
			// Find definition in this scope
			for i := len(defs) - 1; i >= 0; i-- {
				if defs[i].Scope == currentScope {
					seen[name] = true
					if defs[i].Tainted {
						tainted = append(tainted, defs[i])
					}
					break
				}
			}
		}
		currentScope = currentScope.Parent
	}

	return tainted
}

// GetScopeChain returns the chain of scopes from current to global
func (sm *ScopeManager) GetScopeChain() []*Scope {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var chain []*Scope
	current := sm.current
	for current != nil {
		chain = append(chain, current)
		current = current.Parent
	}
	return chain
}

// GetScopeQualifiedName returns the fully qualified scope name
func (sm *ScopeManager) GetScopeQualifiedName() string {
	chain := sm.GetScopeChain()
	if len(chain) == 0 {
		return "global"
	}

	// Reverse chain (global first)
	names := make([]string, 0, len(chain))
	for i := len(chain) - 1; i >= 0; i-- {
		if chain[i].Type != ScopeGlobal {
			names = append(names, chain[i].Name)
		}
	}

	if len(names) == 0 {
		return "global"
	}
	return strings.Join(names, ".")
}

// generateScopeID generates a unique scope ID
func (sm *ScopeManager) generateScopeID(name string) string {
	base := sm.current.ID
	if base == "" || base == "global" {
		return name
	}
	return base + "." + name
}

// Reset resets the scope manager to initial state
func (sm *ScopeManager) Reset() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	global := &Scope{
		ID:       "global",
		Type:     ScopeGlobal,
		Name:     "global",
		Parent:   nil,
		Children: make([]*Scope, 0),
	}

	sm.scopes = []*Scope{global}
	sm.current = global
	sm.variables = make(map[string][]*ScopedVariable)
}

// Clone creates a copy of the scope manager state (for parallel analysis)
func (sm *ScopeManager) Clone() *ScopeManager {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	clone := NewScopeManager()

	// Deep copy scopes (simplified - just copy variable taint state)
	for name, defs := range sm.variables {
		for _, def := range defs {
			if def.Tainted {
				clone.DeclareVariable(name, true, def.Source, def.Depth, def.Location)
			}
		}
	}

	return clone
}
