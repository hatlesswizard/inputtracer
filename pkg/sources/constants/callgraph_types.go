// Package constants provides centralized type constants for the tracer.
package constants

// CallGraphNodeType represents the type of a call graph node
type CallGraphNodeType int

const (
	CGNodeTypeRegular    CallGraphNodeType = iota
	CGNodeTypeEntryPoint                   // HTTP handlers, main functions, CLI entry points
	CGNodeTypeSource                       // Input source functions
)
