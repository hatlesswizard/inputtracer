// Package extractor provides utilities to extract traceable PHP expressions from code snippets
package extractor

import (
	"regexp"
	"strings"
)

// ExtractedExpression represents a PHP expression extracted from code
type ExtractedExpression struct {
	Expression    string   `json:"expression"`
	Line          int      `json:"line"`
	Type          string   `json:"type"` // "superglobal", "method_call", "property_access", "variable", "sql_embedded"
	Context       string   `json:"context,omitempty"`
	Key           string   `json:"key,omitempty"`           // The extracted key (e.g., "aid" from $mybb->input['aid'])
	VarName       string   `json:"var_name,omitempty"`      // Variable name (e.g., "mybb")
	PropertyName  string   `json:"property_name,omitempty"` // Property name (e.g., "input")
	IsEscaped     bool     `json:"is_escaped,omitempty"`    // Whether wrapped in escape_string()
	EscapeWrapper string   `json:"escape_wrapper,omitempty"` // The escape function used
}

// ExpressionExtractor extracts traceable PHP expressions from code
type ExpressionExtractor struct {
	// Pre-compiled patterns
	superglobalPattern   *regexp.Regexp
	methodCallPattern    *regexp.Regexp
	propertyPattern      *regexp.Regexp
	propertyArrayPattern *regexp.Regexp

	// New SQL-specific patterns
	sqlInterpolationPattern   *regexp.Regexp // '{$mybb->input['aid']}'
	sqlConcatPattern          *regexp.Regexp // "' . $mybb->input['aid'] . '"
	escapedExprPattern        *regexp.Regexp // escape_string($mybb->input['title'])
	getInputMethodPattern     *regexp.Regexp // $mybb->get_input('aid')
	curlyBraceInterpolation   *regexp.Regexp // {$var->prop['key']} in double-quoted strings
}

// New creates a new ExpressionExtractor
func New() *ExpressionExtractor {
	return &ExpressionExtractor{
		// Superglobals: $_GET['key'], $_POST['key'], etc.
		superglobalPattern: regexp.MustCompile(`\$_(GET|POST|COOKIE|REQUEST|SERVER|FILES|SESSION|ENV)\[['"]([\w\-]+)['"]\]`),

		// Method calls: $obj->method(...) - captures variable and method name
		methodCallPattern: regexp.MustCompile(`\$(\w+)->(\w+)\s*\(`),

		// Property with array access: $obj->property['key']
		propertyArrayPattern: regexp.MustCompile(`\$(\w+)->(\w+)\[['"]([\w\-]+)['"]\]`),

		// Simple property: $obj->property (not followed by [ or ()
		propertyPattern: regexp.MustCompile(`\$(\w+)->(\w+)(?:[^\[\(]|$)`),
	}
}

// ExtractFromCode extracts all traceable expressions from PHP code lines
func (e *ExpressionExtractor) ExtractFromCode(lines []string) []ExtractedExpression {
	var results []ExtractedExpression
	seen := make(map[string]bool) // Deduplicate

	for lineNum, line := range lines {
		// Extract from this line
		extracted := e.extractFromLine(line, lineNum+1)
		for _, expr := range extracted {
			if !seen[expr.Expression] {
				seen[expr.Expression] = true
				results = append(results, expr)
			}
		}
	}

	return results
}

// ExtractFromDiffContext extracts expressions from diff context (handles +/- prefixes)
func (e *ExpressionExtractor) ExtractFromDiffContext(context []string) []ExtractedExpression {
	var results []ExtractedExpression
	seen := make(map[string]bool)

	for lineNum, line := range context {
		// Strip diff prefix if present
		cleanLine := line
		if strings.HasPrefix(line, "+") || strings.HasPrefix(line, "-") || strings.HasPrefix(line, " ") {
			cleanLine = line[1:]
		}

		extracted := e.extractFromLine(cleanLine, lineNum+1)
		for _, expr := range extracted {
			if !seen[expr.Expression] {
				seen[expr.Expression] = true
				results = append(results, expr)
			}
		}
	}

	return results
}

// extractFromLine extracts expressions from a single line
func (e *ExpressionExtractor) extractFromLine(line string, lineNum int) []ExtractedExpression {
	var results []ExtractedExpression

	// Skip comments, empty lines, and non-code
	trimmed := strings.TrimSpace(line)
	if trimmed == "" || strings.HasPrefix(trimmed, "//") || strings.HasPrefix(trimmed, "/*") ||
		strings.HasPrefix(trimmed, "*") || strings.HasPrefix(trimmed, "<?") ||
		strings.HasPrefix(trimmed, "define(") || strings.HasPrefix(trimmed, "require") ||
		strings.HasPrefix(trimmed, "include") {
		return results
	}

	// 1. Extract superglobals: $_GET['key'], $_POST['key'], etc.
	for _, match := range e.superglobalPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 3 {
			expr := "$_" + match[1] + "['" + match[2] + "']"
			results = append(results, ExtractedExpression{
				Expression: expr,
				Line:       lineNum,
				Type:       "superglobal",
				Context:    trimmed,
			})
		}
	}

	// 2. Extract property with array access: $obj->property['key']
	for _, match := range e.propertyArrayPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 4 {
			expr := "$" + match[1] + "->" + match[2] + "['" + match[3] + "']"
			results = append(results, ExtractedExpression{
				Expression: expr,
				Line:       lineNum,
				Type:       "property_access",
				Context:    trimmed,
			})
		}
	}

	// 3. Extract method calls using smart extraction
	methodResults := e.extractMethodCalls(line, lineNum, trimmed)
	results = append(results, methodResults...)

	return results
}

// extractMethodCalls extracts complete method call expressions with their arguments
func (e *ExpressionExtractor) extractMethodCalls(line string, lineNum int, context string) []ExtractedExpression {
	var results []ExtractedExpression

	// Find all method call starts: $var->method(
	for _, match := range e.methodCallPattern.FindAllStringSubmatchIndex(line, -1) {
		if len(match) >= 6 {
			// match[0]-match[1] is the full match
			// match[2]-match[3] is the var name
			// match[4]-match[5] is the method name

			varName := line[match[2]:match[3]]
			methodName := line[match[4]:match[5]]

			// Now find the complete method call including arguments
			startIdx := match[0]
			openParenIdx := strings.Index(line[startIdx:], "(")
			if openParenIdx == -1 {
				continue
			}

			// Find matching close paren
			afterParen := line[startIdx+openParenIdx+1:]
			closeParenIdx, ok := findMatchingParen(afterParen)
			if !ok {
				continue
			}

			// Extract the full expression
			endIdx := startIdx + openParenIdx + 1 + closeParenIdx + 1
			fullExpr := line[startIdx:endIdx]

			// Skip database query methods as they're usually sinks, not sources
			// But include input-related methods
			if isInputMethod(varName, methodName) || isInterestingMethod(methodName) {
				results = append(results, ExtractedExpression{
					Expression: fullExpr,
					Line:       lineNum,
					Type:       "method_call",
					Context:    context,
				})
			}
		}
	}

	return results
}

// findMatchingParen finds the matching closing paren, respecting nesting and strings
func findMatchingParen(s string) (int, bool) {
	depth := 1
	inString := false
	var stringChar byte = 0

	for i := 0; i < len(s); i++ {
		c := s[i]

		// Handle escape sequences in strings
		if inString && c == '\\' && i+1 < len(s) {
			i++
			continue
		}

		// Handle string boundaries
		if c == '"' || c == '\'' {
			if !inString {
				inString = true
				stringChar = c
			} else if c == stringChar {
				inString = false
				stringChar = 0
			}
			continue
		}

		// Track parenthesis depth (only outside strings)
		if !inString {
			if c == '(' {
				depth++
			} else if c == ')' {
				depth--
				if depth == 0 {
					return i, true
				}
			}
		}
	}

	return -1, false
}

// isInputMethod checks if this method is known to return user input
func isInputMethod(varName, methodName string) bool {
	// MyBB patterns
	if varName == "mybb" {
		switch methodName {
		case "get_input", "get_input_array", "input", "cookies":
			return true
		}
	}

	// Request patterns
	if varName == "request" {
		switch methodName {
		case "variable", "get", "post", "cookie", "server", "header":
			return true
		}
	}

	// Generic patterns
	switch methodName {
	case "get", "post", "cookie", "header", "param", "input", "request", "query":
		return true
	}

	return false
}

// isInterestingMethod checks if this method is worth tracing for security
func isInterestingMethod(methodName string) bool {
	// Database methods that might use user input
	switch methodName {
	case "query", "simple_select", "write_query", "delete_query", "update_query",
		"escape_string", "prepare", "execute", "fetch", "fetch_array", "fetch_field":
		return true
	}

	// File operations
	switch methodName {
	case "read", "write", "file_get_contents", "include", "require", "fopen":
		return true
	}

	// Command execution
	switch methodName {
	case "exec", "shell_exec", "system", "passthru":
		return true
	}

	return false
}

// ExtractFromSnippet extracts ALL traceable expressions from a PHP code snippet
// This is the main entry point for snippet-only analysis
func (e *ExpressionExtractor) ExtractFromSnippet(snippet string) []ExtractedExpression {
	var results []ExtractedExpression
	seen := make(map[string]bool)

	// Split snippet into lines
	lines := strings.Split(snippet, "\n")

	for lineNum, line := range lines {
		// 1. Extract superglobals directly
		superglobals := e.extractSuperglobals(line, lineNum+1)
		for _, expr := range superglobals {
			if !seen[expr.Expression] {
				seen[expr.Expression] = true
				results = append(results, expr)
			}
		}

		// 2. Extract property access with array keys (e.g., $mybb->input['aid'])
		propAccess := e.extractPropertyArrayAccess(line, lineNum+1)
		for _, expr := range propAccess {
			if !seen[expr.Expression] {
				seen[expr.Expression] = true
				results = append(results, expr)
			}
		}

		// 3. Extract method calls (e.g., $mybb->get_input('aid'))
		methodCalls := e.extractAllMethodCalls(line, lineNum+1)
		for _, expr := range methodCalls {
			if !seen[expr.Expression] {
				seen[expr.Expression] = true
				results = append(results, expr)
			}
		}

		// 4. Extract expressions from SQL strings (interpolated variables)
		sqlEmbedded := e.extractSQLEmbedded(line, lineNum+1)
		for _, expr := range sqlEmbedded {
			if !seen[expr.Expression] {
				seen[expr.Expression] = true
				results = append(results, expr)
			}
		}

		// 5. Extract expressions from concatenation
		concatExprs := e.extractConcatenated(line, lineNum+1)
		for _, expr := range concatExprs {
			if !seen[expr.Expression] {
				seen[expr.Expression] = true
				results = append(results, expr)
			}
		}

		// 6. Extract expressions wrapped in escape functions
		escapedExprs := e.extractEscaped(line, lineNum+1)
		for _, expr := range escapedExprs {
			if !seen[expr.Expression] {
				seen[expr.Expression] = true
				results = append(results, expr)
			}
		}
	}

	return results
}

// extractSuperglobals extracts PHP superglobal expressions
func (e *ExpressionExtractor) extractSuperglobals(line string, lineNum int) []ExtractedExpression {
	var results []ExtractedExpression

	// Pattern: $_GET['key'], $_POST['key'], etc.
	pattern := regexp.MustCompile(`\$_(GET|POST|COOKIE|REQUEST|SERVER|FILES|SESSION|ENV)\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`)

	for _, match := range pattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 3 {
			sgType := match[1]
			key := match[2]
			expr := "$_" + sgType + "['" + key + "']"

			results = append(results, ExtractedExpression{
				Expression:   expr,
				Line:         lineNum,
				Type:         "superglobal",
				Key:          key,
				PropertyName: sgType,
				Context:      strings.TrimSpace(line),
			})
		}
	}

	return results
}

// extractPropertyArrayAccess extracts $obj->property['key'] patterns
func (e *ExpressionExtractor) extractPropertyArrayAccess(line string, lineNum int) []ExtractedExpression {
	var results []ExtractedExpression

	// Pattern: $var->property['key'] or $var->property["key"]
	pattern := regexp.MustCompile(`\$(\w+)->(\w+)\s*\[\s*['"]([^'"]+)['"]\s*\]`)

	for _, match := range pattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 4 {
			varName := match[1]
			propName := match[2]
			key := match[3]
			expr := "$" + varName + "->" + propName + "['" + key + "']"

			results = append(results, ExtractedExpression{
				Expression:   expr,
				Line:         lineNum,
				Type:         "property_access",
				VarName:      varName,
				PropertyName: propName,
				Key:          key,
				Context:      strings.TrimSpace(line),
			})
		}
	}

	return results
}

// extractAllMethodCalls extracts all method call patterns
func (e *ExpressionExtractor) extractAllMethodCalls(line string, lineNum int) []ExtractedExpression {
	var results []ExtractedExpression

	// Pattern: $var->method('arg') or $var->method("arg")
	pattern := regexp.MustCompile(`\$(\w+)->(\w+)\s*\(\s*['"]([^'"]*)['"]\s*(?:,\s*[^)]+)?\s*\)`)

	for _, match := range pattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 4 {
			varName := match[1]
			methodName := match[2]
			arg := match[3]
			expr := "$" + varName + "->" + methodName + "('" + arg + "')"

			results = append(results, ExtractedExpression{
				Expression:   expr,
				Line:         lineNum,
				Type:         "method_call",
				VarName:      varName,
				PropertyName: methodName, // Using PropertyName to store method name
				Key:          arg,
				Context:      strings.TrimSpace(line),
			})
		}
	}

	return results
}

// extractSQLEmbedded extracts expressions embedded in SQL strings
func (e *ExpressionExtractor) extractSQLEmbedded(line string, lineNum int) []ExtractedExpression {
	var results []ExtractedExpression

	// Pattern 1: '{$var->prop['key']}' - curly brace interpolation in SQL
	curlyPattern := regexp.MustCompile(`\{\s*\$(\w+)->(\w+)\s*\[\s*['"]([^'"]+)['"]\s*\]\s*\}`)

	for _, match := range curlyPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 4 {
			varName := match[1]
			propName := match[2]
			key := match[3]
			expr := "$" + varName + "->" + propName + "['" + key + "']"

			results = append(results, ExtractedExpression{
				Expression:   expr,
				Line:         lineNum,
				Type:         "sql_embedded",
				VarName:      varName,
				PropertyName: propName,
				Key:          key,
				Context:      strings.TrimSpace(line),
			})
		}
	}

	// Pattern 2: Simple property in curly braces {$var->prop}
	simpleCurlyPattern := regexp.MustCompile(`\{\s*\$(\w+)->(\w+)\s*\}`)

	for _, match := range simpleCurlyPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 3 {
			varName := match[1]
			propName := match[2]
			expr := "$" + varName + "->" + propName

			results = append(results, ExtractedExpression{
				Expression:   expr,
				Line:         lineNum,
				Type:         "sql_embedded",
				VarName:      varName,
				PropertyName: propName,
				Context:      strings.TrimSpace(line),
			})
		}
	}

	// Pattern 3: "...$var->prop..." without curly braces
	noCurlyPattern := regexp.MustCompile(`"\s*[^"]*\$(\w+)->(\w+)\s*\[\s*['"]([^'"]+)['"]\s*\]`)

	for _, match := range noCurlyPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 4 {
			varName := match[1]
			propName := match[2]
			key := match[3]
			expr := "$" + varName + "->" + propName + "['" + key + "']"

			// Check if not already captured by curly pattern
			alreadyFound := false
			for _, r := range results {
				if r.Expression == expr {
					alreadyFound = true
					break
				}
			}

			if !alreadyFound {
				results = append(results, ExtractedExpression{
					Expression:   expr,
					Line:         lineNum,
					Type:         "sql_embedded",
					VarName:      varName,
					PropertyName: propName,
					Key:          key,
					Context:      strings.TrimSpace(line),
				})
			}
		}
	}

	return results
}

// extractConcatenated extracts expressions from string concatenation
func (e *ExpressionExtractor) extractConcatenated(line string, lineNum int) []ExtractedExpression {
	var results []ExtractedExpression

	// Pattern: "' . $var->prop['key'] . '" or similar concatenations
	concatPattern := regexp.MustCompile(`\.\s*\$(\w+)->(\w+)\s*\[\s*['"]([^'"]+)['"]\s*\]\s*\.`)

	for _, match := range concatPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 4 {
			varName := match[1]
			propName := match[2]
			key := match[3]
			expr := "$" + varName + "->" + propName + "['" + key + "']"

			results = append(results, ExtractedExpression{
				Expression:   expr,
				Line:         lineNum,
				Type:         "concatenated",
				VarName:      varName,
				PropertyName: propName,
				Key:          key,
				Context:      strings.TrimSpace(line),
			})
		}
	}

	// Pattern: Simple property concatenation '. $var->prop .'
	simpleConcatPattern := regexp.MustCompile(`\.\s*\$(\w+)->(\w+)\s*\.`)

	for _, match := range simpleConcatPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 3 {
			varName := match[1]
			propName := match[2]
			expr := "$" + varName + "->" + propName

			results = append(results, ExtractedExpression{
				Expression:   expr,
				Line:         lineNum,
				Type:         "concatenated",
				VarName:      varName,
				PropertyName: propName,
				Context:      strings.TrimSpace(line),
			})
		}
	}

	return results
}

// extractEscaped extracts expressions wrapped in escape functions
func (e *ExpressionExtractor) extractEscaped(line string, lineNum int) []ExtractedExpression {
	var results []ExtractedExpression

	// Pattern: escape_string($var->prop['key']) or $db->escape_string($var->prop['key'])
	escapePattern := regexp.MustCompile(`(\w*escape\w*)\s*\(\s*\$(\w+)->(\w+)\s*\[\s*['"]([^'"]+)['"]\s*\]\s*\)`)

	for _, match := range escapePattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 5 {
			escapeFunc := match[1]
			varName := match[2]
			propName := match[3]
			key := match[4]
			expr := "$" + varName + "->" + propName + "['" + key + "']"

			results = append(results, ExtractedExpression{
				Expression:    expr,
				Line:          lineNum,
				Type:          "escaped",
				VarName:       varName,
				PropertyName:  propName,
				Key:           key,
				IsEscaped:     true,
				EscapeWrapper: escapeFunc,
				Context:       strings.TrimSpace(line),
			})
		}
	}

	// Pattern: escape functions with simple property
	simpleEscapePattern := regexp.MustCompile(`(\w*escape\w*)\s*\(\s*\$(\w+)->(\w+)\s*\)`)

	for _, match := range simpleEscapePattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 4 {
			escapeFunc := match[1]
			varName := match[2]
			propName := match[3]
			expr := "$" + varName + "->" + propName

			results = append(results, ExtractedExpression{
				Expression:    expr,
				Line:          lineNum,
				Type:          "escaped",
				VarName:       varName,
				PropertyName:  propName,
				IsEscaped:     true,
				EscapeWrapper: escapeFunc,
				Context:       strings.TrimSpace(line),
			})
		}
	}

	// Pattern: escape with variable
	escapeVarPattern := regexp.MustCompile(`(\w*escape\w*)\s*\(\s*\$(\w+)\s*\)`)

	for _, match := range escapeVarPattern.FindAllStringSubmatch(line, -1) {
		if len(match) >= 3 {
			escapeFunc := match[1]
			varName := match[2]
			expr := "$" + varName

			results = append(results, ExtractedExpression{
				Expression:    expr,
				Line:          lineNum,
				Type:          "escaped",
				VarName:       varName,
				IsEscaped:     true,
				EscapeWrapper: escapeFunc,
				Context:       strings.TrimSpace(line),
			})
		}
	}

	return results
}
