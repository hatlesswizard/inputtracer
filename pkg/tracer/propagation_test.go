package tracer

import (
	"testing"
)

func newTestPropagator() *TaintPropagator {
	return &TaintPropagator{
		state:    NewFullAnalysisState(),
		language: "php",
	}
}

func TestMatchesVariable(t *testing.T) {
	tp := newTestPropagator()

	tests := []struct {
		name    string
		lang    string
		value   string
		varName string
		want    bool
	}{
		// Exact match
		{"php exact match", "php", "$input", "$input", true},
		{"php no match different var", "php", "$input", "$output", false},

		// Property access
		{"php dot access", "php", "$obj.prop", "$obj", true},
		{"php bracket access", "php", "$arr['key']", "$arr", true},
		{"php no match on dot", "php", "$obj.prop", "$other", false},

		// Expression containment (PHP)
		{"php in concatenation", "php", "$prefix . $input . $suffix", "$input", true},
		{"php in function call", "php", "htmlspecialchars($data)", "$data", true},
		{"php in array literal", "php", "array($input, $other)", "$input", true},
		{"php in ternary", "php", "$val ? $val : 'default'", "$val", true},
		{"php in complex expr", "php", "$args = array_merge($defaults, $args)", "$args", true},

		// Substring rejection (PHP)
		{"php no substring match", "php", "$order_id = 5", "$order", false},
		{"php no substring match reversed", "php", "$order = 5", "$order_id", false},

		// Standard variables (Go)
		{"go in expression", "go", "bar + foo * baz", "foo", true},
		{"go no substring", "go", "foobar = 1", "foo", false},
		{"go exact", "go", "x", "x", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tp.language = tt.lang
			got := tp.matchesVariable(tt.value, tt.varName)
			if got != tt.want {
				t.Errorf("matchesVariable(%q, %q) = %v, want %v", tt.value, tt.varName, got, tt.want)
			}
		})
	}
}
