package php

import (
	"regexp"

	"github.com/hatlesswizard/inputtracer/pkg/sources/core"
)

func init() {
	registerWordPressPatterns()
}

func registerWordPressPatterns() {
	registry := core.GetRegistry()

	// WordPress REST API request methods
	wpRestMethods := []struct {
		method   string
		category core.SourceType
		labels   []core.InputLabel
		desc     string
	}{
		{"get_param", core.SourceHTTPRequest, []core.InputLabel{core.LabelUserInput}, "REST API single parameter"},
		{"get_params", core.SourceHTTPRequest, []core.InputLabel{core.LabelUserInput}, "REST API all parameters"},
		{"get_query_params", core.SourceHTTPGet, []core.InputLabel{core.LabelHTTPGet, core.LabelUserInput}, "REST API query params"},
		{"get_body_params", core.SourceHTTPPost, []core.InputLabel{core.LabelHTTPPost, core.LabelUserInput}, "REST API body params"},
		{"get_json_params", core.SourceHTTPBody, []core.InputLabel{core.LabelHTTPBody, core.LabelUserInput}, "REST API JSON body"},
		{"get_body", core.SourceHTTPBody, []core.InputLabel{core.LabelHTTPBody, core.LabelUserInput}, "REST API raw body"},
		{"get_file_params", core.SourceHTTPFile, []core.InputLabel{core.LabelFile, core.LabelUserInput}, "REST API file uploads"},
		{"get_header", core.SourceHTTPHeader, []core.InputLabel{core.LabelHTTPHeader, core.LabelUserInput}, "REST API single header"},
		{"get_headers", core.SourceHTTPHeader, []core.InputLabel{core.LabelHTTPHeader, core.LabelUserInput}, "REST API all headers"},
	}

	for _, m := range wpRestMethods {
		registry.Register(&core.InputPattern{
			Name:        "wordpress_rest_" + m.method,
			Description: "WordPress " + m.desc,
			Category:    m.category,
			Labels:      m.labels,
			Language:    "php",
			Framework:   "wordpress",
			MethodName:  m.method,
			Regex:       regexp.MustCompile(`->` + m.method + `\s*\(`),
		})
	}

	// WordPress functions that receive user input
	wpInputFunctions := []struct {
		function string
		category core.SourceType
		labels   []core.InputLabel
		desc     string
	}{
		// Query variables
		{"get_query_var", core.SourceHTTPGet, []core.InputLabel{core.LabelHTTPGet, core.LabelUserInput}, "URL query variable"},
		{"get_search_query", core.SourceHTTPGet, []core.InputLabel{core.LabelHTTPGet, core.LabelUserInput}, "Search query from URL"},

		// Sanitization functions (first param is user input)
		{"sanitize_text_field", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user text input"},
		{"sanitize_textarea_field", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user textarea"},
		{"sanitize_email", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user email"},
		{"sanitize_file_name", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user filename"},
		{"sanitize_html_class", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user HTML class"},
		{"sanitize_key", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user key"},
		{"sanitize_title", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user title"},
		{"sanitize_user", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes username"},
		{"sanitize_url", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Sanitizes user URL"},

		// Escaping functions (first param is user input)
		{"esc_html", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes HTML from user"},
		{"esc_attr", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes attribute from user"},
		{"esc_url", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes URL from user"},
		{"esc_js", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes JS from user"},
		{"esc_textarea", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes textarea from user"},
		{"esc_sql", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Escapes SQL from user"},

		// Type conversion (implies user input)
		{"absint", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Converts user input to absolute int"},
		{"intval", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Converts user input to int"},
		{"wp_unslash", core.SourceUserInput, []core.InputLabel{core.LabelUserInput}, "Unslashes user input"},

		// File uploads
		{"wp_handle_upload", core.SourceHTTPFile, []core.InputLabel{core.LabelFile, core.LabelUserInput}, "Handles file upload"},
		{"media_handle_upload", core.SourceHTTPFile, []core.InputLabel{core.LabelFile, core.LabelUserInput}, "Handles media upload"},
		{"wp_handle_sideload", core.SourceHTTPFile, []core.InputLabel{core.LabelFile, core.LabelUserInput}, "Handles sideload upload"},
	}

	for _, f := range wpInputFunctions {
		registry.Register(&core.InputPattern{
			Name:        "wordpress_" + f.function,
			Description: "WordPress " + f.desc,
			Category:    f.category,
			Labels:      f.labels,
			Language:    "php",
			Framework:   "wordpress",
			ExactMatch:  f.function,
			Regex:       regexp.MustCompile(`\b` + f.function + `\s*\(`),
			ParamIndex:  0, // First parameter receives input
		})
	}

	// WordPress non-input functions (explicitly excluded)
	wpNonInput := []string{
		"get_option",
		"get_site_option",
		"get_transient",
		"get_site_transient",
		"get_post_meta",
		"get_user_meta",
		"get_term_meta",
		"get_comment_meta",
		"get_metadata",
		"wp_cache_get",
		"get_post",
		"get_page",
		"get_user_by",
		"get_userdata",
		"get_term",
		"get_term_by",
		"get_category",
		"get_tag",
		"get_the_ID",
		"get_the_title",
		"get_the_content",
		"get_the_excerpt",
		"get_permalink",
		"get_bloginfo",
		"get_template_directory",
		"get_stylesheet_directory",
		"home_url",
		"admin_url",
		"site_url",
		"content_url",
		"plugins_url",
	}

	for _, f := range wpNonInput {
		registry.RegisterNonInput(f)
	}

	// Register AJAX action context patterns
	registry.Register(&core.InputPattern{
		Name:        "wordpress_ajax_action",
		Description: "WordPress AJAX action handler receives POST data",
		Category:    core.SourceHTTPPost,
		Labels:      []core.InputLabel{core.LabelHTTPPost, core.LabelUserInput},
		Language:    "php",
		Framework:   "wordpress",
		Regex:       regexp.MustCompile(`add_action\s*\(\s*['"]wp_ajax_(nopriv_)?`),
	})

	// Register admin POST action context
	registry.Register(&core.InputPattern{
		Name:        "wordpress_admin_post",
		Description: "WordPress admin POST action handler",
		Category:    core.SourceHTTPPost,
		Labels:      []core.InputLabel{core.LabelHTTPPost, core.LabelUserInput},
		Language:    "php",
		Framework:   "wordpress",
		Regex:       regexp.MustCompile(`add_action\s*\(\s*['"]admin_post_(nopriv_)?`),
	})
}

// WordPressIndicators returns file patterns that indicate WordPress
var WordPressIndicators = []string{
	"wp-config.php",
	"wp-content/",
	"wp-includes/",
	"wp-admin/",
	"wp-load.php",
	"wp-settings.php",
}

// WordPressFunctionIndicators returns functions that indicate WordPress
var WordPressFunctionIndicators = []string{
	"add_action",
	"add_filter",
	"do_action",
	"apply_filters",
	"register_activation_hook",
	"register_deactivation_hook",
	"wp_enqueue_script",
	"wp_enqueue_style",
}

// WordPressConstantIndicators returns constants that indicate WordPress
var WordPressConstantIndicators = []string{
	"ABSPATH",
	"WPINC",
	"WP_CONTENT_DIR",
	"WP_PLUGIN_DIR",
	"TEMPLATEPATH",
	"STYLESHEETPATH",
}

// IsWordPress checks if the given indicators suggest WordPress
func IsWordPress(files []string, functions []string, constants []string) bool {
	// Check for WordPress files
	for _, file := range files {
		for _, indicator := range WordPressIndicators {
			if file == indicator || containsPath(file, indicator) {
				return true
			}
		}
	}

	// Check for WordPress functions
	wpFuncCount := 0
	for _, fn := range functions {
		for _, indicator := range WordPressFunctionIndicators {
			if fn == indicator {
				wpFuncCount++
				if wpFuncCount >= 2 {
					return true
				}
			}
		}
	}

	// Check for WordPress constants
	for _, c := range constants {
		for _, indicator := range WordPressConstantIndicators {
			if c == indicator {
				return true
			}
		}
	}

	return false
}

func containsPath(path, indicator string) bool {
	return len(path) >= len(indicator) &&
		(path == indicator ||
			(len(path) > len(indicator) &&
				(path[len(path)-len(indicator)-1] == '/' || path[len(path)-len(indicator)-1] == '\\')))
}
