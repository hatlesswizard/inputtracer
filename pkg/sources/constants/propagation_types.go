// Package constants provides centralized type constants for the tracer.
package constants

// PropagationStepType defines the type of propagation step
type PropagationStepType string

const (
	StepAssignment    PropagationStepType = "assignment"
	StepParameterPass PropagationStepType = "parameter_pass"
	StepReturn        PropagationStepType = "return"
	StepConcatenation PropagationStepType = "concatenation"
	StepArrayAccess   PropagationStepType = "array_access"
	StepObjectAccess  PropagationStepType = "object_access"
	StepDestructure   PropagationStepType = "destructure"
)
