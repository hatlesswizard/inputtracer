package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// phalconDefinitions returns definitions for Phalcon PHP framework input methods.
// Phalcon's framework source is written in Zephir (.zep), which tree-sitter cannot
// parse. These patterns target the PHP CALL SITES in user code (e.g.
// $this->request->getPut('name')), not the framework internals.
func phalconDefinitions() []common.Definition {
	return []common.Definition{
		{
			Name:         "->getPut()",
			Pattern:      `->\s*getPut\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Phalcon PUT parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getPut\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getPatch()",
			Pattern:      `->\s*getPatch\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Phalcon PATCH parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getPatch\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getFilteredPost()",
			Pattern:      `->\s*getFilteredPost\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Phalcon filtered POST",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getFilteredPost\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getFilteredQuery()",
			Pattern:      `->\s*getFilteredQuery\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Phalcon filtered query",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getFilteredQuery\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getFilteredPut()",
			Pattern:      `->\s*getFilteredPut\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Phalcon filtered PUT",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getFilteredPut\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "->getJsonRawBody()",
			Pattern:     `->\s*getJsonRawBody\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Phalcon JSON raw body",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getBasicAuth()",
			Pattern:     `->\s*getBasicAuth\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Phalcon HTTP Basic auth",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getDigestAuth()",
			Pattern:     `->\s*getDigestAuth\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Phalcon HTTP Digest auth",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getClientAddress()",
			Pattern:     `->\s*getClientAddress\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Phalcon client IP address",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getRawBody()",
			Pattern:     `->\s*getRawBody\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Phalcon raw HTTP body",
			NodeTypes:   []string{"member_call_expression"},
		},
	}
}
