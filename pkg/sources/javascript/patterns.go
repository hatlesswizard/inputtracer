// Package javascript provides centralized JavaScript patterns for semantic analysis
package javascript

import "regexp"

// =============================================================================
// SEMANTIC ANALYSIS PATTERNS
// Used by semantic analyzers for flow analysis
// =============================================================================

var (
	// MapGetPattern matches .get('key') or .get("key")
	// Used to extract keys from Map/object .get() calls
	MapGetPattern = regexp.MustCompile(`\.get\(['"](\w+)['"]\)`)

	// BracketPropertyPattern matches ['key'] or ["key"] at start of string
	// Used to extract property names from bracket notation
	BracketPropertyPattern = regexp.MustCompile(`^\[['"](\w+)['"]\]`)

	// DotPropertyPattern matches .property at start of string
	// Used to extract property names from dot notation
	DotPropertyPattern = regexp.MustCompile(`^\.(\w+)`)

	// ThisPropertyAssignPattern matches this.property = ...
	// Used to detect constructor parameter flow to properties
	ThisPropertyAssignPattern = regexp.MustCompile(`this\.(\w+)\s*=`)
)

// BuildThisPropertyAssignPattern creates a pattern for this.property = ... paramName
func BuildThisPropertyAssignPattern(paramName string) *regexp.Regexp {
	return regexp.MustCompile(`this\.(\w+)\s*=.*` + regexp.QuoteMeta(paramName))
}

// ExtractMapKey extracts the key from a .get('key') expression
func ExtractMapKey(expr string) string {
	matches := MapGetPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ExtractBracketKey extracts the key from bracket notation ['key']
func ExtractBracketKey(expr string) string {
	matches := BracketPropertyPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ExtractDotProperty extracts property name from .property notation
func ExtractDotProperty(expr string) string {
	matches := DotPropertyPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}
