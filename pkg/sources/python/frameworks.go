// Package python - frameworks.go provides Python framework pattern registry
// All Python framework patterns should be registered here
package python

import (
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// Registry is the global Python framework pattern registry
var Registry = common.NewFrameworkPatternRegistry("python")

// Flask patterns
var flaskPatterns = []*common.FrameworkPattern{
	{
		ID:              "flask_request_args",
		Framework:       "flask",
		Language:        "python",
		Name:            "Flask request.args",
		Description:     "Flask request object containing query string parameters",
		ClassPattern:    "^Request$",
		PropertyPattern: "^args$",
		AccessPattern:   "dict",
		SourceType:      common.SourceHTTPGet,
		CarrierClass:    "Request",
		CarrierProperty: "args",
		PopulatedFrom:   []string{"query string"},
		Tags:            []string{"framework", "flask"},
	},
	{
		ID:              "flask_request_form",
		Framework:       "flask",
		Language:        "python",
		Name:            "Flask request.form",
		Description:     "Flask POST form data",
		ClassPattern:    "^Request$",
		PropertyPattern: "^form$",
		AccessPattern:   "dict",
		SourceType:      common.SourceHTTPPost,
		CarrierClass:    "Request",
		CarrierProperty: "form",
		PopulatedFrom:   []string{"form data"},
		Tags:            []string{"framework", "flask"},
	},
	{
		ID:              "flask_request_data",
		Framework:       "flask",
		Language:        "python",
		Name:            "Flask request.data",
		Description:     "Flask raw request body",
		ClassPattern:    "^Request$",
		PropertyPattern: "^data$",
		SourceType:      common.SourceHTTPBody,
		CarrierClass:    "Request",
		CarrierProperty: "data",
		PopulatedFrom:   []string{"HTTP body"},
		Tags:            []string{"framework", "flask"},
	},
	{
		ID:              "flask_request_json",
		Framework:       "flask",
		Language:        "python",
		Name:            "Flask request.json",
		Description:     "Flask JSON request body",
		ClassPattern:    "^Request$",
		PropertyPattern: "^json$",
		SourceType:      common.SourceHTTPJSON,
		CarrierClass:    "Request",
		CarrierProperty: "json",
		PopulatedFrom:   []string{"JSON body"},
		Tags:            []string{"framework", "flask"},
	},
	{
		ID:              "flask_request_files",
		Framework:       "flask",
		Language:        "python",
		Name:            "Flask request.files",
		Description:     "Flask uploaded files",
		ClassPattern:    "^Request$",
		PropertyPattern: "^files$",
		AccessPattern:   "dict",
		SourceType:      common.SourceHTTPFile,
		CarrierClass:    "Request",
		CarrierProperty: "files",
		PopulatedFrom:   []string{"uploaded files"},
		Tags:            []string{"framework", "flask"},
	},
	{
		ID:              "flask_request_cookies",
		Framework:       "flask",
		Language:        "python",
		Name:            "Flask request.cookies",
		Description:     "Flask cookies",
		ClassPattern:    "^Request$",
		PropertyPattern: "^cookies$",
		AccessPattern:   "dict",
		SourceType:      common.SourceHTTPCookie,
		CarrierClass:    "Request",
		CarrierProperty: "cookies",
		PopulatedFrom:   []string{"cookies"},
		Tags:            []string{"framework", "flask"},
	},
	{
		ID:              "flask_request_headers",
		Framework:       "flask",
		Language:        "python",
		Name:            "Flask request.headers",
		Description:     "Flask HTTP headers",
		ClassPattern:    "^Request$",
		PropertyPattern: "^headers$",
		AccessPattern:   "dict",
		SourceType:      common.SourceHTTPHeader,
		CarrierClass:    "Request",
		CarrierProperty: "headers",
		PopulatedFrom:   []string{"HTTP headers"},
		Tags:            []string{"framework", "flask"},
	},
	{
		ID:              "flask_request_values",
		Framework:       "flask",
		Language:        "python",
		Name:            "Flask request.values",
		Description:     "Flask combined GET/POST parameters",
		ClassPattern:    "^Request$",
		PropertyPattern: "^values$",
		AccessPattern:   "dict",
		SourceType:      common.SourceUserInput,
		CarrierClass:    "Request",
		CarrierProperty: "values",
		PopulatedFrom:   []string{"query string", "form data"},
		Tags:            []string{"framework", "flask"},
	},
}

// Django patterns
var djangoPatterns = []*common.FrameworkPattern{
	{
		ID:              "django_request_get",
		Framework:       "django",
		Language:        "python",
		Name:            "Django request.GET",
		Description:     "Django GET parameters",
		ClassPattern:    "^(Http)?Request$",
		PropertyPattern: "^GET$",
		AccessPattern:   "dict",
		SourceType:      common.SourceHTTPGet,
		CarrierClass:    "HttpRequest",
		CarrierProperty: "GET",
		PopulatedFrom:   []string{"query string"},
		Tags:            []string{"framework", "django"},
	},
	{
		ID:              "django_request_post",
		Framework:       "django",
		Language:        "python",
		Name:            "Django request.POST",
		Description:     "Django POST parameters",
		ClassPattern:    "^(Http)?Request$",
		PropertyPattern: "^POST$",
		AccessPattern:   "dict",
		SourceType:      common.SourceHTTPPost,
		CarrierClass:    "HttpRequest",
		CarrierProperty: "POST",
		PopulatedFrom:   []string{"form data"},
		Tags:            []string{"framework", "django"},
	},
	{
		ID:              "django_request_body",
		Framework:       "django",
		Language:        "python",
		Name:            "Django request.body",
		Description:     "Django raw request body",
		ClassPattern:    "^(Http)?Request$",
		PropertyPattern: "^body$",
		SourceType:      common.SourceHTTPBody,
		CarrierClass:    "HttpRequest",
		CarrierProperty: "body",
		PopulatedFrom:   []string{"HTTP body"},
		Tags:            []string{"framework", "django"},
	},
	{
		ID:              "django_request_files",
		Framework:       "django",
		Language:        "python",
		Name:            "Django request.FILES",
		Description:     "Django uploaded files",
		ClassPattern:    "^(Http)?Request$",
		PropertyPattern: "^FILES$",
		AccessPattern:   "dict",
		SourceType:      common.SourceHTTPFile,
		CarrierClass:    "HttpRequest",
		CarrierProperty: "FILES",
		PopulatedFrom:   []string{"uploaded files"},
		Tags:            []string{"framework", "django"},
	},
	{
		ID:              "django_request_cookies",
		Framework:       "django",
		Language:        "python",
		Name:            "Django request.COOKIES",
		Description:     "Django cookies",
		ClassPattern:    "^(Http)?Request$",
		PropertyPattern: "^COOKIES$",
		AccessPattern:   "dict",
		SourceType:      common.SourceHTTPCookie,
		CarrierClass:    "HttpRequest",
		CarrierProperty: "COOKIES",
		PopulatedFrom:   []string{"cookies"},
		Tags:            []string{"framework", "django"},
	},
	{
		ID:              "django_request_meta",
		Framework:       "django",
		Language:        "python",
		Name:            "Django request.META",
		Description:     "Django request metadata (includes headers)",
		ClassPattern:    "^(Http)?Request$",
		PropertyPattern: "^META$",
		AccessPattern:   "dict",
		SourceType:      common.SourceHTTPHeader,
		CarrierClass:    "HttpRequest",
		CarrierProperty: "META",
		PopulatedFrom:   []string{"HTTP headers", "server variables"},
		Tags:            []string{"framework", "django"},
	},
	{
		ID:              "django_request_headers",
		Framework:       "django",
		Language:        "python",
		Name:            "Django request.headers",
		Description:     "Django HTTP headers (Django 2.2+)",
		ClassPattern:    "^(Http)?Request$",
		PropertyPattern: "^headers$",
		AccessPattern:   "dict",
		SourceType:      common.SourceHTTPHeader,
		CarrierClass:    "HttpRequest",
		CarrierProperty: "headers",
		PopulatedFrom:   []string{"HTTP headers"},
		Tags:            []string{"framework", "django"},
	},
}

// FastAPI patterns
var fastapiPatterns = []*common.FrameworkPattern{
	{
		ID:            "fastapi_request",
		Framework:     "fastapi",
		Language:      "python",
		Name:          "FastAPI Request",
		Description:   "FastAPI request object",
		ClassPattern:  "^Request$",
		SourceType:    common.SourceHTTPBody,
		CarrierClass:  "Request",
		PopulatedFrom: []string{"HTTP request"},
		Tags:          []string{"framework", "fastapi"},
	},
	{
		ID:            "fastapi_query",
		Framework:     "fastapi",
		Language:      "python",
		Name:          "FastAPI Query()",
		Description:   "FastAPI query parameter decorator",
		MethodPattern: "^Query$",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"query string"},
		Tags:          []string{"framework", "fastapi", "decorator"},
	},
	{
		ID:            "fastapi_path",
		Framework:     "fastapi",
		Language:      "python",
		Name:          "FastAPI Path()",
		Description:   "FastAPI path parameter decorator",
		MethodPattern: "^Path$",
		SourceType:    common.SourceHTTPPath,
		PopulatedFrom: []string{"URL path"},
		Tags:          []string{"framework", "fastapi", "decorator"},
	},
	{
		ID:            "fastapi_body",
		Framework:     "fastapi",
		Language:      "python",
		Name:          "FastAPI Body()",
		Description:   "FastAPI body parameter decorator",
		MethodPattern: "^Body$",
		SourceType:    common.SourceHTTPBody,
		PopulatedFrom: []string{"HTTP body"},
		Tags:          []string{"framework", "fastapi", "decorator"},
	},
	{
		ID:            "fastapi_header",
		Framework:     "fastapi",
		Language:      "python",
		Name:          "FastAPI Header()",
		Description:   "FastAPI header parameter decorator",
		MethodPattern: "^Header$",
		SourceType:    common.SourceHTTPHeader,
		PopulatedFrom: []string{"HTTP headers"},
		Tags:          []string{"framework", "fastapi", "decorator"},
	},
	{
		ID:            "fastapi_cookie",
		Framework:     "fastapi",
		Language:      "python",
		Name:          "FastAPI Cookie()",
		Description:   "FastAPI cookie parameter decorator",
		MethodPattern: "^Cookie$",
		SourceType:    common.SourceHTTPCookie,
		PopulatedFrom: []string{"cookies"},
		Tags:          []string{"framework", "fastapi", "decorator"},
	},
	{
		ID:            "fastapi_form",
		Framework:     "fastapi",
		Language:      "python",
		Name:          "FastAPI Form()",
		Description:   "FastAPI form parameter decorator",
		MethodPattern: "^Form$",
		SourceType:    common.SourceHTTPPost,
		PopulatedFrom: []string{"form data"},
		Tags:          []string{"framework", "fastapi", "decorator"},
	},
	{
		ID:            "fastapi_file",
		Framework:     "fastapi",
		Language:      "python",
		Name:          "FastAPI File()",
		Description:   "FastAPI file upload decorator",
		MethodPattern: "^File$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"uploaded files"},
		Tags:          []string{"framework", "fastapi", "decorator"},
	},
	{
		ID:            "fastapi_uploadfile",
		Framework:     "fastapi",
		Language:      "python",
		Name:          "FastAPI UploadFile",
		Description:   "FastAPI uploaded file type",
		ClassPattern:  "^UploadFile$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"uploaded files"},
		Tags:          []string{"framework", "fastapi"},
	},
}

// argparse patterns
var argparsePatterns = []*common.FrameworkPattern{
	{
		ID:            "argparse_parse_args",
		Framework:     "argparse",
		Language:      "python",
		Name:          "argparse parse_args()",
		Description:   "Command line arguments parsed by argparse",
		MethodPattern: "^parse_args$",
		SourceType:    common.SourceCLIArg,
		PopulatedFrom: []string{"sys.argv"},
		Tags:          []string{"cli", "argparse"},
	},
	{
		ID:            "argparse_parse_known_args",
		Framework:     "argparse",
		Language:      "python",
		Name:          "argparse parse_known_args()",
		Description:   "Partial command line argument parsing",
		MethodPattern: "^parse_known_args$",
		SourceType:    common.SourceCLIArg,
		PopulatedFrom: []string{"sys.argv"},
		Tags:          []string{"cli", "argparse"},
	},
}

// Click patterns (bonus - common CLI framework)
var clickPatterns = []*common.FrameworkPattern{
	{
		ID:            "click_argument",
		Framework:     "click",
		Language:      "python",
		Name:          "Click @argument",
		Description:   "Click CLI argument decorator",
		MethodPattern: "^argument$",
		SourceType:    common.SourceCLIArg,
		PopulatedFrom: []string{"command line"},
		Tags:          []string{"cli", "click", "decorator"},
	},
	{
		ID:            "click_option",
		Framework:     "click",
		Language:      "python",
		Name:          "Click @option",
		Description:   "Click CLI option decorator",
		MethodPattern: "^option$",
		SourceType:    common.SourceCLIArg,
		PopulatedFrom: []string{"command line"},
		Tags:          []string{"cli", "click", "decorator"},
	},
}

// GetAllPatterns returns all registered framework patterns
func GetAllPatterns() []*common.FrameworkPattern {
	return Registry.GetAll()
}

// GetPatternsByFramework returns patterns for a specific framework
func GetPatternsByFramework(framework string) []*common.FrameworkPattern {
	return Registry.GetByFramework(framework)
}

// GetPatternByID returns a pattern by its ID
func GetPatternByID(id string) *common.FrameworkPattern {
	return Registry.GetByID(id)
}

func init() {
	// Register all patterns
	Registry.RegisterAll(flaskPatterns)
	Registry.RegisterAll(djangoPatterns)
	Registry.RegisterAll(fastapiPatterns)
	Registry.RegisterAll(argparsePatterns)
	Registry.RegisterAll(clickPatterns)

	// Register framework detectors
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "flask",
		Indicators: []string{"app.py", "flask", "from flask import"},
	})

	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "django",
		Indicators: []string{"manage.py", "settings.py", "urls.py", "wsgi.py"},
	})

	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "fastapi",
		Indicators: []string{"main.py", "from fastapi import"},
	})
}
