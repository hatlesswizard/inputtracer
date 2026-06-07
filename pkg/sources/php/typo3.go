package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// typo3Definitions returns definitions for TYPO3 CMS's GeneralUtility input
// methods. TYPO3 uses static methods _GP(), _GET(), and _POST() on the
// GeneralUtility class as its canonical way to read request parameters.
func typo3Definitions() []common.Definition {
	return []common.Definition{
		// GeneralUtility::_GP('key') — reads from merged GET+POST (POST takes precedence).
		{
			Name:         "GeneralUtility::_GP() (TYPO3)",
			Pattern:      `\bGeneralUtility\s*::\s*_GP\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description:  "TYPO3 GeneralUtility::_GP()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `_GP\s*\(\s*['"]([^'"]+)['"]`,
		},
		// GeneralUtility::_GET('key') — reads from GET parameters only.
		{
			Name:         "GeneralUtility::_GET() (TYPO3)",
			Pattern:      `\bGeneralUtility\s*::\s*_GET\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "TYPO3 GeneralUtility::_GET()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `_GET\s*\(\s*['"]([^'"]+)['"]`,
		},
		// GeneralUtility::_POST('key') — reads from POST parameters only.
		{
			Name:         "GeneralUtility::_POST() (TYPO3)",
			Pattern:      `\bGeneralUtility\s*::\s*_POST\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "TYPO3 GeneralUtility::_POST()",
			NodeTypes:    []string{"scoped_call_expression"},
			KeyExtractor: `_POST\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
