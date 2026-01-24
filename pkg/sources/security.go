// Package sources - security.go provides centralized security pattern definitions
// All validation, sanitization, and auth patterns should be defined here
package sources

import "regexp"

// SecurityPatternType categorizes security functions
type SecurityPatternType string

const (
	PatternValidation     SecurityPatternType = "validation"
	PatternSanitization   SecurityPatternType = "sanitization"
	PatternAuthentication SecurityPatternType = "authentication"
	PatternAuthorization  SecurityPatternType = "authorization"
	PatternNullCheck      SecurityPatternType = "null_check"
	PatternTypeCheck      SecurityPatternType = "type_check"
)

// SecurityPatternEffect describes how a pattern affects data flow
type SecurityPatternEffect string

const (
	EffectAllows    SecurityPatternEffect = "allows"
	EffectBlocks    SecurityPatternEffect = "blocks"
	EffectSanitizes SecurityPatternEffect = "sanitizes"
	EffectValidates SecurityPatternEffect = "validates"
	EffectUnknown   SecurityPatternEffect = "unknown"
)

// SecurityPattern describes a security-related function or pattern
type SecurityPattern struct {
	Name       string
	PatternStr string
	SecType    string                // Security type string e.g. "regex_validation"
	Type       SecurityPatternType
	Effect     SecurityPatternEffect
	Confidence float64
}

// LanguageSecurityPatterns holds all security patterns for a language
type LanguageSecurityPatterns struct {
	Language          string
	ValidationFuncs   []SecurityPattern
	SanitizationFuncs map[string]bool
	AuthPatterns      map[string]string // name -> regex pattern string
}

// securityRegistry holds patterns for all languages
var securityRegistry = make(map[string]*LanguageSecurityPatterns)

// PHP Security Patterns
// Replaces hardcoded patterns in condition/extractor.go registerPHPPatterns()
var PHPSecurityPatterns = &LanguageSecurityPatterns{
	Language: "php",
	ValidationFuncs: []SecurityPattern{
		{Name: "preg_match", PatternStr: `preg_match\s*\(\s*['"].*['"]`, SecType: "regex_validation", Type: PatternValidation, Effect: EffectValidates, Confidence: 0.9},
		{Name: "filter_var", PatternStr: `filter_var\s*\(`, SecType: "filter_validation", Type: PatternValidation, Effect: EffectValidates, Confidence: 0.9},
		{Name: "ctype_", PatternStr: `ctype_\w+\s*\(`, SecType: "ctype_validation", Type: PatternValidation, Effect: EffectValidates, Confidence: 0.85},
		{Name: "is_numeric", PatternStr: `is_numeric\s*\(`, SecType: "numeric_validation", Type: PatternTypeCheck, Effect: EffectValidates, Confidence: 0.8},
		{Name: "is_int", PatternStr: `is_int\s*\(`, SecType: "type_validation", Type: PatternTypeCheck, Effect: EffectValidates, Confidence: 0.8},
		{Name: "isset", PatternStr: `isset\s*\(`, SecType: "null_check", Type: PatternNullCheck, Effect: EffectAllows, Confidence: 0.6},
		{Name: "empty", PatternStr: `empty\s*\(`, SecType: "empty_check", Type: PatternNullCheck, Effect: EffectBlocks, Confidence: 0.6},
	},
	SanitizationFuncs: map[string]bool{
		"htmlspecialchars":          true,
		"htmlentities":              true,
		"addslashes":                true,
		"mysql_real_escape_string":  true,
		"mysqli_real_escape_string": true,
		"pg_escape_string":          true,
		"strip_tags":                true,
		"escapeshellarg":            true,
		"escapeshellcmd":            true,
		"intval":                    true,
		"floatval":                  true,
	},
	AuthPatterns: map[string]string{
		"logged_in":  `(?i)(is_logged_in|logged_in|is_authenticated|isLoggedIn|isAuthenticated)`,
		"admin":      `(?i)(is_admin|isAdmin|has_admin|hasAdmin|admin_check)`,
		"permission": `(?i)(has_permission|can_|has_access|hasPermission|checkPermission)`,
		"token":      `(?i)(verify_token|check_token|csrf_token|validateToken)`,
	},
}

// JavaScript Security Patterns
var JavaScriptSecurityPatterns = &LanguageSecurityPatterns{
	Language: "javascript",
	ValidationFuncs: []SecurityPattern{
		{Name: "regex_test", PatternStr: `\.test\s*\(`, SecType: "regex_validation", Type: PatternValidation, Effect: EffectValidates, Confidence: 0.9},
		{Name: "validator", PatternStr: `validator\.\w+\s*\(`, SecType: "validator_library", Type: PatternValidation, Effect: EffectValidates, Confidence: 0.85},
		{Name: "typeof", PatternStr: `typeof\s+\w+\s*(===?|!==?)`, SecType: "type_check", Type: PatternTypeCheck, Effect: EffectValidates, Confidence: 0.7},
		{Name: "instanceof", PatternStr: `\s+instanceof\s+`, SecType: "instance_check", Type: PatternTypeCheck, Effect: EffectValidates, Confidence: 0.7},
	},
	SanitizationFuncs: map[string]bool{
		"escape":             true,
		"encodeURIComponent": true,
		"encodeURI":          true,
		"sanitize":           true,
		"DOMPurify.sanitize": true,
		"parseInt":           true,
		"parseFloat":         true,
	},
	AuthPatterns: map[string]string{
		"auth":  `(?i)(isAuthenticated|isLoggedIn|req\.user|req\.session\.user)`,
		"admin": `(?i)(isAdmin|req\.user\.admin|hasRole)`,
	},
}

// Python Security Patterns
var PythonSecurityPatterns = &LanguageSecurityPatterns{
	Language: "python",
	ValidationFuncs: []SecurityPattern{
		{Name: "re_match", PatternStr: `re\.(match|search|fullmatch)\s*\(`, SecType: "regex_validation", Type: PatternValidation, Effect: EffectValidates, Confidence: 0.9},
		{Name: "isinstance", PatternStr: `isinstance\s*\(`, SecType: "type_validation", Type: PatternTypeCheck, Effect: EffectValidates, Confidence: 0.8},
	},
	SanitizationFuncs: map[string]bool{
		"escape":       true,
		"quote":        true,
		"html.escape":  true,
		"bleach.clean": true,
		"int":          true,
		"float":        true,
		"str":          true,
	},
	AuthPatterns: map[string]string{
		"login": `(?i)(login_required|is_authenticated|current_user)`,
		"admin": `(?i)(is_admin|is_superuser|has_perm)`,
	},
}

// Go Security Patterns
var GoSecurityPatterns = &LanguageSecurityPatterns{
	Language: "go",
	ValidationFuncs: []SecurityPattern{
		{Name: "regexp", PatternStr: `regexp\.(Match|MatchString)\s*\(`, SecType: "regex_validation", Type: PatternValidation, Effect: EffectValidates, Confidence: 0.9},
		{Name: "validator", PatternStr: `validator\.\w+\s*\(`, SecType: "validator_library", Type: PatternValidation, Effect: EffectValidates, Confidence: 0.85},
	},
	SanitizationFuncs: map[string]bool{
		"html.EscapeString":  true,
		"url.QueryEscape":    true,
		"strconv.Atoi":       true,
		"strconv.ParseInt":   true,
		"strconv.ParseFloat": true,
	},
	AuthPatterns: map[string]string{
		"auth": `(?i)(IsAuthenticated|GetUser|RequireAuth)`,
	},
}

// Java Security Patterns
var JavaSecurityPatterns = &LanguageSecurityPatterns{
	Language: "java",
	ValidationFuncs: []SecurityPattern{
		{Name: "matches", PatternStr: `\.matches\s*\(`, SecType: "regex_validation", Type: PatternValidation, Effect: EffectValidates, Confidence: 0.9},
		{Name: "instanceof", PatternStr: `\s+instanceof\s+`, SecType: "instance_check", Type: PatternTypeCheck, Effect: EffectValidates, Confidence: 0.8},
	},
	SanitizationFuncs: map[string]bool{
		"StringEscapeUtils.escapeHtml":     true,
		"ESAPI.encoder().encodeForHTML":    true,
		"Integer.parseInt":                 true,
		"Long.parseLong":                   true,
		"Double.parseDouble":               true,
	},
	AuthPatterns: map[string]string{
		"auth": `(?i)(isAuthenticated|getPrincipal|hasRole|hasAuthority)`,
	},
}

func init() {
	// Register all language security patterns
	securityRegistry["php"] = PHPSecurityPatterns
	securityRegistry["javascript"] = JavaScriptSecurityPatterns
	securityRegistry["typescript"] = JavaScriptSecurityPatterns // TypeScript uses JS patterns
	securityRegistry["python"] = PythonSecurityPatterns
	securityRegistry["go"] = GoSecurityPatterns
	securityRegistry["java"] = JavaSecurityPatterns
}

// GetSecurityPatterns returns security patterns for a language
func GetSecurityPatterns(language string) *LanguageSecurityPatterns {
	return securityRegistry[language]
}

// IsSanitizationFunc checks if a function name is a known sanitization function
func IsSanitizationFunc(language, funcName string) bool {
	if patterns := securityRegistry[language]; patterns != nil {
		return patterns.SanitizationFuncs[funcName]
	}
	return false
}

// GetSanitizationFuncs returns all sanitization functions for a language
func GetSanitizationFuncs(language string) map[string]bool {
	if patterns := securityRegistry[language]; patterns != nil {
		return patterns.SanitizationFuncs
	}
	return make(map[string]bool)
}

// GetAuthPatterns returns compiled auth patterns for a language
func GetAuthPatterns(language string) map[string]*regexp.Regexp {
	result := make(map[string]*regexp.Regexp)
	if patterns := securityRegistry[language]; patterns != nil {
		for name, patternStr := range patterns.AuthPatterns {
			if re, err := regexp.Compile(patternStr); err == nil {
				result[name] = re
			}
		}
	}
	return result
}

// GetValidationFuncs returns validation function patterns for a language
func GetValidationFuncs(language string) []SecurityPattern {
	if patterns := securityRegistry[language]; patterns != nil {
		return patterns.ValidationFuncs
	}
	return nil
}
