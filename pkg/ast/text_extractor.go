// Package ast - text_extractor.go provides language-aware extraction of
// assignment targets and values from Tree-Sitter AST nodes.
//
// This logic belongs in pkg/ast (language-agnostic AST facts) rather than
// in pkg/tracer (orchestration), because it is pure structural knowledge
// about how each language's grammar encodes assignments.
package ast

import (
	"strings"

	sitter "github.com/smacker/go-tree-sitter"
)

// ExtractAssignmentParts extracts the target variable name and the value
// expression text from an assignment AST node for the given language.
// Returns ("", "") when the node is not a recognisable assignment.
//
// The node parameter accepts the ast.Node interface; callers that hold a
// *sitter.Node can pass it directly because *sitter.Node satisfies ast.Node.
func ExtractAssignmentParts(node Node, src []byte, language string) (target, value string) {
	if isNilNode(node) {
		return "", ""
	}
	// Internal helpers require *sitter.Node to traverse child nodes.
	concrete, ok := node.(*sitter.Node)
	if !ok || concrete == nil {
		return "", ""
	}
	switch language {
	case "php":
		return extractPHPAssignment(concrete, src)
	case "javascript", "typescript", "tsx":
		return extractJSAssignment(concrete, src)
	case "python":
		return extractPythonAssignment(concrete, src)
	case "go":
		return extractGoAssignment(concrete, src)
	case "java":
		return extractJavaAssignment(concrete, src)
	case "c", "cpp":
		return extractCAssignment(concrete, src)
	case "c_sharp":
		return extractCSharpAssignment(concrete, src)
	case "ruby":
		return extractRubyAssignment(concrete, src)
	case "rust":
		return extractRustAssignment(concrete, src)
	}

	// Generic fallback — look for = operator
	nodeType := node.Type()
	text := string(src[node.StartByte():node.EndByte()])
	if strings.Contains(nodeType, "assignment") || strings.Contains(text, "=") {
		parts := strings.SplitN(text, "=", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}
	return "", ""
}

// extractPHPAssignment extracts target/value from a PHP assignment node.
// PHP: $var = value;
func extractPHPAssignment(node *sitter.Node, src []byte) (string, string) {
	for i := 0; i < int(node.ChildCount()); i++ {
		child := node.Child(i)
		if child.Type() == "variable_name" || child.Type() == "member_access_expression" {
			target := string(src[child.StartByte():child.EndByte()])
			// Find value (sibling after =)
			for j := i + 1; j < int(node.ChildCount()); j++ {
				sibling := node.Child(j)
				if sibling.Type() != "=" {
					value := string(src[sibling.StartByte():sibling.EndByte()])
					return target, value
				}
			}
		}
	}
	return "", ""
}

// extractJSAssignment extracts target/value from a JS/TS assignment node.
// JS: let/const/var name = value; or name = value;
func extractJSAssignment(node *sitter.Node, src []byte) (string, string) {
	nodeType := node.Type()
	if nodeType == "variable_declarator" || nodeType == "assignment_expression" {
		var target, value string
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			childType := child.Type()
			if childType == "identifier" || childType == "member_expression" {
				if target == "" {
					target = string(src[child.StartByte():child.EndByte()])
				}
			} else if childType != "=" && childType != "type_annotation" && target != "" {
				value = string(src[child.StartByte():child.EndByte()])
				break
			}
		}
		return target, value
	}
	return "", ""
}

// extractPythonAssignment extracts target/value from a Python assignment node.
// Python: name = value
func extractPythonAssignment(node *sitter.Node, src []byte) (string, string) {
	if node.Type() == "assignment" || node.Type() == "augmented_assignment" {
		var target, value string
		for i := 0; i < int(node.ChildCount()); i++ {
			child := node.Child(i)
			childType := child.Type()
			if childType == "identifier" || childType == "attribute" || childType == "subscript" {
				if target == "" {
					target = string(src[child.StartByte():child.EndByte()])
				}
			} else if childType != "=" && childType != "type" && target != "" {
				value = string(src[child.StartByte():child.EndByte()])
				break
			}
		}
		return target, value
	}
	return "", ""
}

// extractGoAssignment extracts target/value from a Go assignment node.
// Go: name := value or name = value
func extractGoAssignment(node *sitter.Node, src []byte) (string, string) {
	nodeType := node.Type()
	if nodeType == "short_var_declaration" || nodeType == "assignment_statement" {
		text := string(src[node.StartByte():node.EndByte()])
		if strings.Contains(text, ":=") {
			parts := strings.SplitN(text, ":=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			}
		} else if strings.Contains(text, "=") {
			parts := strings.SplitN(text, "=", 2)
			if len(parts) == 2 {
				return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
			}
		}
	}
	return "", ""
}

// extractJavaAssignment extracts target/value from a Java assignment node.
// Java: Type name = value; or name = value;
func extractJavaAssignment(node *sitter.Node, src []byte) (string, string) {
	text := string(src[node.StartByte():node.EndByte()])
	if strings.Contains(text, "=") {
		parts := strings.SplitN(text, "=", 2)
		if len(parts) == 2 {
			// Left side might be "Type name" or just "name"
			leftParts := strings.Fields(strings.TrimSpace(parts[0]))
			target := leftParts[len(leftParts)-1]
			return target, strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
		}
	}
	return "", ""
}

// extractCAssignment extracts target/value from a C/C++ assignment node.
// C/C++: type name = value; or name = value;
func extractCAssignment(node *sitter.Node, src []byte) (string, string) {
	return extractJavaAssignment(node, src) // similar pattern
}

// extractCSharpAssignment extracts target/value from a C# assignment node.
// C#: Type name = value; or name = value;
func extractCSharpAssignment(node *sitter.Node, src []byte) (string, string) {
	return extractJavaAssignment(node, src) // similar pattern
}

// extractRubyAssignment extracts target/value from a Ruby assignment node.
// Ruby: name = value
func extractRubyAssignment(node *sitter.Node, src []byte) (string, string) {
	text := string(src[node.StartByte():node.EndByte()])
	if strings.Contains(text, "=") && !strings.Contains(text, "==") {
		parts := strings.SplitN(text, "=", 2)
		if len(parts) == 2 {
			return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
		}
	}
	return "", ""
}

// extractRustAssignment extracts target/value from a Rust assignment node.
// Rust: let name = value; or let mut name = value;
func extractRustAssignment(node *sitter.Node, src []byte) (string, string) {
	text := string(src[node.StartByte():node.EndByte()])
	if strings.Contains(text, "=") {
		// Remove "let" and "mut" keywords
		text = strings.TrimPrefix(strings.TrimSpace(text), "let")
		text = strings.TrimPrefix(strings.TrimSpace(text), "mut")
		parts := strings.SplitN(text, "=", 2)
		if len(parts) == 2 {
			// Handle type annotations
			target := strings.TrimSpace(parts[0])
			if colonIdx := strings.Index(target, ":"); colonIdx > 0 {
				target = strings.TrimSpace(target[:colonIdx])
			}
			return target, strings.TrimSpace(strings.TrimSuffix(parts[1], ";"))
		}
	}
	return "", ""
}
