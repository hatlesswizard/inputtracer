// Package constants provides centralized type constants for the tracer.
package constants

// PathNodeType represents the type of a path node
type PathNodeType string

const (
	PathNodeSource    PathNodeType = "source"    // Input source
	PathNodeCall      PathNodeType = "call"      // Function call
	PathNodeReturn    PathNodeType = "return"    // Return from function
	PathNodeAssign    PathNodeType = "assign"    // Variable assignment
	PathNodeCondition PathNodeType = "condition" // Conditional branch
	PathNodeTransform PathNodeType = "transform" // Data transformation
)

// PruneReason explains why a path was pruned
type PruneReason string

const (
	PruneNone        PruneReason = ""
	PruneMaxDepth    PruneReason = "max_depth_exceeded"
	PruneCycle       PruneReason = "cycle_detected"
	PruneInfeasible  PruneReason = "infeasible_condition"
	PruneDead        PruneReason = "dead_code"
	PruneUnreachable PruneReason = "unreachable"
	PruneLowPriority PruneReason = "low_priority"
)
