package sources

// JavaMatcher matches Java user input sources
type JavaMatcher struct {
	*BaseMatcher
}

// NewJavaMatcher creates a new Java source matcher
func NewJavaMatcher() *JavaMatcher {
	sources := []Definition{
		// Servlet API
		{
			Name:         "request.getParameter()",
			Pattern:      `\.getParameter\s*\(`,
			Language:     "java",
			Labels:       []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelUserInput},
			Description:  "HTTP request parameter",
			NodeTypes:    []string{"method_invocation"},
			KeyExtractor: `\.getParameter\s*\(\s*"([^"]+)"`,
		},
		{
			Name:        "request.getParameterValues()",
			Pattern:     `\.getParameterValues\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelUserInput},
			Description: "HTTP request parameter array",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getParameterMap()",
			Pattern:     `\.getParameterMap\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelUserInput},
			Description: "All HTTP request parameters",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:         "request.getHeader()",
			Pattern:      `\.getHeader\s*\(`,
			Language:     "java",
			Labels:       []InputLabel{LabelHTTPHeader, LabelUserInput},
			Description:  "HTTP request header",
			NodeTypes:    []string{"method_invocation"},
			KeyExtractor: `\.getHeader\s*\(\s*"([^"]+)"`,
		},
		{
			Name:        "request.getHeaders()",
			Pattern:     `\.getHeaders\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPHeader, LabelUserInput},
			Description: "HTTP request headers",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getInputStream()",
			Pattern:     `\.getInputStream\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPBody, LabelUserInput},
			Description: "HTTP request body stream",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getReader()",
			Pattern:     `\.getReader\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPBody, LabelUserInput},
			Description: "HTTP request body reader",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getCookies()",
			Pattern:     `\.getCookies\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPCookie, LabelUserInput},
			Description: "HTTP cookies",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getQueryString()",
			Pattern:     `\.getQueryString\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "HTTP query string",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getRequestURI()",
			Pattern:     `\.getRequestURI\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "HTTP request URI",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getRequestURL()",
			Pattern:     `\.getRequestURL\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "HTTP request URL",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getPathInfo()",
			Pattern:     `\.getPathInfo\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "HTTP path info",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getPart()",
			Pattern:     `\.getPart\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelFile, LabelUserInput},
			Description: "Multipart form part",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "request.getParts()",
			Pattern:     `\.getParts\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelFile, LabelUserInput},
			Description: "All multipart form parts",
			NodeTypes:   []string{"method_invocation"},
		},

		// Spring MVC annotations (detected in annotations)
		{
			Name:        "@RequestParam",
			Pattern:     `@RequestParam`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelUserInput},
			Description: "Spring request parameter",
			NodeTypes:   []string{"annotation", "marker_annotation"},
		},
		{
			Name:        "@PathVariable",
			Pattern:     `@PathVariable`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "Spring path variable",
			NodeTypes:   []string{"annotation", "marker_annotation"},
		},
		{
			Name:        "@RequestBody",
			Pattern:     `@RequestBody`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPBody, LabelUserInput},
			Description: "Spring request body",
			NodeTypes:   []string{"annotation", "marker_annotation"},
		},
		{
			Name:        "@RequestHeader",
			Pattern:     `@RequestHeader`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPHeader, LabelUserInput},
			Description: "Spring request header",
			NodeTypes:   []string{"annotation", "marker_annotation"},
		},
		{
			Name:        "@CookieValue",
			Pattern:     `@CookieValue`,
			Language:    "java",
			Labels:      []InputLabel{LabelHTTPCookie, LabelUserInput},
			Description: "Spring cookie value",
			NodeTypes:   []string{"annotation", "marker_annotation"},
		},

		// CLI
		{
			Name:        "args[]",
			Pattern:     `\bargs\s*\[`,
			Language:    "java",
			Labels:      []InputLabel{LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"array_access"},
		},

		// Environment
		{
			Name:         "System.getenv()",
			Pattern:      `System\.getenv\s*\(`,
			Language:     "java",
			Labels:       []InputLabel{LabelEnvironment},
			Description:  "Environment variable",
			NodeTypes:    []string{"method_invocation"},
			KeyExtractor: `System\.getenv\s*\(\s*"([^"]+)"`,
		},
		{
			Name:         "System.getProperty()",
			Pattern:      `System\.getProperty\s*\(`,
			Language:     "java",
			Labels:       []InputLabel{LabelEnvironment},
			Description:  "System property",
			NodeTypes:    []string{"method_invocation"},
			KeyExtractor: `System\.getProperty\s*\(\s*"([^"]+)"`,
		},

		// Console input
		{
			Name:        "Scanner.next()",
			Pattern:     `\.next\s*\(\s*\)`,
			Language:    "java",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Scanner next token",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Scanner.nextLine()",
			Pattern:     `\.nextLine\s*\(\s*\)`,
			Language:    "java",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Scanner next line",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Scanner.nextInt()",
			Pattern:     `\.nextInt\s*\(\s*\)`,
			Language:    "java",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Scanner next integer",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "BufferedReader.readLine()",
			Pattern:     `\.readLine\s*\(\s*\)`,
			Language:    "java",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Read line from reader",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Console.readLine()",
			Pattern:     `console\.readLine\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Console read line",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Console.readPassword()",
			Pattern:     `console\.readPassword\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Console read password",
			NodeTypes:   []string{"method_invocation"},
		},

		// File operations
		{
			Name:        "Files.readAllLines()",
			Pattern:     `Files\.readAllLines\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelFile},
			Description: "Read all lines from file",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Files.readAllBytes()",
			Pattern:     `Files\.readAllBytes\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelFile},
			Description: "Read all bytes from file",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Files.readString()",
			Pattern:     `Files\.readString\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelFile},
			Description: "Read string from file",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "FileInputStream",
			Pattern:     `new\s+FileInputStream\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelFile},
			Description: "File input stream",
			NodeTypes:   []string{"object_creation_expression"},
		},
		{
			Name:        "FileReader",
			Pattern:     `new\s+FileReader\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelFile},
			Description: "File reader",
			NodeTypes:   []string{"object_creation_expression"},
		},

		// Database
		{
			Name:        "ResultSet.getString()",
			Pattern:     `\.getString\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelDatabase},
			Description: "Database string result",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "ResultSet.getInt()",
			Pattern:     `\.getInt\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelDatabase},
			Description: "Database integer result",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "ResultSet.getObject()",
			Pattern:     `\.getObject\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelDatabase},
			Description: "Database object result",
			NodeTypes:   []string{"method_invocation"},
		},

		// Network
		{
			Name:        "URL.openStream()",
			Pattern:     `\.openStream\s*\(\s*\)`,
			Language:    "java",
			Labels:      []InputLabel{LabelNetwork},
			Description: "URL input stream",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "HttpURLConnection.getInputStream()",
			Pattern:     `\.getInputStream\s*\(\s*\)`,
			Language:    "java",
			Labels:      []InputLabel{LabelNetwork},
			Description: "HTTP connection input",
			NodeTypes:   []string{"method_invocation"},
		},

		// JSON parsing
		{
			Name:        "ObjectMapper.readValue()",
			Pattern:     `\.readValue\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Jackson JSON parsing",
			NodeTypes:   []string{"method_invocation"},
		},
		{
			Name:        "Gson.fromJson()",
			Pattern:     `\.fromJson\s*\(`,
			Language:    "java",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Gson JSON parsing",
			NodeTypes:   []string{"method_invocation"},
		},
	}

	return &JavaMatcher{
		BaseMatcher: NewBaseMatcher("java", sources),
	}
}
