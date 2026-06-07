package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// mantisbtDefinitions returns definitions for MantisBT bug tracker's gpc_*
// family of input getter functions. MantisBT uses "GPC" (Get/Post/Cookie)
// as a naming convention for its input abstraction layer.
func mantisbtDefinitions() []common.Definition {
	return []common.Definition{
		// gpc_get_string('key') — reads a string parameter from GET/POST/COOKIE.
		{
			Name:         "gpc_get_string() (MantisBT)",
			Pattern:      `\bgpc_get_string\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "MantisBT string input getter",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bgpc_get_string\s*\(\s*['"]([^'"]+)['"]`,
		},
		// gpc_get_int('key') — reads an integer parameter from GET/POST/COOKIE.
		{
			Name:         "gpc_get_int() (MantisBT)",
			Pattern:      `\bgpc_get_int\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "MantisBT integer input getter",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bgpc_get_int\s*\(\s*['"]([^'"]+)['"]`,
		},
		// gpc_get_bool('key') — reads a boolean parameter from GET/POST/COOKIE.
		{
			Name:         "gpc_get_bool() (MantisBT)",
			Pattern:      `\bgpc_get_bool\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "MantisBT boolean input getter",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bgpc_get_bool\s*\(\s*['"]([^'"]+)['"]`,
		},
		// gpc_get_custom_field() — reads a custom field value from request.
		{
			Name:        "gpc_get_custom_field() (MantisBT)",
			Pattern:     `\bgpc_get_custom_field\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "MantisBT custom field input",
			NodeTypes:   []string{"function_call_expression"},
		},
		// gpc_get_file('key') — reads a file upload from the request.
		{
			Name:        "gpc_get_file() (MantisBT)",
			Pattern:     `\bgpc_get_file\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "MantisBT file input getter",
			NodeTypes:   []string{"function_call_expression"},
		},
		// gpc_get('key') — generic getter for any GPC parameter.
		{
			Name:         "gpc_get() (MantisBT)",
			Pattern:      `\bgpc_get\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "MantisBT generic input getter",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bgpc_get\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
