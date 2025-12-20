// Package sink provides a registry of security-sensitive functions (sinks)
// and their reachability analysis, inspired by ATLANTIS's sink-centered approach.
//
// Unlike simple pattern matching, this registry:
// - Categorizes sinks by vulnerability type (SQLi, XSS, command injection, etc.)
// - Tracks which parameters are dangerous (e.g., first arg to exec())
// - Provides severity ratings
// - Integrates with call graph for reachability analysis
package sink

import (
	"regexp"
	"strings"
	"sync"
)

// VulnType represents the type of vulnerability a sink can lead to
type VulnType string

const (
	VulnSQLi          VulnType = "sql_injection"
	VulnCommandInj    VulnType = "command_injection"
	VulnXSS           VulnType = "xss"
	VulnPathTraversal VulnType = "path_traversal"
	VulnFileInclusion VulnType = "file_inclusion"
	VulnCodeExec      VulnType = "code_execution"
	VulnSSRF          VulnType = "ssrf"
	VulnXXE           VulnType = "xxe"
	VulnDeserialization VulnType = "deserialization"
	VulnLDAPI         VulnType = "ldap_injection"
	VulnXPathInj      VulnType = "xpath_injection"
	VulnTemplateInj   VulnType = "template_injection"
	VulnOpenRedirect  VulnType = "open_redirect"
	VulnInfoDisclosure VulnType = "info_disclosure"
)

// Severity represents the severity level of a sink
type Severity int

const (
	SeverityCritical Severity = 4
	SeverityHigh     Severity = 3
	SeverityMedium   Severity = 2
	SeverityLow      Severity = 1
	SeverityInfo     Severity = 0
)

// Sink represents a security-sensitive function
type Sink struct {
	// Identification
	Name         string   `json:"name"`          // e.g., "mysql_query", "exec", "eval"
	Language     string   `json:"language"`      // Target language
	Pattern      string   `json:"pattern"`       // Regex pattern to match
	CompiledPattern *regexp.Regexp `json:"-"` // Compiled regex (not serialized)

	// Classification
	VulnTypes    []VulnType `json:"vuln_types"`   // What vulnerabilities this can cause
	Severity     Severity   `json:"severity"`     // How severe if exploited

	// Parameter tracking (which args are dangerous)
	DangerousParams []int  `json:"dangerous_params"` // 0-indexed param positions that are dangerous
	AllParamsDangerous bool `json:"all_params_dangerous"` // If true, any param is dangerous

	// Call patterns
	IsMethod     bool   `json:"is_method"`     // True if this is a method call
	ClassName    string `json:"class_name,omitempty"` // For method calls

	// Documentation
	Description  string `json:"description"`
	CWE          string `json:"cwe,omitempty"` // CWE identifier
	Remediation  string `json:"remediation,omitempty"`
}

// Match represents a detected sink in code
type Match struct {
	Sink         *Sink    `json:"sink"`
	FilePath     string   `json:"file_path"`
	Line         int      `json:"line"`
	Column       int      `json:"column"`
	Snippet      string   `json:"snippet"`
	FunctionName string   `json:"function_name,omitempty"` // Enclosing function
	ClassName    string   `json:"class_name,omitempty"`    // Enclosing class

	// Taint status (populated during analysis)
	HasTaintedInput bool     `json:"has_tainted_input"`
	TaintedParams   []int    `json:"tainted_params,omitempty"`
	TaintSources    []string `json:"taint_sources,omitempty"`
}

// Registry manages all sink definitions
type Registry struct {
	mu       sync.RWMutex
	sinks    map[string][]*Sink // language -> sinks
	compiled map[string]bool    // track which sinks have been compiled
}

// NewRegistry creates a new sink registry with default definitions
func NewRegistry() *Registry {
	r := &Registry{
		sinks:    make(map[string][]*Sink),
		compiled: make(map[string]bool),
	}
	r.registerDefaults()
	return r
}

// Register adds a sink definition
func (r *Registry) Register(s *Sink) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Compile pattern
	if s.Pattern != "" && s.CompiledPattern == nil {
		if re, err := regexp.Compile(s.Pattern); err == nil {
			s.CompiledPattern = re
		}
	}

	r.sinks[s.Language] = append(r.sinks[s.Language], s)
}

// GetSinks returns all sinks for a language
func (r *Registry) GetSinks(language string) []*Sink {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.sinks[language]
}

// GetSinksByVuln returns all sinks that can cause a specific vulnerability type
func (r *Registry) GetSinksByVuln(vulnType VulnType) []*Sink {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []*Sink
	for _, sinks := range r.sinks {
		for _, s := range sinks {
			for _, vt := range s.VulnTypes {
				if vt == vulnType {
					result = append(result, s)
					break
				}
			}
		}
	}
	return result
}

// MatchText checks if text matches any sink pattern for the given language
func (r *Registry) MatchText(language, text string) []*Match {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matches []*Match
	for _, sink := range r.sinks[language] {
		if sink.CompiledPattern != nil && sink.CompiledPattern.MatchString(text) {
			matches = append(matches, &Match{
				Sink:    sink,
				Snippet: truncate(text, 100),
			})
		}
	}
	return matches
}

// IsSinkCall checks if a function/method call is a sink
func (r *Registry) IsSinkCall(language, funcName string, className string) *Sink {
	r.mu.RLock()
	defer r.mu.RUnlock()

	funcName = strings.TrimSpace(funcName)
	className = strings.TrimSpace(className)

	for _, sink := range r.sinks[language] {
		// Direct name match
		if sink.Name == funcName {
			// If sink requires class, check it
			if sink.ClassName != "" {
				if className == sink.ClassName {
					return sink
				}
			} else {
				return sink
			}
		}

		// Pattern match
		if sink.CompiledPattern != nil {
			target := funcName
			if className != "" {
				target = className + "::" + funcName
			}
			if sink.CompiledPattern.MatchString(target) {
				return sink
			}
		}
	}
	return nil
}

// IsDangerousParam checks if a specific parameter position is dangerous for a sink
func (r *Registry) IsDangerousParam(sink *Sink, paramIndex int) bool {
	if sink == nil {
		return false
	}
	if sink.AllParamsDangerous {
		return true
	}
	for _, idx := range sink.DangerousParams {
		if idx == paramIndex {
			return true
		}
	}
	return false
}

// Stats returns statistics about the registry
func (r *Registry) Stats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	totalSinks := 0
	byLanguage := make(map[string]int)
	bySeverity := make(map[string]int)
	byVuln := make(map[string]int)

	for lang, sinks := range r.sinks {
		totalSinks += len(sinks)
		byLanguage[lang] = len(sinks)

		for _, s := range sinks {
			switch s.Severity {
			case SeverityCritical:
				bySeverity["critical"]++
			case SeverityHigh:
				bySeverity["high"]++
			case SeverityMedium:
				bySeverity["medium"]++
			case SeverityLow:
				bySeverity["low"]++
			default:
				bySeverity["info"]++
			}

			for _, vt := range s.VulnTypes {
				byVuln[string(vt)]++
			}
		}
	}

	return map[string]interface{}{
		"total":       totalSinks,
		"by_language": byLanguage,
		"by_severity": bySeverity,
		"by_vuln":     byVuln,
	}
}

// registerDefaults registers default sink definitions
func (r *Registry) registerDefaults() {
	// PHP Sinks
	r.registerPHPSinks()

	// JavaScript Sinks
	r.registerJavaScriptSinks()

	// Python Sinks
	r.registerPythonSinks()

	// Go Sinks
	r.registerGoSinks()

	// Java Sinks
	r.registerJavaSinks()
}

func (r *Registry) registerPHPSinks() {
	phpSinks := []*Sink{
		// SQL Injection
		{Name: "mysql_query", Language: "php", Pattern: `mysql_query\s*\(`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-89",
			Description: "Legacy MySQL query execution"},
		{Name: "mysqli_query", Language: "php", Pattern: `mysqli_query\s*\(`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityCritical,
			DangerousParams: []int{1}, CWE: "CWE-89",
			Description: "MySQLi query execution"},
		{Name: "mysqli::query", Language: "php", Pattern: `\$\w+\s*->\s*query\s*\(`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityCritical,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-89",
			Description: "MySQLi OOP query execution"},
		{Name: "PDO::query", Language: "php", Pattern: `\$\w+\s*->\s*query\s*\(`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityHigh,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-89",
			Description: "PDO query execution"},
		{Name: "pg_query", Language: "php", Pattern: `pg_query\s*\(`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityCritical,
			DangerousParams: []int{0, 1}, CWE: "CWE-89",
			Description: "PostgreSQL query execution"},

		// Command Injection
		{Name: "exec", Language: "php", Pattern: `\bexec\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Execute shell command"},
		{Name: "shell_exec", Language: "php", Pattern: `shell_exec\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Execute shell command"},
		{Name: "system", Language: "php", Pattern: `\bsystem\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Execute system command"},
		{Name: "passthru", Language: "php", Pattern: `passthru\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Execute and output raw"},
		{Name: "popen", Language: "php", Pattern: `popen\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Open process"},
		{Name: "proc_open", Language: "php", Pattern: `proc_open\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Execute process with pipes"},
		{Name: "backtick", Language: "php", Pattern: "`[^`]+`",
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			AllParamsDangerous: true, CWE: "CWE-78",
			Description: "Backtick command execution"},

		// Code Execution
		{Name: "eval", Language: "php", Pattern: `\beval\s*\(`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-94",
			Description: "Evaluate PHP code"},
		{Name: "assert", Language: "php", Pattern: `\bassert\s*\(`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-94",
			Description: "Assert with code execution"},
		{Name: "create_function", Language: "php", Pattern: `create_function\s*\(`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityCritical,
			DangerousParams: []int{1}, CWE: "CWE-94",
			Description: "Create anonymous function"},
		{Name: "preg_replace_e", Language: "php", Pattern: `preg_replace\s*\(\s*['"][^'"]*e[^'"]*['"]`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityCritical,
			DangerousParams: []int{1, 2}, CWE: "CWE-94",
			Description: "preg_replace with /e modifier"},

		// File Inclusion
		{Name: "include", Language: "php", Pattern: `\binclude\s+`,
			VulnTypes: []VulnType{VulnFileInclusion, VulnPathTraversal}, Severity: SeverityCritical,
			AllParamsDangerous: true, CWE: "CWE-98",
			Description: "Include file"},
		{Name: "include_once", Language: "php", Pattern: `include_once\s+`,
			VulnTypes: []VulnType{VulnFileInclusion, VulnPathTraversal}, Severity: SeverityCritical,
			AllParamsDangerous: true, CWE: "CWE-98",
			Description: "Include file once"},
		{Name: "require", Language: "php", Pattern: `\brequire\s+`,
			VulnTypes: []VulnType{VulnFileInclusion, VulnPathTraversal}, Severity: SeverityCritical,
			AllParamsDangerous: true, CWE: "CWE-98",
			Description: "Require file"},
		{Name: "require_once", Language: "php", Pattern: `require_once\s+`,
			VulnTypes: []VulnType{VulnFileInclusion, VulnPathTraversal}, Severity: SeverityCritical,
			AllParamsDangerous: true, CWE: "CWE-98",
			Description: "Require file once"},

		// File Operations
		{Name: "file_get_contents", Language: "php", Pattern: `file_get_contents\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal, VulnSSRF}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-22",
			Description: "Read file contents"},
		{Name: "file_put_contents", Language: "php", Pattern: `file_put_contents\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityHigh,
			DangerousParams: []int{0, 1}, CWE: "CWE-22",
			Description: "Write file contents"},
		{Name: "fopen", Language: "php", Pattern: `fopen\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal, VulnSSRF}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-22",
			Description: "Open file"},
		{Name: "readfile", Language: "php", Pattern: `readfile\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-22",
			Description: "Read and output file"},
		{Name: "unlink", Language: "php", Pattern: `unlink\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-22",
			Description: "Delete file"},
		{Name: "move_uploaded_file", Language: "php", Pattern: `move_uploaded_file\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityMedium,
			DangerousParams: []int{1}, CWE: "CWE-22",
			Description: "Move uploaded file"},

		// XSS
		{Name: "echo", Language: "php", Pattern: `\becho\s+`,
			VulnTypes: []VulnType{VulnXSS}, Severity: SeverityMedium,
			AllParamsDangerous: true, CWE: "CWE-79",
			Description: "Echo output"},
		{Name: "print", Language: "php", Pattern: `\bprint\s+`,
			VulnTypes: []VulnType{VulnXSS}, Severity: SeverityMedium,
			AllParamsDangerous: true, CWE: "CWE-79",
			Description: "Print output"},
		{Name: "printf", Language: "php", Pattern: `printf\s*\(`,
			VulnTypes: []VulnType{VulnXSS}, Severity: SeverityMedium,
			DangerousParams: []int{0}, CWE: "CWE-79",
			Description: "Formatted print"},

		// Deserialization
		{Name: "unserialize", Language: "php", Pattern: `unserialize\s*\(`,
			VulnTypes: []VulnType{VulnDeserialization}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-502",
			Description: "Deserialize PHP data"},

		// SSRF
		{Name: "curl_exec", Language: "php", Pattern: `curl_exec\s*\(`,
			VulnTypes: []VulnType{VulnSSRF}, Severity: SeverityHigh,
			AllParamsDangerous: true, CWE: "CWE-918",
			Description: "cURL execution"},

		// Header Injection
		{Name: "header", Language: "php", Pattern: `\bheader\s*\(`,
			VulnTypes: []VulnType{VulnOpenRedirect}, Severity: SeverityMedium,
			DangerousParams: []int{0}, CWE: "CWE-601",
			Description: "Set HTTP header"},

		// XXE
		{Name: "simplexml_load_string", Language: "php", Pattern: `simplexml_load_string\s*\(`,
			VulnTypes: []VulnType{VulnXXE}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-611",
			Description: "Parse XML string"},
		{Name: "DOMDocument::loadXML", Language: "php", Pattern: `loadXML\s*\(`,
			VulnTypes: []VulnType{VulnXXE}, Severity: SeverityHigh,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-611",
			Description: "Load XML into DOM"},
	}

	for _, s := range phpSinks {
		r.Register(s)
	}
}

func (r *Registry) registerJavaScriptSinks() {
	jsSinks := []*Sink{
		// Code Execution
		{Name: "eval", Language: "javascript", Pattern: `\beval\s*\(`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-94",
			Description: "Evaluate JavaScript code"},
		{Name: "Function", Language: "javascript", Pattern: `\bnew\s+Function\s*\(`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityCritical,
			AllParamsDangerous: true, CWE: "CWE-94",
			Description: "Create function from string"},
		{Name: "setTimeout", Language: "javascript", Pattern: `setTimeout\s*\(\s*['"]`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-94",
			Description: "setTimeout with string"},
		{Name: "setInterval", Language: "javascript", Pattern: `setInterval\s*\(\s*['"]`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-94",
			Description: "setInterval with string"},

		// XSS (DOM)
		{Name: "innerHTML", Language: "javascript", Pattern: `\.innerHTML\s*=`,
			VulnTypes: []VulnType{VulnXSS}, Severity: SeverityHigh,
			AllParamsDangerous: true, CWE: "CWE-79",
			Description: "Set innerHTML"},
		{Name: "outerHTML", Language: "javascript", Pattern: `\.outerHTML\s*=`,
			VulnTypes: []VulnType{VulnXSS}, Severity: SeverityHigh,
			AllParamsDangerous: true, CWE: "CWE-79",
			Description: "Set outerHTML"},
		{Name: "document.write", Language: "javascript", Pattern: `document\.write\s*\(`,
			VulnTypes: []VulnType{VulnXSS}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-79",
			Description: "Document write"},
		{Name: "document.writeln", Language: "javascript", Pattern: `document\.writeln\s*\(`,
			VulnTypes: []VulnType{VulnXSS}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-79",
			Description: "Document writeln"},
		{Name: "insertAdjacentHTML", Language: "javascript", Pattern: `\.insertAdjacentHTML\s*\(`,
			VulnTypes: []VulnType{VulnXSS}, Severity: SeverityHigh,
			DangerousParams: []int{1}, IsMethod: true, CWE: "CWE-79",
			Description: "Insert adjacent HTML"},

		// Command Injection (Node.js)
		{Name: "exec", Language: "javascript", Pattern: `child_process.*\.exec\s*\(|exec\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Execute shell command"},
		{Name: "execSync", Language: "javascript", Pattern: `\.execSync\s*\(|execSync\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Execute shell command synchronously"},
		{Name: "spawn", Language: "javascript", Pattern: `child_process.*\.spawn\s*\(|spawn\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityHigh,
			DangerousParams: []int{0, 1}, CWE: "CWE-78",
			Description: "Spawn process"},

		// SQL Injection (if using raw queries)
		{Name: "query", Language: "javascript", Pattern: `\.query\s*\(\s*['"\x60]`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityCritical,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-89",
			Description: "Raw SQL query"},

		// Path Traversal
		{Name: "readFile", Language: "javascript", Pattern: `fs\.readFile\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-22",
			Description: "Read file"},
		{Name: "readFileSync", Language: "javascript", Pattern: `fs\.readFileSync\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-22",
			Description: "Read file synchronously"},
		{Name: "writeFile", Language: "javascript", Pattern: `fs\.writeFile\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityHigh,
			DangerousParams: []int{0, 1}, CWE: "CWE-22",
			Description: "Write file"},

		// Open Redirect
		{Name: "location.href", Language: "javascript", Pattern: `location\.href\s*=`,
			VulnTypes: []VulnType{VulnOpenRedirect}, Severity: SeverityMedium,
			AllParamsDangerous: true, CWE: "CWE-601",
			Description: "Set location href"},
		{Name: "location.replace", Language: "javascript", Pattern: `location\.replace\s*\(`,
			VulnTypes: []VulnType{VulnOpenRedirect}, Severity: SeverityMedium,
			DangerousParams: []int{0}, CWE: "CWE-601",
			Description: "Location replace"},
		{Name: "window.open", Language: "javascript", Pattern: `window\.open\s*\(`,
			VulnTypes: []VulnType{VulnOpenRedirect}, Severity: SeverityMedium,
			DangerousParams: []int{0}, CWE: "CWE-601",
			Description: "Window open"},
	}

	for _, s := range jsSinks {
		r.Register(s)
	}

	// Also register for TypeScript
	for _, s := range jsSinks {
		tsSink := *s
		tsSink.Language = "typescript"
		r.Register(&tsSink)
	}
}

func (r *Registry) registerPythonSinks() {
	pythonSinks := []*Sink{
		// Code Execution
		{Name: "eval", Language: "python", Pattern: `\beval\s*\(`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-94",
			Description: "Evaluate Python expression"},
		{Name: "exec", Language: "python", Pattern: `\bexec\s*\(`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-94",
			Description: "Execute Python code"},
		{Name: "compile", Language: "python", Pattern: `\bcompile\s*\(`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-94",
			Description: "Compile Python code"},

		// Command Injection
		{Name: "os.system", Language: "python", Pattern: `os\.system\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Execute system command"},
		{Name: "os.popen", Language: "python", Pattern: `os\.popen\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Open pipe to command"},
		{Name: "subprocess.call", Language: "python", Pattern: `subprocess\.call\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Subprocess call"},
		{Name: "subprocess.run", Language: "python", Pattern: `subprocess\.run\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Subprocess run"},
		{Name: "subprocess.Popen", Language: "python", Pattern: `subprocess\.Popen\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Subprocess Popen"},

		// SQL Injection
		{Name: "execute", Language: "python", Pattern: `\.execute\s*\(\s*['"]`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityCritical,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-89",
			Description: "Execute raw SQL"},
		{Name: "executemany", Language: "python", Pattern: `\.executemany\s*\(`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityHigh,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-89",
			Description: "Execute many SQL statements"},

		// Deserialization
		{Name: "pickle.loads", Language: "python", Pattern: `pickle\.loads\s*\(`,
			VulnTypes: []VulnType{VulnDeserialization}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-502",
			Description: "Deserialize pickle data"},
		{Name: "pickle.load", Language: "python", Pattern: `pickle\.load\s*\(`,
			VulnTypes: []VulnType{VulnDeserialization}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-502",
			Description: "Load pickle from file"},
		{Name: "yaml.load", Language: "python", Pattern: `yaml\.load\s*\(`,
			VulnTypes: []VulnType{VulnDeserialization}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-502",
			Description: "Unsafe YAML load"},

		// Template Injection
		{Name: "Template", Language: "python", Pattern: `Template\s*\(\s*['"].*\$`,
			VulnTypes: []VulnType{VulnTemplateInj}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-94",
			Description: "String template with user input"},

		// XXE
		{Name: "xml.etree.ElementTree.parse", Language: "python", Pattern: `ElementTree\.parse\s*\(`,
			VulnTypes: []VulnType{VulnXXE}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-611",
			Description: "Parse XML file"},
	}

	for _, s := range pythonSinks {
		r.Register(s)
	}
}

func (r *Registry) registerGoSinks() {
	goSinks := []*Sink{
		// Command Injection
		{Name: "exec.Command", Language: "go", Pattern: `exec\.Command\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0, 1}, CWE: "CWE-78",
			Description: "Execute command"},
		{Name: "exec.CommandContext", Language: "go", Pattern: `exec\.CommandContext\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{1, 2}, CWE: "CWE-78",
			Description: "Execute command with context"},

		// SQL Injection
		{Name: "db.Query", Language: "go", Pattern: `\.Query\s*\(\s*[^,\)]+\+`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityCritical,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-89",
			Description: "SQL query with concatenation"},
		{Name: "db.Exec", Language: "go", Pattern: `\.Exec\s*\(\s*[^,\)]+\+`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityCritical,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-89",
			Description: "SQL exec with concatenation"},

		// Path Traversal
		{Name: "os.Open", Language: "go", Pattern: `os\.Open\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-22",
			Description: "Open file"},
		{Name: "os.ReadFile", Language: "go", Pattern: `os\.ReadFile\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-22",
			Description: "Read file"},
		{Name: "ioutil.ReadFile", Language: "go", Pattern: `ioutil\.ReadFile\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-22",
			Description: "Read file (deprecated)"},

		// Template Injection
		{Name: "template.HTML", Language: "go", Pattern: `template\.HTML\s*\(`,
			VulnTypes: []VulnType{VulnXSS}, Severity: SeverityMedium,
			DangerousParams: []int{0}, CWE: "CWE-79",
			Description: "Unescaped HTML in template"},

		// SSRF
		{Name: "http.Get", Language: "go", Pattern: `http\.Get\s*\(`,
			VulnTypes: []VulnType{VulnSSRF}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-918",
			Description: "HTTP GET request"},
		{Name: "http.Post", Language: "go", Pattern: `http\.Post\s*\(`,
			VulnTypes: []VulnType{VulnSSRF}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-918",
			Description: "HTTP POST request"},
	}

	for _, s := range goSinks {
		r.Register(s)
	}
}

func (r *Registry) registerJavaSinks() {
	javaSinks := []*Sink{
		// Command Injection
		{Name: "Runtime.exec", Language: "java", Pattern: `Runtime\.getRuntime\(\)\.exec\s*\(|\.exec\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			DangerousParams: []int{0}, CWE: "CWE-78",
			Description: "Execute runtime command"},
		{Name: "ProcessBuilder", Language: "java", Pattern: `new\s+ProcessBuilder\s*\(`,
			VulnTypes: []VulnType{VulnCommandInj}, Severity: SeverityCritical,
			AllParamsDangerous: true, CWE: "CWE-78",
			Description: "Process builder"},

		// SQL Injection
		{Name: "Statement.executeQuery", Language: "java", Pattern: `\.executeQuery\s*\(\s*[^")]+\+`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityCritical,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-89",
			Description: "Execute query with concatenation"},
		{Name: "Statement.execute", Language: "java", Pattern: `\.execute\s*\(\s*[^")]+\+`,
			VulnTypes: []VulnType{VulnSQLi}, Severity: SeverityCritical,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-89",
			Description: "Execute with concatenation"},

		// Code Execution
		{Name: "ScriptEngine.eval", Language: "java", Pattern: `ScriptEngine.*\.eval\s*\(`,
			VulnTypes: []VulnType{VulnCodeExec}, Severity: SeverityCritical,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-94",
			Description: "Script engine evaluation"},

		// Deserialization
		{Name: "ObjectInputStream.readObject", Language: "java", Pattern: `ObjectInputStream.*\.readObject\s*\(`,
			VulnTypes: []VulnType{VulnDeserialization}, Severity: SeverityCritical,
			AllParamsDangerous: true, IsMethod: true, CWE: "CWE-502",
			Description: "Deserialize object"},

		// XXE
		{Name: "DocumentBuilder.parse", Language: "java", Pattern: `DocumentBuilder.*\.parse\s*\(`,
			VulnTypes: []VulnType{VulnXXE}, Severity: SeverityHigh,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-611",
			Description: "Parse XML document"},
		{Name: "SAXParser.parse", Language: "java", Pattern: `SAXParser.*\.parse\s*\(`,
			VulnTypes: []VulnType{VulnXXE}, Severity: SeverityHigh,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-611",
			Description: "SAX parse XML"},

		// Path Traversal
		{Name: "File", Language: "java", Pattern: `new\s+File\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityMedium,
			DangerousParams: []int{0}, CWE: "CWE-22",
			Description: "Create file object"},
		{Name: "FileInputStream", Language: "java", Pattern: `new\s+FileInputStream\s*\(`,
			VulnTypes: []VulnType{VulnPathTraversal}, Severity: SeverityHigh,
			DangerousParams: []int{0}, CWE: "CWE-22",
			Description: "Open file input stream"},

		// LDAP Injection
		{Name: "DirContext.search", Language: "java", Pattern: `DirContext.*\.search\s*\(`,
			VulnTypes: []VulnType{VulnLDAPI}, Severity: SeverityHigh,
			DangerousParams: []int{1}, IsMethod: true, CWE: "CWE-90",
			Description: "LDAP search"},

		// XPath Injection
		{Name: "XPath.evaluate", Language: "java", Pattern: `XPath.*\.evaluate\s*\(`,
			VulnTypes: []VulnType{VulnXPathInj}, Severity: SeverityHigh,
			DangerousParams: []int{0}, IsMethod: true, CWE: "CWE-643",
			Description: "XPath evaluation"},
	}

	for _, s := range javaSinks {
		r.Register(s)
	}
}

// Helper function
func truncate(s string, maxLen int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}
