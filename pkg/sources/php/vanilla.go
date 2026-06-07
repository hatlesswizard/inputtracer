package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// vanillaDefinitions returns definitions for Vanilla Forums' request object
// methods. Vanilla Forums uses instance methods on a request/input object to
// read GET, POST, and merged parameters.
func vanillaDefinitions() []common.Definition {
	return []common.Definition{
		// ->getValue('key') — reads a named value from the request.
		{
			Name:         "->getValue() (Vanilla)",
			Pattern:      `->getValue\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Vanilla getValue from request",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->getValue\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->getValueFrom('source', 'key') — reads a value from a specific source.
		{
			Name:         "->getValueFrom() (Vanilla)",
			Pattern:      `->getValueFrom\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Vanilla getValueFrom source",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->getValueFrom\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->getFormValue('key') — reads a form (POST) value from the request.
		{
			Name:         "->getFormValue() (Vanilla)",
			Pattern:      `->getFormValue\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Vanilla form value",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->getFormValue\s*\(\s*['"]([^'"]+)['"]`,
		},
		// ->merged() — returns merged GET+POST parameters.
		{
			Name:        "->merged() (Vanilla)",
			Pattern:     `->merged\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Vanilla merged GET+POST",
			NodeTypes:   []string{"member_call_expression"},
		},
	}
}
