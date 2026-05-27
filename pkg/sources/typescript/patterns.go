// Package typescript provides centralized TypeScript patterns for semantic analysis
package typescript

import "regexp"

// =============================================================================
// SEMANTIC ANALYSIS PATTERNS
// Used by semantic analyzers for flow analysis
// =============================================================================

var (
	// ThisPropertyAssignPattern matches this.property = ...
	// Used to detect constructor parameter flow to properties
	ThisPropertyAssignPattern = regexp.MustCompile(`this\.(\w+)\s*=`)

	// BracketKeyAccessPattern matches ['key'], ["key"], or [`key`] (template literal)
	// Used to extract keys from bracket notation including template literals
	BracketKeyAccessPattern = regexp.MustCompile(`\[['"\x60](\w+)['"\x60]\]`)

	// RequestPropertyChainPattern matches .body.prop, .query.prop, .params.prop, etc.
	// Used to extract nested property access on request objects
	RequestPropertyChainPattern = regexp.MustCompile(`\.(body|query|params|headers|cookies)\.(\w+)`)

	// DecoratorPatternPrefix is the prefix for TypeScript/NestJS decorators
	DecoratorPatternPrefix = `@`
)

// BuildThisPropertyAssignPattern creates a pattern for this.property = ... paramName
func BuildThisPropertyAssignPattern(paramName string) *regexp.Regexp {
	return regexp.MustCompile(`this\.(\w+)\s*=.*\b` + regexp.QuoteMeta(paramName) + `\b`)
}

// BuildPropertyPattern creates a pattern for .propertyName with word boundary
func BuildPropertyPattern(pattern string) *regexp.Regexp {
	return regexp.MustCompile(`\.` + pattern + `\b`)
}

// BuildDecoratorPattern creates a pattern for @decoratorName(
func BuildDecoratorPattern(decoratorPattern string) *regexp.Regexp {
	return regexp.MustCompile(DecoratorPatternPrefix + decoratorPattern + `\(`)
}

// ExtractBracketKey extracts the key from bracket notation ['key'], ["key"], or [`key`]
func ExtractBracketKey(expr string) string {
	matches := BracketKeyAccessPattern.FindStringSubmatch(expr)
	if len(matches) > 1 {
		return matches[1]
	}
	return ""
}

// ExtractRequestChainProperty extracts the property from request.body.prop style access
// Returns (category, property) e.g., ("body", "userId")
func ExtractRequestChainProperty(expr string) (string, string) {
	matches := RequestPropertyChainPattern.FindStringSubmatch(expr)
	if len(matches) > 2 {
		return matches[1], matches[2]
	}
	return "", ""
}
