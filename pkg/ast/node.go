// Package ast - node.go defines the Node interface that abstracts over the
// concrete *sitter.Node type from github.com/smacker/go-tree-sitter.
//
// # Design rationale
//
// With 308+ usages of *sitter.Node spread across pkg/tracer, pkg/semantic, and
// pkg/sources, a full migration to the interface in a single pass would be
// high-risk. The chosen approach is therefore incremental:
//
//  1. The Node interface is defined here, covering every method that pkg/ast
//     itself calls on tree-sitter nodes.
//  2. pkg/ast public function signatures accept Node (or *sitter.Node where the
//     return type forces it — see "Pragmatic exceptions" below).
//  3. pkg/tracer and pkg/semantic continue to use *sitter.Node for now.
//     Callers that migrate to Node can use Wrap() to convert.
//
// # Pragmatic exceptions
//
// Child, NamedChild, and Parent return *sitter.Node in the upstream library.
// The interface preserves these concrete return types so that call-sites do not
// need to type-assert after every child traversal. When a future release of the
// upstream library (or a fork) returns a Node interface, these signatures can be
// updated without changing the rest of the codebase.
//
// # Gradual migration plan
//
// Future work: update pkg/tracer/propagation.go, pkg/tracer/scope.go, and the
// pkg/semantic/analyzer/* files to accept ast.Node instead of *sitter.Node.
// The Wrap helper makes that transition zero-cost: Wrap(n) is a no-op because
// *sitter.Node already satisfies the interface.
package ast

import sitter "github.com/smacker/go-tree-sitter"

// Node is the abstract syntax tree node interface used throughout pkg/ast.
// It is satisfied by *sitter.Node and by any test double that implements the
// same surface area, making pkg/ast independently testable without a live
// tree-sitter parser.
//
// All methods mirror the tree-sitter Go binding signatures exactly so that
// existing *sitter.Node values can be passed without wrapping or conversion.
type Node interface {
	// Positional information
	StartByte() uint32
	EndByte() uint32
	StartPoint() sitter.Point
	EndPoint() sitter.Point

	// Type returns the grammar node type (e.g. "assignment_expression").
	Type() string

	// Content returns the source text spanned by the node.
	Content(input []byte) string

	// Tree navigation — pragmatically return *sitter.Node (see package doc).
	ChildCount() uint32
	Child(idx int) *sitter.Node
	NamedChildCount() uint32
	NamedChild(idx int) *sitter.Node
	Parent() *sitter.Node
}

// Compile-time assertion: *sitter.Node must satisfy Node.
// If the upstream library ever removes or renames a method, this line will
// produce a clear build error pointing back to the interface definition.
var _ Node = (*sitter.Node)(nil)

// Wrap converts a concrete *sitter.Node to ast.Node.
// The conversion is a compile-time no-op: *sitter.Node already satisfies the
// Node interface because all required methods are defined on sitter.Node with
// value receivers (promoted to the pointer type by Go).
//
// Use Wrap at call-sites that will gradually migrate from *sitter.Node to
// ast.Node, or in tests that need to pass a concrete node to a function that
// accepts ast.Node.
func Wrap(n *sitter.Node) Node { return n }
