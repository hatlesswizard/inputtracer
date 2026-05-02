package symbolic

import (
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/sources/patterns"
)

func (e *ExecutionEngine) parseExpression(expr string) *ParsedExpression {
	parsed := &ParsedExpression{
		Type:    ExprTypeUnknown,
		RawExpr: expr,
	}

	expr = strings.TrimSpace(expr)

	// GAP #4 FIX: Try chained expression parsing first
	// This handles expressions like: $obj->method()->property or $obj->method1()->method2('arg')
	if strings.Count(expr, "->") > 1 {
		if chainedParsed := e.parseChainedExpression(expr); chainedParsed != nil {
			return chainedParsed
		}
	}

	// GAP #1 FIX: Try superglobal pattern first: $_GET['key'], $_POST['key'], etc.
	// Pattern: $_SUPERGLOBAL['key'] or $_SUPERGLOBAL["key"]
	if matches := patterns.SuperglobalAccessPattern.FindStringSubmatch(expr); len(matches) >= 3 {
		parsed.Type = ExprTypeSuperglobal
		parsed.SuperglobalName = "$_" + matches[1]
		parsed.AccessKey = matches[2]
		parsed.IsSuperglobal = true
		parsed.VarName = parsed.SuperglobalName // For compatibility
		return parsed
	}

	// GAP #3 FIX: Try static method call pattern: Class::method('arg') with proper paren handling
	if className, methodName, argsStr, ok := e.extractStaticMethodCall(expr); ok {
		parsed.Type = ExprTypeStaticCall
		parsed.ClassName = className
		parsed.MethodName = methodName
		if argsStr != "" {
			parsed.Arguments = e.parseArguments(argsStr)
			if len(parsed.Arguments) > 0 {
				parsed.AccessKey = strings.Trim(parsed.Arguments[0], "'\"")
			}
		}
		return parsed
	}

	// GAP #3 FIX: Try static property/constant pattern: Class::$property or Class::CONSTANT
	if matches := patterns.StaticPropertyPattern.FindStringSubmatch(expr); len(matches) >= 3 {
		parsed.Type = ExprTypeStaticProperty
		parsed.ClassName = matches[1]
		parsed.PropertyName = matches[2]
		return parsed
	}

	// Try method call pattern using smart extraction that handles nested parens
	// This handles: $var->method('arg'), $var->method(func($x)), $var->method("string with (parens)")
	if varName, methodName, argsStr, ok := e.extractMethodCall(expr); ok {
		parsed.Type = ExprTypeMethodCall
		parsed.VarName = varName
		parsed.MethodName = methodName
		if argsStr != "" {
			parsed.Arguments = e.parseArguments(argsStr)
			if len(parsed.Arguments) > 0 {
				parsed.AccessKey = strings.Trim(parsed.Arguments[0], "'\"")
			}
		}
		return parsed
	}

	// Try property access pattern: $var->property or $var->property['key']
	if matches := patterns.PropertyAccessPattern.FindStringSubmatch(expr); len(matches) >= 3 {
		parsed.Type = ExprTypePropertyAccess
		parsed.VarName = "$" + matches[1]
		parsed.PropertyName = matches[2]
		if len(matches) >= 4 {
			parsed.AccessKey = matches[3]
		}
		return parsed
	}

	// GAP #2 FIX: Try simple local variable pattern: $varname
	// This must come LAST as it's the most generic pattern
	if matches := patterns.LocalVariablePattern.FindStringSubmatch(expr); len(matches) >= 2 {
		parsed.Type = ExprTypeLocalVariable
		parsed.VarName = "$" + matches[1]
		return parsed
	}

	return parsed
}

// parseChainedExpression parses expressions like $obj->method()->property or $obj->method1()->method2('arg')
// GAP #4 FIX: Support chained method calls
func (e *ExecutionEngine) parseChainedExpression(expr string) *ParsedExpression {
	expr = strings.TrimSpace(expr)

	// Must start with a variable
	if !strings.HasPrefix(expr, "$") {
		return nil
	}

	// Find the base variable name
	varNameEnd := strings.Index(expr, "->")
	if varNameEnd == -1 {
		return nil
	}

	basePart := expr[1:varNameEnd] // Remove $ prefix
	if !patterns.WordPattern.MatchString(basePart) {
		return nil
	}
	varName := "$" + basePart

	// Parse the chain steps
	remainder := expr[varNameEnd:]
	var steps []ChainStep

	// Patterns for parsing chain steps (property access only - method calls use extractChainMethodCall)
	propWithKeyPattern := patterns.ChainPropertyWithKeyPattern
	propPattern := patterns.ChainSimplePropertyPattern

	for len(remainder) > 0 && strings.HasPrefix(remainder, "->") {
		var step ChainStep
		matched := false

		// Try method call first: ->method(args) with proper paren handling
		if methodName, argsStr, consumed, ok := e.extractChainMethodCall(remainder); ok {
			step.Type = ExprTypeMethodCall
			step.Name = methodName
			if argsStr != "" {
				step.Arguments = e.parseArguments(argsStr)
				if len(step.Arguments) > 0 {
					step.AccessKey = strings.Trim(step.Arguments[0], "'\"")
				}
			}
			steps = append(steps, step)
			remainder = remainder[consumed:]
			matched = true
		}

		// Try property with key: ->property['key']
		if !matched {
			if matches := propWithKeyPattern.FindStringSubmatch(remainder); len(matches) >= 3 {
				step.Type = ExprTypePropertyAccess
				step.Name = matches[1]
				step.AccessKey = matches[2]
				steps = append(steps, step)
				remainder = remainder[len(matches[0]):]
				matched = true
			}
		}

		// Try simple property: ->property
		if !matched {
			if matches := propPattern.FindStringSubmatch(remainder); len(matches) >= 2 {
				step.Type = ExprTypePropertyAccess
				step.Name = matches[1]
				steps = append(steps, step)
				remainder = remainder[len(matches[0]):]
				matched = true
			}
		}

		// No pattern matched, fail
		if !matched {
			return nil
		}
	}

	// Must have at least 2 steps to be a chain
	if len(steps) < 2 {
		return nil
	}

	// We have leftover text that didn't match
	if len(strings.TrimSpace(remainder)) > 0 {
		return nil
	}

	// Build the parsed expression
	// For tracing purposes, we use the first step info
	parsed := &ParsedExpression{
		VarName:    varName,
		IsChained:  true,
		ChainSteps: steps,
	}

	// Set the type based on the first step
	if steps[0].Type == ExprTypeMethodCall {
		parsed.Type = ExprTypeMethodCall
		parsed.MethodName = steps[0].Name
		parsed.Arguments = steps[0].Arguments
		parsed.AccessKey = steps[0].AccessKey
	} else {
		parsed.Type = ExprTypePropertyAccess
		parsed.PropertyName = steps[0].Name
		parsed.AccessKey = steps[0].AccessKey
	}

	return parsed
}

// parseArguments splits method arguments with proper handling of nested parentheses and strings
// This handles complex cases like: "users", "uid='".$mybb->get_input('uid')."'"
func (e *ExecutionEngine) parseArguments(argsStr string) []string {
	var args []string
	var current strings.Builder
	depth := 0
	inString := false
	var stringChar byte = 0

	for i := 0; i < len(argsStr); i++ {
		c := argsStr[i]

		// Handle escape sequences in strings
		if inString && c == '\\' && i+1 < len(argsStr) {
			current.WriteByte(c)
			i++
			current.WriteByte(argsStr[i])
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
			current.WriteByte(c)
			continue
		}

		// Track parenthesis depth (only outside strings)
		if !inString {
			if c == '(' || c == '[' {
				depth++
			} else if c == ')' || c == ']' {
				depth--
			} else if c == ',' && depth == 0 {
				// Argument separator at top level
				arg := strings.TrimSpace(current.String())
				if arg != "" {
					args = append(args, arg)
				}
				current.Reset()
				continue
			}
		}

		current.WriteByte(c)
	}

	// Add the last argument
	if current.Len() > 0 {
		arg := strings.TrimSpace(current.String())
		if arg != "" {
			args = append(args, arg)
		}
	}

	return args
}

// extractMethodCall extracts $var->method(args) with proper handling of nested parentheses
// Returns: varName, methodName, argsStr, success
// This handles complex expressions like:
//   - $db->simple_select("users", "uid='".$mybb->get_input('uid')."'")
//   - $db->write_query("ALTER TABLE posts ADD INDEX (tid)")
//   - $db->escape_string(trim($input))
func (e *ExecutionEngine) extractMethodCall(expr string) (string, string, string, bool) {
	expr = strings.TrimSpace(expr)

	// Must start with $
	if !strings.HasPrefix(expr, "$") {
		return "", "", "", false
	}

	// Find the variable name (ends at ->)
	arrowIdx := strings.Index(expr, "->")
	if arrowIdx == -1 {
		return "", "", "", false
	}

	varName := expr[:arrowIdx]
	// Validate variable name: $word
	if !patterns.DollarVariablePattern.MatchString(varName) {
		return "", "", "", false
	}

	// Find the method name (from after -> to the opening paren)
	remainder := expr[arrowIdx+2:]
	parenIdx := strings.Index(remainder, "(")
	if parenIdx == -1 {
		return "", "", "", false // Not a method call (no parens)
	}

	methodName := remainder[:parenIdx]
	// Validate method name: word characters only
	if !patterns.WordPattern.MatchString(methodName) {
		return "", "", "", false
	}

	// Extract arguments by finding the matching closing paren
	argsStart := parenIdx + 1
	afterParen := remainder[argsStart:]

	// Find the matching closing paren, respecting nesting and strings
	argsEnd, ok := findMatchingParen(afterParen)
	if !ok {
		return "", "", "", false
	}

	argsStr := afterParen[:argsEnd]

	// Verify that after the closing paren there's nothing (or just whitespace)
	// This prevents partial matching of longer expressions
	afterArgs := strings.TrimSpace(afterParen[argsEnd+1:])
	if afterArgs != "" {
		// There's more after the method call - this might be a chained call
		// For now, just take what we have
		// In future, we could handle chaining here
	}

	return varName, methodName, argsStr, true
}

// extractStaticMethodCall extracts Class::method(args)
// Returns: className, methodName, argsStr, success
func (e *ExecutionEngine) extractStaticMethodCall(expr string) (string, string, string, bool) {
	expr = strings.TrimSpace(expr)

	// Find the :: separator
	colonIdx := strings.Index(expr, "::")
	if colonIdx == -1 {
		return "", "", "", false
	}

	className := expr[:colonIdx]
	if !patterns.WordPattern.MatchString(className) {
		return "", "", "", false
	}

	remainder := expr[colonIdx+2:]

	// Find method name (ends at opening paren)
	parenIdx := strings.Index(remainder, "(")
	if parenIdx == -1 {
		return "", "", "", false // Not a method call
	}

	methodName := remainder[:parenIdx]
	if !patterns.WordPattern.MatchString(methodName) {
		return "", "", "", false
	}

	// Find matching closing paren
	afterParen := remainder[parenIdx+1:]
	argsEnd, ok := findMatchingParen(afterParen)
	if !ok {
		return "", "", "", false
	}

	argsStr := afterParen[:argsEnd]

	// Verify nothing after the method call
	afterArgs := strings.TrimSpace(afterParen[argsEnd+1:])
	if afterArgs != "" {
		return "", "", "", false
	}

	return className, methodName, argsStr, true
}

// extractChainMethodCall extracts a method call from a chain step starting with ->
// Input: "->method(args)..."
// Returns: methodName, argsStr, bytesConsumed, success
func (e *ExecutionEngine) extractChainMethodCall(s string) (string, string, int, bool) {
	// Must start with ->
	if !strings.HasPrefix(s, "->") {
		return "", "", 0, false
	}

	remainder := s[2:] // Skip ->

	// Find method name (word characters until '(')
	parenIdx := strings.Index(remainder, "(")
	if parenIdx == -1 {
		return "", "", 0, false
	}

	methodName := remainder[:parenIdx]
	if !patterns.WordPattern.MatchString(methodName) {
		return "", "", 0, false
	}

	// Find matching closing paren
	afterParen := remainder[parenIdx+1:]
	argsEnd, ok := findMatchingParen(afterParen)
	if !ok {
		return "", "", 0, false
	}

	argsStr := afterParen[:argsEnd]
	// Total consumed: 2 (for ->) + parenIdx + 1 (for open paren) + argsEnd + 1 (for close paren)
	consumed := 2 + parenIdx + 1 + argsEnd + 1

	return methodName, argsStr, consumed, true
}

// findMatchingParen finds the position of the closing paren that matches the implicit opening paren
// Input is the string AFTER the opening paren
// Returns the index of the matching ')' and success flag
func findMatchingParen(s string) (int, bool) {
	depth := 1
	inString := false
	var stringChar byte = 0

	for i := 0; i < len(s); i++ {
		c := s[i]

		// Handle escape sequences in strings
		if inString && c == '\\' && i+1 < len(s) {
			i++ // Skip the escaped character
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

// traceMethodCall traces a method call expression like $mybb->get_input('timezone')
