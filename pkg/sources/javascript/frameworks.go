// Package javascript - frameworks.go provides JavaScript framework pattern registry and universal patterns
// All JavaScript framework patterns should be registered here
package javascript

import (
	"regexp"

	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// Registry is the global JavaScript framework pattern registry
var Registry = common.NewFrameworkPatternRegistry("javascript")

// Universal patterns for detecting input across ANY JavaScript framework
var (
	// InputMethodPattern matches method names that ALWAYS indicate user input
	// e.g., get, json, text, param, params, query, body, headers, cookies
	InputMethodPattern = regexp.MustCompile(`(?i)^(get|json|text|param|params|query|body|headers?|cookies?|all)$`)

	// InputPropertyPattern matches property names that typically hold user input
	// e.g., body, query, params, headers, cookies, value, search, hash
	InputPropertyPattern = regexp.MustCompile(`(?i)^(body|query|params?|headers?|cookies?|value|search|hash|href|response(Text|XML)?)$`)

	// InputObjectPattern matches object/variable names that suggest an input carrier
	// e.g., req, request, ctx, context, event
	InputObjectPattern = regexp.MustCompile(`(?i)^(req|request|ctx|context|event|xhr|params|searchParams)$`)

	// DOMSourcePattern matches DOM properties that are user-controllable
	DOMSourcePattern = regexp.MustCompile(`(?i)(location\.(search|hash|href)|document\.(cookie|URL|referrer)|\.value\b)`)

	// NetworkResponsePattern matches network response properties
	NetworkResponsePattern = regexp.MustCompile(`(?i)(response(Text|XML)?|\.json\(\)|\.text\(\))`)
)

// InputPropertyPatterns contains universal property access patterns
// These match .property access on input objects
var InputPropertyPatterns = []string{
	".body",      // Express, Koa, Fastify
	".query",     // Express, Koa
	".params",    // Express, Koa
	".headers",   // Express, generic
	".cookies",   // Express
	".value",     // DOM form inputs
	".search",    // location.search
	".hash",      // location.hash
	".href",      // location.href
	".cookie",    // document.cookie
	".referrer",  // document.referrer
	".response",  // XHR
	".responseText", // XHR
	".responseXML",  // XHR
}

// InputMethodPatterns contains universal method call patterns
// These match .method() calls that return user input
var InputMethodPatterns = []string{
	// URLSearchParams
	".get(",
	".getAll(",
	// Fetch API response
	".json(",
	".text(",
	".formData(",
	".blob(",
	".arrayBuffer(",
	// FormData
	".get(",
	".getAll(",
	// Node.js fs
	".readFile(",
	".readFileSync(",
	".read(",
	".readSync(",
	// readline
	".question(",
}

// IsInputPropertyAccess checks if an expression matches an input property pattern
func IsInputPropertyAccess(expr string) bool {
	for _, pattern := range InputPropertyPatterns {
		if contains(expr, pattern) {
			return true
		}
	}
	return false
}

// IsInputMethodCall checks if an expression matches an input method pattern
func IsInputMethodCall(expr string) bool {
	for _, pattern := range InputMethodPatterns {
		if contains(expr, pattern) {
			return true
		}
	}
	return false
}

// IsInputMethod checks if a method name always indicates user input
func IsInputMethod(methodName string) bool {
	return InputMethodPattern.MatchString(methodName)
}

// IsInputProperty checks if a property name typically holds user input
func IsInputProperty(propertyName string) bool {
	return InputPropertyPattern.MatchString(propertyName)
}

// IsInputObject checks if a variable/object name suggests an input carrier
func IsInputObject(objectName string) bool {
	return InputObjectPattern.MatchString(objectName)
}

// IsDOMSource checks if an expression accesses a DOM source
func IsDOMSource(expr string) bool {
	return DOMSourcePattern.MatchString(expr)
}

// IsNetworkResponse checks if an expression accesses a network response
func IsNetworkResponse(expr string) bool {
	return NetworkResponsePattern.MatchString(expr)
}

// GetInputConfidence returns confidence score based on method context
func GetInputConfidence(methodName string, objectName string) float64 {
	// High confidence for explicit input methods on request objects
	if InputMethodPattern.MatchString(methodName) && InputObjectPattern.MatchString(objectName) {
		return 0.95
	}
	// High confidence for DOM sources
	if DOMSourcePattern.MatchString(methodName) || DOMSourcePattern.MatchString(objectName) {
		return 0.9
	}
	// Medium confidence for network responses
	if NetworkResponsePattern.MatchString(methodName) {
		return 0.8
	}
	return 0.5
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

// helper functions
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// DOM/Browser patterns (not framework-specific)
var domPatterns = []*common.FrameworkPattern{
	{
		ID:              "dom_location_search",
		Framework:       "browser",
		Language:        "javascript",
		Name:            "location.search",
		Description:     "URL query string from browser location",
		PropertyPattern: "^search$",
		SourceType:      common.SourceHTTPGet,
		PopulatedFrom:   []string{"URL query string"},
		Confidence:      0.95,
		Tags:            []string{"browser", "dom"},
	},
	{
		ID:              "dom_location_hash",
		Framework:       "browser",
		Language:        "javascript",
		Name:            "location.hash",
		Description:     "URL fragment/hash from browser location",
		PropertyPattern: "^hash$",
		SourceType:      common.SourceUserInput,
		PopulatedFrom:   []string{"URL fragment"},
		Confidence:      0.95,
		Tags:            []string{"browser", "dom"},
	},
	{
		ID:              "dom_location_href",
		Framework:       "browser",
		Language:        "javascript",
		Name:            "location.href",
		Description:     "Full URL from browser location",
		PropertyPattern: "^href$",
		SourceType:      common.SourceUserInput,
		PopulatedFrom:   []string{"URL"},
		Confidence:      0.9,
		Tags:            []string{"browser", "dom"},
	},
	{
		ID:              "dom_document_cookie",
		Framework:       "browser",
		Language:        "javascript",
		Name:            "document.cookie",
		Description:     "Browser cookies from document",
		PropertyPattern: "^cookie$",
		SourceType:      common.SourceHTTPCookie,
		PopulatedFrom:   []string{"HTTP Cookie header"},
		Confidence:      0.95,
		Tags:            []string{"browser", "dom"},
	},
	{
		ID:              "dom_document_referrer",
		Framework:       "browser",
		Language:        "javascript",
		Name:            "document.referrer",
		Description:     "Referring URL from document",
		PropertyPattern: "^referrer$",
		SourceType:      common.SourceUserInput,
		PopulatedFrom:   []string{"HTTP Referer header"},
		Confidence:      0.85,
		Tags:            []string{"browser", "dom"},
	},
	{
		ID:              "dom_document_url",
		Framework:       "browser",
		Language:        "javascript",
		Name:            "document.URL",
		Description:     "Document URL",
		PropertyPattern: "^URL$",
		SourceType:      common.SourceUserInput,
		PopulatedFrom:   []string{"URL"},
		Confidence:      0.9,
		Tags:            []string{"browser", "dom"},
	},
	{
		ID:              "dom_element_value",
		Framework:       "browser",
		Language:        "javascript",
		Name:            "element.value",
		Description:     "Form input element value",
		PropertyPattern: "^value$",
		SourceType:      common.SourceUserInput,
		PopulatedFrom:   []string{"Form input"},
		Confidence:      0.95,
		Tags:            []string{"browser", "dom", "form"},
	},
}

// Fetch API patterns
var fetchPatterns = []*common.FrameworkPattern{
	{
		ID:            "fetch_json",
		Framework:     "fetch",
		Language:      "javascript",
		Name:          "fetch().json()",
		Description:   "Fetch API JSON response parsing",
		MethodPattern: "^json$",
		SourceType:    common.SourceNetwork,
		PopulatedFrom: []string{"HTTP response body"},
		Confidence:    0.8,
		Tags:          []string{"browser", "network"},
	},
	{
		ID:            "fetch_text",
		Framework:     "fetch",
		Language:      "javascript",
		Name:          "fetch().text()",
		Description:   "Fetch API text response",
		MethodPattern: "^text$",
		SourceType:    common.SourceNetwork,
		PopulatedFrom: []string{"HTTP response body"},
		Confidence:    0.8,
		Tags:          []string{"browser", "network"},
	},
	{
		ID:            "fetch_formdata",
		Framework:     "fetch",
		Language:      "javascript",
		Name:          "fetch().formData()",
		Description:   "Fetch API FormData response",
		MethodPattern: "^formData$",
		SourceType:    common.SourceNetwork,
		PopulatedFrom: []string{"HTTP response body"},
		Confidence:    0.8,
		Tags:          []string{"browser", "network"},
	},
	{
		ID:            "fetch_blob",
		Framework:     "fetch",
		Language:      "javascript",
		Name:          "fetch().blob()",
		Description:   "Fetch API Blob response",
		MethodPattern: "^blob$",
		SourceType:    common.SourceNetwork,
		PopulatedFrom: []string{"HTTP response body"},
		Confidence:    0.8,
		Tags:          []string{"browser", "network"},
	},
	{
		ID:            "fetch_arraybuffer",
		Framework:     "fetch",
		Language:      "javascript",
		Name:          "fetch().arrayBuffer()",
		Description:   "Fetch API ArrayBuffer response",
		MethodPattern: "^arrayBuffer$",
		SourceType:    common.SourceNetwork,
		PopulatedFrom: []string{"HTTP response body"},
		Confidence:    0.8,
		Tags:          []string{"browser", "network"},
	},
}

// XMLHttpRequest patterns
var xhrPatterns = []*common.FrameworkPattern{
	{
		ID:              "xhr_response",
		Framework:       "xhr",
		Language:        "javascript",
		Name:            "XMLHttpRequest.response",
		Description:     "XMLHttpRequest response body",
		PropertyPattern: "^response$",
		SourceType:      common.SourceNetwork,
		PopulatedFrom:   []string{"HTTP response body"},
		Confidence:      0.8,
		Tags:            []string{"browser", "network"},
	},
	{
		ID:              "xhr_response_text",
		Framework:       "xhr",
		Language:        "javascript",
		Name:            "XMLHttpRequest.responseText",
		Description:     "XMLHttpRequest text response",
		PropertyPattern: "^responseText$",
		SourceType:      common.SourceNetwork,
		PopulatedFrom:   []string{"HTTP response body"},
		Confidence:      0.8,
		Tags:            []string{"browser", "network"},
	},
	{
		ID:              "xhr_response_xml",
		Framework:       "xhr",
		Language:        "javascript",
		Name:            "XMLHttpRequest.responseXML",
		Description:     "XMLHttpRequest XML response",
		PropertyPattern: "^responseXML$",
		SourceType:      common.SourceNetwork,
		PopulatedFrom:   []string{"HTTP response body"},
		Confidence:      0.8,
		Tags:            []string{"browser", "network"},
	},
}

// URLSearchParams patterns
var urlSearchParamsPatterns = []*common.FrameworkPattern{
	{
		ID:            "urlsearchparams_get",
		Framework:     "web-api",
		Language:      "javascript",
		Name:          "URLSearchParams.get()",
		Description:   "URLSearchParams query parameter access",
		MethodPattern: "^get$",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"URL query string"},
		Confidence:    0.9,
		Tags:          []string{"browser", "url"},
	},
	{
		ID:            "urlsearchparams_getall",
		Framework:     "web-api",
		Language:      "javascript",
		Name:          "URLSearchParams.getAll()",
		Description:   "URLSearchParams query parameter access (all values)",
		MethodPattern: "^getAll$",
		SourceType:    common.SourceHTTPGet,
		PopulatedFrom: []string{"URL query string"},
		Confidence:    0.9,
		Tags:          []string{"browser", "url"},
	},
}

// FormData patterns
var formDataPatterns = []*common.FrameworkPattern{
	{
		ID:            "formdata_get",
		Framework:     "web-api",
		Language:      "javascript",
		Name:          "FormData.get()",
		Description:   "FormData value access",
		MethodPattern: "^get$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"Form input"},
		Confidence:    0.9,
		Tags:          []string{"browser", "form"},
	},
	{
		ID:            "formdata_getall",
		Framework:     "web-api",
		Language:      "javascript",
		Name:          "FormData.getAll()",
		Description:   "FormData value access (all values)",
		MethodPattern: "^getAll$",
		SourceType:    common.SourceUserInput,
		PopulatedFrom: []string{"Form input"},
		Confidence:    0.9,
		Tags:          []string{"browser", "form"},
	},
}

// Node.js patterns (not framework-specific)
var nodePatterns = []*common.FrameworkPattern{
	{
		ID:              "node_process_argv",
		Framework:       "node",
		Language:        "javascript",
		Name:            "process.argv",
		Description:     "Node.js command line arguments",
		PropertyPattern: "^argv$",
		SourceType:      common.SourceCLIArg,
		PopulatedFrom:   []string{"Command line"},
		Confidence:      0.95,
		Tags:            []string{"node", "cli"},
	},
	{
		ID:              "node_process_env",
		Framework:       "node",
		Language:        "javascript",
		Name:            "process.env",
		Description:     "Node.js environment variables",
		PropertyPattern: "^env$",
		SourceType:      common.SourceEnvVar,
		PopulatedFrom:   []string{"Environment"},
		Confidence:      0.95,
		Tags:            []string{"node", "environment"},
	},
	{
		ID:            "node_fs_readfile",
		Framework:     "node",
		Language:      "javascript",
		Name:          "fs.readFile()",
		Description:   "Node.js file read",
		MethodPattern: "^readFile(Sync)?$",
		SourceType:    common.SourceFile,
		PopulatedFrom: []string{"File system"},
		Confidence:    0.85,
		Tags:          []string{"node", "file"},
	},
	{
		ID:            "node_readline_question",
		Framework:     "node",
		Language:      "javascript",
		Name:          "readline.question()",
		Description:   "Node.js readline input",
		MethodPattern: "^question$",
		SourceType:    common.SourceStdin,
		PopulatedFrom: []string{"Standard input"},
		Confidence:    0.9,
		Tags:          []string{"node", "stdin"},
	},
}

func init() {
	// Register all universal patterns
	Registry.RegisterAll(domPatterns)
	Registry.RegisterAll(fetchPatterns)
	Registry.RegisterAll(xhrPatterns)
	Registry.RegisterAll(urlSearchParamsPatterns)
	Registry.RegisterAll(formDataPatterns)
	Registry.RegisterAll(nodePatterns)

	// Register browser framework detector
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "browser",
		Indicators: []string{"index.html", "public/index.html", "src/index.html"},
	})

	// Register Node.js framework detector
	common.RegisterFrameworkDetector(&common.FrameworkDetector{
		Framework:  "node",
		Indicators: []string{"package.json"},
	})
}
