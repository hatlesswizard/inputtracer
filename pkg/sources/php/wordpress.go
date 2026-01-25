// Package php - wordpress.go provides WordPress-specific input patterns
// WordPress primarily uses standard PHP superglobals for input, but has
// some wrapper functions and the REST API for receiving user data
package php

import (
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// WordPress user input sources - these are the ACTUAL entry points
// where user data enters the WordPress application
var wordpressPatterns = []*common.FrameworkPattern{
	// =====================================================
	// WP_REST_Request - WordPress REST API Input
	// This is the modern way to receive input in WordPress
	// =====================================================
	{
		ID:            "wordpress_rest_get_param",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->get_param()",
		Description:   "REST API request parameter (from URL, query, or body)",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^get_param$",
		SourceType:    common.SourceUserInput,
		CarrierClass:  "WP_REST_Request",
		PopulatedFrom: []string{"$_GET", "$_POST", "php://input"},
		Tags:          []string{"rest-api", "modern"},
	},
	{
		ID:            "wordpress_rest_get_params",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->get_params()",
		Description:   "REST API all parameters",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^get_params$",
		SourceType:    common.SourceUserInput,
		CarrierClass:  "WP_REST_Request",
		PopulatedFrom: []string{"$_GET", "$_POST", "php://input"},
		Tags:          []string{"rest-api", "modern"},
	},
	{
		ID:            "wordpress_rest_get_query_params",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->get_query_params()",
		Description:   "REST API query string parameters",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^get_query_params$",
		SourceType:    common.SourceHTTPGet,
		CarrierClass:  "WP_REST_Request",
		PopulatedFrom: []string{"$_GET"},
		Tags:          []string{"rest-api", "modern"},
	},
	{
		ID:            "wordpress_rest_get_body_params",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->get_body_params()",
		Description:   "REST API POST body parameters",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^get_body_params$",
		SourceType:    common.SourceHTTPPost,
		CarrierClass:  "WP_REST_Request",
		PopulatedFrom: []string{"$_POST"},
		Tags:          []string{"rest-api", "modern"},
	},
	{
		ID:            "wordpress_rest_get_json_params",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->get_json_params()",
		Description:   "REST API JSON body parameters",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^get_json_params$",
		SourceType:    common.SourceHTTPJSON,
		CarrierClass:  "WP_REST_Request",
		PopulatedFrom: []string{"php://input"},
		Tags:          []string{"rest-api", "modern", "json"},
	},
	{
		ID:            "wordpress_rest_get_body",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->get_body()",
		Description:   "REST API raw request body",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^get_body$",
		SourceType:    common.SourceHTTPBody,
		CarrierClass:  "WP_REST_Request",
		PopulatedFrom: []string{"php://input"},
		Tags:          []string{"rest-api", "modern"},
	},
	{
		ID:            "wordpress_rest_get_file_params",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->get_file_params()",
		Description:   "REST API file upload parameters",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^get_file_params$",
		SourceType:    common.SourceHTTPFile,
		CarrierClass:  "WP_REST_Request",
		PopulatedFrom: []string{"$_FILES"},
		Tags:          []string{"rest-api", "modern", "upload"},
	},
	{
		ID:            "wordpress_rest_get_header",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->get_header()",
		Description:   "REST API request header",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^get_header$",
		SourceType:    common.SourceHTTPHeader,
		CarrierClass:  "WP_REST_Request",
		PopulatedFrom: []string{"$_SERVER"},
		Tags:          []string{"rest-api", "modern"},
	},
	{
		ID:            "wordpress_rest_get_headers",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->get_headers()",
		Description:   "REST API all request headers",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^get_headers$",
		SourceType:    common.SourceHTTPHeader,
		CarrierClass:  "WP_REST_Request",
		PopulatedFrom: []string{"$_SERVER"},
		Tags:          []string{"rest-api", "modern"},
	},
	{
		ID:            "wordpress_rest_get_url_params",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->get_url_params()",
		Description:   "REST API URL path parameters",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^get_url_params$",
		SourceType:    common.SourceHTTPPath,
		CarrierClass:  "WP_REST_Request",
		Tags:          []string{"rest-api", "modern"},
	},

	// =====================================================
	// AJAX Request Input - admin-ajax.php handling
	// =====================================================
	{
		ID:            "wordpress_ajax_action",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress AJAX $_REQUEST['action']",
		Description:   "AJAX action parameter from request",
		AccessPattern: "superglobal",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_REQUEST"},
		Tags:          []string{"ajax", "legacy"},
	},

	// =====================================================
	// Form Submission Input - Traditional WordPress forms
	// =====================================================
	{
		ID:            "wordpress_post_data",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress $_POST form data",
		Description:   "Form POST data submitted to WordPress",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPPost,
		PopulatedFrom: []string{"$_POST"},
		Tags:          []string{"form", "traditional"},
	},
	{
		ID:            "wordpress_get_data",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress $_GET query parameters",
		Description:   "URL query parameters in WordPress",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"$_GET"},
		Tags:          []string{"query", "traditional"},
	},

	// =====================================================
	// WordPress-specific input retrieval functions
	// These are functions that explicitly retrieve user input
	// =====================================================
	{
		ID:            "wordpress_filter_input",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "filter_input() in WordPress context",
		Description:   "PHP filter_input retrieves external variable",
		MethodPattern: "^filter_input$",
		SourceType:    common.SourceUserInput,
		Tags:          []string{"filter", "validation"},
	},
	{
		ID:            "wordpress_filter_input_array",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "filter_input_array() in WordPress context",
		Description:   "PHP filter_input_array retrieves multiple external variables",
		MethodPattern: "^filter_input_array$",
		SourceType:    common.SourceUserInput,
		Tags:          []string{"filter", "validation"},
	},

	// =====================================================
	// Cookie handling
	// =====================================================
	{
		ID:            "wordpress_cookie_data",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress $_COOKIE data",
		Description:   "Cookie data in WordPress",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPCookie,
		PopulatedFrom: []string{"$_COOKIE"},
		Tags:          []string{"cookie", "session"},
	},

	// =====================================================
	// File uploads
	// =====================================================
	{
		ID:            "wordpress_file_upload",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress $_FILES upload",
		Description:   "File upload data in WordPress",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"$_FILES"},
		Tags:          []string{"upload", "media"},
	},
	{
		ID:            "wordpress_wp_handle_upload",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_handle_upload()",
		Description:   "WordPress file upload handler",
		MethodPattern: "^wp_handle_upload$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"$_FILES"},
		Tags:          []string{"upload", "media"},
	},

	// =====================================================
	// Server/Header information that contains user data
	// =====================================================
	{
		ID:            "wordpress_server_request_uri",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress REQUEST_URI",
		Description:   "Request URI containing user-controlled path",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPPath,
		PopulatedFrom: []string{"$_SERVER['REQUEST_URI']"},
		Tags:          []string{"server", "path"},
	},
	{
		ID:            "wordpress_server_query_string",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress QUERY_STRING",
		Description:   "Raw query string from URL",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"$_SERVER['QUERY_STRING']"},
		Tags:          []string{"server", "query"},
	},
	{
		ID:            "wordpress_server_http_referer",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress HTTP_REFERER",
		Description:   "HTTP Referer header (user-controlled)",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPHeader,
		PopulatedFrom: []string{"$_SERVER['HTTP_REFERER']"},
		Tags:          []string{"server", "header"},
	},
	{
		ID:            "wordpress_server_user_agent",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress HTTP_USER_AGENT",
		Description:   "User-Agent header (user-controlled)",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPHeader,
		PopulatedFrom: []string{"$_SERVER['HTTP_USER_AGENT']"},
		Tags:          []string{"server", "header"},
	},

	// =====================================================
	// WordPress Input Processing Functions
	// These functions WRAP superglobals and are commonly used
	// to access user input in a "WordPress way"
	// =====================================================

	// wp_unslash() - WordPress "magic quotes" all superglobals
	// so wp_unslash() is required to get the actual input value
	// Pattern: wp_unslash($_POST['data']) or wp_unslash($_GET['param'])
	{
		ID:            "wordpress_wp_unslash",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_unslash()",
		Description:   "WordPress unslash - strips magic quotes from input. Common pattern: wp_unslash($_POST['data'])",
		MethodPattern: "^wp_unslash$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST", "$_COOKIE"},
		Tags:          []string{"input-processing", "sanitization"},
	},

	// stripslashes_deep() - Often used with superglobals
	{
		ID:            "wordpress_stripslashes_deep",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "stripslashes_deep()",
		Description:   "WordPress recursive stripslashes - used to process input arrays",
		MethodPattern: "^stripslashes_deep$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"input-processing", "sanitization"},
	},

	// =====================================================
	// WordPress Sanitization Functions
	// These are commonly used to wrap $_GET/$_POST access
	// Pattern: sanitize_text_field($_POST['field'])
	// =====================================================
	{
		ID:            "wordpress_sanitize_text_field",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "sanitize_text_field()",
		Description:   "WordPress text sanitizer - commonly wraps $_POST/$_GET. Pattern: sanitize_text_field(wp_unslash($_POST['data']))",
		MethodPattern: "^sanitize_text_field$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization", "common"},
	},
	{
		ID:            "wordpress_sanitize_textarea_field",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "sanitize_textarea_field()",
		Description:   "WordPress textarea sanitizer - preserves newlines",
		MethodPattern: "^sanitize_textarea_field$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization"},
	},
	{
		ID:            "wordpress_sanitize_email",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "sanitize_email()",
		Description:   "WordPress email sanitizer",
		MethodPattern: "^sanitize_email$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization"},
	},
	{
		ID:            "wordpress_sanitize_file_name",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "sanitize_file_name()",
		Description:   "WordPress filename sanitizer",
		MethodPattern: "^sanitize_file_name$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_FILES"},
		Tags:          []string{"sanitization", "upload"},
	},
	{
		ID:            "wordpress_sanitize_key",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "sanitize_key()",
		Description:   "WordPress key sanitizer - lowercase alphanumeric",
		MethodPattern: "^sanitize_key$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization"},
	},
	{
		ID:            "wordpress_sanitize_title",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "sanitize_title()",
		Description:   "WordPress title sanitizer",
		MethodPattern: "^sanitize_title$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization"},
	},
	{
		ID:            "wordpress_sanitize_url",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "sanitize_url()",
		Description:   "WordPress URL sanitizer (formerly esc_url_raw)",
		MethodPattern: "^sanitize_url$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization"},
	},
	{
		ID:            "wordpress_esc_url_raw",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "esc_url_raw()",
		Description:   "WordPress raw URL sanitizer (for database)",
		MethodPattern: "^esc_url_raw$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization"},
	},
	{
		ID:            "wordpress_absint",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "absint()",
		Description:   "WordPress absolute integer - sanitizes to positive int",
		MethodPattern: "^absint$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization", "integer"},
	},
	{
		ID:            "wordpress_intval",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "intval() in WordPress",
		Description:   "PHP intval commonly used in WordPress for input",
		MethodPattern: "^intval$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization", "integer"},
	},

	// =====================================================
	// wp_parse_args() - Often used with $_GET/$_POST
	// Pattern: wp_parse_args($_GET, $defaults)
	// =====================================================
	{
		ID:            "wordpress_wp_parse_args",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_parse_args()",
		Description:   "WordPress args parser - merges user input with defaults. Pattern: wp_parse_args($_GET, $defaults)",
		MethodPattern: "^wp_parse_args$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"args", "parsing"},
	},
	{
		ID:            "wordpress_shortcode_atts",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "shortcode_atts()",
		Description:   "WordPress shortcode attributes - processes user-provided shortcode attrs",
		MethodPattern: "^shortcode_atts$",
		SourceType:    common.SourceUserInput,
		Tags:          []string{"shortcode", "parsing"},
	},

	// =====================================================
	// WP_REST_Request Array Access
	// WordPress REST API supports ArrayAccess: $request['param']
	// =====================================================
	{
		ID:            "wordpress_rest_array_access",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request array access",
		Description:   "REST API parameter via array access: $request['param']",
		ClassPattern:  "^WP_REST_Request$",
		AccessPattern: "array",
		SourceType:    common.SourceUserInput,
		CarrierClass:  "WP_REST_Request",
		PopulatedFrom: []string{"$_GET", "$_POST", "php://input"},
		Tags:          []string{"rest-api", "array-access"},
	},
	{
		ID:            "wordpress_rest_offsetget",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->offsetGet()",
		Description:   "REST API ArrayAccess offsetGet method",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^offsetGet$",
		SourceType:    common.SourceUserInput,
		CarrierClass:  "WP_REST_Request",
		PopulatedFrom: []string{"$_GET", "$_POST", "php://input"},
		Tags:          []string{"rest-api", "array-access"},
	},
	{
		ID:            "wordpress_rest_has_param",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->has_param()",
		Description:   "REST API parameter existence check - indicates param access",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^has_param$",
		SourceType:    common.SourceUserInput,
		CarrierClass:  "WP_REST_Request",
		Tags:          []string{"rest-api"},
	},
	{
		ID:            "wordpress_rest_get_default_params",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_REST_Request->get_default_params()",
		Description:   "REST API default parameters merged with input",
		ClassPattern:  "^WP_REST_Request$",
		MethodPattern: "^get_default_params$",
		SourceType:    common.SourceUserInput,
		CarrierClass:  "WP_REST_Request",
		Tags:          []string{"rest-api"},
	},

	// =====================================================
	// AJAX Security Functions
	// These verify nonces but their presence indicates input handling
	// =====================================================
	{
		ID:            "wordpress_check_ajax_referer",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "check_ajax_referer()",
		Description:   "WordPress AJAX nonce verification - indicates AJAX handler with input",
		MethodPattern: "^check_ajax_referer$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_POST", "$_GET", "$_REQUEST"},
		Tags:          []string{"ajax", "security", "nonce"},
	},
	{
		ID:            "wordpress_wp_verify_nonce",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_verify_nonce()",
		Description:   "WordPress nonce verification - indicates form/AJAX handler",
		MethodPattern: "^wp_verify_nonce$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_POST", "$_GET", "$_REQUEST"},
		Tags:          []string{"security", "nonce"},
	},

	// =====================================================
	// WordPress Media/Upload Functions
	// =====================================================
	{
		ID:            "wordpress_media_handle_upload",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "media_handle_upload()",
		Description:   "WordPress media upload handler",
		MethodPattern: "^media_handle_upload$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"$_FILES"},
		Tags:          []string{"upload", "media"},
	},
	{
		ID:            "wordpress_wp_handle_sideload",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_handle_sideload()",
		Description:   "WordPress sideload handler (URL downloads)",
		MethodPattern: "^wp_handle_sideload$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"$_FILES"},
		Tags:          []string{"upload", "sideload"},
	},

	// =====================================================
	// Additional Server Variables (user-controllable)
	// =====================================================
	{
		ID:            "wordpress_server_php_self",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress PHP_SELF",
		Description:   "PHP_SELF - vulnerable to XSS if not escaped",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPPath,
		PopulatedFrom: []string{"$_SERVER['PHP_SELF']"},
		Tags:          []string{"server", "path", "xss-risk"},
	},
	{
		ID:            "wordpress_server_path_info",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress PATH_INFO",
		Description:   "PATH_INFO - extra path data after script",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPPath,
		PopulatedFrom: []string{"$_SERVER['PATH_INFO']"},
		Tags:          []string{"server", "path"},
	},
	{
		ID:            "wordpress_server_http_host",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress HTTP_HOST",
		Description:   "HTTP Host header - can be spoofed",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPHeader,
		PopulatedFrom: []string{"$_SERVER['HTTP_HOST']"},
		Tags:          []string{"server", "header"},
	},
	{
		ID:            "wordpress_server_content_type",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress CONTENT_TYPE",
		Description:   "Content-Type header from request",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPHeader,
		PopulatedFrom: []string{"$_SERVER['CONTENT_TYPE']"},
		Tags:          []string{"server", "header"},
	},
	{
		ID:            "wordpress_server_http_authorization",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress HTTP_AUTHORIZATION",
		Description:   "HTTP Authorization header - contains credentials",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPHeader,
		PopulatedFrom: []string{"$_SERVER['HTTP_AUTHORIZATION']"},
		Tags:          []string{"server", "header", "auth"},
	},
	{
		ID:            "wordpress_server_http_x_forwarded",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WordPress HTTP_X_FORWARDED_FOR",
		Description:   "X-Forwarded-For header - proxy/client IP (spoofable)",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPHeader,
		PopulatedFrom: []string{"$_SERVER['HTTP_X_FORWARDED_FOR']"},
		Tags:          []string{"server", "header", "proxy"},
	},

	// =====================================================
	// WordPress Query Variables
	// WP_Query processes user input from URL
	// =====================================================
	{
		ID:            "wordpress_get_query_var",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "get_query_var()",
		Description:   "WordPress query variable - from URL/rewrite rules",
		MethodPattern: "^get_query_var$",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"$_GET", "$_SERVER['REQUEST_URI']"},
		Tags:          []string{"query", "url"},
	},
	{
		ID:            "wordpress_wp_query_get",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "WP_Query->get()",
		Description:   "WordPress query object variable access",
		ClassPattern:  "^WP_Query$",
		MethodPattern: "^get$",
		SourceType:    common.SourceHTTPGet,
		Tags:          []string{"query"},
	},

	// =====================================================
	// Additional WordPress Sanitization Functions
	// These commonly wrap $_GET/$_POST access
	// =====================================================
	{
		ID:            "wordpress_sanitize_user",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "sanitize_user()",
		Description:   "WordPress username sanitizer",
		MethodPattern: "^sanitize_user$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization", "user"},
	},
	{
		ID:            "wordpress_sanitize_html_class",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "sanitize_html_class()",
		Description:   "WordPress HTML class sanitizer",
		MethodPattern: "^sanitize_html_class$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization"},
	},
	{
		ID:            "wordpress_sanitize_mime_type",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "sanitize_mime_type()",
		Description:   "WordPress MIME type sanitizer",
		MethodPattern: "^sanitize_mime_type$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_FILES", "$_POST"},
		Tags:          []string{"sanitization", "upload"},
	},
	{
		ID:            "wordpress_wp_kses",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_kses()",
		Description:   "WordPress KSES HTML filter - strips disallowed HTML",
		MethodPattern: "^wp_kses$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization", "html"},
	},
	{
		ID:            "wordpress_wp_kses_post",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_kses_post()",
		Description:   "WordPress KSES for post content",
		MethodPattern: "^wp_kses_post$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization", "html"},
	},
	{
		ID:            "wordpress_wp_kses_data",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_kses_data()",
		Description:   "WordPress KSES for untrusted data",
		MethodPattern: "^wp_kses_data$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization", "html"},
	},
	{
		ID:            "wordpress_map_deep",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "map_deep()",
		Description:   "WordPress recursive array sanitizer - commonly wraps $_POST arrays",
		MethodPattern: "^map_deep$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_GET", "$_POST", "$_REQUEST"},
		Tags:          []string{"sanitization", "array"},
	},

	// =====================================================
	// WordPress HTTP API - Incoming request data
	// =====================================================
	{
		ID:            "wordpress_wp_remote_retrieve_body",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_remote_retrieve_body()",
		Description:   "WordPress HTTP API response body - external data (NOT direct user input)",
		MethodPattern: "^wp_remote_retrieve_body$",
		SourceType:    common.SourceNetwork,
		Tags:          []string{"http", "external"},
	},

	// =====================================================
	// WordPress Form Handling
	// =====================================================
	{
		ID:            "wordpress_isset_post",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "isset($_POST[...])",
		Description:   "WordPress form submission check - indicates POST handling",
		AccessPattern: "superglobal",
		SourceType:    common.SourceHTTPPost,
		PopulatedFrom: []string{"$_POST"},
		Tags:          []string{"form", "check"},
	},
	{
		ID:            "wordpress_empty_request",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "empty($_REQUEST[...])",
		Description:   "WordPress request data check",
		AccessPattern: "superglobal",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_REQUEST"},
		Tags:          []string{"check"},
	},

	// =====================================================
	// WordPress Admin AJAX Actions
	// =====================================================
	{
		ID:            "wordpress_doing_ajax",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_doing_ajax()",
		Description:   "WordPress AJAX context check - indicates AJAX handler",
		MethodPattern: "^wp_doing_ajax$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"$_REQUEST"},
		Tags:          []string{"ajax", "context"},
	},

	// =====================================================
	// WordPress Multipart Form Data
	// =====================================================
	{
		ID:            "wordpress_wp_check_filetype",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_check_filetype()",
		Description:   "WordPress file type check - validates uploaded file",
		MethodPattern: "^wp_check_filetype$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"$_FILES"},
		Tags:          []string{"upload", "validation"},
	},
	{
		ID:            "wordpress_wp_check_filetype_and_ext",
		Framework:     "wordpress",
		Language:      "php",
		Name:          "wp_check_filetype_and_ext()",
		Description:   "WordPress file type and extension check",
		MethodPattern: "^wp_check_filetype_and_ext$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"$_FILES"},
		Tags:          []string{"upload", "validation"},
	},
}

func init() {
	Registry.RegisterAll(wordpressPatterns)

	// Register framework detector
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "wordpress",
		Indicators: []string{"wp-config.php", "wp-includes/version.php", "wp-admin/admin.php"},
	})
}
