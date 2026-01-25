// Package php provides PHP-specific patterns for input tracing.
package php

import "regexp"

// TaintPatterns contains pre-compiled regex patterns for PHP taint analysis
var TaintPatterns = struct {
	// ThisArrayPattern matches $this->prop[$key] = assignments
	ThisArrayPattern *regexp.Regexp

	// DynamicPropPattern matches $this->$key = $val assignments
	DynamicPropPattern *regexp.Regexp

	// ReturnThisPattern matches return $this->prop statements
	ReturnThisPattern *regexp.Regexp

	// SuperglobalKeyPattern extracts keys from superglobal access
	SuperglobalKeyPattern *regexp.Regexp

	// LoopVariablePattern matches foreach loop variable assignments
	LoopVariablePattern *regexp.Regexp

	// ForeachValueOnlyPattern matches foreach($x as $value) without key
	ForeachValueOnlyPattern *regexp.Regexp

	// ThisPropertyOptionalArrayPattern matches $this->prop or $this->prop[...]
	ThisPropertyOptionalArrayPattern *regexp.Regexp

	// ReturnThisPropertyArrayPattern matches return $this->prop[...]
	ReturnThisPropertyArrayPattern *regexp.Regexp
}{
	ThisArrayPattern:      regexp.MustCompile(`\$this->(\w+)\[\$\w+\]\s*=`),
	DynamicPropPattern:    regexp.MustCompile(`\$this->\$(\w+)\s*=`),
	ReturnThisPattern:     regexp.MustCompile(`return\s+\$this->(\w+)`),
	SuperglobalKeyPattern: regexp.MustCompile(`\$_[A-Z]+\s*\[\s*['"]([^'"]+)['"]\s*\]`),
	LoopVariablePattern:              regexp.MustCompile(`as\s+\$(\w+)\s*=>\s*\$(\w+)`),
	ForeachValueOnlyPattern:          regexp.MustCompile(`as\s+\$(\w+)\s*\)`),
	ThisPropertyOptionalArrayPattern: regexp.MustCompile(`\$this->(\w+)(?:\[[^\]]*\])?`),
	ReturnThisPropertyArrayPattern:   regexp.MustCompile(`return\s+\$this->(\w+)\[`),
}

// PHPFileExtension is the file extension for PHP files
const PHPFileExtension = ".php"

// PHPFileExtensions contains all PHP file extensions
var PHPFileExtensions = []string{".php", ".php5", ".php7", ".phtml"}

// IsPHPFile checks if a file path is a PHP file
func IsPHPFile(path string) bool {
	for _, ext := range PHPFileExtensions {
		if len(path) > len(ext) && path[len(path)-len(ext):] == ext {
			return true
		}
	}
	return false
}

// PHPNodeTypes contains PHP-specific AST node type strings
var PHPNodeTypes = struct {
	// Class and function nodes
	ClassDeclaration    string
	MethodDeclaration   string
	FunctionDefinition  string
	PropertyDeclaration string
	DeclarationList     string

	// Variable and expression nodes
	VariableName             string
	SubscriptExpression      string
	MemberAccessExpression   string
	MemberCallExpression     string
	FunctionCallExpression   string
	ScopedCallExpression     string
	AssignmentExpression     string
	BinaryExpression         string
	ParenthesizedExpression  string
	EncapsedString           string

	// Parameter types
	SimpleParameter            string
	VariadicParameter          string
	PropertyPromotionParameter string

	// Statement nodes
	ForeachStatement string
	ReturnStatement  string

	// Field names
	FieldName      string
	FieldBaseClause string
	FieldBody      string
	FieldObject    string
	FieldIndex     string
	FieldLeft      string
	FieldRight     string
	FieldFunction  string

	// Modifier types
	VisibilityModifier string
	StaticModifier     string

	// Visibility values
	VisibilityPublic    string
	VisibilityProtected string
	VisibilityPrivate   string
}{
	ClassDeclaration:    "class_declaration",
	MethodDeclaration:   "method_declaration",
	FunctionDefinition:  "function_definition",
	PropertyDeclaration: "property_declaration",
	DeclarationList:     "declaration_list",

	VariableName:             "variable_name",
	SubscriptExpression:      "subscript_expression",
	MemberAccessExpression:   "member_access_expression",
	MemberCallExpression:     "member_call_expression",
	FunctionCallExpression:   "function_call_expression",
	ScopedCallExpression:     "scoped_call_expression",
	AssignmentExpression:     "assignment_expression",
	BinaryExpression:         "binary_expression",
	ParenthesizedExpression:  "parenthesized_expression",
	EncapsedString:           "encapsed_string",

	SimpleParameter:            "simple_parameter",
	VariadicParameter:          "variadic_parameter",
	PropertyPromotionParameter: "property_promotion_parameter",

	ForeachStatement: "foreach_statement",
	ReturnStatement:  "return_statement",

	FieldName:       "name",
	FieldBaseClause: "base_clause",
	FieldBody:       "body",
	FieldObject:     "object",
	FieldIndex:      "index",
	FieldLeft:       "left",
	FieldRight:      "right",
	FieldFunction:   "function",

	VisibilityModifier: "visibility_modifier",
	StaticModifier:     "static_modifier",

	VisibilityPublic:    "public",
	VisibilityProtected: "protected",
	VisibilityPrivate:   "private",
}

// PHPInputProperties contains common PHP input carrier property names
var PHPInputProperties = []string{"input", "cookies", "query", "request", "files", "server", "headers"}

// PHPInputFunctions contains common PHP functions that read input
var PHPInputFunctions = []string{
	"file_get_contents",
	"fread",
	"fgets",
	"fgetc",
	"stream_get_contents",
	"readfile",
}

// PHPInputConstant is the php://input constant value
const PHPInputConstant = "php://input"

// PHPConstructorName is the name of PHP constructor methods
const PHPConstructorName = "__construct"
