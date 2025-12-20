package sources

// RustMatcher matches Rust user input sources
type RustMatcher struct {
	*BaseMatcher
}

// NewRustMatcher creates a new Rust source matcher
func NewRustMatcher() *RustMatcher {
	sources := []Definition{
		// CLI arguments
		{
			Name:        "std::env::args()",
			Pattern:     `std::env::args\s*\(\s*\)|env::args\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []InputLabel{LabelCLI},
			Description: "Command line arguments iterator",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::env::args_os()",
			Pattern:     `std::env::args_os\s*\(\s*\)|env::args_os\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []InputLabel{LabelCLI},
			Description: "OS command line arguments",
			NodeTypes:   []string{"call_expression"},
		},

		// Environment variables
		{
			Name:         "std::env::var()",
			Pattern:      `std::env::var\s*\(|env::var\s*\(`,
			Language:     "rust",
			Labels:       []InputLabel{LabelEnvironment},
			Description:  "Get environment variable",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `env::var\s*\(\s*"([^"]+)"`,
		},
		{
			Name:        "std::env::var_os()",
			Pattern:     `std::env::var_os\s*\(|env::var_os\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelEnvironment},
			Description: "Get environment variable as OsString",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::env::vars()",
			Pattern:     `std::env::vars\s*\(\s*\)|env::vars\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []InputLabel{LabelEnvironment},
			Description: "All environment variables",
			NodeTypes:   []string{"call_expression"},
		},

		// Standard input
		{
			Name:        "std::io::stdin().read_line()",
			Pattern:     `stdin\(\)\.read_line\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Read line from stdin",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::io::stdin().read()",
			Pattern:     `stdin\(\)\.read\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Read from stdin",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::io::stdin().read_to_string()",
			Pattern:     `stdin\(\)\.read_to_string\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Read stdin to string",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::io::stdin().read_to_end()",
			Pattern:     `stdin\(\)\.read_to_end\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Read stdin to end",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "BufRead::read_line()",
			Pattern:     `\.read_line\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Read line from buffered reader",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "BufRead::lines()",
			Pattern:     `\.lines\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Lines iterator from buffered reader",
			NodeTypes:   []string{"call_expression"},
		},

		// File operations
		{
			Name:        "std::fs::read_to_string()",
			Pattern:     `std::fs::read_to_string\s*\(|fs::read_to_string\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelFile},
			Description: "Read file to string",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "std::fs::read()",
			Pattern:     `std::fs::read\s*\(|fs::read\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelFile},
			Description: "Read file to bytes",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "File::open().read()",
			Pattern:     `\.read\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelFile},
			Description: "Read from file",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "File::open().read_to_string()",
			Pattern:     `\.read_to_string\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelFile},
			Description: "Read file to string",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "File::open().read_to_end()",
			Pattern:     `\.read_to_end\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelFile},
			Description: "Read file to end",
			NodeTypes:   []string{"call_expression"},
		},

		// Actix-web
		{
			Name:        "web::Query",
			Pattern:     `web::Query`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "Actix query parameters",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "web::Form",
			Pattern:     `web::Form`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPPost, LabelUserInput},
			Description: "Actix form data",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "web::Path",
			Pattern:     `web::Path`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "Actix path parameters",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "web::Json",
			Pattern:     `web::Json`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPBody, LabelUserInput},
			Description: "Actix JSON body",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "web::Bytes",
			Pattern:     `web::Bytes`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPBody, LabelUserInput},
			Description: "Actix raw body bytes",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "HttpRequest::headers()",
			Pattern:     `\.headers\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPHeader, LabelUserInput},
			Description: "Actix request headers",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "HttpRequest::cookie()",
			Pattern:     `\.cookie\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPCookie, LabelUserInput},
			Description: "Actix cookie",
			NodeTypes:   []string{"call_expression"},
		},

		// Axum
		{
			Name:        "axum::extract::Query",
			Pattern:     `extract::Query|Query<`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "Axum query parameters",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "axum::extract::Form",
			Pattern:     `extract::Form|Form<`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPPost, LabelUserInput},
			Description: "Axum form data",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "axum::extract::Path",
			Pattern:     `extract::Path|Path<`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "Axum path parameters",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "axum::extract::Json",
			Pattern:     `extract::Json|Json<`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPBody, LabelUserInput},
			Description: "Axum JSON body",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "axum::extract::RawBody",
			Pattern:     `extract::RawBody|RawBody`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPBody, LabelUserInput},
			Description: "Axum raw body",
			NodeTypes:   []string{"type_identifier", "scoped_identifier"},
		},
		{
			Name:        "TypedHeader",
			Pattern:     `TypedHeader<`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPHeader, LabelUserInput},
			Description: "Axum typed header",
			NodeTypes:   []string{"type_identifier"},
		},

		// Rocket
		{
			Name:        "FromForm",
			Pattern:     `#\[derive\([^)]*FromForm`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPPost, LabelUserInput},
			Description: "Rocket form derive",
			NodeTypes:   []string{"attribute_item"},
		},
		{
			Name:        "FromRequest",
			Pattern:     `#\[derive\([^)]*FromRequest`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelUserInput},
			Description: "Rocket request derive",
			NodeTypes:   []string{"attribute_item"},
		},
		{
			Name:        "rocket::form::Form",
			Pattern:     `Form<`,
			Language:    "rust",
			Labels:      []InputLabel{LabelHTTPPost, LabelUserInput},
			Description: "Rocket form data",
			NodeTypes:   []string{"type_identifier"},
		},

		// CLI parsing (clap)
		{
			Name:        "clap::Parser",
			Pattern:     `#\[derive\([^)]*Parser`,
			Language:    "rust",
			Labels:      []InputLabel{LabelCLI},
			Description: "Clap CLI parser derive",
			NodeTypes:   []string{"attribute_item"},
		},
		{
			Name:        "clap::Arg",
			Pattern:     `Arg::new\s*\(|\.arg\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelCLI},
			Description: "Clap argument",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "clap::ArgMatches::value_of()",
			Pattern:     `\.value_of\s*\(|\.get_one\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelCLI},
			Description: "Clap get argument value",
			NodeTypes:   []string{"call_expression"},
		},

		// Network (reqwest)
		{
			Name:        "reqwest::get()",
			Pattern:     `reqwest::get\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelNetwork},
			Description: "Reqwest HTTP GET",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "Client::get()",
			Pattern:     `\.get\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelNetwork},
			Description: "HTTP client GET",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "Response::text()",
			Pattern:     `\.text\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []InputLabel{LabelNetwork},
			Description: "HTTP response text",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "Response::json()",
			Pattern:     `\.json\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []InputLabel{LabelNetwork},
			Description: "HTTP response JSON",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "Response::bytes()",
			Pattern:     `\.bytes\s*\(\s*\)`,
			Language:    "rust",
			Labels:      []InputLabel{LabelNetwork},
			Description: "HTTP response bytes",
			NodeTypes:   []string{"call_expression"},
		},

		// Database (sqlx, diesel)
		{
			Name:        "sqlx::query().fetch()",
			Pattern:     `\.fetch\s*\(|\.fetch_one\s*\(|\.fetch_all\s*\(|\.fetch_optional\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelDatabase},
			Description: "SQLx query fetch",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "diesel::query",
			Pattern:     `\.load\s*\(|\.first\s*\(|\.get_result\s*\(|\.get_results\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelDatabase},
			Description: "Diesel query execution",
			NodeTypes:   []string{"call_expression"},
		},

		// JSON parsing (serde_json)
		{
			Name:        "serde_json::from_str()",
			Pattern:     `serde_json::from_str\s*\(|from_str\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelUserInput},
			Description: "JSON parse from string",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "serde_json::from_slice()",
			Pattern:     `serde_json::from_slice\s*\(|from_slice\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelUserInput},
			Description: "JSON parse from bytes",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "serde_json::from_reader()",
			Pattern:     `serde_json::from_reader\s*\(|from_reader\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "JSON parse from reader",
			NodeTypes:   []string{"call_expression"},
		},

		// TOML/YAML parsing
		{
			Name:        "toml::from_str()",
			Pattern:     `toml::from_str\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "TOML parse",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "serde_yaml::from_str()",
			Pattern:     `serde_yaml::from_str\s*\(`,
			Language:    "rust",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "YAML parse",
			NodeTypes:   []string{"call_expression"},
		},
	}

	return &RustMatcher{
		BaseMatcher: NewBaseMatcher("rust", sources),
	}
}
