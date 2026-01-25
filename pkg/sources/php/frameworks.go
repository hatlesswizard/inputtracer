// Package php - frameworks.go provides PHP framework pattern registry and string-based patterns
// Regex patterns are defined in patterns.go - this file contains string-based pattern lists
// and framework detection functionality
package php

import (
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// Registry is the global PHP framework pattern registry
var Registry = common.NewFrameworkPatternRegistry("php")

// Note: Regex patterns (InputMethodPattern, InputPropertyPattern, etc.) are defined in patterns.go
// This file contains string-based pattern lists for quick substring matching

// InputPropertyPatterns contains universal property access patterns
// These match ->property[ array access on input objects
var InputPropertyPatterns = []string{
	"->input[",     // Generic input array
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

// Note: IsContextDependentMethod, IsInputMethod, IsInputProperty, IsInputObject
// are defined in patterns.go using the centralized regex patterns.
// GetInputConfidence is also defined in patterns.go.

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

// =============================================================================
// FRAMEWORK DETECTION
// Centralized framework detection patterns (moved from pkg/semantic/analyzer/php)
// =============================================================================

// FrameworkDetectionPatterns maps framework names to detection patterns
// Each framework has patterns to match in: imports, class names, and source code
var FrameworkDetectionPatterns = map[string]FrameworkDetection{
	"laravel": {
		ImportPatterns:  []string{"illuminate", "laravel"},
		ClassPatterns:   []string{},
		SourcePatterns:  []string{"Illuminate\\", "Laravel\\"},
	},
	"symfony": {
		ImportPatterns:  []string{"symfony"},
		ClassPatterns:   []string{},
		SourcePatterns:  []string{"Symfony\\"},
	},
	"codeigniter": {
		ImportPatterns:  []string{"codeigniter", "ci_"},
		ClassPatterns:   []string{"CI_Controller", "CI_Model"},
		SourcePatterns:  []string{"CodeIgniter\\"},
	},
	"wordpress": {
		ImportPatterns:  []string{},
		ClassPatterns:   []string{"WP_REST_Request", "WP_Query"},
		SourcePatterns:  []string{"wp_", "WP_", "WordPress", "get_option("},
	},
	"mybb": {
		ImportPatterns:  []string{},
		ClassPatterns:   []string{"mybb", "MyBB"},
		SourcePatterns:  []string{"$mybb->", "MyBB"},
	},
	"drupal": {
		ImportPatterns:  []string{"drupal"},
		ClassPatterns:   []string{},
		SourcePatterns:  []string{"Drupal\\", "drupal_"},
	},
	"yii": {
		ImportPatterns:  []string{"yii"},
		ClassPatterns:   []string{"CController", "CModel"},
		SourcePatterns:  []string{"Yii::"},
	},
	"cakephp": {
		ImportPatterns:  []string{"cake"},
		ClassPatterns:   []string{"AppController", "AppModel"},
		SourcePatterns:  []string{"Cake\\"},
	},
	"zend": {
		ImportPatterns:  []string{"zend", "laminas"},
		ClassPatterns:   []string{},
		SourcePatterns:  []string{"Zend\\", "Laminas\\"},
	},
	"slim": {
		ImportPatterns:  []string{"slim"},
		ClassPatterns:   []string{},
		SourcePatterns:  []string{"Slim\\"},
	},
	"lumen": {
		ImportPatterns:  []string{"lumen"},
		ClassPatterns:   []string{},
		SourcePatterns:  []string{"Laravel\\Lumen\\"},
	},
	"phpbb": {
		ImportPatterns:  []string{"phpbb"},
		ClassPatterns:   []string{"phpbb_"},
		SourcePatterns:  []string{"phpbb\\", "phpBB"},
	},
	"mediawiki": {
		ImportPatterns:  []string{"mediawiki"},
		ClassPatterns:   []string{"SpecialPage", "ApiBase"},
		SourcePatterns:  []string{"MediaWiki\\", "wfMessage("},
	},
	"joomla": {
		ImportPatterns:  []string{"joomla"},
		ClassPatterns:   []string{"JController", "JModel"},
		SourcePatterns:  []string{"Joomla\\", "JFactory::"},
	},
	"magento": {
		ImportPatterns:  []string{"magento"},
		ClassPatterns:   []string{"Mage_"},
		SourcePatterns:  []string{"Magento\\", "Mage::"},
	},
	"prestashop": {
		ImportPatterns:  []string{"prestashop"},
		ClassPatterns:   []string{"Module", "AdminController"},
		SourcePatterns:  []string{"PrestaShop\\"},
	},
}

// FrameworkDetection contains patterns for detecting a framework
type FrameworkDetection struct {
	ImportPatterns  []string // Patterns to match in import/use statements
	ClassPatterns   []string // Patterns to match in class names
	SourcePatterns  []string // Patterns to match in source code
}

// DetectFrameworkFromImports detects frameworks based on import statements
func DetectFrameworkFromImports(imports []string) []string {
	var frameworks []string
	seen := make(map[string]bool)

	for _, imp := range imports {
		lowerImp := toLower(imp)
		for framework, detection := range FrameworkDetectionPatterns {
			if seen[framework] {
				continue
			}
			for _, pattern := range detection.ImportPatterns {
				if containsString(lowerImp, pattern) {
					frameworks = append(frameworks, framework)
					seen[framework] = true
					break
				}
			}
		}
	}

	return frameworks
}

// DetectFrameworkFromClasses detects frameworks based on class names
func DetectFrameworkFromClasses(classNames []string) []string {
	var frameworks []string
	seen := make(map[string]bool)

	for _, className := range classNames {
		lowerClass := toLower(className)
		for framework, detection := range FrameworkDetectionPatterns {
			if seen[framework] {
				continue
			}
			for _, pattern := range detection.ClassPatterns {
				if containsString(lowerClass, toLower(pattern)) || className == pattern {
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
			if containsString(source, pattern) {
				frameworks = append(frameworks, framework)
				seen[framework] = true
				break
			}
		}
	}

	return frameworks
}

// DetectFrameworks detects all frameworks using all detection methods
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

	// Check class names
	for _, f := range DetectFrameworkFromClasses(classNames) {
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

// helper functions for framework detection
func toLower(s string) string {
	// Simple lowercase for ASCII
	result := make([]byte, len(s))
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			result[i] = c + 32
		} else {
			result[i] = c
		}
	}
	return string(result)
}

func containsString(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr) >= 0
}

func findSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func init() {
	// Initialize the registry
}
