// Package constants provides centralized type constants for the tracer.
// All type constants should be defined here and imported by other packages.
package constants

// ConditionType classifies the type of condition
type ConditionType string

const (
	CondTypeComparison  ConditionType = "comparison"   // ==, !=, <, >, etc.
	CondTypeNullCheck   ConditionType = "null_check"   // isset, empty, is_null
	CondTypeTypeCheck   ConditionType = "type_check"   // is_string, instanceof
	CondTypeLengthCheck ConditionType = "length_check" // strlen, count
	CondTypeLogical     ConditionType = "logical"      // &&, ||, !
	CondTypeUnknown     ConditionType = "unknown"
)

// ConditionEffect describes how a condition affects data flow
type ConditionEffect string

const (
	EffectAllows  ConditionEffect = "allows"  // Condition allows flow if true
	EffectBlocks  ConditionEffect = "blocks"  // Condition blocks flow if true
	EffectUnknown ConditionEffect = "unknown"
)
