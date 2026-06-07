package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// moodleDefinitions returns definitions for Moodle LMS's parameter retrieval
// functions. Moodle uses required_param() and optional_param() as its canonical
// way to read and type-check GET/POST parameters (instead of accessing
// superglobals directly).
func moodleDefinitions() []common.Definition {
	return []common.Definition{
		// required_param('key', PARAM_*) — reads a required parameter; throws if missing.
		{
			Name:         "required_param() (Moodle)",
			Pattern:      `\brequired_param\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Moodle required parameter",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\brequired_param\s*\(\s*['"]([^'"]+)['"]`,
		},
		// optional_param('key', default, PARAM_*) — reads an optional parameter with default.
		{
			Name:         "optional_param() (Moodle)",
			Pattern:      `\boptional_param\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Moodle optional parameter",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\boptional_param\s*\(\s*['"]([^'"]+)['"]`,
		},
		// required_param_array('key', PARAM_*) — reads a required array parameter.
		{
			Name:         "required_param_array() (Moodle)",
			Pattern:      `\brequired_param_array\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Moodle required param array",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\brequired_param_array\s*\(\s*['"]([^'"]+)['"]`,
		},
		// optional_param_array('key', default, PARAM_*) — reads an optional array parameter.
		{
			Name:         "optional_param_array() (Moodle)",
			Pattern:      `\boptional_param_array\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Moodle optional param array",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\boptional_param_array\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
