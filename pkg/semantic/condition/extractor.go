// Package condition provides key condition extraction for branch analysis.
// Inspired by ATLANTIS's approach to understanding what conditions must be
// true for tainted data to reach a sink, enabling exploitability assessment.
//
// Key conditions help determine:
// - What guards exist on the path to a sink
// - Whether security checks (validation, sanitization) are present
// - What constraints exist on user input
// - If a vulnerability is actually exploitable
package condition

import (
	"regexp"
	"strings"
)

// ConditionType classifies the type of condition
type ConditionType string

const (
	CondTypeComparison    ConditionType = "comparison"    // ==, !=, <, >, etc.
	CondTypeNullCheck     ConditionType = "null_check"    // isset, empty, is_null
	CondTypeTypeCheck     ConditionType = "type_check"    // is_string, instanceof
	CondTypeValidation    ConditionType = "validation"    // preg_match, filter_var
	CondTypeSanitization  ConditionType = "sanitization"  // htmlspecialchars, addslashes
	CondTypeAuthentication ConditionType = "authentication" // logged_in, is_admin
	CondTypeAuthorization ConditionType = "authorization"  // has_permission, can_access
	CondTypeLengthCheck   ConditionType = "length_check"   // strlen, count
	CondTypeLogical       ConditionType = "logical"        // &&, ||, !
	CondTypeUnknown       ConditionType = "unknown"
)

// ConditionEffect describes how a condition affects data flow
type ConditionEffect string

const (
	EffectAllows   ConditionEffect = "allows"   // Condition allows flow if true
	EffectBlocks   ConditionEffect = "blocks"   // Condition blocks flow if true
	EffectSanitizes ConditionEffect = "sanitizes" // Condition implies sanitization
	EffectValidates ConditionEffect = "validates" // Condition implies validation
	EffectUnknown  ConditionEffect = "unknown"
)

// KeyCondition represents a condition that guards code execution
type KeyCondition struct {
	// Location
	FilePath   string `json:"file_path"`
	Line       int    `json:"line"`
	Column     int    `json:"column"`

	// Condition details
	Expression    string        `json:"expression"`     // The condition expression
	Type          ConditionType `json:"type"`
	Effect        ConditionEffect `json:"effect"`
	IsNegated     bool          `json:"is_negated"`     // If inside else or has !

	// Variables involved
	Variables     []string      `json:"variables"`      // Variables referenced
	TaintedVars   []string      `json:"tainted_vars"`   // Which are tainted

	// Scope
	GuardsLines   []int         `json:"guards_lines"`   // Lines guarded by this condition
	NestingDepth  int           `json:"nesting_depth"`  // How nested this condition is
	ParentCondition *KeyCondition `json:"parent_condition,omitempty"`

	// Security relevance
	IsSecurity    bool          `json:"is_security"`    // Is this a security check?
	SecurityType  string        `json:"security_type,omitempty"` // Type of security check
	Confidence    float64       `json:"confidence"`     // Confidence in classification
}

// ConditionPath represents a path through conditions to reach a point
type ConditionPath struct {
	Conditions []*KeyCondition `json:"conditions"`
	TargetLine int             `json:"target_line"`
	TargetExpr string          `json:"target_expr,omitempty"`
	Feasible   bool            `json:"feasible"`      // Is this path feasible?
	Reason     string          `json:"reason,omitempty"` // Why infeasible if not
}

// Extractor extracts key conditions from code
type Extractor struct {
	// Pattern databases
	securityPatterns   map[string]*securityPattern
	validationPatterns map[string]*regexp.Regexp
	sanitizationFuncs  map[string]bool
	authPatterns       map[string]*regexp.Regexp

	// Language-specific settings
	language string
}

type securityPattern struct {
	pattern  *regexp.Regexp
	secType  string
	effect   ConditionEffect
	condType ConditionType
}

// NewExtractor creates a new condition extractor for a language
func NewExtractor(language string) *Extractor {
	e := &Extractor{
		language:           language,
		securityPatterns:   make(map[string]*securityPattern),
		validationPatterns: make(map[string]*regexp.Regexp),
		sanitizationFuncs:  make(map[string]bool),
		authPatterns:       make(map[string]*regexp.Regexp),
	}
	e.registerDefaults()
	return e
}

// registerDefaults registers default patterns for the language
func (e *Extractor) registerDefaults() {
	switch e.language {
	case "php":
		e.registerPHPPatterns()
	case "javascript", "typescript":
		e.registerJavaScriptPatterns()
	case "python":
		e.registerPythonPatterns()
	case "go":
		e.registerGoPatterns()
	case "java":
		e.registerJavaPatterns()
	}
}

func (e *Extractor) registerPHPPatterns() {
	// Validation functions
	e.securityPatterns["preg_match"] = &securityPattern{
		pattern:  regexp.MustCompile(`preg_match\s*\(\s*['"].*['"]`),
		secType:  "regex_validation",
		effect:   EffectValidates,
		condType: CondTypeValidation,
	}
	e.securityPatterns["filter_var"] = &securityPattern{
		pattern:  regexp.MustCompile(`filter_var\s*\(`),
		secType:  "filter_validation",
		effect:   EffectValidates,
		condType: CondTypeValidation,
	}
	e.securityPatterns["ctype_"] = &securityPattern{
		pattern:  regexp.MustCompile(`ctype_\w+\s*\(`),
		secType:  "ctype_validation",
		effect:   EffectValidates,
		condType: CondTypeValidation,
	}
	e.securityPatterns["is_numeric"] = &securityPattern{
		pattern:  regexp.MustCompile(`is_numeric\s*\(`),
		secType:  "numeric_validation",
		effect:   EffectValidates,
		condType: CondTypeTypeCheck,
	}
	e.securityPatterns["is_int"] = &securityPattern{
		pattern:  regexp.MustCompile(`is_int\s*\(`),
		secType:  "type_validation",
		effect:   EffectValidates,
		condType: CondTypeTypeCheck,
	}

	// Null/empty checks
	e.securityPatterns["isset"] = &securityPattern{
		pattern:  regexp.MustCompile(`isset\s*\(`),
		secType:  "null_check",
		effect:   EffectAllows,
		condType: CondTypeNullCheck,
	}
	e.securityPatterns["empty"] = &securityPattern{
		pattern:  regexp.MustCompile(`empty\s*\(`),
		secType:  "empty_check",
		effect:   EffectBlocks,
		condType: CondTypeNullCheck,
	}

	// Sanitization functions
	e.sanitizationFuncs["htmlspecialchars"] = true
	e.sanitizationFuncs["htmlentities"] = true
	e.sanitizationFuncs["addslashes"] = true
	e.sanitizationFuncs["mysql_real_escape_string"] = true
	e.sanitizationFuncs["mysqli_real_escape_string"] = true
	e.sanitizationFuncs["pg_escape_string"] = true
	e.sanitizationFuncs["strip_tags"] = true
	e.sanitizationFuncs["escapeshellarg"] = true
	e.sanitizationFuncs["escapeshellcmd"] = true
	e.sanitizationFuncs["intval"] = true
	e.sanitizationFuncs["floatval"] = true

	// Auth patterns
	e.authPatterns["logged_in"] = regexp.MustCompile(`(?i)(is_logged_in|logged_in|is_authenticated|isLoggedIn|isAuthenticated)`)
	e.authPatterns["admin"] = regexp.MustCompile(`(?i)(is_admin|isAdmin|has_admin|hasAdmin|admin_check)`)
	e.authPatterns["permission"] = regexp.MustCompile(`(?i)(has_permission|can_|has_access|hasPermission|checkPermission)`)
	e.authPatterns["token"] = regexp.MustCompile(`(?i)(verify_token|check_token|csrf_token|validateToken)`)
}

func (e *Extractor) registerJavaScriptPatterns() {
	// Validation
	e.securityPatterns["regex_test"] = &securityPattern{
		pattern:  regexp.MustCompile(`\.test\s*\(`),
		secType:  "regex_validation",
		effect:   EffectValidates,
		condType: CondTypeValidation,
	}
	e.securityPatterns["validator"] = &securityPattern{
		pattern:  regexp.MustCompile(`validator\.\w+\s*\(`),
		secType:  "validator_library",
		effect:   EffectValidates,
		condType: CondTypeValidation,
	}

	// Type checks
	e.securityPatterns["typeof"] = &securityPattern{
		pattern:  regexp.MustCompile(`typeof\s+\w+\s*(===?|!==?)`),
		secType:  "type_check",
		effect:   EffectValidates,
		condType: CondTypeTypeCheck,
	}
	e.securityPatterns["instanceof"] = &securityPattern{
		pattern:  regexp.MustCompile(`\s+instanceof\s+`),
		secType:  "instance_check",
		effect:   EffectValidates,
		condType: CondTypeTypeCheck,
	}

	// Sanitization
	e.sanitizationFuncs["escape"] = true
	e.sanitizationFuncs["encodeURIComponent"] = true
	e.sanitizationFuncs["encodeURI"] = true
	e.sanitizationFuncs["sanitize"] = true
	e.sanitizationFuncs["DOMPurify.sanitize"] = true
	e.sanitizationFuncs["parseInt"] = true
	e.sanitizationFuncs["parseFloat"] = true

	// Auth
	e.authPatterns["auth"] = regexp.MustCompile(`(?i)(isAuthenticated|isLoggedIn|req\.user|req\.session\.user)`)
	e.authPatterns["admin"] = regexp.MustCompile(`(?i)(isAdmin|req\.user\.admin|hasRole)`)
}

func (e *Extractor) registerPythonPatterns() {
	// Validation
	e.securityPatterns["re_match"] = &securityPattern{
		pattern:  regexp.MustCompile(`re\.(match|search|fullmatch)\s*\(`),
		secType:  "regex_validation",
		effect:   EffectValidates,
		condType: CondTypeValidation,
	}
	e.securityPatterns["isinstance"] = &securityPattern{
		pattern:  regexp.MustCompile(`isinstance\s*\(`),
		secType:  "type_validation",
		effect:   EffectValidates,
		condType: CondTypeTypeCheck,
	}

	// Sanitization
	e.sanitizationFuncs["escape"] = true
	e.sanitizationFuncs["quote"] = true
	e.sanitizationFuncs["html.escape"] = true
	e.sanitizationFuncs["bleach.clean"] = true
	e.sanitizationFuncs["int"] = true
	e.sanitizationFuncs["float"] = true
	e.sanitizationFuncs["str"] = true

	// Auth
	e.authPatterns["login"] = regexp.MustCompile(`(?i)(login_required|is_authenticated|current_user)`)
	e.authPatterns["admin"] = regexp.MustCompile(`(?i)(is_admin|is_superuser|has_perm)`)
}

func (e *Extractor) registerGoPatterns() {
	// Validation
	e.securityPatterns["regexp"] = &securityPattern{
		pattern:  regexp.MustCompile(`regexp\.(Match|MatchString)\s*\(`),
		secType:  "regex_validation",
		effect:   EffectValidates,
		condType: CondTypeValidation,
	}
	e.securityPatterns["validator"] = &securityPattern{
		pattern:  regexp.MustCompile(`validator\.\w+\s*\(`),
		secType:  "validator_library",
		effect:   EffectValidates,
		condType: CondTypeValidation,
	}

	// Sanitization
	e.sanitizationFuncs["html.EscapeString"] = true
	e.sanitizationFuncs["url.QueryEscape"] = true
	e.sanitizationFuncs["strconv.Atoi"] = true
	e.sanitizationFuncs["strconv.ParseInt"] = true

	// Auth
	e.authPatterns["auth"] = regexp.MustCompile(`(?i)(IsAuthenticated|GetUser|RequireAuth)`)
}

func (e *Extractor) registerJavaPatterns() {
	// Validation
	e.securityPatterns["matches"] = &securityPattern{
		pattern:  regexp.MustCompile(`\.matches\s*\(`),
		secType:  "regex_validation",
		effect:   EffectValidates,
		condType: CondTypeValidation,
	}
	e.securityPatterns["instanceof"] = &securityPattern{
		pattern:  regexp.MustCompile(`\s+instanceof\s+`),
		secType:  "instance_check",
		effect:   EffectValidates,
		condType: CondTypeTypeCheck,
	}

	// Sanitization
	e.sanitizationFuncs["StringEscapeUtils.escapeHtml"] = true
	e.sanitizationFuncs["ESAPI.encoder().encodeForHTML"] = true
	e.sanitizationFuncs["Integer.parseInt"] = true
	e.sanitizationFuncs["Long.parseLong"] = true

	// Auth
	e.authPatterns["auth"] = regexp.MustCompile(`(?i)(isAuthenticated|getPrincipal|hasRole|hasAuthority)`)
}

// ExtractFromCode extracts key conditions from code text
func (e *Extractor) ExtractFromCode(code string, filePath string) []*KeyCondition {
	var conditions []*KeyCondition

	lines := strings.Split(code, "\n")
	nestingStack := make([]*KeyCondition, 0)

	for lineNum, line := range lines {
		// Detect condition statements
		if cond := e.detectCondition(line, lineNum+1, filePath, nestingStack); cond != nil {
			// Calculate guarded lines (simple heuristic)
			cond.GuardsLines = e.estimateGuardedLines(lines, lineNum)
			cond.NestingDepth = len(nestingStack)
			if len(nestingStack) > 0 {
				cond.ParentCondition = nestingStack[len(nestingStack)-1]
			}
			conditions = append(conditions, cond)

			// Update nesting stack
			if strings.Contains(line, "{") {
				nestingStack = append(nestingStack, cond)
			}
		}

		// Track block endings
		if strings.Contains(line, "}") && len(nestingStack) > 0 {
			// Pop from stack (simple heuristic - count braces)
			openBraces := strings.Count(line, "{")
			closeBraces := strings.Count(line, "}")
			for i := 0; i < closeBraces-openBraces && len(nestingStack) > 0; i++ {
				nestingStack = nestingStack[:len(nestingStack)-1]
			}
		}
	}

	return conditions
}

// detectCondition checks if a line contains a condition
func (e *Extractor) detectCondition(line string, lineNum int, filePath string, stack []*KeyCondition) *KeyCondition {
	trimmed := strings.TrimSpace(line)

	// Skip non-condition lines
	if !e.isConditionLine(trimmed) {
		return nil
	}

	// Extract the condition expression
	expr := e.extractConditionExpression(trimmed)
	if expr == "" {
		return nil
	}

	cond := &KeyCondition{
		FilePath:   filePath,
		Line:       lineNum,
		Expression: expr,
		Variables:  e.extractVariables(expr),
		IsNegated:  strings.Contains(trimmed, "else") || strings.HasPrefix(strings.TrimSpace(expr), "!"),
	}

	// Classify the condition
	e.classifyCondition(cond)

	return cond
}

// isConditionLine checks if line contains a condition statement
func (e *Extractor) isConditionLine(line string) bool {
	patterns := []string{
		`^\s*if\s*\(`,           // if (condition)
		`^\s*if\s+[^(].*:`,      // Python: if condition:
		`^\s*}\s*else\s*if\s*\(`,
		`^\s*else\s*if\s*\(`,
		`^\s*elif\s+`,           // Python elif
		`^\s*elseif\s*\(`,
		`^\s*}\s*elseif\s*\(`,
		`\?\s*.*\s*:`,           // Ternary
		`^\s*switch\s*\(`,
		`^\s*case\s+`,
	}

	for _, p := range patterns {
		if matched, _ := regexp.MatchString(p, line); matched {
			return true
		}
	}
	return false
}

// extractConditionExpression extracts the condition from a line
func (e *Extractor) extractConditionExpression(line string) string {
	// Extract content between if/elseif and closing paren/colon
	patterns := []struct {
		prefix string
		re     *regexp.Regexp
	}{
		{"if_paren", regexp.MustCompile(`if\s*\((.+)\)\s*[{:]?`)},
		{"if_python", regexp.MustCompile(`if\s+(.+?)\s*:\s*$`)},        // Python: if condition:
		{"elif_python", regexp.MustCompile(`elif\s+(.+?)\s*:\s*$`)},    // Python: elif condition:
		{"elseif", regexp.MustCompile(`(?:else\s*if|elseif)\s*\((.+)\)\s*[{:]?`)},
		{"switch", regexp.MustCompile(`switch\s*\((.+?)\)\s*{?`)},
		{"case", regexp.MustCompile(`case\s+(.+?)\s*:`)},
		{"ternary", regexp.MustCompile(`(.+?)\s*\?\s*.+\s*:`)},
	}

	for _, p := range patterns {
		if matches := p.re.FindStringSubmatch(line); len(matches) > 1 {
			return strings.TrimSpace(matches[1])
		}
	}

	return ""
}

// extractVariables extracts variable names from an expression
func (e *Extractor) extractVariables(expr string) []string {
	var vars []string
	seen := make(map[string]bool)

	var patterns []*regexp.Regexp
	switch e.language {
	case "php":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`\$[a-zA-Z_][a-zA-Z0-9_]*`),
			regexp.MustCompile(`\$_[A-Z]+\s*\[\s*['"]([^'"]+)['"]\s*\]`),
		}
	case "javascript", "typescript":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`\b[a-zA-Z_$][a-zA-Z0-9_$]*\b`),
		}
	case "python":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`),
		}
	case "go", "java":
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`\b[a-zA-Z_][a-zA-Z0-9_]*\b`),
		}
	default:
		patterns = []*regexp.Regexp{
			regexp.MustCompile(`\b[a-zA-Z_$][a-zA-Z0-9_$]*\b`),
		}
	}

	// Filter out keywords
	keywords := map[string]bool{
		"if": true, "else": true, "elseif": true, "switch": true, "case": true,
		"true": true, "false": true, "null": true, "nil": true, "undefined": true,
		"and": true, "or": true, "not": true, "is": true, "in": true,
		"function": true, "return": true, "var": true, "let": true, "const": true,
		"new": true, "this": true, "self": true, "isset": true, "empty": true,
		"instanceof": true, "typeof": true,
	}

	for _, p := range patterns {
		matches := p.FindAllString(expr, -1)
		for _, m := range matches {
			m = strings.TrimSpace(m)
			if !seen[m] && !keywords[strings.ToLower(m)] && len(m) > 1 {
				vars = append(vars, m)
				seen[m] = true
			}
		}
	}

	return vars
}

// classifyCondition classifies a condition and determines its security relevance
func (e *Extractor) classifyCondition(cond *KeyCondition) {
	expr := cond.Expression

	// Check security patterns
	for name, sp := range e.securityPatterns {
		if sp.pattern.MatchString(expr) {
			cond.Type = sp.condType
			cond.Effect = sp.effect
			cond.IsSecurity = true
			cond.SecurityType = sp.secType
			cond.Confidence = 0.9
			return
		}
		_ = name
	}

	// Check auth patterns
	for authType, pattern := range e.authPatterns {
		if pattern.MatchString(expr) {
			cond.Type = CondTypeAuthentication
			cond.Effect = EffectAllows
			cond.IsSecurity = true
			cond.SecurityType = authType + "_check"
			cond.Confidence = 0.85
			return
		}
	}

	// Check for sanitization in condition context
	for funcName := range e.sanitizationFuncs {
		if strings.Contains(expr, funcName) {
			cond.Type = CondTypeSanitization
			cond.Effect = EffectSanitizes
			cond.IsSecurity = true
			cond.SecurityType = "sanitization"
			cond.Confidence = 0.8
			return
		}
	}

	// Check for comparison operators
	if matched, _ := regexp.MatchString(`[<>=!]=?`, expr); matched {
		cond.Type = CondTypeComparison
		cond.Effect = EffectUnknown
		cond.Confidence = 0.5
		return
	}

	// Check for length/count checks
	if matched, _ := regexp.MatchString(`(?i)(strlen|length|count|size)\s*\(`, expr); matched {
		cond.Type = CondTypeLengthCheck
		cond.Effect = EffectValidates
		cond.Confidence = 0.6
		return
	}

	// Default
	cond.Type = CondTypeUnknown
	cond.Effect = EffectUnknown
	cond.Confidence = 0.3
}

// estimateGuardedLines estimates which lines are guarded by a condition
func (e *Extractor) estimateGuardedLines(lines []string, condLineIdx int) []int {
	var guarded []int
	braceCount := 0
	started := false

	for i := condLineIdx; i < len(lines) && i < condLineIdx+100; i++ {
		line := lines[i]
		openBraces := strings.Count(line, "{")
		closeBraces := strings.Count(line, "}")

		if openBraces > 0 && !started {
			started = true
		}

		braceCount += openBraces - closeBraces

		if started && braceCount > 0 {
			guarded = append(guarded, i+1) // 1-indexed
		}

		if started && braceCount <= 0 {
			break
		}
	}

	return guarded
}

// GetConditionPathToLine finds all conditions that guard a specific line
func (e *Extractor) GetConditionPathToLine(conditions []*KeyCondition, targetLine int) *ConditionPath {
	path := &ConditionPath{
		Conditions: make([]*KeyCondition, 0),
		TargetLine: targetLine,
		Feasible:   true,
	}

	for _, cond := range conditions {
		for _, guardedLine := range cond.GuardsLines {
			if guardedLine == targetLine {
				path.Conditions = append(path.Conditions, cond)
				break
			}
		}
	}

	// Check feasibility (simple check for contradictions)
	e.checkPathFeasibility(path)

	return path
}

// checkPathFeasibility checks if a condition path is feasible
func (e *Extractor) checkPathFeasibility(path *ConditionPath) {
	// Simple contradiction detection
	varStates := make(map[string]map[string]bool) // var -> (must_be_null, must_be_not_null, etc.)

	for _, cond := range path.Conditions {
		for _, v := range cond.Variables {
			if varStates[v] == nil {
				varStates[v] = make(map[string]bool)
			}

			// Track null states
			if cond.Type == CondTypeNullCheck {
				if cond.IsNegated {
					// !isset or !empty means we need null
					if varStates[v]["not_null"] {
						path.Feasible = false
						path.Reason = "Contradictory null check on " + v
						return
					}
					varStates[v]["null"] = true
				} else {
					// isset or !empty means not null
					if varStates[v]["null"] {
						path.Feasible = false
						path.Reason = "Contradictory null check on " + v
						return
					}
					varStates[v]["not_null"] = true
				}
			}
		}
	}
}

// HasSecurityGuard checks if any condition provides security protection
func (e *Extractor) HasSecurityGuard(conditions []*KeyCondition) (bool, []*KeyCondition) {
	var guards []*KeyCondition
	for _, cond := range conditions {
		if cond.IsSecurity {
			guards = append(guards, cond)
		}
	}
	return len(guards) > 0, guards
}

// FindSanitizationConditions finds conditions that imply sanitization
func (e *Extractor) FindSanitizationConditions(conditions []*KeyCondition) []*KeyCondition {
	var result []*KeyCondition
	for _, cond := range conditions {
		if cond.Type == CondTypeSanitization || cond.Effect == EffectSanitizes {
			result = append(result, cond)
		}
	}
	return result
}

// FindValidationConditions finds conditions that validate input
func (e *Extractor) FindValidationConditions(conditions []*KeyCondition) []*KeyCondition {
	var result []*KeyCondition
	for _, cond := range conditions {
		if cond.Type == CondTypeValidation || cond.Effect == EffectValidates {
			result = append(result, cond)
		}
	}
	return result
}

// FindAuthConditions finds authentication/authorization conditions
func (e *Extractor) FindAuthConditions(conditions []*KeyCondition) []*KeyCondition {
	var result []*KeyCondition
	for _, cond := range conditions {
		if cond.Type == CondTypeAuthentication || cond.Type == CondTypeAuthorization {
			result = append(result, cond)
		}
	}
	return result
}

// SummarizeConditions creates a summary of conditions
func (e *Extractor) SummarizeConditions(conditions []*KeyCondition) map[string]interface{} {
	summary := map[string]interface{}{
		"total":           len(conditions),
		"security_guards": 0,
		"validations":     0,
		"sanitizations":   0,
		"auth_checks":     0,
		"by_type":         make(map[string]int),
		"by_effect":       make(map[string]int),
	}

	byType := summary["by_type"].(map[string]int)
	byEffect := summary["by_effect"].(map[string]int)

	for _, cond := range conditions {
		byType[string(cond.Type)]++
		byEffect[string(cond.Effect)]++

		if cond.IsSecurity {
			summary["security_guards"] = summary["security_guards"].(int) + 1
		}
		if cond.Type == CondTypeValidation {
			summary["validations"] = summary["validations"].(int) + 1
		}
		if cond.Type == CondTypeSanitization {
			summary["sanitizations"] = summary["sanitizations"].(int) + 1
		}
		if cond.Type == CondTypeAuthentication || cond.Type == CondTypeAuthorization {
			summary["auth_checks"] = summary["auth_checks"].(int) + 1
		}
	}

	return summary
}
