// Package python provides centralized Python patterns for semantic analysis
package python

import "regexp"

// =============================================================================
// SEMANTIC ANALYSIS PATTERNS
// Used by semantic analyzers for flow analysis
// =============================================================================

var (
	// SelfPropertyAssignPattern matches self.property = ...
	// Used to detect __init__ parameter flow to properties
	SelfPropertyAssignPattern = regexp.MustCompile(`self\.(\w+)\s*=`)

	// DictKeyAccessPattern matches ['key'] or ["key"]
	// Used to extract dictionary keys from expressions
	DictKeyAccessPattern = regexp.MustCompile(`\[['"](\w+)['"]\]`)

	// DictGetPattern matches .get('key') or .get("key")
	// Used to extract keys from dict.get() calls
	DictGetPattern = regexp.MustCompile(`\.get\(['"](\w+)['"]\)`)
)

// BuildSelfPropertyAssignPattern creates a pattern for self.property = ... paramName
func BuildSelfPropertyAssignPattern(paramName string) *regexp.Regexp {
	return regexp.MustCompile(`self\.(\w+)\s*=.*\b` + regexp.QuoteMeta(paramName) + `\b`)
}

// BuildPropertyPattern creates a pattern to match a property pattern with word boundary
func BuildPropertyPattern(pattern string) *regexp.Regexp {
	return regexp.MustCompile(`\b` + pattern)
}

// BuildMethodCallPattern creates a pattern for .methodName(
func BuildMethodCallPattern(methodPattern string) *regexp.Regexp {
	return regexp.MustCompile(`\.` + methodPattern + `\(`)
}

// ExtractDictKey extracts the key from dict['key'] or dict["key"] expression
func ExtractDictKey(expr string) string {
	matches := DictKeyAccessPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ExtractDictGetKey extracts the key from dict.get('key') expression
func ExtractDictGetKey(expr string) string {
	matches := DictGetPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
