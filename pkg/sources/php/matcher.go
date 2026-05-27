package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// Matcher matches PHP user input sources
type Matcher struct {
	*common.BaseMatcher
}

// NewMatcher creates a new PHP source matcher combining all definition groups.
func NewMatcher() *Matcher {
	defs := superglobalDefinitions()
	defs = append(defs, streamDefinitions()...)
	defs = append(defs, headerDefinitions()...)
	defs = append(defs, frameworkDefinitions()...)
	defs = append(defs, restAPIDefinitions()...)
	defs = append(defs, wordpressHookDefinitions()...)
	defs = append(defs, wordpressSanitizerDefinitions()...)
	defs = append(defs, wordpressMetaDefinitions()...)
	defs = append(defs, wordpressRESTMethodDefinitions()...)
	defs = append(defs, drupalFormAPIDefinitions()...)
	defs = append(defs, joomlaInputDefinitions()...)
	defs = append(defs, opencartRequestDefinitions()...)
	// Builtin-discovery groups (patterns driven by Haiku gap analysis of real PHP apps).
	defs = append(defs, filterInputDefinitions()...)
	defs = append(defs, phpBuiltinWrapperDefinitions()...)
	defs = append(defs, legacyInputDefinitions()...)
	defs = append(defs, phpExistenceCheckDefinitions()...)
	defs = append(defs, phpStreamInputDefinitions()...)
	defs = append(defs, phpCliInputDefinitions()...)
	defs = append(defs, phpCastInputDefinitions()...)
	defs = append(defs, phpArrayOperationDefinitions()...)
	defs = append(defs, phpDynamicInputDefinitions()...)
	defs = append(defs, phpSessionInputDefinitions()...)
	defs = append(defs, phpEnvironmentInputDefinitions()...)
	defs = append(defs, phpHTTPAuthDefinitions()...)
	defs = append(defs, phpFileUploadDefinitions()...)
	defs = append(defs, phpRequestFactoryDefinitions()...)
	defs = append(defs, phpPSR7SingleParamDefinitions()...)
	defs = append(defs, phpSymfonyParameterBagDefinitions()...)
	return &Matcher{
		BaseMatcher: common.NewBaseMatcher("php", defs),
	}
}

// superglobalDefinitions returns definitions for PHP superglobals ($_GET, $_POST, etc.)
// including both bracket-access forms and bare array forms, plus wrappers like extract().
func superglobalDefinitions() []common.Definition {
	return []common.Definition{
		// --- Subscript access ---
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
		// $_ENV — server configuration, NOT request data
		{
			Name:         "$_ENV",
			Pattern:      `\$_ENV\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "Environment variables (server config, NOT request data)",
			NodeTypes:    []string{"subscript_expression", "variable_name"},
			KeyExtractor: `\$_ENV\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		// $_SESSION — stored server-side, NOT sent in the request
		{
			Name:         "$_SESSION",
			Pattern:      `\$_SESSION\s*\[`,
			Language:     "php",
			Labels:       []common.InputLabel{},
			Description:  "Session data (stored server-side, NOT sent in request)",
			NodeTypes:    []string{"subscript_expression", "variable_name"},
			KeyExtractor: `\$_SESSION\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},

		// --- Bare array forms (passed as argument or used in foreach) ---
		{
			Name:               "$_GET (bare)",
			Pattern:            `^\$_GET$`,
			Language:           "php",
			Labels:             []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:        "HTTP GET array passed as argument or in expression",
			NodeTypes:          []string{"variable_name"},
			ExcludeParentTypes: []string{"subscript_expression"},
		},
		{
			Name:               "$_POST (bare)",
			Pattern:            `^\$_POST$`,
			Language:           "php",
			Labels:             []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:        "HTTP POST array passed as argument or in expression",
			NodeTypes:          []string{"variable_name"},
			ExcludeParentTypes: []string{"subscript_expression"},
		},
		{
			Name:               "$_REQUEST (bare)",
			Pattern:            `^\$_REQUEST$`,
			Language:           "php",
			Labels:             []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description:        "HTTP REQUEST array passed as argument",
			NodeTypes:          []string{"variable_name"},
			ExcludeParentTypes: []string{"subscript_expression"},
		},
		{
			Name:               "$_COOKIE (bare)",
			Pattern:            `^\$_COOKIE$`,
			Language:           "php",
			Labels:             []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:        "HTTP COOKIE array passed as argument",
			NodeTypes:          []string{"variable_name"},
			ExcludeParentTypes: []string{"subscript_expression"},
		},
		{
			Name:               "$_FILES (bare)",
			Pattern:            `^\$_FILES$`,
			Language:           "php",
			Labels:             []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description:        "FILES array passed as argument",
			NodeTypes:          []string{"variable_name"},
			ExcludeParentTypes: []string{"subscript_expression"},
		},
		{
			Name:               "$_SERVER (bare)",
			Pattern:            `^\$_SERVER$`,
			Language:           "php",
			Labels:             []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:        "SERVER array passed as argument",
			NodeTypes:          []string{"variable_name"},
			ExcludeParentTypes: []string{"subscript_expression"},
		},

		// --- Superglobal wrapper functions ---
		{
			Name:        "extract($_POST)",
			Pattern:     `\bextract\s*\(\s*\$_POST`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "extract() from $_POST creates tainted variables",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "extract($_GET)",
			Pattern:     `\bextract\s*\(\s*\$_GET`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "extract() from $_GET creates tainted variables",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "extract($_REQUEST)",
			Pattern:     `\bextract\s*\(\s*\$_REQUEST`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "extract() from $_REQUEST creates tainted variables",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "wp_parse_args() with superglobal",
			Pattern:     `\bwp_parse_args\s*\(\s*\$_(GET|POST|REQUEST)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "wp_parse_args() with superglobal input",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "stripslashes_deep() with superglobal",
			Pattern:     `\bstripslashes_deep\s*\(\s*\$_(GET|POST|REQUEST)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "stripslashes_deep() wrapping superglobal",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "wp_unslash() with superglobal",
			Pattern:     `\bwp_unslash\s*\(\s*\$_(GET|POST|REQUEST)`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "wp_unslash() wrapping superglobal",
			NodeTypes:   []string{"function_call_expression"},
		},
	}
}

// streamDefinitions returns definitions for raw HTTP body streams, file operations,
// environment variables, and CLI input.
func streamDefinitions() []common.Definition {
	return []common.Definition{
		// Raw HTTP body
		{
			Name:        "php://input (file_get_contents)",
			Pattern:     `file_get_contents\s*\(\s*['"]php://input['"]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Raw POST body via file_get_contents",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "php://input (fopen)",
			Pattern:     `fopen\s*\(\s*['"]php://input['"]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Raw POST body via fopen stream",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "php://stdin",
			Pattern:     `fopen\s*\(\s*['"]php://stdin['"]`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Standard input stream",
			NodeTypes:   []string{"function_call_expression"},
		},

		// File operations (NOT user input from HTTP request)
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

		// Environment / CLI
		{
			Name:        "getenv",
			Pattern:     `getenv\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelEnvironment},
			Description: "Get environment variable",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "$argv",
			Pattern:     `\$argv`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"variable_name"},
		},
	}
}

// headerDefinitions returns definitions for HTTP header retrieval functions.
func headerDefinitions() []common.Definition {
	return []common.Definition{
		{
			Name:        "getallheaders()",
			Pattern:     `\bgetallheaders\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Get all HTTP request headers (alias of apache_request_headers)",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "apache_request_headers()",
			Pattern:     `\bapache_request_headers\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Get all HTTP request headers (Apache SAPI)",
			NodeTypes:   []string{"function_call_expression"},
		},
	}
}

// frameworkDefinitions returns definitions for framework-agnostic input patterns:
// PSR-7, Laravel/Symfony-style methods, and generic object input accessors.
func frameworkDefinitions() []common.Definition {
	return []common.Definition{
		// Generic input getters
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
			Name:           "->get_var() (non-wpdb)",
			Pattern:        `->\s*get_var\s*\(`,
			Language:       "php",
			Labels:         []common.InputLabel{common.LabelUserInput},
			Description:    "Generic variable getter method (excludes $wpdb which is database)",
			NodeTypes:      []string{"member_call_expression"},
			KeyExtractor:   `->\s*get_var\s*\(\s*['"]([^'"]+)`,
			ExcludePattern: `\$wpdb\s*->\s*get_var`,
		},
		{
			Name:         "->variable()",
			Pattern:      `->\s*variable\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Generic variable getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*variable\s*\(\s*['"]([^'"]+)`,
		},

		// PSR-7 HTTP Message Interface
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

		// Laravel/Symfony-style input methods
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
			Pattern:     `->\s*all\s*\(`,
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

		// Generic GET method
		{
			Name:         "->get()",
			Pattern:      `->\s*get\s*\(\s*['"]`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Generic/Symfony style GET parameter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*get\s*\(\s*['"]([^'"]+)`,
		},

		// Object property array access patterns
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

		// HTTP body content (Symfony/Drupal JSON API)
		{
			Name:        "->getContent()",
			Pattern:     `->\s*getContent\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Symfony HttpFoundation Request::getContent() — raw request body (used in Drupal JSON:API)",
			NodeTypes:   []string{"member_call_expression"},
		},

		// Magento/CodeIgniter-style single parameter getter
		{
			Name:         "->getParam()",
			Pattern:      `->\s*getParam\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "Magento/CodeIgniter-style single request parameter getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getParam\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "->getParams()",
			Pattern:     `->\s*getParams\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Magento/CodeIgniter-style all request parameters getter",
			NodeTypes:   []string{"member_call_expression"},
		},

		// Magento/CodeIgniter-style typed getters (distinct from ->post(), ->query(), etc.)
		{
			Name:         "->getPost()",
			Pattern:      `->\s*getPost\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Magento/CodeIgniter-style POST parameter getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getPost\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getQuery()",
			Pattern:      `->\s*getQuery\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Magento/CodeIgniter-style query string parameter getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getQuery\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getCookie()",
			Pattern:      `->\s*getCookie\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:  "Magento/Slim-style cookie getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getCookie\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getServer()",
			Pattern:      `->\s*getServer\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "Magento-style server variable getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getServer\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "->getFiles()",
			Pattern:      `->\s*getFiles\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description:  "Magento-style file upload getter",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*getFiles\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}

// restAPIDefinitions returns definitions for REST API input (WP_REST_Request),
// external network sources, and database sources.
func restAPIDefinitions() []common.Definition {
	return []common.Definition{
		// WP_REST_Request ArrayAccess — $request['param'] equivalent to get_param()
		{
			Name:         "$request['...'] (WP_REST_Request ArrayAccess)",
			Pattern:      `\$(?:request|req|api_request|rest_request)\s*\[\s*['"]`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput, common.LabelHTTPGet, common.LabelHTTPPost},
			Description:  "WP_REST_Request ArrayAccess - equivalent to get_param()",
			NodeTypes:    []string{"subscript_expression"},
			KeyExtractor: `\[\s*['"]([^'"]+)['"]`,
		},

		// External network (NOT the current HTTP request)
		{
			Name:        "curl_exec()",
			Pattern:     `\bcurl_exec\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "cURL execute - returns external network data (NOT HTTP request input)",
			NodeTypes:   []string{"function_call_expression"},
		},
		{
			Name:        "curl_multi_getcontent()",
			Pattern:     `\bcurl_multi_getcontent\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "cURL multi get content - returns external network data",
			NodeTypes:   []string{"function_call_expression"},
		},

		// Database sources (NOT user input from the HTTP request)
		{
			Name:        "$wpdb->get_var()",
			Pattern:     `\$wpdb\s*->\s*get_var\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelDatabase},
			Description: "WordPress database query - returns single variable (NOT user input)",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "$wpdb->get_results()",
			Pattern:     `\$wpdb\s*->\s*get_results\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelDatabase},
			Description: "WordPress database query - returns result set (NOT user input)",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "$wpdb->get_row()",
			Pattern:     `\$wpdb\s*->\s*get_row\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelDatabase},
			Description: "WordPress database query - returns single row (NOT user input)",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "get_option()",
			Pattern:     `\bget_option\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelDatabase},
			Description: "WordPress options API - reads from wp_options table (NOT user input)",
			NodeTypes:   []string{"function_call_expression"},
		},
	}
}
