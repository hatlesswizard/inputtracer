package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// joomlaInputDefinitions returns definitions for Joomla's JInput class typed getter methods
// and legacy JRequest static methods. JInput wraps PHP superglobals with type-filtered access.
func joomlaInputDefinitions() []common.Definition {
	return []common.Definition{
		// Typed getter methods on JInput / Joomla\Input\Input
		{
			Name:         "->getString()",
			Pattern:      `->\s*getString\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla JInput::getString() — retrieves string input parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getString\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getInt()",
			Pattern:      `->\s*getInt\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla JInput::getInt() — retrieves integer-cast input parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getInt\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getUInt()",
			Pattern:      `->\s*getUInt\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla JInput::getUInt() — retrieves unsigned integer input parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getUInt\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getFloat()",
			Pattern:      `->\s*getFloat\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla JInput::getFloat() — retrieves float-cast input parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getFloat\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getBool()",
			Pattern:      `->\s*getBool\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla JInput::getBool() — retrieves boolean-cast input parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getBool\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getCmd()",
			Pattern:      `->\s*getCmd\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla JInput::getCmd() — retrieves command-safe (alphanumeric+dash+dot) input",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getCmd\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getWord()",
			Pattern:      `->\s*getWord\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla JInput::getWord() — retrieves word (letters only) input parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getWord\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getAlnum()",
			Pattern:      `->\s*getAlnum\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla JInput::getAlnum() — retrieves alphanumeric input parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getAlnum\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getBase64()",
			Pattern:      `->\s*getBase64\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla JInput::getBase64() — retrieves base64-encoded input parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getBase64\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getHtml()",
			Pattern:      `->\s*getHtml\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla JInput::getHtml() — retrieves HTML-allowed input parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getHtml\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getRaw()",
			Pattern:      `->\s*getRaw\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla JInput::getRaw() — retrieves raw unfiltered input (no sanitization applied)",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getRaw\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "->getArray()",
			Pattern:     `->\s*getArray\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Joomla JInput::getArray() — retrieves multiple input parameters as an array",
			NodeTypes:   []string{"member_call_expression"},
		},

		// Legacy JRequest static methods (deprecated since Joomla 3.x, still widely used in old code)
		{
			Name:         "JRequest::getVar()",
			Pattern:      `\bJRequest\s*::\s*getVar\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla legacy JRequest::getVar() — reads GET/POST/request data (deprecated)",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `\bJRequest\s*::\s*getVar\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "JRequest::get()",
			Pattern:      `\bJRequest\s*::\s*get\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Joomla legacy JRequest::get() — reads filtered request data (deprecated)",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `\bJRequest\s*::\s*get\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
