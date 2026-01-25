// Package constants provides centralized type constants for the tracer.
package constants

// FlowNodeType represents the type of a node in the data flow graph
type FlowNodeType string

const (
	NodeSource   FlowNodeType = "source"   // Original input source (e.g., $_GET, req.body)
	NodeCarrier  FlowNodeType = "carrier"  // Object/variable that holds user data
	NodeVariable FlowNodeType = "variable" // Regular variable in the flow
	NodeFunction FlowNodeType = "function" // Function/method call
	NodeProperty FlowNodeType = "property" // Object property access
	NodeParam    FlowNodeType = "param"    // Function parameter
	NodeReturn   FlowNodeType = "return"   // Return value
)

// FlowEdgeType represents how data flows between nodes
type FlowEdgeType string

const (
	EdgeAssignment  FlowEdgeType = "assignment"   // $x = $y
	EdgeParameter   FlowEdgeType = "parameter"    // func($x)
	EdgeReturn      FlowEdgeType = "return"       // return $x
	EdgeProperty    FlowEdgeType = "property"     // $obj->prop = $x
	EdgeArraySet    FlowEdgeType = "array_set"    // $arr['key'] = $x
	EdgeArrayGet    FlowEdgeType = "array_get"    // $x = $arr['key']
	EdgeMethodCall  FlowEdgeType = "method_call"  // $obj->method($x)
	EdgeConstructor FlowEdgeType = "constructor"  // new Class($x)
	EdgeFramework   FlowEdgeType = "framework"    // Framework-specific flow
	EdgeConcatenate FlowEdgeType = "concatenate"  // $x . $y
	EdgeDestructure FlowEdgeType = "destructure"  // const {a, b} = obj
	EdgeIteration   FlowEdgeType = "iteration"    // foreach/for loop
	EdgeConditional FlowEdgeType = "conditional"  // if/else branch
	EdgeCall        FlowEdgeType = "call"         // Function call
	EdgeDataFlow    FlowEdgeType = "data_flow"    // Generic data flow
)
