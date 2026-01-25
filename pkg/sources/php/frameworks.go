// Package php - frameworks.go provides PHP framework pattern registry and universal patterns
// All PHP framework patterns should be registered here
package php

import (
	"regexp"

	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// Registry is the global PHP framework pattern registry
var Registry = common.NewFrameworkPatternRegistry("php")

// Universal patterns for detecting input across ANY PHP framework
var (
	// InputMethodPattern matches method names that ALWAYS indicate user input
	// e.g., input, get_input, getInput, get_var, variable, getQueryParams, getParsedBody
	InputMethodPattern = regexp.MustCompile(`(?i)^(get_?)?(input|var|variable|query_?params?|parsed_?body|cookie_?params?|server_?params?|uploaded_?files?|headers?|all)$|^(get_?)?(post|cookie|param)s?$`)

	// InputPropertyPattern matches property names that typically hold user input
	// e.g., input, request, params, query, cookies, headers, body, data
	InputPropertyPattern = regexp.MustCompile(`(?i)^(input|request|params?|query|cookies?|headers?|body|data|args?|post|get|files?|server|attributes?|payload)s?$`)

	// InputObjectPattern matches object/variable names that suggest an input carrier
	// e.g., $request, $input, $mybb, $ctx
	InputObjectPattern = regexp.MustCompile(`(?i)(request|input|req|params?|http|ctx|context|mybb|getRequest\(\)|getApplication\(\))`)

	// ExcludeMethodPattern matches methods to exclude from input detection (false positives)
	ExcludeMethodPattern = regexp.MustCompile(`(?i)^(getData|getBody|getContent|fetch|find|load|read)$`)

	// ContextDependentMethodPattern matches generic getters that need context
	// Only detect as input when called on a request-like object
	ContextDependentMethodPattern = regexp.MustCompile(`(?i)^(get_?)?(val|text|int|bool|array|raw_?val|check)$`)
)

// InputPropertyPatterns contains universal property access patterns
// These match ->property[ array access on input objects
var InputPropertyPatterns = []string{
	"->input[",     // MyBB, generic
	"->data[",      // Generic data array
	"->request[",   // Symfony, generic
	"->params[",    // Generic params
	"->cookies[",   // Cookie arrays
	"->query[",     // Symfony query bag
	"->post[",      // POST data arrays
	"->get[",       // GET data arrays
	"->files[",     // File uploads
	"->server[",    // Server vars
	"->headers[",   // Headers
	"->attributes[", // PSR-7 attributes
	"->payload[",   // API payloads
	"->args[",      // Arguments
}

// InputMethodPatterns contains universal method call patterns
// These match ->method( calls that return user input
var InputMethodPatterns = []string{
	// Generic input getters
	"->get_input(",
	"->getInput(",
	"->get_var(",
	"->getVar(",
	"->variable(",
	"->input(",
	"->query(",
	"->post(",
	"->cookie(",
	"->header(",
	"->file(",
	"->get(",
	"->all(",
	// PSR-7 methods
	"->getQueryParams(",
	"->getParsedBody(",
	"->getCookieParams(",
	"->getUploadedFiles(",
	"->getServerParams(",
	"->getHeaders(",
	"->getHeader(",
	"->getHeaderLine(",
	"->getAttribute(",
}

// IsInputPropertyAccess checks if an expression matches an input property pattern
func IsInputPropertyAccess(expr string) bool {
	for _, pattern := range InputPropertyPatterns {
		if contains(expr, pattern) {
			return true
		}
	}
	return false
}

// IsInputMethodCall checks if an expression matches an input method pattern
func IsInputMethodCall(expr string) bool {
	for _, pattern := range InputMethodPatterns {
		if contains(expr, pattern) {
			return true
		}
	}
	return false
}

// IsContextDependentMethod checks if a method needs context to determine if it's input
func IsContextDependentMethod(methodName string) bool {
	return ContextDependentMethodPattern.MatchString(methodName)
}

// IsInputMethod checks if a method name always indicates user input
func IsInputMethod(methodName string) bool {
	return InputMethodPattern.MatchString(methodName) && !ExcludeMethodPattern.MatchString(methodName)
}

// IsInputProperty checks if a property name typically holds user input
func IsInputProperty(propertyName string) bool {
	return InputPropertyPattern.MatchString(propertyName)
}

// IsInputObject checks if a variable/object name suggests an input carrier
func IsInputObject(objectName string) bool {
	return InputObjectPattern.MatchString(objectName)
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

// GetAllPatterns returns all registered framework patterns
func GetAllPatterns() []*common.FrameworkPattern {
	return Registry.GetAll()
}

// GetPatternsByFramework returns patterns for a specific framework
func GetPatternsByFramework(framework string) []*common.FrameworkPattern {
	return Registry.GetByFramework(framework)
}

// GetPatternByID returns a pattern by its ID
func GetPatternByID(id string) *common.FrameworkPattern {
	return Registry.GetByID(id)
}

// helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func containsAny(s string, substrs []string) bool {
	for _, substr := range substrs {
		if contains(s, substr) {
			return true
		}
	}
	return false
}

func init() {
	// Initialize the registry
}
