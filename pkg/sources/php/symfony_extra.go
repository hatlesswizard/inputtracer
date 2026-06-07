package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// symfonyExtraDefinitions returns definitions for Symfony HttpFoundation request
// methods not covered by the auto-generated symfony.go. This includes JSON
// payload access, raw URI/path getters, authentication helpers, and host/scheme
// accessors available on Symfony\Component\HttpFoundation\Request.
func symfonyExtraDefinitions() []common.Definition {
	return []common.Definition{
		// ── JSON / raw body access ────────────────────────────────────────────
		{
			Name:        "->getPayload()",
			Pattern:     `->\s*getPayload\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Symfony JSON request payload",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->toArray()",
			Pattern:     `->\s*toArray\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Symfony request as array",
			NodeTypes:   []string{"member_call_expression"},
		},

		// ── HTTP Basic auth accessors ─────────────────────────────────────────
		{
			Name:        "->getUser()",
			Pattern:     `->\s*getUser\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "HTTP Basic auth username",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getPassword()",
			Pattern:     `->\s*getPassword\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "HTTP Basic auth password",
			NodeTypes:   []string{"member_call_expression"},
		},

		// ── Query string / URI accessors ──────────────────────────────────────
		{
			Name:        "->getQueryString()",
			Pattern:     `->\s*getQueryString\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "Symfony raw query string",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getRequestUri()",
			Pattern:     `->\s*getRequestUri\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Symfony full request URI",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getPathInfo()",
			Pattern:     `->\s*getPathInfo\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Symfony URL path info",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getUri()",
			Pattern:     `->\s*getUri\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Symfony full URI",
			NodeTypes:   []string{"member_call_expression"},
		},

		// ── HTTP method ───────────────────────────────────────────────────────
		{
			Name:        "->getMethod()",
			Pattern:     `->\s*getMethod\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Symfony HTTP method",
			NodeTypes:   []string{"member_call_expression"},
		},

		// ── Host / scheme accessors ───────────────────────────────────────────
		{
			Name:        "->getHost()",
			Pattern:     `->\s*getHost\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Symfony request host",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getSchemeAndHttpHost()",
			Pattern:     `->\s*getSchemeAndHttpHost\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Symfony scheme+host",
			NodeTypes:   []string{"member_call_expression"},
		},

		// ParameterBag typed getters (forward-detection Definitions for methods
		// that were only registered as FrameworkPatterns in symfony.go)
		{
			Name:         "->filter()",
			Pattern:      `->\s*filter\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Symfony ParameterBag filter accessor",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*filter\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getAlpha()",
			Pattern:      `->\s*getAlpha\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Symfony ParameterBag alpha-only accessor",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getAlpha\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getBoolean()",
			Pattern:      `->\s*getBoolean\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Symfony ParameterBag boolean accessor",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getBoolean\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getDigits()",
			Pattern:      `->\s*getDigits\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Symfony ParameterBag digits-only accessor",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getDigits\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getEnum()",
			Pattern:      `->\s*getEnum\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Symfony ParameterBag enum accessor",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getEnum\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getString()",
			Pattern:      `->\s*getString\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Symfony ParameterBag string accessor",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getString\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getInt()",
			Pattern:      `->\s*getInt\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Symfony ParameterBag integer accessor",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getInt\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
