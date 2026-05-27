// Package constants provides centralized type constants for the tracer.
package constants

// ScopeType represents the type of scope
type ScopeType string

const (
	ScopeGlobal   ScopeType = "global"
	ScopeFile     ScopeType = "file"
	ScopeModule   ScopeType = "module"
	ScopeClass    ScopeType = "class"
	ScopeFunction ScopeType = "function"
	ScopeBlock    ScopeType = "block"
)
