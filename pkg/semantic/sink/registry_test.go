package sink

import (
	"testing"
)

func TestNewRegistry(t *testing.T) {
	r := NewRegistry()

	// Check that defaults are registered
	phpSinks := r.GetSinks("php")
	if len(phpSinks) == 0 {
		t.Error("expected PHP sinks to be registered")
	}

	jsSinks := r.GetSinks("javascript")
	if len(jsSinks) == 0 {
		t.Error("expected JavaScript sinks to be registered")
	}

	pythonSinks := r.GetSinks("python")
	if len(pythonSinks) == 0 {
		t.Error("expected Python sinks to be registered")
	}

	goSinks := r.GetSinks("go")
	if len(goSinks) == 0 {
		t.Error("expected Go sinks to be registered")
	}

	javaSinks := r.GetSinks("java")
	if len(javaSinks) == 0 {
		t.Error("expected Java sinks to be registered")
	}
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()

	customSink := &Sink{
		Name:        "custom_danger",
		Language:    "custom",
		Pattern:     `customDanger\s*\(`,
		VulnTypes:   []VulnType{VulnCodeExec},
		Severity:    SeverityCritical,
		Description: "Custom dangerous function",
	}

	r.Register(customSink)

	sinks := r.GetSinks("custom")
	if len(sinks) != 1 {
		t.Fatalf("expected 1 custom sink, got %d", len(sinks))
	}

	if sinks[0].Name != "custom_danger" {
		t.Errorf("expected sink name 'custom_danger', got '%s'", sinks[0].Name)
	}

	// Verify pattern was compiled
	if sinks[0].CompiledPattern == nil {
		t.Error("expected pattern to be compiled")
	}
}

func TestRegistry_GetSinksByVuln(t *testing.T) {
	r := NewRegistry()

	// Get all SQL injection sinks
	sqliSinks := r.GetSinksByVuln(VulnSQLi)
	if len(sqliSinks) == 0 {
		t.Error("expected SQL injection sinks")
	}

	// Verify they all have SQLi as a vuln type
	for _, sink := range sqliSinks {
		found := false
		for _, vt := range sink.VulnTypes {
			if vt == VulnSQLi {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("sink %s should have VulnSQLi", sink.Name)
		}
	}

	// Get command injection sinks
	cmdSinks := r.GetSinksByVuln(VulnCommandInj)
	if len(cmdSinks) == 0 {
		t.Error("expected command injection sinks")
	}
}

func TestRegistry_MatchText(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		language string
		code     string
		expected int  // minimum number of matches
		sinkName string // expected sink name (one of them)
	}{
		{"php", "mysql_query($sql)", 1, "mysql_query"},
		{"php", "exec($cmd)", 1, "exec"},
		{"php", "eval($_GET['code'])", 1, "eval"},
		{"php", "unserialize($data)", 1, "unserialize"},
		{"javascript", "eval(userInput)", 1, "eval"},
		{"javascript", "document.write(html)", 1, "document.write"},
		{"python", "os.system(cmd)", 1, "os.system"},
		{"python", "pickle.loads(data)", 1, "pickle.loads"},
		{"go", "exec.Command(cmd)", 1, "exec.Command"},
		{"java", "Runtime.getRuntime().exec(cmd)", 1, "Runtime.exec"},
	}

	for _, tc := range tests {
		t.Run(tc.language+"_"+tc.sinkName, func(t *testing.T) {
			matches := r.MatchText(tc.language, tc.code)
			if len(matches) < tc.expected {
				t.Errorf("expected at least %d matches for '%s', got %d",
					tc.expected, tc.code, len(matches))
			}

			// Verify one of the matches has the expected sink name
			found := false
			for _, m := range matches {
				if m.Sink.Name == tc.sinkName {
					found = true
					break
				}
			}
			if !found && len(matches) > 0 {
				names := make([]string, len(matches))
				for i, m := range matches {
					names[i] = m.Sink.Name
				}
				t.Errorf("expected to find sink '%s', got %v", tc.sinkName, names)
			}
		})
	}
}

func TestRegistry_IsSinkCall(t *testing.T) {
	r := NewRegistry()

	tests := []struct {
		language  string
		funcName  string
		className string
		expected  bool
	}{
		{"php", "mysql_query", "", true},
		{"php", "exec", "", true},
		{"php", "safe_function", "", false},
		{"javascript", "eval", "", true},
		{"javascript", "safeFn", "", false},
		{"python", "eval", "", true},
		{"go", "exec.Command", "", true},
		{"java", "Runtime.exec", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.language+"_"+tc.funcName, func(t *testing.T) {
			sink := r.IsSinkCall(tc.language, tc.funcName, tc.className)
			if tc.expected && sink == nil {
				t.Errorf("expected '%s' to be recognized as a sink", tc.funcName)
			}
			if !tc.expected && sink != nil {
				t.Errorf("expected '%s' to NOT be recognized as a sink", tc.funcName)
			}
		})
	}
}

func TestRegistry_IsDangerousParam(t *testing.T) {
	r := NewRegistry()

	// Find mysql_query sink
	sink := r.IsSinkCall("php", "mysql_query", "")
	if sink == nil {
		t.Fatal("mysql_query should be a sink")
	}

	// First param (index 0) should be dangerous
	if !r.IsDangerousParam(sink, 0) {
		t.Error("param 0 of mysql_query should be dangerous")
	}

	// Second param (index 1) should not be dangerous (only param 0 is)
	if r.IsDangerousParam(sink, 1) {
		t.Error("param 1 of mysql_query should not be dangerous")
	}

	// Find eval sink (all params dangerous)
	evalSink := r.IsSinkCall("php", "eval", "")
	if evalSink == nil {
		t.Fatal("eval should be a sink")
	}

	if !r.IsDangerousParam(evalSink, 0) {
		t.Error("param 0 of eval should be dangerous")
	}
}

func TestRegistry_Stats(t *testing.T) {
	r := NewRegistry()

	stats := r.Stats()

	// Check that we have stats
	if stats["total"].(int) == 0 {
		t.Error("expected total sinks > 0")
	}

	byLanguage, ok := stats["by_language"].(map[string]int)
	if !ok {
		t.Fatal("expected by_language map")
	}

	if byLanguage["php"] == 0 {
		t.Error("expected PHP sinks > 0")
	}

	bySeverity, ok := stats["by_severity"].(map[string]int)
	if !ok {
		t.Fatal("expected by_severity map")
	}

	if bySeverity["critical"] == 0 {
		t.Error("expected critical sinks > 0")
	}

	byVuln, ok := stats["by_vuln"].(map[string]int)
	if !ok {
		t.Fatal("expected by_vuln map")
	}

	if byVuln[string(VulnSQLi)] == 0 {
		t.Error("expected SQL injection sinks > 0")
	}

	t.Logf("Registry stats: %+v", stats)
}

func TestVulnTypes(t *testing.T) {
	// Verify all vuln types are defined
	vulnTypes := []VulnType{
		VulnSQLi,
		VulnCommandInj,
		VulnXSS,
		VulnPathTraversal,
		VulnFileInclusion,
		VulnCodeExec,
		VulnSSRF,
		VulnXXE,
		VulnDeserialization,
		VulnLDAPI,
		VulnXPathInj,
		VulnTemplateInj,
		VulnOpenRedirect,
		VulnInfoDisclosure,
	}

	for _, vt := range vulnTypes {
		if string(vt) == "" {
			t.Errorf("vuln type should not be empty")
		}
	}
}

func TestSeverityLevels(t *testing.T) {
	// Verify severity ordering
	if SeverityCritical <= SeverityHigh {
		t.Error("Critical should be higher than High")
	}
	if SeverityHigh <= SeverityMedium {
		t.Error("High should be higher than Medium")
	}
	if SeverityMedium <= SeverityLow {
		t.Error("Medium should be higher than Low")
	}
	if SeverityLow <= SeverityInfo {
		t.Error("Low should be higher than Info")
	}
}

func TestPHPSinksComprehensive(t *testing.T) {
	r := NewRegistry()

	// Test PHP-specific patterns
	phpPatterns := []struct {
		code     string
		vuln     VulnType
		severity Severity
	}{
		// SQL Injection
		{`$result = mysql_query("SELECT * FROM users WHERE id=" . $id);`, VulnSQLi, SeverityCritical},
		{`$stmt = $pdo->query($sql);`, VulnSQLi, SeverityHigh},

		// Command Injection
		{`exec("ls " . $dir);`, VulnCommandInj, SeverityCritical},
		{`system($command);`, VulnCommandInj, SeverityCritical},
		{`shell_exec($cmd);`, VulnCommandInj, SeverityCritical},
		{"`cat $file`", VulnCommandInj, SeverityCritical},

		// Code Execution
		{`eval($_POST['code']);`, VulnCodeExec, SeverityCritical},
		{`assert($userInput);`, VulnCodeExec, SeverityCritical},

		// File Inclusion
		{`include $_GET['page'];`, VulnFileInclusion, SeverityCritical},
		{`require $template;`, VulnFileInclusion, SeverityCritical},

		// Deserialization
		{`$obj = unserialize($_COOKIE['data']);`, VulnDeserialization, SeverityCritical},

		// XSS
		{`echo $userInput;`, VulnXSS, SeverityMedium},
	}

	for _, tc := range phpPatterns {
		t.Run(tc.code[:min(30, len(tc.code))], func(t *testing.T) {
			matches := r.MatchText("php", tc.code)
			if len(matches) == 0 {
				t.Errorf("expected match for: %s", tc.code)
				return
			}

			// Check at least one match has expected vuln type
			found := false
			for _, m := range matches {
				for _, vt := range m.Sink.VulnTypes {
					if vt == tc.vuln {
						found = true
						// Also check severity
						if m.Sink.Severity < tc.severity {
							t.Errorf("expected severity >= %d, got %d for %s",
								tc.severity, m.Sink.Severity, tc.code)
						}
						break
					}
				}
			}
			if !found {
				t.Errorf("expected vuln type %s for: %s", tc.vuln, tc.code)
			}
		})
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
