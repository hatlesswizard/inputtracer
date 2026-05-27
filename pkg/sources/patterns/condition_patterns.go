// Package patterns provides centralized regex patterns for code analysis.
package patterns

import "regexp"

// ConditionLinePatterns matches lines containing condition statements
var ConditionLinePatterns = []*regexp.Regexp{
	regexp.MustCompile(`^\s*if\s*\(`),            // if (condition)
	regexp.MustCompile(`^\s*if\s+[^(].*:`),       // Python: if condition:
	regexp.MustCompile(`^\s*}\s*else\s*if\s*\(`), // } else if (
	regexp.MustCompile(`^\s*else\s*if\s*\(`),     // else if (
	regexp.MustCompile(`^\s*elif\s+`),            // Python elif
	regexp.MustCompile(`^\s*elseif\s*\(`),        // PHP elseif
	regexp.MustCompile(`^\s*}\s*elseif\s*\(`),    // } elseif (
	regexp.MustCompile(`\?\s*.*\s*:`),            // Ternary
	regexp.MustCompile(`^\s*switch\s*\(`),        // switch (
	regexp.MustCompile(`^\s*case\s+`),            // case
}

// ConditionExpressionPatterns extract condition expressions from code
var ConditionExpressionPatterns = map[string]*regexp.Regexp{
	"if_paren":    regexp.MustCompile(`if\s*\((.+)\)\s*[{:]?`),
	"if_python":   regexp.MustCompile(`if\s+(.+?)\s*:\s*$`),
	"elif_python": regexp.MustCompile(`elif\s+(.+?)\s*:\s*$`),
	"elseif":      regexp.MustCompile(`(?:else\s*if|elseif)\s*\((.+)\)\s*[{:]?`),
	"switch":      regexp.MustCompile(`switch\s*\((.+?)\)\s*{?`),
	"case":        regexp.MustCompile(`case\s+(.+?)\s*:`),
	"ternary":     regexp.MustCompile(`(.+?)\s*\?\s*.+\s*:`),
}

// NullCheckPattern matches null/empty check expressions
var NullCheckPattern = regexp.MustCompile(`(?i)(isset|empty|is_null|null|\bnil\b|undefined)`)

// TypeCheckPattern matches type check expressions
var TypeCheckPattern = regexp.MustCompile(`(?i)(is_string|is_int|is_array|instanceof|typeof)`)

// LengthCheckPattern matches length/count check expressions
var LengthCheckPattern = regexp.MustCompile(`(?i)(strlen|length|count|size)\s*\(`)

// ComparisonPattern matches comparison operators
var ComparisonPattern = regexp.MustCompile(`[<>=!]=?`)

// LogicalOperatorPattern matches logical operators
var LogicalOperatorPattern = regexp.MustCompile(`(&&|\|\||!|and|or|not)`)

// IsConditionLine checks if a line matches any condition pattern
func IsConditionLine(line string) bool {
	for _, p := range ConditionLinePatterns {
		if p.MatchString(line) {
			return true
		}
	}
	return false
}

// ExtractConditionExpression extracts the condition from a line
func ExtractConditionExpression(line string) string {
	for _, re := range ConditionExpressionPatterns {
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			return matches[1]
		}
	}
	return ""
}
