// Package php - frameworks.go provides PHP framework pattern registry and string-based patterns
// String-based pattern lists are derived dynamically from registered framework patterns
package php

import (
	"strings"
	"sync"

	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// Registry is the global PHP framework pattern registry
var Registry = common.NewFrameworkPatternRegistry("php")

// Pattern caches - built lazily from registered framework patterns
var (
	inputMethodPatterns   []string
	inputPropertyPatterns []string
	patternsOnce          sync.Once
)

// buildPatternsFromRegistry extracts method and property patterns from registered framework patterns
func buildPatternsFromRegistry() {
	methodSet := make(map[string]bool)
	propertySet := make(map[string]bool)

	for _, pattern := range Registry.GetAll() {
		// Extract method names from MethodPattern (e.g., "^input$" -> "input")
		if pattern.MethodPattern != "" {
			methodName := stripRegexAnchors(pattern.MethodPattern)
			if methodName != "" {
				methodSet["->" + methodName + "("] = true
			}
		}

		// Extract property names from PropertyPattern (e.g., "^query$" -> "query")
		if pattern.PropertyPattern != "" {
			propName := stripRegexAnchors(pattern.PropertyPattern)
			if propName != "" {
				propertySet["->" + propName + "["] = true
			}
		}

		// Also use CarrierProperty for property patterns
		if pattern.CarrierProperty != "" {
			propertySet["->" + pattern.CarrierProperty + "["] = true
		}
	}

	// Convert sets to slices
	inputMethodPatterns = make([]string, 0, len(methodSet))
	for pattern := range methodSet {
		inputMethodPatterns = append(inputMethodPatterns, pattern)
	}

	inputPropertyPatterns = make([]string, 0, len(propertySet))
	for pattern := range propertySet {
		inputPropertyPatterns = append(inputPropertyPatterns, pattern)
	}
}

// stripRegexAnchors removes ^ and $ anchors from a regex pattern
func stripRegexAnchors(pattern string) string {
	result := pattern
	if len(result) > 0 && result[0] == '^' {
		result = result[1:]
	}
	if len(result) > 0 && result[len(result)-1] == '$' {
		result = result[:len(result)-1]
	}
	return result
}

// GetInputMethodPatterns returns method patterns derived from registered framework patterns
// Built lazily on first access to ensure all framework patterns are registered
func GetInputMethodPatterns() []string {
	patternsOnce.Do(buildPatternsFromRegistry)
	return inputMethodPatterns
}

// GetInputPropertyPatterns returns property patterns derived from registered framework patterns
// Built lazily on first access to ensure all framework patterns are registered
func GetInputPropertyPatterns() []string {
	patternsOnce.Do(buildPatternsFromRegistry)
	return inputPropertyPatterns
}

// IsInputPropertyAccess checks if an expression matches an input property pattern
func IsInputPropertyAccess(expr string) bool {
	for _, pattern := range GetInputPropertyPatterns() {
		if strings.Contains(expr, pattern) {
			return true
		}
	}
	return false
}

// IsInputMethodCall checks if an expression matches an input method pattern
func IsInputMethodCall(expr string) bool {
	for _, pattern := range GetInputMethodPatterns() {
		if strings.Contains(expr, pattern) {
			return true
		}
	}
	return false
}

// Note: IsContextDependentMethod, IsInputMethod, IsInputProperty, IsInputObject
// are defined in patterns.go using the centralized regex patterns.

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


// =============================================================================
// FRAMEWORK DETECTION
// Centralized framework detection patterns (moved from pkg/semantic/analyzer/php)
// =============================================================================

// FrameworkDetectionPatterns maps framework names to detection patterns
// Each framework has patterns to match in imports and source code
// Note: Only Laravel and Symfony are supported
var FrameworkDetectionPatterns = map[string]FrameworkDetection{
	"laravel": {
		ImportPatterns: []string{"illuminate", "laravel"},
		SourcePatterns: []string{"Illuminate\\", "Laravel\\"},
	},
	"symfony": {
		ImportPatterns: []string{"symfony"},
		SourcePatterns: []string{"Symfony\\"},
	},
}

// FrameworkDetection contains patterns for detecting a framework
type FrameworkDetection struct {
	ImportPatterns []string // Patterns to match in import/use statements
	SourcePatterns []string // Patterns to match in source code
}

// DetectFrameworkFromImports detects frameworks based on import statements
func DetectFrameworkFromImports(imports []string) []string {
	var frameworks []string
	seen := make(map[string]bool)

	for _, imp := range imports {
		lowerImp := strings.ToLower(imp)
		for framework, detection := range FrameworkDetectionPatterns {
			if seen[framework] {
				continue
			}
			for _, pattern := range detection.ImportPatterns {
				if strings.Contains(lowerImp, pattern) {
					frameworks = append(frameworks, framework)
					seen[framework] = true
					break
				}
			}
		}
	}

	return frameworks
}

// DetectFrameworkFromSource detects frameworks based on source code content
func DetectFrameworkFromSource(source string) []string {
	var frameworks []string
	seen := make(map[string]bool)

	for framework, detection := range FrameworkDetectionPatterns {
		if seen[framework] {
			continue
		}
		for _, pattern := range detection.SourcePatterns {
			if strings.Contains(source, pattern) {
				frameworks = append(frameworks, framework)
				seen[framework] = true
				break
			}
		}
	}

	return frameworks
}

// DetectFrameworks detects all frameworks using import and source detection methods
// The classNames parameter is kept for API compatibility but is unused
func DetectFrameworks(imports []string, classNames []string, source string) []string {
	seen := make(map[string]bool)
	var frameworks []string

	// Check imports
	for _, f := range DetectFrameworkFromImports(imports) {
		if !seen[f] {
			frameworks = append(frameworks, f)
			seen[f] = true
		}
	}

	// Check source
	for _, f := range DetectFrameworkFromSource(source) {
		if !seen[f] {
			frameworks = append(frameworks, f)
			seen[f] = true
		}
	}

	return frameworks
}


