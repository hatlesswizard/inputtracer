package php

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// wordpressHookDefinitions returns definitions for WordPress hook registration patterns
// that signal user-controlled data entry points (AJAX handlers, shortcodes, filters, actions).
func wordpressHookDefinitions() []common.Definition {
	return []common.Definition{
		{
			Name:           "add_action(wp_ajax_*)",
			Pattern:        `\badd_action\s*\(\s*['"]wp_ajax_`,
			ExcludePattern: `\badd_action\s*\(\s*['"]wp_ajax_nopriv_`,
			Language:       "php",
			Labels:         []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:    "WordPress AJAX handler for logged-in users — callback receives $_POST data",
			NodeTypes:      []string{"function_call_expression"},
			KeyExtractor:   `\badd_action\s*\(\s*['"]wp_ajax_([^'"]+)['"]`,
		},
		{
			Name:         "add_action(wp_ajax_nopriv_*)",
			Pattern:      `\badd_action\s*\(\s*['"]wp_ajax_nopriv_`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "WordPress AJAX handler for logged-out users — callback receives $_POST data",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\badd_action\s*\(\s*['"]wp_ajax_nopriv_([^'"]+)['"]`,
		},
		{
			Name:         "add_shortcode()",
			Pattern:      `\badd_shortcode\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "WordPress shortcode registration — callback $atts contains user-supplied attributes from post content",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\badd_shortcode\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "apply_filters() on user-data hooks",
			Pattern:      `\bapply_filters\s*\(\s*['"](?:the_content|comment_text|widget_text|get_search_query|request|pre_get_posts|the_title|get_the_excerpt|the_excerpt|comment_author|get_comment_author)['"]`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "WordPress filter on known user-controlled data (content, comments, search query, request)",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bapply_filters\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "do_action() on request lifecycle hooks",
			Pattern:      `\bdo_action\s*\(\s*['"](?:init|wp_loaded|wp|parse_request|pre_get_posts|admin_init|admin_post|wp_ajax|template_redirect)['"]`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "WordPress action at request lifecycle point where $_GET/$_POST is in scope",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bdo_action\s*\(\s*['"]([^'"]+)['"]`,
		},
	}
}

// wordpressSanitizerDefinitions returns definitions for WordPress sanitizer functions
// wrapping superglobals ($_GET, $_POST, $_REQUEST). These wrappers do not remove the
// user-controlled nature of the data — the input still originates from the HTTP request.
func wordpressSanitizerDefinitions() []common.Definition {
	return []common.Definition{
		{
			Name:         "sanitize_text_field($_GET/$_POST/$_REQUEST)",
			Pattern:      `\bsanitize_text_field\s*\(\s*\$_(GET|POST|REQUEST)`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "sanitize_text_field() wrapping a superglobal — data still originates from HTTP request",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bsanitize_text_field\s*\(\s*\$_(?:GET|POST|REQUEST)\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "absint($_GET/$_POST/$_REQUEST)",
			Pattern:      `\babsint\s*\(\s*\$_(GET|POST|REQUEST)`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "absint() wrapping a superglobal — data still originates from HTTP request",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\babsint\s*\(\s*\$_(?:GET|POST|REQUEST)\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "intval($_GET/$_POST/$_REQUEST)",
			Pattern:      `\bintval\s*\(\s*\$_(GET|POST|REQUEST)`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "intval() wrapping a superglobal — data still originates from HTTP request",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bintval\s*\(\s*\$_(?:GET|POST|REQUEST)\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "esc_attr($_GET/$_POST/$_REQUEST)",
			Pattern:      `\besc_attr\s*\(\s*\$_(GET|POST|REQUEST)`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "esc_attr() wrapping a superglobal — data still originates from HTTP request",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\besc_attr\s*\(\s*\$_(?:GET|POST|REQUEST)\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "esc_html($_GET/$_POST/$_REQUEST)",
			Pattern:      `\besc_html\s*\(\s*\$_(GET|POST|REQUEST)`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "esc_html() wrapping a superglobal — data still originates from HTTP request",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\besc_html\s*\(\s*\$_(?:GET|POST|REQUEST)\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "sanitize_email($_GET/$_POST/$_REQUEST)",
			Pattern:      `\bsanitize_email\s*\(\s*\$_(GET|POST|REQUEST)`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "sanitize_email() wrapping a superglobal — data still originates from HTTP request",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bsanitize_email\s*\(\s*\$_(?:GET|POST|REQUEST)\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
		{
			Name:         "sanitize_key($_GET/$_POST/$_REQUEST)",
			Pattern:      `\bsanitize_key\s*\(\s*\$_(GET|POST|REQUEST)`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput},
			Description:  "sanitize_key() wrapping a superglobal — data still originates from HTTP request",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bsanitize_key\s*\(\s*\$_(?:GET|POST|REQUEST)\s*\[\s*['"]?([^'"\]]+)['"]?\s*\]`,
		},
	}
}

// wordpressMetaDefinitions returns definitions for WordPress meta API functions that read
// values from the database. Although stored server-side, these values may have originally
// been set by users and therefore carry a database-origin label.
func wordpressMetaDefinitions() []common.Definition {
	return []common.Definition{
		{
			Name:         "get_post_meta()",
			Pattern:      `\bget_post_meta\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelDatabase},
			Description:  "WordPress get_post_meta() — reads post meta from database (user may have set this value)",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bget_post_meta\s*\(\s*[^,]+,\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "get_user_meta()",
			Pattern:      `\bget_user_meta\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelDatabase},
			Description:  "WordPress get_user_meta() — reads user meta from database (user may have set this value)",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bget_user_meta\s*\(\s*[^,]+,\s*['"]([^'"]+)['"]`,
		},
		{
			Name:         "get_term_meta()",
			Pattern:      `\bget_term_meta\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelDatabase},
			Description:  "WordPress get_term_meta() — reads term meta from database (user may have set this value)",
			NodeTypes:    []string{"function_call_expression"},
			KeyExtractor: `\bget_term_meta\s*\(\s*[^,]+,\s*['"]([^'"]+)['"]`,
		},
	}
}

// wordpressRESTMethodDefinitions returns definitions for WP_REST_Request method calls.
// These cover the full surface of how WordPress REST API handlers read request data.
func wordpressRESTMethodDefinitions() []common.Definition {
	return []common.Definition{
		{
			Name:         "->get_param()",
			Pattern:      `->\s*get_param\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelUserInput, common.LabelHTTPGet, common.LabelHTTPPost},
			Description:  "WP_REST_Request::get_param() — retrieves a single request parameter (GET or POST)",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*get_param\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "->get_params()",
			Pattern:     `->\s*get_params\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelHTTPGet, common.LabelHTTPPost},
			Description: "WP_REST_Request::get_params() — retrieves all merged request parameters",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->get_query_params()",
			Pattern:     `->\s*get_query_params\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "WP_REST_Request::get_query_params() — retrieves URL query string parameters",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->get_body_params()",
			Pattern:     `->\s*get_body_params\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "WP_REST_Request::get_body_params() — retrieves POST body parameters",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->get_json_params()",
			Pattern:     `->\s*get_json_params\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "WP_REST_Request::get_json_params() — retrieves decoded JSON request body",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:        "->get_body()",
			Pattern:     `->\s*get_body\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "WP_REST_Request::get_body() — retrieves the raw request body string",
			NodeTypes:   []string{"member_call_expression"},
		},
		{
			Name:         "->get_header()",
			Pattern:      `->\s*get_header\s*\(`,
			Language:     "php",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "WP_REST_Request::get_header() — retrieves a single HTTP request header",
			NodeTypes:    []string{"member_call_expression"},
			KeyExtractor: `->\s*get_header\s*\(\s*['"]([^'"]+)['"]`,
		},
		{
			Name:        "->get_headers()",
			Pattern:     `->\s*get_headers\s*\(`,
			Language:    "php",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "WP_REST_Request::get_headers() — retrieves all HTTP request headers",
			NodeTypes:   []string{"member_call_expression"},
		},
	}
}
