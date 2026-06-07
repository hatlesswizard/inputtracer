package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// codeigniterDefinitions returns definitions for CodeIgniter 4 and CodeIgniter 3
// request input methods. CI4 uses an IncomingRequest object with typed getters;
// CI3 uses the Input library accessed via $this->input->method().
func codeigniterDefinitions() []common.Definition {
	return []common.Definition{
		// ══════════════════════════════════════════════════════════════════════
		// CodeIgniter 4 — IncomingRequest methods
		// ══════════════════════════════════════════════════════════════════════

		{
			Name:         "->getVar()",
			Pattern:      `->\s*getVar\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "CI4 request variable",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getVar\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getGet()",
			Pattern:      `->\s*getGet\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "CI4 GET parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getGet\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getPost()",
			Pattern:      `->\s*getPost\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "CI4 POST parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getPost\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "->getPostGet()",
			Pattern:     `->\s*getPostGet\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelHTTPGet, common.LabelUserInput},
			Description: "CI4 POST then GET",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getGetPost()",
			Pattern:     `->\s*getGetPost\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "CI4 GET then POST",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getJSON()",
			Pattern:     `->\s*getJSON\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "CI4 JSON body",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getJsonVar()",
			Pattern:     `->\s*getJsonVar\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "CI4 JSON body variable",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getRawInput()",
			Pattern:     `->\s*getRawInput\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "CI4 raw input",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getRawInputVar()",
			Pattern:     `->\s*getRawInputVar\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "CI4 raw input variable",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getUserAgent()",
			Pattern:     `->\s*getUserAgent\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "CI4 user agent",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:         "->getFile()",
			Pattern:      `->\s*getFile\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description:  "CI4 uploaded file",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getFile\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "->getSegment()",
			Pattern:     `->\s*getSegment\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "CI4 URI segment",
			NodeTypes:   []string{"member_call_expression"},
		},

		// ══════════════════════════════════════════════════════════════════════
		// CodeIgniter 3 — Input library methods ($this->input->method())
		// ══════════════════════════════════════════════════════════════════════

		{
			Name:         "->server()",
			Pattern:      `->\s*server\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "CI3 server variable",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*server\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "->input_stream()",
			Pattern:     `->\s*input_stream\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "CI3 input stream",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->post_get()",
			Pattern:     `->\s*post_get\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelHTTPGet, common.LabelUserInput},
			Description: "CI3 POST then GET",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->get_post()",
			Pattern:     `->\s*get_post\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "CI3 GET then POST",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->ip_address()",
			Pattern:     `->\s*ip_address\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "CI3 client IP",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->user_agent()",
			Pattern:     `->\s*user_agent\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "CI3 user agent",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:         "->get_request_header()",
			Pattern:      `->\s*get_request_header\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "CI3 request header",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*get_request_header\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}
