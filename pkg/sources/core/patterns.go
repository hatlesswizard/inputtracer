package core

import (
	"regexp"
	"sync"
)

// UniversalPatterns holds pre-compiled regex patterns used across all languages
type UniversalPatterns struct {
	// Input method detection
	InputMethod   *regexp.Regexp
	InputProperty *regexp.Regexp
	InputObject   *regexp.Regexp
	ExcludeMethod *regexp.Regexp

	// Key/property access
	ArrayKeyAccess *regexp.Regexp
	PropertyAccess *regexp.Regexp
	MethodCall     *regexp.Regexp

	// Assignment patterns
	SimpleAssign   *regexp.Regexp
	PropertyAssign *regexp.Regexp
}

var (
	universalPatterns *UniversalPatterns
	patternsOnce      sync.Once
)

// GetUniversalPatterns returns the singleton universal patterns instance
func GetUniversalPatterns() *UniversalPatterns {
	patternsOnce.Do(func() {
		universalPatterns = &UniversalPatterns{
			// Methods that indicate user input retrieval
			InputMethod: regexp.MustCompile(`(?i)^(get_?)?(input|var|variable|query_?params?|parsed_?body|cookie_?params?|server_?params?|uploaded_?files?|headers?|all|body|content)$|^(get_?)?(post|cookie|param|query|header)s?$`),

			// Properties that hold user input
			InputProperty: regexp.MustCompile(`(?i)^(input|request|params?|query|cookies?|headers?|body|data|args?|post|get|files?|server|attributes?|payload)s?$`),

			// Objects that carry user input
			InputObject: regexp.MustCompile(`(?i)(request|input|req|params?|http|ctx|context)`),

			// Methods to exclude (false positives)
			ExcludeMethod: regexp.MustCompile(`(?i)^(getData|getBody|getContent|fetch|find|load|read|save|store|cache|log|debug|info|warn|error)$`),

			// Key access patterns: ['key'] or ["key"]
			ArrayKeyAccess: regexp.MustCompile(`\[['"](\w+)['"]\]`),

			// Property access: .property or ->property
			PropertyAccess: regexp.MustCompile(`(?:->|\.)(\w+)`),

			// Method call: .method( or ->method(
			MethodCall: regexp.MustCompile(`(?:->|\.)(\w+)\s*\(`),

			// Simple assignment: var =
			SimpleAssign: regexp.MustCompile(`(\$?\w+)\s*=\s*`),

			// Property assignment: obj.prop = or obj->prop =
			PropertyAssign: regexp.MustCompile(`(\$?\w+)(?:->|\.)(\w+)\s*=\s*`),
		}
	})
	return universalPatterns
}

// IsInputMethod checks if a method name indicates input retrieval
func IsInputMethod(methodName string) bool {
	return GetUniversalPatterns().InputMethod.MatchString(methodName)
}

// IsInputProperty checks if a property name holds input data
func IsInputProperty(propName string) bool {
	return GetUniversalPatterns().InputProperty.MatchString(propName)
}

// IsInputObject checks if an object name carries input
func IsInputObject(objName string) bool {
	return GetUniversalPatterns().InputObject.MatchString(objName)
}

// IsExcludedMethod checks if a method should be excluded from input detection
func IsExcludedMethod(methodName string) bool {
	return GetUniversalPatterns().ExcludeMethod.MatchString(methodName)
}

// ExtractKey extracts the key from array/property access expressions
func ExtractKey(expr string) string {
	if match := GetUniversalPatterns().ArrayKeyAccess.FindStringSubmatch(expr); len(match) > 1 {
		return match[1]
	}
	return ""
}
