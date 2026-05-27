package ast

import (
	"reflect"

	sitter "github.com/smacker/go-tree-sitter"
)

// isNilNode reports whether n is nil, accounting for the Go interface nil trap:
// a Node interface value is non-nil even when it wraps a nil *sitter.Node
// pointer. Using reflect.Value.IsNil() catches both cases without panic.
func isNilNode(n Node) bool {
	if n == nil {
		return true
	}
	v := reflect.ValueOf(n)
	return v.Kind() == reflect.Ptr && v.IsNil()
}

// isPunctuation reports whether a Tree-Sitter node type is syntactic
// punctuation that should be skipped during argument/parameter iteration.
// The set covers the tokens universally used as delimiters across all
// supported languages: commas, parentheses, and braces.
func isPunctuation(t string) bool {
	switch t {
	case ",", "(", ")", "{", "}":
		return true
	}
	return false
}

// IterateChildren calls fn for each non-punctuation child of node.
// It is the canonical way to walk argument lists and parameter lists
// in Tree-Sitter ASTs, avoiding the repeated inline loop with punctuation
// guards scattered across the codebase.
func IterateChildren(node *sitter.Node, fn func(child *sitter.Node)) {
	if node == nil {
		return
	}
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child == nil || isPunctuation(child.Type()) {
			continue
		}
		fn(child)
	}
}
