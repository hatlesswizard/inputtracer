package sources

// JavaScriptMatcher matches JavaScript user input sources
type JavaScriptMatcher struct {
	*BaseMatcher
}

// NewJavaScriptMatcher creates a new JavaScript source matcher
func NewJavaScriptMatcher() *JavaScriptMatcher {
	sources := []Definition{
		// Express.js / Node.js HTTP
		{
			Name:         "req.body",
			Pattern:      `req\.body(?:\.\w+|\[)`,
			Language:     "javascript",
			Labels:       []InputLabel{LabelHTTPBody, LabelHTTPPost, LabelUserInput},
			Description:  "Express POST body",
			NodeTypes:    []string{"member_expression", "subscript_expression"},
			KeyExtractor: `req\.body\.(\w+)|req\.body\[['"](\w+)['"]\]`,
		},
		{
			Name:         "req.query",
			Pattern:      `req\.query(?:\.\w+|\[)`,
			Language:     "javascript",
			Labels:       []InputLabel{LabelHTTPGet, LabelUserInput},
			Description:  "Express query parameters",
			NodeTypes:    []string{"member_expression", "subscript_expression"},
			KeyExtractor: `req\.query\.(\w+)|req\.query\[['"](\w+)['"]\]`,
		},
		{
			Name:         "req.params",
			Pattern:      `req\.params(?:\.\w+|\[)`,
			Language:     "javascript",
			Labels:       []InputLabel{LabelHTTPGet, LabelUserInput},
			Description:  "Express URL parameters",
			NodeTypes:    []string{"member_expression", "subscript_expression"},
			KeyExtractor: `req\.params\.(\w+)|req\.params\[['"](\w+)['"]\]`,
		},
		{
			Name:         "req.headers",
			Pattern:      `req\.headers(?:\.\w+|\[)`,
			Language:     "javascript",
			Labels:       []InputLabel{LabelHTTPHeader, LabelUserInput},
			Description:  "HTTP request headers",
			NodeTypes:    []string{"member_expression", "subscript_expression"},
			KeyExtractor: `req\.headers\.(\w+)|req\.headers\[['"]([^'"]+)['"]\]`,
		},
		{
			Name:         "req.cookies",
			Pattern:      `req\.cookies(?:\.\w+|\[)`,
			Language:     "javascript",
			Labels:       []InputLabel{LabelHTTPCookie, LabelUserInput},
			Description:  "HTTP cookies",
			NodeTypes:    []string{"member_expression", "subscript_expression"},
			KeyExtractor: `req\.cookies\.(\w+)|req\.cookies\[['"](\w+)['"]\]`,
		},

		// Browser APIs
		{
			Name:        "document.location",
			Pattern:     `document\.location`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Current page URL",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "window.location",
			Pattern:     `window\.location`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Window URL",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "location.href",
			Pattern:     `location\.href`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Full URL",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "location.search",
			Pattern:     `location\.search`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "URL query string",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "location.hash",
			Pattern:     `location\.hash`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelUserInput},
			Description: "URL fragment",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "document.URL",
			Pattern:     `document\.URL`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Document URL",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "document.referrer",
			Pattern:     `document\.referrer`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Referring URL",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "document.cookie",
			Pattern:     `document\.cookie`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelHTTPCookie, LabelUserInput},
			Description: "Document cookies",
			NodeTypes:   []string{"member_expression"},
		},

		// Form inputs
		{
			Name:        "element.value",
			Pattern:     `\.value\b`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Form input value",
			NodeTypes:   []string{"member_expression"},
		},

		// XHR and Fetch
		{
			Name:        "XMLHttpRequest.response",
			Pattern:     `\.response(?:Text|XML)?`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelNetwork},
			Description: "XHR response",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "fetch().json()",
			Pattern:     `\.json\s*\(\s*\)`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelNetwork},
			Description: "Fetch JSON response",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fetch().text()",
			Pattern:     `\.text\s*\(\s*\)`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelNetwork},
			Description: "Fetch text response",
			NodeTypes:   []string{"call_expression"},
		},

		// Node.js CLI
		{
			Name:        "process.argv",
			Pattern:     `process\.argv`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "process.env",
			Pattern:     `process\.env(?:\.\w+|\[)`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelEnvironment},
			Description: "Environment variables",
			NodeTypes:   []string{"member_expression", "subscript_expression"},
			KeyExtractor: `process\.env\.(\w+)|process\.env\[['"](\w+)['"]\]`,
		},

		// Node.js file system
		{
			Name:        "fs.readFile",
			Pattern:     `fs\.readFile(?:Sync)?\s*\(`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelFile},
			Description: "File read",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fs.read",
			Pattern:     `fs\.read(?:Sync)?\s*\(`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelFile},
			Description: "File read",
			NodeTypes:   []string{"call_expression"},
		},

		// readline
		{
			Name:        "readline",
			Pattern:     `readline\.question|rl\.question`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Readline input",
			NodeTypes:   []string{"call_expression"},
		},

		// URL and URLSearchParams
		{
			Name:        "URLSearchParams.get",
			Pattern:     `\.get\s*\(`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "URL search params",
			NodeTypes:   []string{"call_expression"},
		},

		// FormData
		{
			Name:        "FormData.get",
			Pattern:     `formData\.get\s*\(|FormData.*\.get\s*\(`,
			Language:    "javascript",
			Labels:      []InputLabel{LabelUserInput},
			Description: "FormData value",
			NodeTypes:   []string{"call_expression"},
		},
	}

	return &JavaScriptMatcher{
		BaseMatcher: NewBaseMatcher("javascript", sources),
	}
}

// TypeScriptMatcher matches TypeScript user input sources (same as JavaScript)
type TypeScriptMatcher struct {
	*BaseMatcher
}

// NewTypeScriptMatcher creates a new TypeScript source matcher
func NewTypeScriptMatcher() *TypeScriptMatcher {
	// Use same sources as JavaScript but with typescript language
	jsMatcher := NewJavaScriptMatcher()
	sources := make([]Definition, len(jsMatcher.sources))
	for i, src := range jsMatcher.sources {
		src.Language = "typescript"
		sources[i] = src
	}

	return &TypeScriptMatcher{
		BaseMatcher: NewBaseMatcher("typescript", sources),
	}
}
