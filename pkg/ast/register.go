package ast

// RegisterAll registers all language extractors with the registry
func RegisterAll(r *Registry) {
	// PHP
	r.Register(NewBaseExtractor("php",
		[]string{"assignment_expression", "augmented_assignment_expression"},
		[]string{"function_call_expression", "method_call_expression", "scoped_call_expression"},
		[]string{"variable_name", "name"},
	))

	// JavaScript
	r.Register(NewBaseExtractor("javascript",
		[]string{"assignment_expression", "variable_declarator", "augmented_assignment_expression"},
		[]string{"call_expression", "new_expression"},
		[]string{"identifier", "property_identifier"},
	))

	// TypeScript
	r.Register(NewBaseExtractor("typescript",
		[]string{"assignment_expression", "variable_declarator", "augmented_assignment_expression"},
		[]string{"call_expression", "new_expression"},
		[]string{"identifier", "property_identifier"},
	))

	// TSX
	r.Register(NewBaseExtractor("tsx",
		[]string{"assignment_expression", "variable_declarator", "augmented_assignment_expression"},
		[]string{"call_expression", "new_expression"},
		[]string{"identifier", "property_identifier"},
	))

	// Python
	r.Register(NewBaseExtractor("python",
		[]string{"assignment", "augmented_assignment", "named_expression"},
		[]string{"call"},
		[]string{"identifier", "attribute"},
	))

	// Go
	r.Register(NewBaseExtractor("go",
		[]string{"assignment_statement", "short_var_declaration"},
		[]string{"call_expression"},
		[]string{"identifier", "selector_expression"},
	))

	// Java
	r.Register(NewBaseExtractor("java",
		[]string{"assignment_expression", "variable_declarator"},
		[]string{"method_invocation", "object_creation_expression"},
		[]string{"identifier"},
	))

	// C
	r.Register(NewBaseExtractor("c",
		[]string{"assignment_expression", "init_declarator"},
		[]string{"call_expression"},
		[]string{"identifier"},
	))

	// C++
	r.Register(NewBaseExtractor("cpp",
		[]string{"assignment_expression", "init_declarator"},
		[]string{"call_expression"},
		[]string{"identifier"},
	))

	// C#
	r.Register(NewBaseExtractor("c_sharp",
		[]string{"assignment_expression", "variable_declarator"},
		[]string{"invocation_expression", "object_creation_expression"},
		[]string{"identifier"},
	))

	// Ruby
	r.Register(NewBaseExtractor("ruby",
		[]string{"assignment", "operator_assignment"},
		[]string{"call", "method_call"},
		[]string{"identifier", "constant"},
	))

	// Rust
	r.Register(NewBaseExtractor("rust",
		[]string{"assignment_expression", "let_declaration"},
		[]string{"call_expression", "method_call_expression"},
		[]string{"identifier"},
	))
}
