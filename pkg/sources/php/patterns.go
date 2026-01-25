// Package php provides centralized PHP patterns for semantic analysis
// All PHP-specific regex patterns should be defined here to avoid duplication
package php

import (
	"regexp"
	"strings"
)

// =============================================================================
// UNIVERSAL INPUT DETECTION PATTERNS
// These patterns detect input sources across ANY PHP framework generically
// =============================================================================

var (
	// InputMethodPattern matches method names that ALWAYS indicate user input
	// Pattern matches:
	// - Explicit input getters: input, get_input, getInput, get_var, variable
	// - HTTP method getters: getPost, getQuery, getCookie, getHeader, etc.
	// - PSR-7 methods: getQueryParams, getParsedBody, getCookieParams, etc.
	// - All input: all()
	InputMethodPattern = regexp.MustCompile(`(?i)^(get_?)?(input|var|variable|query_?params?|parsed_?body|cookie_?params?|server_?params?|uploaded_?files?|headers?|all)$|^(get_?)?(post|cookie|param)s?$`)

	// InputPropertyPattern matches property names that typically hold user input
	// (for array access patterns like ->input['key'])
	// Matches: input, request, params, query, cookies, headers, body, data, args, post, get, files, server
	InputPropertyPattern = regexp.MustCompile(`(?i)^(input|request|params?|query|cookies?|headers?|body|data|args?|post|get|files?|server|attributes?|payload)s?$`)

	// InputObjectPattern matches object/variable names that suggest the object is an input carrier
	// Also matches chain calls like "->getRequest()" or "Factory::getApplication()->getInput()"
	InputObjectPattern = regexp.MustCompile(`(?i)(request|input|req|params?|http|ctx|context|getRequest\(\)|getApplication\(\))`)

	// ExcludeMethodPattern matches method names to EXCLUDE from input detection (false positive prevention)
	// These are methods that might match patterns but aren't typically user input
	ExcludeMethodPattern = regexp.MustCompile(`(?i)^(getData|getBody|getContent|fetch|find|load|read)$`)

	// ContextDependentMethodPattern matches methods like getVal, getText, getInt, getBool
	// used in MediaWiki on request objects but also on many other objects
	// Only detect these when the object looks like a request
	ContextDependentMethodPattern = regexp.MustCompile(`(?i)^(get_?)?(val|text|int|bool|array|raw_?val|check)$`)
)

// =============================================================================
// SQL EMBEDDED PATTERNS
// Used for extracting expressions embedded in SQL strings
// =============================================================================

var (
	// SQLCurlyBracePattern matches '{$var->prop['key']}' - curly brace interpolation in SQL
	SQLCurlyBracePattern = regexp.MustCompile(`\{\s*\$(\w+)->(\w+)\s*\[\s*['"]([^'"]+)['"]\s*\]\s*\}`)

	// SQLSimpleCurlyPattern matches simple property in curly braces {$var->prop}
	SQLSimpleCurlyPattern = regexp.MustCompile(`\{\s*\$(\w+)->(\w+)\s*\}`)

	// SQLNoCurlyPattern matches "...$var->prop..." without curly braces in strings
	SQLNoCurlyPattern = regexp.MustCompile(`"\s*[^"]*\$(\w+)->(\w+)\s*\[\s*['"]([^'"]+)['"]\s*\]`)
)

// =============================================================================
// STRING CONCATENATION PATTERNS
// Used for extracting expressions from string concatenation
// =============================================================================

var (
	// ConcatPattern matches "' . $var->prop['key'] . '" or similar concatenations
	ConcatPattern = regexp.MustCompile(`\.\s*\$(\w+)->(\w+)\s*\[\s*['"]([^'"]+)['"]\s*\]\s*\.`)

	// SimpleConcatPattern matches simple property concatenation '. $var->prop .'
	SimpleConcatPattern = regexp.MustCompile(`\.\s*\$(\w+)->(\w+)\s*\.`)
)

// =============================================================================
// ESCAPE FUNCTION PATTERNS
// Used for extracting expressions wrapped in escape functions
// =============================================================================

var (
	// EscapeWithPropArrayPattern matches escape_string($var->prop['key']) or $db->escape_string($var->prop['key'])
	EscapeWithPropArrayPattern = regexp.MustCompile(`(\w*escape\w*)\s*\(\s*\$(\w+)->(\w+)\s*\[\s*['"]([^'"]+)['"]\s*\]\s*\)`)

	// EscapeSimplePropPattern matches escape functions with simple property
	EscapeSimplePropPattern = regexp.MustCompile(`(\w*escape\w*)\s*\(\s*\$(\w+)->(\w+)\s*\)`)

	// EscapeVarPattern matches escape with variable
	EscapeVarPattern = regexp.MustCompile(`(\w*escape\w*)\s*\(\s*\$(\w+)\s*\)`)
)

// =============================================================================
// GLOBALS AND DI PATTERNS
// Used in symbolic execution
// =============================================================================

var (
	// GlobalsPattern matches $GLOBALS['varname'] or $GLOBALS["varname"]
	GlobalsPattern = regexp.MustCompile(`\$GLOBALS\[['"](\w+)['"]\]`)

	// DIContainerPattern matches DI container pattern: $var->get('service')
	DIContainerPattern = regexp.MustCompile(`\$\w+->get\(['"]([^'"]+)['"]\)`)
)

// =============================================================================
// PHPDOC TYPE HINT PATTERNS
// Used for extracting type hints from PHPDoc comments
// =============================================================================

// TypeHintPatterns returns patterns for PHPDoc @var type hints
// Pattern 1: /* @var $varname \namespace\classname */
// Pattern 2: /* @var \namespace\classname $varname */
func GetTypeHintPatterns(varName string) []*regexp.Regexp {
	return []*regexp.Regexp{
		// /* @var $request \phpbb\request\request_interface */
		regexp.MustCompile(`@var\s+\$` + regexp.QuoteMeta(varName) + `\s+\\?([\w\\]+)`),
		// /* @var \phpbb\request\request_interface $request */
		regexp.MustCompile(`@var\s+\\?([\w\\]+)\s+\$` + regexp.QuoteMeta(varName)),
	}
}

// =============================================================================
// SYMBOLIC EXECUTION PATTERNS
// Used in symbolic tracing through objects
// =============================================================================

var (
	// ThisMethodCallPattern matches $this->methodName($arg)
	ThisMethodCallPattern = regexp.MustCompile(`\$this->(\w+)\(([^)]*)\)`)

	// PropertyAssignLoopPattern builds a pattern for $this->property[$key] = $val
	// Use BuildPropertyAssignLoopPattern for dynamic keys
	PropertyAssignLoopPatternTemplate = `\$this->(%s)\[\$%s\]\s*=\s*\$%s`

	// ForeachPattern matches foreach($array as $key => $val)
	ForeachPattern = regexp.MustCompile(`foreach\s*\(\s*\$(\w+)\s+as\s+\$(\w+)\s*=>\s*\$(\w+)\s*\)`)

	// DirectAssignPatternTemplate for $this->property = $something
	// Use BuildDirectAssignPattern for specific properties
	DirectAssignPatternTemplate = `\$this->%s\s*=\s*([^;]+)`

	// ConditionalPatternTemplate for if($_SUPERGLOBAL[anything])
	// Use BuildConditionalPattern for specific superglobals
	ConditionalPatternTemplate = `if\s*\(\s*%s\[['"]?(\w+)['"]?\]`
)

// BuildPropertyAssignLoopPattern creates a pattern for property assignment in loop
func BuildPropertyAssignLoopPattern(propertyName, keyVar, valVar string) *regexp.Regexp {
	pattern := `\$this->` + regexp.QuoteMeta(propertyName) + `\[\$` + keyVar + `\]\s*=\s*\$` + valVar
	return regexp.MustCompile(pattern)
}

// BuildDirectAssignPattern creates a pattern for direct property assignment
func BuildDirectAssignPattern(propertyName string) *regexp.Regexp {
	pattern := `\$this->` + regexp.QuoteMeta(propertyName) + `\s*=\s*([^;]+)`
	return regexp.MustCompile(pattern)
}

// BuildConditionalPattern creates a pattern for conditional based on superglobal
func BuildConditionalPattern(superglobal string) *regexp.Regexp {
	pattern := `if\s*\(\s*` + regexp.QuoteMeta(superglobal) + `\[['"]?(\w+)['"]?\]`
	return regexp.MustCompile(pattern)
}

// =============================================================================
// HELPER FUNCTIONS
// =============================================================================

// IsInputMethod returns true if the method name matches input method patterns
// and is not excluded (false positive prevention)
func IsInputMethod(methodName string) bool {
	return InputMethodPattern.MatchString(methodName) && !ExcludeMethodPattern.MatchString(methodName)
}

// IsInputProperty returns true if the property name matches input property patterns
func IsInputProperty(propName string) bool {
	return InputPropertyPattern.MatchString(propName)
}

// IsInputObject returns true if the object/variable name suggests an input carrier
func IsInputObject(objName string) bool {
	return InputObjectPattern.MatchString(objName)
}

// IsExcludedMethod returns true if the method should be excluded from input detection
func IsExcludedMethod(methodName string) bool {
	return ExcludeMethodPattern.MatchString(methodName)
}

// IsContextDependentMethod returns true if the method needs object context to determine if it's input
func IsContextDependentMethod(methodName string) bool {
	return ContextDependentMethodPattern.MatchString(methodName)
}

// GetInputConfidence returns confidence score based on method context
func GetInputConfidence(methodName string, objectName string) float64 {
	// High confidence for explicit input methods
	if InputMethodPattern.MatchString(methodName) {
		return 0.95
	}
	// Medium confidence for context-dependent methods on request objects
	if ContextDependentMethodPattern.MatchString(methodName) && InputObjectPattern.MatchString(objectName) {
		return 0.8
	}
	return 0.5
}

// MatchesInputCarrier returns true if the expression matches patterns suggesting user input
// This checks object name, property name, and method name combinations
func MatchesInputCarrier(objName, propOrMethodName string, isMethod bool) bool {
	// Check if object itself is an input carrier
	if IsInputObject(objName) {
		return true
	}

	// For methods, check if it's an input method (not excluded)
	if isMethod {
		if IsExcludedMethod(propOrMethodName) {
			return false
		}
		if IsInputMethod(propOrMethodName) {
			return true
		}
		// Context-dependent methods only match if object looks like input
		if IsContextDependentMethod(propOrMethodName) && IsInputObject(objName) {
			return true
		}
		return false
	}

	// For properties, check if it's an input property
	return IsInputProperty(propOrMethodName)
}

// ExtractSQLEmbeddedExpressions extracts expressions from SQL strings
func ExtractSQLEmbeddedExpressions(line string) []SQLEmbeddedMatch {
	var results []SQLEmbeddedMatch

	// Pattern 1: Curly brace interpolation
	for _, match := range SQLCurlyBracePattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 4 {
			results = append(results, SQLEmbeddedMatch{
				VarName:      match[1],
				PropertyName: match[2],
				Key:          match[3],
				Pattern:      "curly_brace",
			})
		}
	}

	// Pattern 2: Simple curly
	for _, match := range SQLSimpleCurlyPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 3 {
			results = append(results, SQLEmbeddedMatch{
				VarName:      match[1],
				PropertyName: match[2],
				Pattern:      "simple_curly",
			})
		}
	}

	// Pattern 3: No curly (direct interpolation)
	for _, match := range SQLNoCurlyPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 4 {
			// Check not already found
			found := false
			for _, r := range results {
				if r.VarName == match[1] && r.PropertyName == match[2] && r.Key == match[3] {
					found = true
					break
				}
			}
			if !found {
				results = append(results, SQLEmbeddedMatch{
					VarName:      match[1],
					PropertyName: match[2],
					Key:          match[3],
					Pattern:      "no_curly",
				})
			}
		}
	}

	return results
}

// SQLEmbeddedMatch represents a matched SQL embedded expression
type SQLEmbeddedMatch struct {
	VarName      string
	PropertyName string
	Key          string
	Pattern      string
}

// ExtractConcatenatedExpressions extracts expressions from string concatenation
func ExtractConcatenatedExpressions(line string) []ConcatMatch {
	var results []ConcatMatch

	// Pattern 1: With array key
	for _, match := range ConcatPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 4 {
			results = append(results, ConcatMatch{
				VarName:      match[1],
				PropertyName: match[2],
				Key:          match[3],
			})
		}
	}

	// Pattern 2: Simple property
	for _, match := range SimpleConcatPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 3 {
			results = append(results, ConcatMatch{
				VarName:      match[1],
				PropertyName: match[2],
			})
		}
	}

	return results
}

// ConcatMatch represents a matched concatenated expression
type ConcatMatch struct {
	VarName      string
	PropertyName string
	Key          string
}

// ExtractEscapedExpressions extracts expressions wrapped in escape functions
func ExtractEscapedExpressions(line string) []EscapeMatch {
	var results []EscapeMatch

	// Pattern 1: With property array
	for _, match := range EscapeWithPropArrayPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 5 {
			results = append(results, EscapeMatch{
				EscapeFunc:   match[1],
				VarName:      match[2],
				PropertyName: match[3],
				Key:          match[4],
			})
		}
	}

	// Pattern 2: Simple property
	for _, match := range EscapeSimplePropPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 4 {
			results = append(results, EscapeMatch{
				EscapeFunc:   match[1],
				VarName:      match[2],
				PropertyName: match[3],
			})
		}
	}

	// Pattern 3: Variable only
	for _, match := range EscapeVarPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 3 {
			results = append(results, EscapeMatch{
				EscapeFunc: match[1],
				VarName:    match[2],
			})
		}
	}

	return results
}

// EscapeMatch represents a matched escaped expression
type EscapeMatch struct {
	EscapeFunc   string
	VarName      string
	PropertyName string
	Key          string
}

// ContainsSuperglobal checks if text contains any PHP superglobal
func ContainsSuperglobal(text string) (bool, string) {
	superglobals := []string{"$_GET", "$_POST", "$_COOKIE", "$_REQUEST", "$_SERVER", "$_FILES", "$_SESSION", "$_ENV"}
	lower := strings.ToLower(text)
	for _, sg := range superglobals {
		if strings.Contains(lower, strings.ToLower(sg)) {
			return true, sg
		}
	}
	return false, ""
}
