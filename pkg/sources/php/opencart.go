package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// opencartRequestDefinitions returns definitions for OpenCart's request object
// property array access patterns. OpenCart exposes request data as array properties
// on the $this->request object (Controller::request field).
func opencartRequestDefinitions() []common.Definition {
	return []common.Definition{
		{
			Name:         "->get[...]",
			Pattern:      `->\s*get\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "OpenCart request->get[] — HTTP GET parameters accessed as array property",
			NodeTypes:    []string{"subscript_expression", "member_access_expression"},
			KeyExtractor: `->\s*get\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "->post[...]",
			Pattern:      `->\s*post\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "OpenCart request->post[] — HTTP POST parameters accessed as array property",
			NodeTypes:    []string{"subscript_expression", "member_access_expression"},
			KeyExtractor: `->\s*post\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "->server[...]",
			Pattern:      `->\s*server\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "OpenCart request->server[] — server/request variables accessed as array property",
			NodeTypes:    []string{"subscript_expression", "member_access_expression"},
			KeyExtractor: `->\s*server\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "->files[...]",
			Pattern:      `->\s*files\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description:  "OpenCart request->files[] — uploaded files accessed as array property",
			NodeTypes:    []string{"subscript_expression", "member_access_expression"},
			KeyExtractor: `->\s*files\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
	}
}
