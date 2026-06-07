package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// freshRSSDefinitions returns definitions for FreshRSS feed reader's request
// parameter methods. FreshRSS uses static methods on Minz_Request to read
// typed HTTP parameters.
func freshRSSDefinitions() []common.Definition {
	return []common.Definition{
		// Minz_Request::paramString('key') — reads a string parameter from the request.
		{
			Name:         "Minz_Request::paramString() (FreshRSS)",
			Pattern:      `\bMinz_Request\s*::\s*paramString\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "FreshRSS paramString()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `paramString\s*\(\s*['"]([^'"]+)['"]`,
		},
		// Minz_Request::paramInt('key') — reads an integer parameter from the request.
		{
			Name:         "Minz_Request::paramInt() (FreshRSS)",
			Pattern:      `\bMinz_Request\s*::\s*paramInt\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "FreshRSS paramInt()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `paramInt\s*\(\s*['"]([^'"]+)['"]`,
		},
		// Minz_Request::paramArray('key') — reads an array parameter from the request.
		{
			Name:         "Minz_Request::paramArray() (FreshRSS)",
			Pattern:      `\bMinz_Request\s*::\s*paramArray\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "FreshRSS paramArray()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `paramArray\s*\(\s*['"]([^'"]+)['"]`,
		},
		// Minz_Request::paramBoolean('key') — reads a boolean parameter from the request.
		{
			Name:         "Minz_Request::paramBoolean() (FreshRSS)",
			Pattern:      `\bMinz_Request\s*::\s*paramBoolean\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "FreshRSS paramBoolean()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `paramBoolean\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
