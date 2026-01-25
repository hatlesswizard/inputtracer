package rust

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// Matcher matches Rust user input sources
type Matcher struct {
	*common.BaseMatcher
}

// NewMatcher creates a new Rust source matcher
func NewMatcher() *Matcher {
	defs := []common.Definition{
		// CLI arguments
		{
			Name:        "std::env::args()",
			Pattern:     `std::env::args\s*\(\s*\)|env::args\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Command line arguments iterator",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::env::args_os()",
			Pattern:     `std::env::args_os\s*\(\s*\)|env::args_os\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "OS command line arguments",
			NodeTypes:   []string{"call_expression"},
		},

		// Environment variables
		{
			Name:         "std::env::var()",
			Pattern:      `std::env::var\s*\(|env::var\s*\(`,
			Language:     "rust",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "Get environment variable",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `env::var\s*\(\s*"([^"]+)"`,
		},
		{
			Name:        "std::env::var_os()",
			Pattern:     `std::env::var_os\s*\(|env::var_os\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelEnvironment},
			Description: "Get environment variable as OsString",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::env::vars()",
			Pattern:     `std::env::vars\s*\(\s*\)|env::vars\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelEnvironment},
			Description: "All environment variables",
			NodeTypes:   []string{"call_expression"},
		},

		// Standard input
		{
			Name:        "std::io::stdin().read_line()",
			Pattern:     `stdin\(\)\.read_line\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Read line from stdin",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::io::stdin().read()",
			Pattern:     `stdin\(\)\.read\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Read from stdin",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::io::stdin().read_to_string()",
			Pattern:     `stdin\(\)\.read_to_string\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Read stdin to string",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::io::stdin().read_to_end()",
			Pattern:     `stdin\(\)\.read_to_end\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Read stdin to end",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "BufRead::read_line()",
			Pattern:     `\.read_line\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelFile},
			Description: "Read line from buffered reader",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "BufRead::lines()",
			Pattern:     `\.lines\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelFile},
			Description: "Lines iterator from buffered reader",
			NodeTypes:   []string{"call_expression"},
		},

		// File operations
		{
			Name:        "std::fs::read_to_string()",
			Pattern:     `std::fs::read_to_string\s*\(|fs::read_to_string\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read file to string",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::fs::read()",
			Pattern:     `std::fs::read\s*\(|fs::read\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read file to bytes",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "File::open().read()",
			Pattern:     `\.read\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read from file",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "File::open().read_to_string()",
			Pattern:     `\.read_to_string\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read file to string",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "File::open().read_to_end()",
			Pattern:     `\.read_to_end\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read file to end",
			NodeTypes:   []string{"call_expression"},
		},

		// Actix-web
		{
			Name:        "web::Query",
			Pattern:     `web::Query`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "Actix query parameters",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "web::Form",
			Pattern:     `web::Form`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "Actix form data",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "web::Path",
			Pattern:     `web::Path`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "Actix path parameters",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "web::Json",
			Pattern:     `web::Json`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Actix JSON body",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "web::Bytes",
			Pattern:     `web::Bytes`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Actix raw body bytes",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "HttpRequest::headers()",
			Pattern:     `\.headers\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Actix request headers",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "HttpRequest::cookie()",
			Pattern:     `\.cookie\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description: "Actix cookie",
			NodeTypes:   []string{"call_expression"},
		},

		// Axum
		{
			Name:        "axum::extract::Query",
			Pattern:     `extract::Query|Query<`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "Axum query parameters",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "axum::extract::Form",
			Pattern:     `extract::Form|Form<`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "Axum form data",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "axum::extract::Path",
			Pattern:     `extract::Path|Path<`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "Axum path parameters",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "axum::extract::Json",
			Pattern:     `extract::Json|Json<`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Axum JSON body",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "axum::extract::RawBody",
			Pattern:     `extract::RawBody|RawBody`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Axum raw body",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "TypedHeader",
			Pattern:     `TypedHeader<`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Axum typed header",
			NodeTypes:   []string{"type_identifier"},
		},

		// Rocket
		{
			Name:        "FromForm",
			Pattern:     `#\[derive\([^)]*FromForm`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "Rocket form derive",
			NodeTypes:   []string{"attribute_item"},
		},
		{
			Name:        "FromRequest",
			Pattern:     `#\[derive\([^)]*FromRequest`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "Rocket request derive",
			NodeTypes:   []string{"attribute_item"},
		},
		{
			Name:        "rocket::form::Form",
			Pattern:     `Form<`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "Rocket form data",
			NodeTypes:   []string{"type_identifier"},
		},

		// CLI parsing (clap)
		{
			Name:        "clap::Parser",
			Pattern:     `#\[derive\([^)]*Parser`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Clap CLI parser derive",
			NodeTypes:   []string{"attribute_item"},
		},
		{
			Name:        "clap::Arg",
			Pattern:     `Arg::new\s*\(|\.arg\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Clap argument",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "clap::ArgMatches::value_of()",
			Pattern:     `\.value_of\s*\(|\.get_one\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Clap get argument value",
			NodeTypes:   []string{"call_expression"},
		},

		// Network (reqwest)
		{
			Name:        "reqwest::get()",
			Pattern:     `reqwest::get\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "Reqwest HTTP GET",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "Client::get()",
			Pattern:     `\.get\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "HTTP client GET",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "Response::text()",
			Pattern:     `\.text\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "HTTP response text",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "Response::json()",
			Pattern:     `\.json\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "HTTP response JSON",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "Response::bytes()",
			Pattern:     `\.bytes\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "HTTP response bytes",
			NodeTypes:   []string{"call_expression"},
		},

		// JSON parsing (serde_json)
		{
			Name:        "serde_json::from_str()",
			Pattern:     `serde_json::from_str\s*\(|from_str\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "JSON parse from string",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "serde_json::from_slice()",
			Pattern:     `serde_json::from_slice\s*\(|from_slice\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "JSON parse from bytes",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "serde_json::from_reader()",
			Pattern:     `serde_json::from_reader\s*\(|from_reader\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelFile},
			Description: "JSON parse from reader",
			NodeTypes:   []string{"call_expression"},
		},

		// TOML/YAML parsing
		{
			Name:        "toml::from_str()",
			Pattern:     `toml::from_str\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelFile},
			Description: "TOML parse",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "serde_yaml::from_str()",
			Pattern:     `serde_yaml::from_str\s*\(`,
			Language:    "rust",
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelFile},
			Description: "YAML parse",
			NodeTypes:   []string{"call_expression"},
		},
	}

	return &Matcher{
		BaseMatcher: common.NewBaseMatcher("rust", defs),
	}
}
