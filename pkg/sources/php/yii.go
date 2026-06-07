package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// yiiDefinitions returns definitions for Yii2 framework and CraftCMS (which
// extends Yii2's Request class) input methods. These patterns target PHP call
// sites in user code.
func yiiDefinitions() []common.Definition {
	return []common.Definition{
		// ── Yii2 Request methods ──────────────────────────────────────────────
		{
			Name:         "->getBodyParam()",
			Pattern:      `->\s*getBodyParam\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelHTTPBody, common.LabelUserInput},
			Description:  "Yii2 body parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getBodyParam\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "->getBodyParams()",
			Pattern:     `->\s*getBodyParams\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelHTTPBody, common.LabelUserInput},
			Description: "Yii2 all body parameters",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getUserIP()",
			Pattern:     `->\s*getUserIP\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Yii2 user IP address",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getAuthUser()",
			Pattern:     `->\s*getAuthUser\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Yii2 HTTP auth username",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getAuthPassword()",
			Pattern:     `->\s*getAuthPassword\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Yii2 HTTP auth password",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getCookies()",
			Pattern:     `->\s*getCookies\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description: "Yii2 cookie collection",
			NodeTypes:   []string{"member_call_expression"},
		},

		// ── CraftCMS additions (extends Yii2 Request) ────────────────────────
		{
			Name:         "->getRequiredBodyParam()",
			Pattern:      `->\s*getRequiredBodyParam\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description:  "CraftCMS required body parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getRequiredBodyParam\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getValidatedBodyParam()",
			Pattern:      `->\s*getValidatedBodyParam\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description:  "CraftCMS validated body parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getValidatedBodyParam\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "->getSegment()",
			Pattern:     `->\s*getSegment\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "CraftCMS URL segment",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getToken()",
			Pattern:     `->\s*getToken\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "CraftCMS request token",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getFullPath()",
			Pattern:     `->\s*getFullPath\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "CraftCMS request full path",
			NodeTypes:   []string{"member_call_expression"},
		},
	}
}
