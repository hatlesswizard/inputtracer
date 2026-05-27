package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// drupalFormAPIDefinitions returns definitions for Drupal's Form API input access patterns.
// Drupal uses a FormState object (FormStateInterface) to carry all submitted form values
// rather than accessing $_POST directly.
func drupalFormAPIDefinitions() []common.Definition {
	return []common.Definition{
		{
			Name:         "$form_state->getValue()",
			Pattern:      `\$form_state\s*->\s*getValue\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Drupal Form API — retrieves a single submitted form field value",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `\$form_state\s*->\s*getValue\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "$form_state->getValues()",
			Pattern:     `\$form_state\s*->\s*getValues\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "Drupal Form API — retrieves all submitted form values as an array",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "$form_state->getUserInput()",
			Pattern:     `\$form_state\s*->\s*getUserInput\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "Drupal Form API — retrieves raw unvalidated user input (before processing)",
			NodeTypes:   []string{"member_call_expression"},
		},
	}
}
