package java

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// Matcher matches Java user input sources
type Matcher struct {
	*common.BaseMatcher
}

// NewMatcher creates a new Java source matcher
func NewMatcher() *Matcher {
	defs := []common.Definition{
		// Servlet API
		{
			Name:         "request.getParameter()",
			Pattern:      `\.getParameter\s*\(`,
			Language:     "java",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description:  "HTTP request parameter",
			NodeTypes:    []string{"method_invocation"},
			KeyExtractor: `\.getParameter\s*\(\s*"([^"]+)"`,
		},
		{
			Name:        "request.getParameterValues()",
			Pattern:     `\.getParameterValues\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "HTTP request parameter array",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getParameterMap()",
			Pattern:     `\.getParameterMap\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "All HTTP request parameters",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:         "request.getHeader()",
			Pattern:      `\.getHeader\s*\(`,
			Language:     "java",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "HTTP request header",
			NodeTypes:    []string{"method_invocation"},
			KeyExtractor: `\.getHeader\s*\(\s*"([^"]+)"`,
		},
		{
			Name:        "request.getHeaders()",
			Pattern:     `\.getHeaders\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "HTTP request headers",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getInputStream()",
			Pattern:     `\.getInputStream\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "HTTP request body stream",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getReader()",
			Pattern:     `\.getReader\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "HTTP request body reader",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getCookies()",
			Pattern:     `\.getCookies\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description: "HTTP cookies",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getQueryString()",
			Pattern:     `\.getQueryString\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "HTTP query string",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getRequestURI()",
			Pattern:     `\.getRequestURI\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "HTTP request URI",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getRequestURL()",
			Pattern:     `\.getRequestURL\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "HTTP request URL",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getPathInfo()",
			Pattern:     `\.getPathInfo\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "HTTP path info",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getPart()",
			Pattern:     `\.getPart\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "Multipart form part",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getParts()",
			Pattern:     `\.getParts\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "All multipart form parts",
			NodeTypes:   []string{"method_invocation"},
		},

		// Spring MVC annotations
		{
			Name:        "@RequestParam",
			Pattern:     `@RequestParam`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "Spring request parameter",
			NodeTypes:   []string{"annotation", "marker_annotation"},
		},
		{
			Name:        "@PathVariable",
			Pattern:     `@PathVariable`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "Spring path variable",
			NodeTypes:   []string{"annotation", "marker_annotation"},
		},
		{
			Name:        "@RequestBody",
			Pattern:     `@RequestBody`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Spring request body",
			NodeTypes:   []string{"annotation", "marker_annotation"},
		},
		{
			Name:        "@RequestHeader",
			Pattern:     `@RequestHeader`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "Spring request header",
			NodeTypes:   []string{"annotation", "marker_annotation"},
		},
		{
			Name:        "@CookieValue",
			Pattern:     `@CookieValue`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description: "Spring cookie value",
			NodeTypes:   []string{"annotation", "marker_annotation"},
		},

		// CLI
		{
			Name:        "args[]",
			Pattern:     `\bargs\s*\[`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"array_access"},
		},

		// Environment
		{
			Name:         "System.getenv()",
			Pattern:      `System\.getenv\s*\(`,
			Language:     "java",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "Environment variable",
			NodeTypes:    []string{"method_invocation"},
			KeyExtractor: `System\.getenv\s*\(\s*"([^"]+)"`,
		},
		{
			Name:         "System.getProperty()",
			Pattern:      `System\.getProperty\s*\(`,
			Language:     "java",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "System property",
			NodeTypes:    []string{"method_invocation"},
			KeyExtractor: `System\.getProperty\s*\(\s*"([^"]+)"`,
		},

		// Console input
		{
			Name:        "Scanner.next()",
			Pattern:     `\.next\s*\(\s*\)`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Scanner next token",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Scanner.nextLine()",
			Pattern:     `\.nextLine\s*\(\s*\)`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Scanner next line",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Scanner.nextInt()",
			Pattern:     `\.nextInt\s*\(\s*\)`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Scanner next integer",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "BufferedReader.readLine()",
			Pattern:     `\.readLine\s*\(\s*\)`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelFile},
			Description: "Read line from reader",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Console.readLine()",
			Pattern:     `console\.readLine\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Console read line",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Console.readPassword()",
			Pattern:     `console\.readPassword\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Console read password",
			NodeTypes:   []string{"method_invocation"},
		},

		// File operations
		{
			Name:        "Files.readAllLines()",
			Pattern:     `Files\.readAllLines\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read all lines from file",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Files.readAllBytes()",
			Pattern:     `Files\.readAllBytes\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read all bytes from file",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Files.readString()",
			Pattern:     `Files\.readString\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read string from file",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "FileInputStream",
			Pattern:     `new\s+FileInputStream\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "File input stream",
			NodeTypes:   []string{"object_creation_expression"},
		},
		{
			Name:        "FileReader",
			Pattern:     `new\s+FileReader\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "File reader",
			NodeTypes:   []string{"object_creation_expression"},
		},

		// Network
		{
			Name:        "URL.openStream()",
			Pattern:     `\.openStream\s*\(\s*\)`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "URL input stream",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "HttpURLConnection.getInputStream()",
			Pattern:     `\.getInputStream\s*\(\s*\)`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "HTTP connection input",
			NodeTypes:   []string{"method_invocation"},
		},

		// JSON parsing
		{
			Name:        "ObjectMapper.readValue()",
			Pattern:     `\.readValue\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Jackson JSON parsing",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Gson.fromJson()",
			Pattern:     `\.fromJson\s*\(`,
			Language:    "java",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Gson JSON parsing",
			NodeTypes:   []string{"method_invocation"},
		},
	}

	return &Matcher{
		BaseMatcher: common.NewBaseMatcher("java", defs),
	}
}
