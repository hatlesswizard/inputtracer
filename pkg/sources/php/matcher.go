package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// Matcher matches PHP user input sources
type Matcher struct {
	*common.BaseMatcher
}

// NewMatcher creates a new PHP source matcher
func NewMatcher() *Matcher {
	defs := []common.Definition{
		// Superglobals - HTTP
		{
			Name:         "$_GET",
			Pattern:      `\$_GET\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "HTTP GET parameters",
			NodeTypes:    []string{"subscript_expression", "variable_name"},
			KeyExtractor: `\$_GET\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "$_POST",
			Pattern:      `\$_POST\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "HTTP POST parameters",
			NodeTypes:    []string{"subscript_expression", "variable_name"},
			KeyExtractor: `\$_POST\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "$_REQUEST",
			Pattern:      `\$_REQUEST\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Combined GET/POST/COOKIE parameters",
			NodeTypes:    []string{"subscript_expression", "variable_name"},
			KeyExtractor: `\$_REQUEST\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "$_COOKIE",
			Pattern:      `\$_COOKIE\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:  "HTTP cookies",
			NodeTypes:    []string{"subscript_expression", "variable_name"},
			KeyExtractor: `\$_COOKIE\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "$_SERVER",
			Pattern:      `\$_SERVER\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "Server and request information",
			NodeTypes:    []string{"subscript_expression", "variable_name"},
			KeyExtractor: `\$_SERVER\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "$_FILES",
			Pattern:      `\$_FILES\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description:  "Uploaded files",
			NodeTypes:    []string{"subscript_expression", "variable_name"},
			KeyExtractor: `\$_FILES\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "$_ENV",
			Pattern:      `\$_ENV\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "Environment variables",
			NodeTypes:    []string{"subscript_expression", "variable_name"},
			KeyExtractor: `\$_ENV\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},

		// Raw input
		{
			Name:        "php://input",
			Pattern:     `file_get_contents\s*\(\s*['"]php://input['"]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Raw POST body",
			NodeTypes:   []string{"function_call_expression"},
		},

		// File functions
		{
			Name:        "file_get_contents",
			Pattern:     `file_get_contents\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "File contents reader",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "fopen",
			Pattern:     `fopen\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "File handle opener",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "fgets",
			Pattern:     `fgets\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read line from file",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "fread",
			Pattern:     `fread\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Binary file read",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "file",
			Pattern:     `\bfile\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read file into array",
			NodeTypes:   []string{"function_call_expression"},
		},

		// getenv
		{
			Name:        "getenv",
			Pattern:     `getenv\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelEnvironment},
			Description: "Get environment variable",
			NodeTypes:   []string{"function_call_expression"},
		},

		// CLI
		{
			Name:        "$argv",
			Pattern:     `\$argv`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"variable_name"},
		},

		// Stream input
		{
			Name:        "php://stdin",
			Pattern:     `fopen\s*\(\s*['"]php://stdin['"]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Standard input stream",
			NodeTypes:   []string{"function_call_expression"},
		},

		// =====================================================
		// UNIVERSAL PATTERNS - Work across all PHP applications
		// =====================================================

		// Bare superglobals (when passed as arguments or used in foreach)
		// These detect $_GET, $_POST etc. when NOT followed by bracket
		{
			Name:        "$_GET (bare)",
			Pattern:     `\$_GET\s*[,)\]]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "HTTP GET array passed as argument or in expression",
			NodeTypes:   []string{"variable_name", "argument"},
		},
		{
			Name:        "$_POST (bare)",
			Pattern:     `\$_POST\s*[,)\]]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "HTTP POST array passed as argument or in expression",
			NodeTypes:   []string{"variable_name", "argument"},
		},
		{
			Name:        "$_REQUEST (bare)",
			Pattern:     `\$_REQUEST\s*[,)\]]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "HTTP REQUEST array passed as argument",
			NodeTypes:   []string{"variable_name", "argument"},
		},
		{
			Name:        "$_COOKIE (bare)",
			Pattern:     `\$_COOKIE\s*[,)\]]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description: "HTTP COOKIE array passed as argument",
			NodeTypes:   []string{"variable_name", "argument"},
		},
		{
			Name:        "$_FILES (bare)",
			Pattern:     `\$_FILES\s*[,)\]]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "FILES array passed as argument",
			NodeTypes:   []string{"variable_name", "argument"},
		},
		{
			Name:        "$_SERVER (bare)",
			Pattern:     `\$_SERVER\s*[,)\]]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "SERVER array passed as argument",
			NodeTypes:   []string{"variable_name", "argument"},
		},

		// Generic input getter methods (universal across frameworks)
		{
			Name:         "->get_input()",
			Pattern:      `->\s*get_input\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Generic input getter method",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*get_input\s*\(\s*['"]([^'"]+)`,
		},
		{
			Name:         "->get_var()",
			Pattern:      `->\s*get_var\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Generic variable getter method (phpBB style)",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*get_var\s*\(\s*['"]([^'"]+)`,
		},
		{
			Name:         "->variable()",
			Pattern:      `->\s*variable\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Generic variable getter (phpBB style)",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*variable\s*\(\s*['"]([^'"]+)`,
		},

		// PSR-7 HTTP Message Interface (universal standard)
		{
			Name:        "->getQueryParams()",
			Pattern:     `->\s*getQueryParams\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "PSR-7 query parameters",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getParsedBody()",
			Pattern:     `->\s*getParsedBody\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "PSR-7 parsed request body",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getCookieParams()",
			Pattern:     `->\s*getCookieParams\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description: "PSR-7 cookie parameters",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getUploadedFiles()",
			Pattern:     `->\s*getUploadedFiles\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "PSR-7 uploaded files",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getServerParams()",
			Pattern:     `->\s*getServerParams\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "PSR-7 server parameters",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getHeaders()",
			Pattern:     `->\s*getHeaders\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "PSR-7 request headers",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->getHeader()",
			Pattern:     `->\s*getHeader\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "PSR-7 specific header",
			NodeTypes:   []string{"member_call_expression"},
		},

		// Laravel/Symfony style input methods (universal patterns)
		{
			Name:         "->input()",
			Pattern:      `->\s*input\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Laravel-style input method",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*input\s*\(\s*['"]([^'"]+)`,
		},
		{
			Name:        "->all()",
			Pattern:     `->\s*all\s*\(\s*\)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Get all input data",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:         "->query()",
			Pattern:      `->\s*query\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Query string getter (Symfony style)",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*query\s*\(\s*['"]([^'"]+)`,
		},
		{
			Name:         "->post()",
			Pattern:      `->\s*post\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "POST data getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*post\s*\(\s*['"]([^'"]+)`,
		},
		{
			Name:         "->cookie()",
			Pattern:      `->\s*cookie\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:  "Cookie getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*cookie\s*\(\s*['"]([^'"]+)`,
		},
		{
			Name:         "->header()",
			Pattern:      `->\s*header\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "Header getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*header\s*\(\s*['"]([^'"]+)`,
		},
		{
			Name:         "->file()",
			Pattern:      `->\s*file\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description:  "File upload getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*file\s*\(\s*['"]([^'"]+)`,
		},

		// Deserialization functions (receives potentially tainted data)
		{
			Name:        "unserialize()",
			Pattern:     `\bunserialize\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "PHP unserialize function - potential object injection",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "*unserialize() (custom)",
			Pattern:     `\b\w*unserialize\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Custom unserialize wrapper (my_unserialize, safe_unserialize, etc.)",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "json_decode()",
			Pattern:     `\bjson_decode\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "JSON decode - parses external data",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "simplexml_load_string()",
			Pattern:     `\bsimplexml_load_string\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "XML parsing - potential XXE",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "yaml_parse()",
			Pattern:     `\byaml_parse\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "YAML parsing - potential code execution",
			NodeTypes:   []string{"function_call_expression"},
		},

		// cURL - network request responses
		{
			Name:        "curl_exec()",
			Pattern:     `\bcurl_exec\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "cURL execute - returns external data",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "curl_multi_getcontent()",
			Pattern:     `\bcurl_multi_getcontent\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "cURL multi get content - returns external data",
			NodeTypes:   []string{"function_call_expression"},
		},

		// CodeIgniter style input
		{
			Name:         "->get()",
			Pattern:      `->\s*get\s*\(\s*['"]`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "CodeIgniter/Symfony style GET parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*get\s*\(\s*['"]([^'"]+)`,
		},

		// Object property array access (universal - any object's input/data array)
		{
			Name:         "->input[]",
			Pattern:      `->\s*input\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Object input array access",
			NodeTypes:    []string{"subscript_expression", "member_access_expression"},
			KeyExtractor: `->\s*input\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "->cookies[]",
			Pattern:      `->\s*cookies\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:  "Object cookies array access",
			NodeTypes:    []string{"subscript_expression", "member_access_expression"},
			KeyExtractor: `->\s*cookies\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "->data[]",
			Pattern:      `->\s*data\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Object data array access",
			NodeTypes:    []string{"subscript_expression", "member_access_expression"},
			KeyExtractor: `->\s*data\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "->params[]",
			Pattern:      `->\s*params\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Object params array access",
			NodeTypes:    []string{"subscript_expression", "member_access_expression"},
			KeyExtractor: `->\s*params\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "->request[]",
			Pattern:      `->\s*request\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Object request array access",
			NodeTypes:    []string{"subscript_expression", "member_access_expression"},
			KeyExtractor: `->\s*request\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
	}

	return &Matcher{
		BaseMatcher: common.NewBaseMatcher("php", defs),
	}
}
