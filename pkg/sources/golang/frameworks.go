// Package golang - frameworks.go provides Go framework pattern registry
// All Go framework patterns should be registered here
package golang

import (
	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// Registry is the global Go framework pattern registry
var Registry = common.NewFrameworkPatternRegistry("go")

// Gin framework patterns
var ginPatterns = []*common.FrameworkPattern{
	{
		ID:            "gin_query",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.Query()",
		Description:   "Gin query parameter",
		MethodPattern: "^Query$",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"query string"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_default_query",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.DefaultQuery()",
		Description:   "Gin query parameter with default value",
		MethodPattern: "^DefaultQuery$",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"query string"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_query_array",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.QueryArray()",
		Description:   "Gin query parameter array",
		MethodPattern: "^QueryArray$",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"query string"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_param",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.Param()",
		Description:   "Gin URL path parameter",
		MethodPattern: "^Param$",
		SourceType:    common.SourceHTTPPath,
		PopulatedFrom: []string{"URL path"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_post_form",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.PostForm()",
		Description:   "Gin POST form value",
		MethodPattern: "^PostForm$",
		SourceType:    common.SourceHTTPPost,
		PopulatedFrom: []string{"form data"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_default_post_form",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.DefaultPostForm()",
		Description:   "Gin POST form value with default",
		MethodPattern: "^DefaultPostForm$",
		SourceType:    common.SourceHTTPPost,
		PopulatedFrom: []string{"form data"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_post_form_array",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.PostFormArray()",
		Description:   "Gin POST form array",
		MethodPattern: "^PostFormArray$",
		SourceType:    common.SourceHTTPPost,
		PopulatedFrom: []string{"form data"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_form_file",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.FormFile()",
		Description:   "Gin uploaded file",
		MethodPattern: "^FormFile$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"uploaded files"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_multipart_form",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.MultipartForm()",
		Description:   "Gin multipart form data",
		MethodPattern: "^MultipartForm$",
		SourceType:    common.SourceHTTPPost,
		PopulatedFrom: []string{"multipart form"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_get_header",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.GetHeader()",
		Description:   "Gin HTTP header value",
		MethodPattern: "^GetHeader$",
		SourceType:    common.SourceHTTPHeader,
		PopulatedFrom: []string{"HTTP headers"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_cookie",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.Cookie()",
		Description:   "Gin cookie value",
		MethodPattern: "^Cookie$",
		SourceType:    common.SourceHTTPCookie,
		PopulatedFrom: []string{"cookies"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_bind_json",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.BindJSON()",
		Description:   "Gin JSON binding",
		MethodPattern: "^BindJSON$",
		SourceType:    common.SourceHTTPJSON,
		PopulatedFrom: []string{"JSON body"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_should_bind_json",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.ShouldBindJSON()",
		Description:   "Gin JSON binding (non-fatal)",
		MethodPattern: "^ShouldBindJSON$",
		SourceType:    common.SourceHTTPJSON,
		PopulatedFrom: []string{"JSON body"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_bind",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.Bind()",
		Description:   "Gin auto binding",
		MethodPattern: "^Bind$",
		SourceType:    common.SourceHTTPBody,
		PopulatedFrom: []string{"HTTP body"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_should_bind",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.ShouldBind()",
		Description:   "Gin auto binding (non-fatal)",
		MethodPattern: "^ShouldBind$",
		SourceType:    common.SourceHTTPBody,
		PopulatedFrom: []string{"HTTP body"},
		Tags:          []string{"framework", "gin"},
	},
	{
		ID:            "gin_get_raw_data",
		Framework:     "gin",
		Language:      "go",
		Name:          "Gin c.GetRawData()",
		Description:   "Gin raw request body",
		MethodPattern: "^GetRawData$",
		SourceType:    common.SourceHTTPBody,
		PopulatedFrom: []string{"HTTP body"},
		Tags:          []string{"framework", "gin"},
	},
}

// Echo framework patterns
var echoPatterns = []*common.FrameworkPattern{
	{
		ID:            "echo_query_param",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.QueryParam()",
		Description:   "Echo query parameter",
		MethodPattern: "^QueryParam$",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"query string"},
		Tags:          []string{"framework", "echo"},
	},
	{
		ID:            "echo_query_params",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.QueryParams()",
		Description:   "Echo all query parameters",
		MethodPattern: "^QueryParams$",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"query string"},
		Tags:          []string{"framework", "echo"},
	},
	{
		ID:            "echo_param",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.Param()",
		Description:   "Echo URL path parameter",
		MethodPattern: "^Param$",
		SourceType:    common.SourceHTTPPath,
		PopulatedFrom: []string{"URL path"},
		Tags:          []string{"framework", "echo"},
	},
	{
		ID:            "echo_form_value",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.FormValue()",
		Description:   "Echo form value",
		MethodPattern: "^FormValue$",
		SourceType:    common.SourceHTTPPost,
		PopulatedFrom: []string{"form data"},
		Tags:          []string{"framework", "echo"},
	},
	{
		ID:            "echo_form_params",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.FormParams()",
		Description:   "Echo all form parameters",
		MethodPattern: "^FormParams$",
		SourceType:    common.SourceHTTPPost,
		PopulatedFrom: []string{"form data"},
		Tags:          []string{"framework", "echo"},
	},
	{
		ID:            "echo_form_file",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.FormFile()",
		Description:   "Echo uploaded file",
		MethodPattern: "^FormFile$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"uploaded files"},
		Tags:          []string{"framework", "echo"},
	},
	{
		ID:            "echo_multipart_form",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.MultipartForm()",
		Description:   "Echo multipart form data",
		MethodPattern: "^MultipartForm$",
		SourceType:    common.SourceHTTPPost,
		PopulatedFrom: []string{"multipart form"},
		Tags:          []string{"framework", "echo"},
	},
	{
		ID:            "echo_request_header",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.Request().Header",
		Description:   "Echo HTTP header",
		PropertyPattern: "^Header$",
		SourceType:      common.SourceHTTPHeader,
		PopulatedFrom:   []string{"HTTP headers"},
		Tags:            []string{"framework", "echo"},
	},
	{
		ID:            "echo_cookie",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.Cookie()",
		Description:   "Echo cookie value",
		MethodPattern: "^Cookie$",
		SourceType:    common.SourceHTTPCookie,
		PopulatedFrom: []string{"cookies"},
		Tags:          []string{"framework", "echo"},
	},
	{
		ID:            "echo_cookies",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.Cookies()",
		Description:   "Echo all cookies",
		MethodPattern: "^Cookies$",
		SourceType:    common.SourceHTTPCookie,
		PopulatedFrom: []string{"cookies"},
		Tags:          []string{"framework", "echo"},
	},
	{
		ID:            "echo_bind",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.Bind()",
		Description:   "Echo auto binding",
		MethodPattern: "^Bind$",
		SourceType:    common.SourceHTTPBody,
		PopulatedFrom: []string{"HTTP body"},
		Tags:          []string{"framework", "echo"},
	},
	{
		ID:            "echo_request_body",
		Framework:     "echo",
		Language:      "go",
		Name:          "Echo c.Request().Body",
		Description:   "Echo raw request body",
		PropertyPattern: "^Body$",
		SourceType:      common.SourceHTTPBody,
		PopulatedFrom:   []string{"HTTP body"},
		Tags:            []string{"framework", "echo"},
	},
}

// Fiber framework patterns
var fiberPatterns = []*common.FrameworkPattern{
	{
		ID:            "fiber_query",
		Framework:     "fiber",
		Language:      "go",
		Name:          "Fiber c.Query()",
		Description:   "Fiber query parameter",
		MethodPattern: "^Query$",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"query string"},
		Tags:          []string{"framework", "fiber"},
	},
	{
		ID:            "fiber_params",
		Framework:     "fiber",
		Language:      "go",
		Name:          "Fiber c.Params()",
		Description:   "Fiber URL path parameter",
		MethodPattern: "^Params$",
		SourceType:    common.SourceHTTPPath,
		PopulatedFrom: []string{"URL path"},
		Tags:          []string{"framework", "fiber"},
	},
	{
		ID:            "fiber_form_value",
		Framework:     "fiber",
		Language:      "go",
		Name:          "Fiber c.FormValue()",
		Description:   "Fiber form value",
		MethodPattern: "^FormValue$",
		SourceType:    common.SourceHTTPPost,
		PopulatedFrom: []string{"form data"},
		Tags:          []string{"framework", "fiber"},
	},
	{
		ID:            "fiber_form_file",
		Framework:     "fiber",
		Language:      "go",
		Name:          "Fiber c.FormFile()",
		Description:   "Fiber uploaded file",
		MethodPattern: "^FormFile$",
		SourceType:    common.SourceHTTPFile,
		PopulatedFrom: []string{"uploaded files"},
		Tags:          []string{"framework", "fiber"},
	},
	{
		ID:            "fiber_body_parser",
		Framework:     "fiber",
		Language:      "go",
		Name:          "Fiber c.BodyParser()",
		Description:   "Fiber body parser",
		MethodPattern: "^BodyParser$",
		SourceType:    common.SourceHTTPBody,
		PopulatedFrom: []string{"HTTP body"},
		Tags:          []string{"framework", "fiber"},
	},
	{
		ID:            "fiber_body",
		Framework:     "fiber",
		Language:      "go",
		Name:          "Fiber c.Body()",
		Description:   "Fiber raw body",
		MethodPattern: "^Body$",
		SourceType:    common.SourceHTTPBody,
		PopulatedFrom: []string{"HTTP body"},
		Tags:          []string{"framework", "fiber"},
	},
	{
		ID:            "fiber_get",
		Framework:     "fiber",
		Language:      "go",
		Name:          "Fiber c.Get()",
		Description:   "Fiber HTTP header value",
		MethodPattern: "^Get$",
		SourceType:    common.SourceHTTPHeader,
		PopulatedFrom: []string{"HTTP headers"},
		Tags:          []string{"framework", "fiber"},
	},
	{
		ID:            "fiber_cookies",
		Framework:     "fiber",
		Language:      "go",
		Name:          "Fiber c.Cookies()",
		Description:   "Fiber cookie value",
		MethodPattern: "^Cookies$",
		SourceType:    common.SourceHTTPCookie,
		PopulatedFrom: []string{"cookies"},
		Tags:          []string{"framework", "fiber"},
	},
}

// Chi framework patterns
var chiPatterns = []*common.FrameworkPattern{
	{
		ID:            "chi_url_param",
		Framework:     "chi",
		Language:      "go",
		Name:          "Chi chi.URLParam()",
		Description:   "Chi URL path parameter",
		MethodPattern: "^URLParam$",
		SourceType:    common.SourceHTTPPath,
		PopulatedFrom: []string{"URL path"},
		Tags:          []string{"framework", "chi"},
	},
	{
		ID:            "chi_url_param_from_ctx",
		Framework:     "chi",
		Language:      "go",
		Name:          "Chi chi.URLParamFromCtx()",
		Description:   "Chi URL path parameter from context",
		MethodPattern: "^URLParamFromCtx$",
		SourceType:    common.SourceHTTPPath,
		PopulatedFrom: []string{"URL path"},
		Tags:          []string{"framework", "chi"},
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
	Registry.RegisterAll(ginPatterns)
	Registry.RegisterAll(echoPatterns)
	Registry.RegisterAll(fiberPatterns)
	Registry.RegisterAll(chiPatterns)

	// Register framework detectors
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "gin",
		Indicators: []string{"github.com/gin-gonic/gin"},
	})

	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "echo",
		Indicators: []string{"github.com/labstack/echo"},
	})

	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "fiber",
		Indicators: []string{"github.com/gofiber/fiber"},
	})

	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "chi",
		Indicators: []string{"github.com/go-chi/chi"},
	})
}
