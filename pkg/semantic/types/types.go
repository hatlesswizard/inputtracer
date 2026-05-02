// Package types defines universal data structures for semantic input tracing
// across all supported programming languages.
//
// Types are organized across four files by concern:
//   - types_flow.go    — FlowNode, FlowEdge, FlowMap and its methods
//   - types_symbol.go  — SymbolTable, ClassDef, FunctionDef and related
//   - types_taint.go   — TaintChain, Assignment, AnalysisState and related
//   - types_result.go  — BackwardTraceResult, BatchTraceResult, FrameworkPattern
package types

import (
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
	"github.com/hatlesswizard/inputtracer/pkg/sources/constants"
)

// ============================================================================
// Re-exported type aliases (for backward compatibility)
// ============================================================================

// FlowNodeType represents the type of a node in the data flow graph.
// Re-exported from pkg/sources/constants.
type FlowNodeType = constants.FlowNodeType

const (
	NodeSource   = constants.NodeSource
	NodeCarrier  = constants.NodeCarrier
	NodeVariable = constants.NodeVariable
	NodeFunction = constants.NodeFunction
	NodeProperty = constants.NodeProperty
	NodeParam    = constants.NodeParam
	NodeReturn   = constants.NodeReturn
)

// FlowEdgeType represents how data flows between nodes.
// Re-exported from pkg/sources/constants.
type FlowEdgeType = constants.FlowEdgeType

const (
	EdgeAssignment  = constants.EdgeAssignment
	EdgeParameter   = constants.EdgeParameter
	EdgeReturn      = constants.EdgeReturn
	EdgeProperty    = constants.EdgeProperty
	EdgeArraySet    = constants.EdgeArraySet
	EdgeArrayGet    = constants.EdgeArrayGet
	EdgeMethodCall  = constants.EdgeMethodCall
	EdgeConstructor = constants.EdgeConstructor
	EdgeFramework   = constants.EdgeFramework
	EdgeConcatenate = constants.EdgeConcatenate
	EdgeDestructure = constants.EdgeDestructure
	EdgeIteration   = constants.EdgeIteration
	EdgeConditional = constants.EdgeConditional
	EdgeCall        = constants.EdgeCall
	EdgeDataFlow    = constants.EdgeDataFlow
)

// SourceType represents the type of input source.
// Re-exported from pkg/sources/common.
type SourceType = common.SourceType

const (
	SourceHTTPGet     = common.SourceHTTPGet
	SourceHTTPPost    = common.SourceHTTPPost
	SourceHTTPBody    = common.SourceHTTPBody
	SourceHTTPJSON    = common.SourceHTTPJSON
	SourceHTTPHeader  = common.SourceHTTPHeader
	SourceHTTPCookie  = common.SourceHTTPCookie
	SourceHTTPPath    = common.SourceHTTPPath
	SourceHTTPFile    = common.SourceHTTPFile
	SourceHTTPRequest = common.SourceHTTPRequest
	SourceSession     = common.SourceSession
	SourceCLIArg      = common.SourceCLIArg
	SourceEnvVar      = common.SourceEnvVar
	SourceStdin       = common.SourceStdin
	SourceFile        = common.SourceFile
	SourceDatabase    = common.SourceDatabase
	SourceNetwork     = common.SourceNetwork
	SourceUserInput   = common.SourceUserInput
	SourceUnknown     = common.SourceUnknown
)
