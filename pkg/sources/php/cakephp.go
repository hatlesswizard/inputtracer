package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// cakephpDefinitions returns definitions for CakePHP request input methods.
// CakePHP's ServerRequest class provides getData() as the primary body accessor,
// getEnv() for environment/server variables, getUploadedFile() for file uploads,
// and getAttribute() for route parameters injected by the router middleware.
func cakephpDefinitions() []common.Definition {
	return []common.Definition{
		// getData() is CakePHP's PRIMARY body data accessor. It was previously
		// blocked by ExcludeMethodPattern in the generic matcher because getData()
		// is a common method name across many classes. Adding it as an explicit
		// Definition makes it fire regardless of the exclude pattern — the
		// Definition path takes precedence over ExcludeMethodPattern exclusions.
		{
			Name:         "->getData()",
			Pattern:      `->\s*getData\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description:  "CakePHP body data accessor",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getData\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getEnv()",
			Pattern:      `->\s*getEnv\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "CakePHP environment variable",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getEnv\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getUploadedFile()",
			Pattern:      `->\s*getUploadedFile\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description:  "CakePHP single uploaded file",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getUploadedFile\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getAttribute()",
			Pattern:      `->\s*getAttribute\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "CakePHP request attribute (route params)",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getAttribute\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
