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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    1.0,
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
		Confidence:    0.95,
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
		Confidence:    0.95,
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
		Confidence:    0.9,
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
		Confidence:    0.9,
		Tags:          []string{"server", "header"},
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
