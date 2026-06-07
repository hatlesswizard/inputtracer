package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// matomoDefinitions returns definitions for Matomo analytics platform's input
// accessors. Matomo uses a static method Common::getRequestVar() as its
// canonical way to read request parameters (GET/POST) with type validation.
func matomoDefinitions() []common.Definition {
	return []common.Definition{
		// Common::getRequestVar('key', default, type) — static accessor for request params.
		{
			Name:         "Common::getRequestVar() (Matomo)",
			Pattern:      `\bCommon\s*::\s*getRequestVar\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Matomo Common::getRequestVar()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `getRequestVar\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
