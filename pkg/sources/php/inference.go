// Package php provides PHP-specific source type inference
package php

import (
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// =============================================================================
// SOURCE TYPE INFERENCE
// Determines the semantic source type based on method/property name patterns
// =============================================================================

// MethodNamePatterns maps patterns in method names to their source types
var MethodNamePatterns = map[string]common.SourceType{
	"cookie":  common.SourceHTTPCookie,
	"header":  common.SourceHTTPHeader,
	"server":  common.SourceHTTPHeader, // $_SERVER is typically HTTP headers
	"post":    common.SourceHTTPPost,
	"body":    common.SourceHTTPPost,
	"parsed":  common.SourceHTTPPost, // getParsedBody
	"query":   common.SourceHTTPGet,
	"get":     common.SourceHTTPGet,
	"file":    common.SourceHTTPBody,
	"upload":  common.SourceHTTPBody,
}

// PropertyNamePatterns maps patterns in property names to their source types
var PropertyNamePatterns = map[string]common.SourceType{
	"cookie":  common.SourceHTTPCookie,
	"cookies": common.SourceHTTPCookie,
	"header":  common.SourceHTTPHeader,
	"headers": common.SourceHTTPHeader,
	"server":  common.SourceHTTPHeader,
	"post":    common.SourceHTTPPost,
	"body":    common.SourceHTTPPost,
	"query":   common.SourceHTTPGet,
	"get":     common.SourceHTTPGet,
	"file":    common.SourceHTTPBody,
	"files":   common.SourceHTTPBody,
}

// InferSourceTypeFromMethodName determines the source type based on method name patterns
// This centralizes the logic previously in pkg/semantic/analyzer/php/analyzer.go
func InferSourceTypeFromMethodName(methodName string) common.SourceType {
	lowerName := strings.ToLower(methodName)

	// Check for specific type hints in the method name
	// Order matters: check more specific patterns first
	for pattern, sourceType := range MethodNamePatterns {
		if strings.Contains(lowerName, pattern) {
			return sourceType
		}
	}

	// Default to generic user input
	return common.SourceUserInput
}

// InferSourceTypeFromPropertyName determines the source type based on property name patterns
// This centralizes the logic previously in pkg/semantic/analyzer/php/analyzer.go
func InferSourceTypeFromPropertyName(propName string) common.SourceType {
	lowerName := strings.ToLower(propName)

	// Check for exact matches first (more specific)
	if sourceType, ok := PropertyNamePatterns[lowerName]; ok {
		return sourceType
	}

	// Check for partial matches (contains pattern)
	for pattern, sourceType := range PropertyNamePatterns {
		if strings.Contains(lowerName, pattern) {
			return sourceType
		}
	}

	// Default to generic user input
	return common.SourceUserInput
}

// InferSourceTypeFromExpression determines source type from a full expression
// e.g., "$request->getCookieParams()" -> SourceHTTPCookie
func InferSourceTypeFromExpression(expr string) common.SourceType {
	lower := strings.ToLower(expr)

	// Check method patterns
	for pattern, sourceType := range MethodNamePatterns {
		if strings.Contains(lower, pattern) {
			return sourceType
		}
	}

	// Check property patterns
	for pattern, sourceType := range PropertyNamePatterns {
		if strings.Contains(lower, pattern) {
			return sourceType
		}
	}

	return common.SourceUserInput
}
