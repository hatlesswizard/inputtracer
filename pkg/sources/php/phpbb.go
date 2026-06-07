package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// phpbbDefinitions returns definitions for phpBB forum's custom request object
// methods. phpBB wraps PHP superglobals in a request class that provides typed
// access to GET/POST/COOKIE/SERVER variables via instance methods.
func phpbbDefinitions() []common.Definition {
	return []common.Definition{
		// ->raw_variable('key') — reads raw (unvalidated) request variable.
		{
			Name:         "->raw_variable() (phpBB)",
			Pattern:      `->raw_variable\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "phpBB raw variable from request",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->raw_variable\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->server('key') — reads server variable (equivalent to $_SERVER).
		{
			Name:        "->server() (phpBB)",
			Pattern:     `->server\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "phpBB server variable",
			NodeTypes:   []string{"member_call_expression"},
		},
		// ->is_set_post('key') — checks if a POST variable is set and returns truthy.
		{
			Name:        "->is_set_post() (phpBB)",
			Pattern:     `->is_set_post\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "phpBB POST check",
			NodeTypes:   []string{"member_call_expression"},
		},
		// ->get_super_global('key') — reads any superglobal by name.
		{
			Name:        "->get_super_global() (phpBB)",
			Pattern:     `->get_super_global\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "phpBB superglobal accessor",
			NodeTypes:   []string{"member_call_expression"},
		},
		// ->variable_names() — returns all variable names from a given superglobal.
		{
			Name:        "->variable_names() (phpBB)",
			Pattern:     `->variable_names\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "phpBB request variable names",
			NodeTypes:   []string{"member_call_expression"},
		},
	}
}
