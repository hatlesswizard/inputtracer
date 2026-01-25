// Package patterns provides centralized regex patterns for code analysis.
package patterns

import "regexp"

// LanguageVariablePatterns provides language-specific patterns for extracting variables
var LanguageVariablePatterns = map[string][]*regexp.Regexp{
	"php": {
		regexp.MustCompile(`\$[a-zA-Z_][a-zA-Z0-9_]*`),
		regexp.MustCompile(`\$_[A-Z]+\s*\[\s*['"]([^'"]+)['"]\s*\]`),
	},
	"javascript": {
		regexp.MustCompile(`\b[a-zA-Z_$][a-zA-Z0-9_$]*\b`),
	},
	"typescript": {
		regexp.MustCompile(`\b[a-zA-Z_$][a-zA-Z0-9_$]*\b`),
	},
	"python": {
		regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`),
	},
	"go": {
		regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`),
	},
	"java": {
		regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`),
	},
	"c": {
		regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`),
	},
	"cpp": {
		regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`),
	},
	"c_sharp": {
		regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`),
	},
	"ruby": {
		regexp.MustCompile(`[@$]?[a-zA-Z_][a-zA-Z0-9_]*`),
	},
	"rust": {
		regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`),
	},
}

// DefaultVariablePattern is used when language is not recognized
var DefaultVariablePattern = regexp.MustCompile(`\b[a-zA-Z_$][a-zA-Z0-9_$]*\b`)

// GetVariablePatterns returns the variable patterns for a language
func GetVariablePatterns(language string) []*regexp.Regexp {
	if patterns, ok := LanguageVariablePatterns[language]; ok {
		return patterns
	}
	return []*regexp.Regexp{DefaultVariablePattern}
}
