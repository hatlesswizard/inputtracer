// Package rust - frameworks.go provides Rust web framework patterns
// Includes patterns for Actix-web, Rocket, Axum, Warp, and Tide
package rust

import (
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// Registry is the global Rust framework pattern registry
var Registry = common.NewFrameworkPatternRegistry("rust")

// Actix-web patterns (powerful, pragmatic, and extremely fast web framework)
var actixPatterns = []*common.FrameworkPattern{
	{
		ID:           "actix_web_query",
		Framework:    "actix-web",
		Language:     "rust",
		Name:         "web::Query<T>",
		Description:  "Actix-web query string extractor",
		ClassPattern: "^web::Query$",
		SourceType:   common.SourceHTTPGet,
		Tags:         []string{"web", "async", "popular"},
	},
	{
		ID:           "actix_web_form",
		Framework:    "actix-web",
		Language:     "rust",
		Name:         "web::Form<T>",
		Description:  "Actix-web form data extractor",
		ClassPattern: "^web::Form$",
		SourceType:   common.SourceHTTPPost,
		Tags:         []string{"web", "async", "popular"},
	},
	{
		ID:           "actix_web_path",
		Framework:    "actix-web",
		Language:     "rust",
		Name:         "web::Path<T>",
		Description:  "Actix-web path parameter extractor",
		ClassPattern: "^web::Path$",
		SourceType:   common.SourceHTTPPath,
		Tags:         []string{"web", "async", "popular"},
	},
	{
		ID:           "actix_web_json",
		Framework:    "actix-web",
		Language:     "rust",
		Name:         "web::Json<T>",
		Description:  "Actix-web JSON body extractor",
		ClassPattern: "^web::Json$",
		SourceType:   common.SourceHTTPJSON,
		Tags:         []string{"web", "async", "popular"},
	},
	{
		ID:           "actix_web_bytes",
		Framework:    "actix-web",
		Language:     "rust",
		Name:         "web::Bytes",
		Description:  "Actix-web raw body bytes",
		ClassPattern: "^web::Bytes$",
		SourceType:   common.SourceHTTPBody,
		Tags:         []string{"web", "async"},
	},
	{
		ID:            "actix_web_headers",
		Framework:     "actix-web",
		Language:      "rust",
		Name:          "HttpRequest::headers()",
		Description:   "Actix-web request headers",
		MethodPattern: "^headers$",
		CarrierClass:  "HttpRequest",
		SourceType:    common.SourceHTTPHeader,
		Tags:          []string{"web", "async"},
	},
	{
		ID:            "actix_web_cookie",
		Framework:     "actix-web",
		Language:      "rust",
		Name:          "HttpRequest::cookie()",
		Description:   "Actix-web cookie access",
		MethodPattern: "^cookie$",
		CarrierClass:  "HttpRequest",
		SourceType:    common.SourceHTTPCookie,
		Tags:          []string{"web", "async"},
	},
	{
		ID:            "actix_web_payload",
		Framework:     "actix-web",
		Language:      "rust",
		Name:          "Payload",
		Description:   "Actix-web streaming payload",
		ClassPattern:  "^Payload$",
		SourceType:    common.SourceHTTPBody,
		Tags:          []string{"web", "async", "streaming"},
	},
}

// Rocket patterns (type-safe, ergonomic web framework)
var rocketPatterns = []*common.FrameworkPattern{
	{
		ID:            "rocket_from_form",
		Framework:     "rocket",
		Language:      "rust",
		Name:          "#[derive(FromForm)]",
		Description:   "Rocket form data derive macro",
		MethodPattern: "^FromForm$",
		SourceType:    common.SourceHTTPPost,
		Tags:          []string{"web", "type-safe"},
	},
	{
		ID:            "rocket_from_request",
		Framework:     "rocket",
		Language:      "rust",
		Name:          "#[derive(FromRequest)]",
		Description:   "Rocket request guard derive macro",
		MethodPattern: "^FromRequest$",
		SourceType:    common.SourceHTTPRequest,
		Tags:          []string{"web", "type-safe"},
	},
	{
		ID:           "rocket_form",
		Framework:    "rocket",
		Language:     "rust",
		Name:         "Form<T>",
		Description:  "Rocket form extractor",
		ClassPattern: "^Form$",
		SourceType:   common.SourceHTTPPost,
		Tags:         []string{"web", "type-safe"},
	},
	{
		ID:           "rocket_json",
		Framework:    "rocket",
		Language:     "rust",
		Name:         "Json<T>",
		Description:  "Rocket JSON body extractor",
		ClassPattern: "^Json$",
		SourceType:   common.SourceHTTPJSON,
		Tags:         []string{"web", "type-safe"},
	},
	{
		ID:           "rocket_data",
		Framework:    "rocket",
		Language:     "rust",
		Name:         "Data",
		Description:  "Rocket raw body data",
		ClassPattern: "^Data$",
		SourceType:   common.SourceHTTPBody,
		Tags:         []string{"web", "type-safe"},
	},
	{
		ID:           "rocket_cookies",
		Framework:    "rocket",
		Language:     "rust",
		Name:         "CookieJar",
		Description:  "Rocket cookie jar",
		ClassPattern: "^CookieJar$",
		SourceType:   common.SourceHTTPCookie,
		Tags:         []string{"web", "type-safe"},
	},
}

// Axum patterns (ergonomic and modular web framework built on Tokio)
var axumPatterns = []*common.FrameworkPattern{
	{
		ID:           "axum_query",
		Framework:    "axum",
		Language:     "rust",
		Name:         "Query<T>",
		Description:  "Axum query string extractor",
		ClassPattern: "^(extract::)?Query$",
		SourceType:   common.SourceHTTPGet,
		Tags:         []string{"web", "tokio", "modular"},
	},
	{
		ID:           "axum_form",
		Framework:    "axum",
		Language:     "rust",
		Name:         "Form<T>",
		Description:  "Axum form data extractor",
		ClassPattern: "^(extract::)?Form$",
		SourceType:   common.SourceHTTPPost,
		Tags:         []string{"web", "tokio", "modular"},
	},
	{
		ID:           "axum_path",
		Framework:    "axum",
		Language:     "rust",
		Name:         "Path<T>",
		Description:  "Axum path parameter extractor",
		ClassPattern: "^(extract::)?Path$",
		SourceType:   common.SourceHTTPPath,
		Tags:         []string{"web", "tokio", "modular"},
	},
	{
		ID:           "axum_json",
		Framework:    "axum",
		Language:     "rust",
		Name:         "Json<T>",
		Description:  "Axum JSON body extractor",
		ClassPattern: "^(extract::)?Json$",
		SourceType:   common.SourceHTTPJSON,
		Tags:         []string{"web", "tokio", "modular"},
	},
	{
		ID:           "axum_raw_body",
		Framework:    "axum",
		Language:     "rust",
		Name:         "RawBody",
		Description:  "Axum raw body extractor",
		ClassPattern: "^(extract::)?RawBody$",
		SourceType:   common.SourceHTTPBody,
		Tags:         []string{"web", "tokio"},
	},
	{
		ID:           "axum_typed_header",
		Framework:    "axum",
		Language:     "rust",
		Name:         "TypedHeader<T>",
		Description:  "Axum typed header extractor",
		ClassPattern: "^TypedHeader$",
		SourceType:   common.SourceHTTPHeader,
		Tags:         []string{"web", "tokio"},
	},
	{
		ID:           "axum_extension",
		Framework:    "axum",
		Language:     "rust",
		Name:         "Extension<T>",
		Description:  "Axum extension extractor (shared state)",
		ClassPattern: "^Extension$",
		SourceType:   common.SourceUserInput,
		Tags:         []string{"web", "tokio", "state"},
	},
}

// Warp patterns (composable, lightweight web framework)
var warpPatterns = []*common.FrameworkPattern{
	{
		ID:            "warp_query",
		Framework:     "warp",
		Language:      "rust",
		Name:          "warp::query()",
		Description:   "Warp query filter",
		MethodPattern: "^query$",
		SourceType:    common.SourceHTTPGet,
		Tags:          []string{"web", "filters", "composable"},
	},
	{
		ID:            "warp_body_form",
		Framework:     "warp",
		Language:      "rust",
		Name:          "warp::body::form()",
		Description:   "Warp form body filter",
		MethodPattern: "^form$",
		SourceType:    common.SourceHTTPPost,
		Tags:          []string{"web", "filters"},
	},
	{
		ID:            "warp_body_json",
		Framework:     "warp",
		Language:      "rust",
		Name:          "warp::body::json()",
		Description:   "Warp JSON body filter",
		MethodPattern: "^json$",
		SourceType:    common.SourceHTTPJSON,
		Tags:          []string{"web", "filters"},
	},
	{
		ID:            "warp_body_bytes",
		Framework:     "warp",
		Language:      "rust",
		Name:          "warp::body::bytes()",
		Description:   "Warp bytes body filter",
		MethodPattern: "^bytes$",
		SourceType:    common.SourceHTTPBody,
		Tags:          []string{"web", "filters"},
	},
	{
		ID:            "warp_path_param",
		Framework:     "warp",
		Language:      "rust",
		Name:          "warp::path::param()",
		Description:   "Warp path parameter filter",
		MethodPattern: "^param$",
		SourceType:    common.SourceHTTPPath,
		Tags:          []string{"web", "filters"},
	},
	{
		ID:            "warp_header",
		Framework:     "warp",
		Language:      "rust",
		Name:          "warp::header()",
		Description:   "Warp header filter",
		MethodPattern: "^header$",
		SourceType:    common.SourceHTTPHeader,
		Tags:          []string{"web", "filters"},
	},
	{
		ID:            "warp_cookie",
		Framework:     "warp",
		Language:      "rust",
		Name:          "warp::cookie()",
		Description:   "Warp cookie filter",
		MethodPattern: "^cookie$",
		SourceType:    common.SourceHTTPCookie,
		Tags:          []string{"web", "filters"},
	},
}

// Tide patterns (minimal and pragmatic async web framework)
var tidePatterns = []*common.FrameworkPattern{
	{
		ID:            "tide_request_body_string",
		Framework:     "tide",
		Language:      "rust",
		Name:          "Request::body_string()",
		Description:   "Tide request body as string",
		MethodPattern: "^body_string$",
		CarrierClass:  "Request",
		SourceType:    common.SourceHTTPBody,
		Tags:          []string{"web", "async", "minimal"},
	},
	{
		ID:            "tide_request_body_json",
		Framework:     "tide",
		Language:      "rust",
		Name:          "Request::body_json()",
		Description:   "Tide request body as JSON",
		MethodPattern: "^body_json$",
		CarrierClass:  "Request",
		SourceType:    common.SourceHTTPJSON,
		Tags:          []string{"web", "async", "minimal"},
	},
	{
		ID:            "tide_request_body_form",
		Framework:     "tide",
		Language:      "rust",
		Name:          "Request::body_form()",
		Description:   "Tide request body as form",
		MethodPattern: "^body_form$",
		CarrierClass:  "Request",
		SourceType:    common.SourceHTTPPost,
		Tags:          []string{"web", "async", "minimal"},
	},
	{
		ID:            "tide_request_query",
		Framework:     "tide",
		Language:      "rust",
		Name:          "Request::query()",
		Description:   "Tide query string",
		MethodPattern: "^query$",
		CarrierClass:  "Request",
		SourceType:    common.SourceHTTPGet,
		Tags:          []string{"web", "async", "minimal"},
	},
	{
		ID:            "tide_request_param",
		Framework:     "tide",
		Language:      "rust",
		Name:          "Request::param()",
		Description:   "Tide path parameter",
		MethodPattern: "^param$",
		CarrierClass:  "Request",
		SourceType:    common.SourceHTTPPath,
		Tags:          []string{"web", "async", "minimal"},
	},
}

// CLI parsing patterns (clap, structopt)
var cliPatterns = []*common.FrameworkPattern{
	{
		ID:            "clap_parser",
		Framework:     "clap",
		Language:      "rust",
		Name:          "#[derive(Parser)]",
		Description:   "Clap command line parser derive",
		MethodPattern: "^Parser$",
		SourceType:    common.SourceCLIArg,
		Tags:          []string{"cli", "parsing"},
	},
	{
		ID:            "clap_args",
		Framework:     "clap",
		Language:      "rust",
		Name:          "#[derive(Args)]",
		Description:   "Clap arguments derive",
		MethodPattern: "^Args$",
		SourceType:    common.SourceCLIArg,
		Tags:          []string{"cli", "parsing"},
	},
	{
		ID:            "clap_value_of",
		Framework:     "clap",
		Language:      "rust",
		Name:          "ArgMatches::value_of()",
		Description:   "Clap get argument value",
		MethodPattern: "^value_of$",
		CarrierClass:  "ArgMatches",
		SourceType:    common.SourceCLIArg,
		Tags:          []string{"cli", "parsing"},
	},
	{
		ID:            "clap_get_one",
		Framework:     "clap",
		Language:      "rust",
		Name:          "ArgMatches::get_one()",
		Description:   "Clap get single argument value",
		MethodPattern: "^get_one$",
		CarrierClass:  "ArgMatches",
		SourceType:    common.SourceCLIArg,
		Tags:          []string{"cli", "parsing"},
	},
	{
		ID:            "structopt_from_args",
		Framework:     "structopt",
		Language:      "rust",
		Name:          "StructOpt::from_args()",
		Description:   "StructOpt parse CLI arguments",
		MethodPattern: "^from_args$",
		SourceType:    common.SourceCLIArg,
		Tags:          []string{"cli", "parsing", "legacy"},
	},
}

func init() {
	Registry.RegisterAll(actixPatterns)
	Registry.RegisterAll(rocketPatterns)
	Registry.RegisterAll(axumPatterns)
	Registry.RegisterAll(warpPatterns)
	Registry.RegisterAll(tidePatterns)
	Registry.RegisterAll(cliPatterns)

	// Register framework detectors
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "actix-web",
		Indicators: []string{"Cargo.toml", "actix-web"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "rocket",
		Indicators: []string{"Cargo.toml", "rocket"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "axum",
		Indicators: []string{"Cargo.toml", "axum"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "warp",
		Indicators: []string{"Cargo.toml", "warp"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "tide",
		Indicators: []string{"Cargo.toml", "tide"},
	})
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "clap",
		Indicators: []string{"Cargo.toml", "clap"},
	})
}
