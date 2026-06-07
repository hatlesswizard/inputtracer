package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// prestashopDefinitions returns definitions for PrestaShop e-commerce platform's
// input accessors. PrestaShop uses static methods on the Tools class to read
// HTTP parameters.
func prestashopDefinitions() []common.Definition {
	return []common.Definition{
		// Tools::getValue('key') — reads a GET or POST parameter by name.
		{
			Name:         "Tools::getValue() (PrestaShop)",
			Pattern:      `\bTools\s*::\s*getValue\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "PrestaShop Tools::getValue()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `getValue\s*\(\s*['"]([^'"]+)['"]`,
		},
		// Tools::getAllValues() — reads all GET/POST parameters as an array.
		{
			Name:        "Tools::getAllValues() (PrestaShop)",
			Pattern:     `\bTools\s*::\s*getAllValues\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "PrestaShop Tools::getAllValues()",
			NodeTypes:   []string{"scoped_call_expression"},
		},
		// Tools::isSubmit('key') — checks if a form was submitted (POST parameter exists).
		{
			Name:         "Tools::isSubmit() (PrestaShop)",
			Pattern:      `\bTools\s*::\s*isSubmit\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "PrestaShop Tools::isSubmit()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `isSubmit\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
