package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// roundcubeDefinitions returns definitions for Roundcube webmail's input
// utility methods. Roundcube uses static methods on rcube_utils to safely
// read and sanitize request input.
func roundcubeDefinitions() []common.Definition {
	return []common.Definition{
		// rcube_utils::get_input_value('key', source) — reads input from specified source.
		{
			Name:         "rcube_utils::get_input_value() (Roundcube)",
			Pattern:      `\brcube_utils\s*::\s*get_input_value\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Roundcube get_input_value()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `get_input_value\s*\(\s*['"]([^'"]+)['"]`,
		},
		// rcube_utils::get_input_string('key', source) — reads input as string.
		{
			Name:         "rcube_utils::get_input_string() (Roundcube)",
			Pattern:      `\brcube_utils\s*::\s*get_input_string\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Roundcube get_input_string()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `get_input_string\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
