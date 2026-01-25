// Package sources - ast_patterns.go provides centralized AST node type patterns
// All AST node type definitions should be referenced from here
package sources

// ASTNodeCategory represents categories of AST nodes
type ASTNodeCategory string

const (
	ASTCategoryFunction   ASTNodeCategory = "function"
	ASTCategoryScope      ASTNodeCategory = "scope"
	ASTCategoryAssignment ASTNodeCategory = "assignment"
	ASTCategoryCall       ASTNodeCategory = "call"
)

// ASTNodeTypes holds node type patterns for different categories
type ASTNodeTypes struct {
	// FunctionTypes are node types that represent function/method definitions
	FunctionTypes []string

	// ScopeTypes are node types that define variable scopes
	ScopeTypes []string

	// AssignmentTypes are node types for assignment operations
	AssignmentTypes []string

	// CallTypes are node types for function/method calls
	CallTypes []string

	// IdentifierTypes are node types for variable/identifier names
	IdentifierTypes []string
}

// UniversalASTNodeTypes contains AST patterns that work across languages
// Replaces hardcoded arrays in propagation.go findContainingFunction() and getCurrentScope()
var UniversalASTNodeTypes = ASTNodeTypes{
	FunctionTypes: []string{
		"function_definition",
		"function_declaration",
		"method_definition",
		"method_declaration",
		"function_item",
		"arrow_function",
		"function_expression",
		"lambda",
		"def",
		"fn_item",
	},
	ScopeTypes: []string{
		"function_definition",
		"function_declaration",
		"method_definition",
		"method_declaration",
		"class_definition",
		"class_declaration",
		"module",
		"program",
		"source_file",
	},
	AssignmentTypes: []string{
		"assignment_expression",
		"assignment_statement",
		"augmented_assignment",
		"variable_declarator",
		"short_var_declaration",
	},
	CallTypes: []string{
		"call_expression",
		"function_call_expression",
		"member_call_expression",
		"method_invocation",
	},
	IdentifierTypes: []string{
		"identifier",
		"variable_name",
		"name",
		"property_identifier",
		"attribute",
		"constant",
	},
}

// LanguageASTNodeTypes provides language-specific AST node types
var LanguageASTNodeTypes = map[string]ASTNodeTypes{
	"php": {
		FunctionTypes: []string{
			"function_definition",
			"method_declaration",
			"arrow_function",
		},
		ScopeTypes: []string{
			"function_definition",
			"method_declaration",
			"class_declaration",
			"program",
		},
		AssignmentTypes: []string{
			"assignment_expression",
			"augmented_assignment_expression",
		},
		CallTypes: []string{
			"function_call_expression",
			"member_call_expression",
			"scoped_call_expression",
		},
		IdentifierTypes: []string{
			"variable_name",
			"name",
		},
	},
	"javascript": {
		FunctionTypes: []string{
			"function_declaration",
			"function_expression",
			"arrow_function",
			"method_definition",
		},
		ScopeTypes: []string{
			"function_declaration",
			"function_expression",
			"arrow_function",
			"method_definition",
			"class_declaration",
			"program",
		},
		AssignmentTypes: []string{
			"assignment_expression",
			"augmented_assignment_expression",
			"variable_declarator",
		},
		CallTypes: []string{
			"call_expression",
			"new_expression",
		},
		IdentifierTypes: []string{
			"identifier",
			"property_identifier",
		},
	},
	"typescript": {
		FunctionTypes: []string{
			"function_declaration",
			"function_expression",
			"arrow_function",
			"method_definition",
		},
		ScopeTypes: []string{
			"function_declaration",
			"function_expression",
			"arrow_function",
			"method_definition",
			"class_declaration",
			"program",
		},
		AssignmentTypes: []string{
			"assignment_expression",
			"augmented_assignment_expression",
			"variable_declarator",
		},
		CallTypes: []string{
			"call_expression",
			"new_expression",
		},
		IdentifierTypes: []string{
			"identifier",
			"property_identifier",
		},
	},
	"tsx": {
		FunctionTypes: []string{
			"function_declaration",
			"function_expression",
			"arrow_function",
			"method_definition",
		},
		ScopeTypes: []string{
			"function_declaration",
			"function_expression",
			"arrow_function",
			"method_definition",
			"class_declaration",
			"program",
		},
		AssignmentTypes: []string{
			"assignment_expression",
			"augmented_assignment_expression",
			"variable_declarator",
		},
		CallTypes: []string{
			"call_expression",
			"new_expression",
		},
		IdentifierTypes: []string{
			"identifier",
			"property_identifier",
		},
	},
	"python": {
		FunctionTypes: []string{
			"function_definition",
			"lambda",
		},
		ScopeTypes: []string{
			"function_definition",
			"class_definition",
			"module",
		},
		AssignmentTypes: []string{
			"assignment",
			"augmented_assignment",
			"named_expression",
		},
		CallTypes: []string{
			"call",
		},
		IdentifierTypes: []string{
			"identifier",
			"attribute",
		},
	},
	"go": {
		FunctionTypes: []string{
			"function_declaration",
			"method_declaration",
			"func_literal",
		},
		ScopeTypes: []string{
			"function_declaration",
			"method_declaration",
			"source_file",
		},
		AssignmentTypes: []string{
			"short_var_declaration",
			"assignment_statement",
		},
		CallTypes: []string{
			"call_expression",
		},
		IdentifierTypes: []string{
			"identifier",
			"selector_expression",
		},
	},
	"java": {
		FunctionTypes: []string{
			"method_declaration",
			"constructor_declaration",
			"lambda_expression",
		},
		ScopeTypes: []string{
			"method_declaration",
			"constructor_declaration",
			"class_declaration",
			"interface_declaration",
			"program",
		},
		AssignmentTypes: []string{
			"assignment_expression",
			"variable_declarator",
		},
		CallTypes: []string{
			"method_invocation",
			"object_creation_expression",
		},
		IdentifierTypes: []string{
			"identifier",
		},
	},
	"c": {
		FunctionTypes: []string{
			"function_definition",
		},
		ScopeTypes: []string{
			"function_definition",
			"translation_unit",
		},
		AssignmentTypes: []string{
			"assignment_expression",
			"init_declarator",
		},
		CallTypes: []string{
			"call_expression",
		},
		IdentifierTypes: []string{
			"identifier",
		},
	},
	"cpp": {
		FunctionTypes: []string{
			"function_definition",
			"lambda_expression",
		},
		ScopeTypes: []string{
			"function_definition",
			"class_specifier",
			"translation_unit",
		},
		AssignmentTypes: []string{
			"assignment_expression",
			"init_declarator",
		},
		CallTypes: []string{
			"call_expression",
		},
		IdentifierTypes: []string{
			"identifier",
		},
	},
	"c_sharp": {
		FunctionTypes: []string{
			"method_declaration",
			"constructor_declaration",
			"lambda_expression",
		},
		ScopeTypes: []string{
			"method_declaration",
			"constructor_declaration",
			"class_declaration",
			"interface_declaration",
			"compilation_unit",
		},
		AssignmentTypes: []string{
			"assignment_expression",
			"variable_declarator",
		},
		CallTypes: []string{
			"invocation_expression",
			"object_creation_expression",
		},
		IdentifierTypes: []string{
			"identifier",
		},
	},
	"ruby": {
		FunctionTypes: []string{
			"method",
			"singleton_method",
			"lambda",
			"block",
		},
		ScopeTypes: []string{
			"method",
			"singleton_method",
			"class",
			"module",
			"program",
		},
		AssignmentTypes: []string{
			"assignment",
			"operator_assignment",
		},
		CallTypes: []string{
			"call",
			"method_call",
		},
		IdentifierTypes: []string{
			"identifier",
			"constant",
		},
	},
	"rust": {
		FunctionTypes: []string{
			"function_item",
			"closure_expression",
		},
		ScopeTypes: []string{
			"function_item",
			"impl_item",
			"mod_item",
			"source_file",
		},
		AssignmentTypes: []string{
			"assignment_expression",
			"let_declaration",
		},
		CallTypes: []string{
			"call_expression",
			"method_call_expression",
		},
		IdentifierTypes: []string{
			"identifier",
		},
	},
}

// IsFunctionNode checks if a node type represents a function definition
func IsFunctionNode(nodeType string) bool {
	for _, ft := range UniversalASTNodeTypes.FunctionTypes {
		if nodeType == ft {
			return true
		}
	}
	return false
}

// IsFunctionNodeForLanguage checks if a node type is a function for a specific language
func IsFunctionNodeForLanguage(nodeType, language string) bool {
	// Check language-specific first
	if langTypes, ok := LanguageASTNodeTypes[language]; ok {
		for _, ft := range langTypes.FunctionTypes {
			if nodeType == ft {
				return true
			}
		}
	}
	// Fall back to universal
	return IsFunctionNode(nodeType)
}

// IsScopeNode checks if a node type defines a scope
func IsScopeNode(nodeType string) bool {
	for _, st := range UniversalASTNodeTypes.ScopeTypes {
		if nodeType == st {
			return true
		}
	}
	return false
}

// IsScopeNodeForLanguage checks if a node type defines a scope for a specific language
func IsScopeNodeForLanguage(nodeType, language string) bool {
	// Check language-specific first
	if langTypes, ok := LanguageASTNodeTypes[language]; ok {
		for _, st := range langTypes.ScopeTypes {
			if nodeType == st {
				return true
			}
		}
	}
	// Fall back to universal
	return IsScopeNode(nodeType)
}

// IsAssignmentNode checks if a node type is an assignment
func IsAssignmentNode(nodeType string) bool {
	for _, at := range UniversalASTNodeTypes.AssignmentTypes {
		if nodeType == at {
			return true
		}
	}
	return false
}

// IsCallNode checks if a node type is a function/method call
func IsCallNode(nodeType string) bool {
	for _, ct := range UniversalASTNodeTypes.CallTypes {
		if nodeType == ct {
			return true
		}
	}
	return false
}

// GetFunctionTypes returns the list of function node types
func GetFunctionTypes() []string {
	return UniversalASTNodeTypes.FunctionTypes
}

// GetScopeTypes returns the list of scope node types
func GetScopeTypes() []string {
	return UniversalASTNodeTypes.ScopeTypes
}

// GetAssignmentTypes returns the list of assignment node types
func GetAssignmentTypes() []string {
	return UniversalASTNodeTypes.AssignmentTypes
}

// GetCallTypes returns the list of call node types
func GetCallTypes() []string {
	return UniversalASTNodeTypes.CallTypes
}

// GetIdentifierTypes returns the list of identifier node types
func GetIdentifierTypes() []string {
	return UniversalASTNodeTypes.IdentifierTypes
}

// GetAssignmentTypesForLanguage returns assignment node types for a specific language
func GetAssignmentTypesForLanguage(language string) []string {
	if langTypes, ok := LanguageASTNodeTypes[language]; ok {
		return langTypes.AssignmentTypes
	}
	return UniversalASTNodeTypes.AssignmentTypes
}

// GetCallTypesForLanguage returns call node types for a specific language
func GetCallTypesForLanguage(language string) []string {
	if langTypes, ok := LanguageASTNodeTypes[language]; ok {
		return langTypes.CallTypes
	}
	return UniversalASTNodeTypes.CallTypes
}

// GetIdentifierTypesForLanguage returns identifier node types for a specific language
func GetIdentifierTypesForLanguage(language string) []string {
	if langTypes, ok := LanguageASTNodeTypes[language]; ok {
		return langTypes.IdentifierTypes
	}
	return UniversalASTNodeTypes.IdentifierTypes
}

// GetASTNodeTypesForLanguage returns the complete ASTNodeTypes for a language
func GetASTNodeTypesForLanguage(language string) ASTNodeTypes {
	if langTypes, ok := LanguageASTNodeTypes[language]; ok {
		return langTypes
	}
	return UniversalASTNodeTypes
}

// IsIdentifierNode checks if a node type is an identifier
func IsIdentifierNode(nodeType string) bool {
	for _, it := range UniversalASTNodeTypes.IdentifierTypes {
		if nodeType == it {
			return true
		}
	}
	return false
}

// IsIdentifierNodeForLanguage checks if a node type is an identifier for a specific language
func IsIdentifierNodeForLanguage(nodeType, language string) bool {
	if langTypes, ok := LanguageASTNodeTypes[language]; ok {
		for _, it := range langTypes.IdentifierTypes {
			if nodeType == it {
				return true
			}
		}
	}
	return IsIdentifierNode(nodeType)
}

// IsAssignmentNodeForLanguage checks if a node type is an assignment for a specific language
func IsAssignmentNodeForLanguage(nodeType, language string) bool {
	if langTypes, ok := LanguageASTNodeTypes[language]; ok {
		for _, at := range langTypes.AssignmentTypes {
			if nodeType == at {
				return true
			}
		}
	}
	return IsAssignmentNode(nodeType)
}

// IsCallNodeForLanguage checks if a node type is a call for a specific language
func IsCallNodeForLanguage(nodeType, language string) bool {
	if langTypes, ok := LanguageASTNodeTypes[language]; ok {
		for _, ct := range langTypes.CallTypes {
			if nodeType == ct {
				return true
			}
		}
	}
	return IsCallNode(nodeType)
}
