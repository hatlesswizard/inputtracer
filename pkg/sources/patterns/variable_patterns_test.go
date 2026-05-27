package patterns

import (
	"regexp"
	"testing"
)

func TestDefaultVariablePattern(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		// PHP variables
		{"php simple var", "$var", true},
		{"php superglobal", "$_POST", true},
		{"php snake_case", "$my_variable", true},
		{"php short name", "$args", true},

		// Ruby variables
		{"ruby instance var", "@name", true},
		{"ruby class var", "@@count", true},
		{"ruby instance snake_case", "@instance_var", true},
		{"ruby class snake_case", "@@class_var", true},

		// Standard identifiers
		{"standard lowercase", "foo", true},
		{"standard underscore prefix", "_private", true},
		{"standard camelCase", "camelCase", true},
		{"standard UPPER_SNAKE", "CONST_VAL", true},
		{"standard plain", "plain_var", true},

		// Invalid
		{"leading digits", "123invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultVariablePattern.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("DefaultVariablePattern.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestGetVariablePatterns(t *testing.T) {
	t.Run("php patterns match dollar-prefixed vars", func(t *testing.T) {
		pats := GetVariablePatterns("php")
		if len(pats) == 0 {
			t.Fatal("GetVariablePatterns(\"php\") returned no patterns")
		}
		phpVarPattern := pats[0]

		tests := []struct {
			input string
			want  bool
		}{
			{"$username", true},
			{"$_GET", true},
			{"$order_id", true},
			{"plain", false},
		}
		for _, tt := range tests {
			got := phpVarPattern.MatchString(tt.input)
			if got != tt.want {
				t.Errorf("PHP pattern.MatchString(%q) = %v, want %v", tt.input, got, tt.want)
			}
		}
	})

	t.Run("unknown language returns default", func(t *testing.T) {
		pats := GetVariablePatterns("unknown_language")
		if len(pats) != 1 {
			t.Fatalf("GetVariablePatterns(\"unknown_language\") returned %d patterns, want 1", len(pats))
		}
		if pats[0] != DefaultVariablePattern {
			t.Error("GetVariablePatterns(\"unknown_language\") did not return DefaultVariablePattern")
		}
	})
}

func TestVariableBoundaryPattern(t *testing.T) {
	tests := []struct {
		name    string
		varName string
		text    string
		want    bool
	}{
		// PHP $-prefixed variables
		{"php var in array literal", "$input", `array($input, $other)`, true},
		{"php var in concatenation", "$name", `$prefix . $name . $suffix`, true},
		{"php var in function arg", "$data", `htmlspecialchars($data)`, true},
		{"php var at start of string", "$args", `$args['key']`, true},
		{"php var in assignment RHS", "$user_input", `$result = $user_input`, true},
		{"php var in ternary", "$val", `$val ? $val : 'default'`, true},
		{"php exact match", "$x", `$x`, true},
		{"php no substring match", "$order", `$order_id = 5`, false},
		{"php no substring match reversed", "$order_id", `$order = 5`, false},

		// Ruby @-prefixed variables
		{"ruby instance var", "@name", `puts @name`, true},
		{"ruby class var", "@@count", `@@count += 1`, true},
		{"ruby instance var in interpolation", "@user", `"Hello #{@user}"`, true},
		{"ruby no substring", "@item", `@item_list.each`, false},

		// Standard word-boundary variables
		{"standard var in expression", "foo", `bar + foo * baz`, true},
		{"standard var in call", "data", `process(data, opts)`, true},
		{"standard var exact", "x", `x`, true},
		{"standard no substring", "foo", `foobar = 1`, false},
		{"standard no substring reversed", "foobar", `foo = 1`, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pat := VariableBoundaryPattern(tt.varName)
			re := regexp.MustCompile(pat)
			got := re.MatchString(tt.text)
			if got != tt.want {
				t.Errorf("VariableBoundaryPattern(%q) on %q: got %v, want %v (pattern: %s)",
					tt.varName, tt.text, got, tt.want, pat)
			}
		})
	}
}
