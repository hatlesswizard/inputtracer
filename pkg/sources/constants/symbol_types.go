// Package constants provides centralized type constants for the tracer.
package constants

// SymbolType represents the type of a code symbol
type SymbolType string

const (
	SymbolFunction  SymbolType = "function"
	SymbolMethod    SymbolType = "method"
	SymbolClass     SymbolType = "class"
	SymbolInterface SymbolType = "interface"
	SymbolVariable  SymbolType = "variable"
	SymbolConstant  SymbolType = "constant"
	SymbolProperty  SymbolType = "property"
	SymbolParameter SymbolType = "parameter"
)
