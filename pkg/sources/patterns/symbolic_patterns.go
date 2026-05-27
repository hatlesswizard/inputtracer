// Package patterns provides centralized regex patterns for code analysis.
// This file contains patterns used by symbolic execution engine.
package patterns

import (
	"regexp"
)

// =============================================================================
// EXPRESSION PARSING PATTERNS
// Used for parsing expressions in symbolic tracing
// =============================================================================

var (
	// SuperglobalAccessPattern matches $_SUPERGLOBAL['key'] or $_SUPERGLOBAL["key"]
	// e.g., $_GET['id'], $_POST["name"], $_REQUEST[$var]
	SuperglobalAccessPattern = regexp.MustCompile(`^\$_(GET|POST|COOKIE|REQUEST|SERVER|FILES|ENV|SESSION)\[['"]?(\w+)['"]?\]$`)

	// StaticPropertyPattern matches Class::$property or Class::CONSTANT
	// e.g., MyClass::$instance, Config::DEBUG
	StaticPropertyPattern = regexp.MustCompile(`^(\w+)::\$?(\w+)$`)

	// PropertyAccessPattern matches $var->property or $var->property['key']
	// e.g., $obj->data, $request->params['id']
	PropertyAccessPattern = regexp.MustCompile(`^\$(\w+)->(\w+)(?:\[['"]?(\w+)['"]?\])?$`)

	// LocalVariablePattern matches simple variable $varname
	// e.g., $id, $username, $data
	LocalVariablePattern = regexp.MustCompile(`^\$(\w+)$`)

	// ChainPropertyWithKeyPattern matches chain property with array access ->property['key']
	// e.g., ->input['id'], ->data["name"]
	ChainPropertyWithKeyPattern = regexp.MustCompile(`^->(\w+)\[['"]?(\w+)['"]?\]`)

	// ChainSimplePropertyPattern matches simple chain property access ->property
	// e.g., ->input, ->data
	ChainSimplePropertyPattern = regexp.MustCompile(`^->(\w+)`)
)

// =============================================================================
// VALIDATION PATTERNS
// Used for validating identifiers and variable names
// =============================================================================

var (
	// WordPattern matches word characters only (identifier validation)
	// e.g., className, method_name, var123
	WordPattern = regexp.MustCompile(`^\w+$`)

	// DollarVariablePattern matches PHP variable with dollar sign
	// e.g., $var, $myVariable
	DollarVariablePattern = regexp.MustCompile(`^\$\w+$`)
)

// =============================================================================
// RETURN STATEMENT PATTERNS
// Used for analyzing method return statements
// =============================================================================

var (
	// ReturnStatementPattern matches return statements
	// e.g., return $value;, return $this->prop;
	ReturnStatementPattern = regexp.MustCompile(`return\s+([^;]+);`)

	// TypeCastPropertyReturnPattern matches (type)$this->property[$param]
	// e.g., (int)$this->input[$name], (string)$this->data[$key]
	TypeCastPropertyReturnPattern = regexp.MustCompile(`\((\w+)\)\s*\$this->(\w+)\[\$(\w+)\]`)

	// PropertyWithParamKeyPattern matches $this->property[$param] without cast
	// e.g., $this->input[$name], $this->data[$key]
	PropertyWithParamKeyPattern = regexp.MustCompile(`\$this->(\w+)\[\$(\w+)\]`)

	// NullCoalescePropertyPattern matches $this->property[$param] ??
	// e.g., $this->input[$name] ?? $default
	NullCoalescePropertyPattern = regexp.MustCompile(`\$this->(\w+)\[\$(\w+)\]\s*\?\?`)

	// TernaryIssetPattern matches isset($this->property[$param]) ? $this->property[$param] : default
	// e.g., isset($this->data[$key]) ? $this->data[$key] : null
	TernaryIssetPattern = regexp.MustCompile(`isset\s*\(\s*\$this->(\w+)\[\$(\w+)\]\s*\)\s*\?\s*\$this->(\w+)\[\$(\w+)\]`)

	// DirectPropertyReturnPattern matches return $this->property
	// e.g., return $this->data;, return $this->input;
	DirectPropertyReturnPattern = regexp.MustCompile(`^\$this->(\w+)$`)
)

// =============================================================================
// MAGIC PROPERTY PATTERNS
// Used for analyzing magic methods and dynamic properties
// =============================================================================

var (
	// BackingPropertyPattern matches return $this->property[$name] in __get
	// e.g., return $this->phrases[$name];
	BackingPropertyPattern = regexp.MustCompile(`return\s+\$this->(\w+)\[\$\w+\]`)

	// DynamicPropertyAssignPattern matches $this->$key = $val
	// e.g., $this->$name = $value;
	DynamicPropertyAssignPattern = regexp.MustCompile(`\$this->\$(\w+)\s*=\s*\$(\w+)`)

	// ForeachWithKVPattern matches foreach($array as $key => $val)
	// e.g., foreach($data as $k => $v)
	ForeachWithKVPattern = regexp.MustCompile(`foreach\s*\(\s*\$(\w+)\s+as\s+\$\w+\s*=>\s*\$\w+`)
)

// =============================================================================
// METHOD INFERENCE PATTERNS
// Used for inferring method return types
// =============================================================================

var (
	// ReturnNewPattern matches return new ClassName(
	// e.g., return new User(, return new Response(
	ReturnNewPattern = regexp.MustCompile(`return\s+new\s+(\w+)\(`)

	// PHPDocReturnPattern matches @return TypeName in PHPDoc
	// e.g., @return User, @return Response
	PHPDocReturnPattern = regexp.MustCompile(`@return\s+(\w+)`)
)

// =============================================================================
// FUNCTION CALL PATTERNS
// Used for parsing function calls in expressions
// =============================================================================

var (
	// FunctionCallPattern matches functionName(args)
	// e.g., strlen($str), generate_post_check()
	FunctionCallPattern = regexp.MustCompile(`^(\w+)\(([^)]*)\)$`)
)

// =============================================================================
// DYNAMIC PATTERN BUILDERS
// Functions that build patterns based on runtime values
// =============================================================================

// BuildVariableAssignPattern creates a pattern for $varname = something;
func BuildVariableAssignPattern(varName string) *regexp.Regexp {
	return regexp.MustCompile(`\$` + regexp.QuoteMeta(varName) + `\s*=\s*([^;]+);`)
}

// BuildPropertyExternalAssignPattern creates a pattern for $var->property = something;
func BuildPropertyExternalAssignPattern(varName, propertyName string) *regexp.Regexp {
	return regexp.MustCompile(`\$` + regexp.QuoteMeta(varName) + `->` + regexp.QuoteMeta(propertyName) + `\s*=\s*([^;]+);`)
}

// BuildPropertyArrayExternalAssignPattern creates a pattern for $var->property['key'] = something;
func BuildPropertyArrayExternalAssignPattern(varName, propertyName string) *regexp.Regexp {
	return regexp.MustCompile(`\$` + regexp.QuoteMeta(varName) + `->` + regexp.QuoteMeta(propertyName) + `\[['"]?\w+['"]?\]\s*=\s*([^;]+);`)
}

// BuildPropertyAssignInLoopPattern creates a pattern for $this->property[$keyVar] = $valVar
func BuildPropertyAssignInLoopPattern(keyVar, valVar string) *regexp.Regexp {
	return regexp.MustCompile(`\$this->(\w+)\[\$` + keyVar + `\]\s*=\s*\$` + valVar)
}
