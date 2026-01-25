package javascript

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// Matcher matches JavaScript user input sources
type Matcher struct {
	*common.BaseMatcher
}

// getDefinitions returns the source definitions for JavaScript/TypeScript
func getDefinitions(language string) []common.Definition {
	return []common.Definition{
		// Express.js / Node.js HTTP
		{
			Name:         "req.body",
			Pattern:      `req\.body(?:\.\w+|\[)`,
			Language:     language,
			Labels:       []common.InputLabel{common.LabelHTTPBody, common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Express POST body",
			NodeTypes:    []string{"member_expression", "subscript_expression"},
			KeyExtractor: `req\.body\.(\w+)|req\.body\[['"](\w+)['"]\]`,
		},
		{
			Name:         "req.query",
			Pattern:      `req\.query(?:\.\w+|\[)`,
			Language:     language,
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Express query parameters",
			NodeTypes:    []string{"member_expression", "subscript_expression"},
			KeyExtractor: `req\.query\.(\w+)|req\.query\[['"](\w+)['"]\]`,
		},
		{
			Name:         "req.params",
			Pattern:      `req\.params(?:\.\w+|\[)`,
			Language:     language,
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Express URL parameters",
			NodeTypes:    []string{"member_expression", "subscript_expression"},
			KeyExtractor: `req\.params\.(\w+)|req\.params\[['"](\w+)['"]\]`,
		},
		{
			Name:         "req.headers",
			Pattern:      `req\.headers(?:\.\w+|\[)`,
			Language:     language,
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "HTTP request headers",
			NodeTypes:    []string{"member_expression", "subscript_expression"},
			KeyExtractor: `req\.headers\.(\w+)|req\.headers\[['"]([^'"]+)['"]\]`,
		},
		{
			Name:         "req.cookies",
			Pattern:      `req\.cookies(?:\.\w+|\[)`,
			Language:     language,
			Labels:       []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:  "HTTP cookies",
			NodeTypes:    []string{"member_expression", "subscript_expression"},
			KeyExtractor: `req\.cookies\.(\w+)|req\.cookies\[['"](\w+)['"]\]`,
		},

		// Browser APIs
		{
			Name:        "document.location",
			Pattern:     `document\.location`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Current page URL",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "window.location",
			Pattern:     `window\.location`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Window URL",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "location.href",
			Pattern:     `location\.href`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Full URL",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "location.search",
			Pattern:     `location\.search`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "URL query string",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "location.hash",
			Pattern:     `location\.hash`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "URL fragment",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "document.URL",
			Pattern:     `document\.URL`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Document URL",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "document.referrer",
			Pattern:     `document\.referrer`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Referring URL",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "document.cookie",
			Pattern:     `document\.cookie`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description: "Document cookies",
			NodeTypes:   []string{"member_expression"},
		},

		// Form inputs
		{
			Name:        "element.value",
			Pattern:     `\.value\b`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Form input value",
			NodeTypes:   []string{"member_expression"},
		},

		// XHR and Fetch
		{
			Name:        "XMLHttpRequest.response",
			Pattern:     `\.response(?:Text|XML)?`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "XHR response",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:        "fetch().json()",
			Pattern:     `\.json\s*\(\s*\)`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "Fetch JSON response",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fetch().text()",
			Pattern:     `\.text\s*\(\s*\)`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "Fetch text response",
			NodeTypes:   []string{"call_expression"},
		},

		// Node.js CLI
		{
			Name:        "process.argv",
			Pattern:     `process\.argv`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"member_expression"},
		},
		{
			Name:         "process.env",
			Pattern:      `process\.env(?:\.\w+|\[)`,
			Language:     language,
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "Environment variables",
			NodeTypes:    []string{"member_expression", "subscript_expression"},
			KeyExtractor: `process\.env\.(\w+)|process\.env\[['"](\w+)['"]\]`,
		},

		// Node.js file system
		{
			Name:        "fs.readFile",
			Pattern:     `fs\.readFile(?:Sync)?\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "File read",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fs.read",
			Pattern:     `fs\.read(?:Sync)?\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "File read",
			NodeTypes:   []string{"call_expression"},
		},

		// readline
		{
			Name:        "readline",
			Pattern:     `readline\.question|rl\.question`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Readline input",
			NodeTypes:   []string{"call_expression"},
		},

		// URL and URLSearchParams
		{
			Name:        "URLSearchParams.get",
			Pattern:     `\.get\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "URL search params",
			NodeTypes:   []string{"call_expression"},
		},

		// FormData
		{
			Name:        "FormData.get",
			Pattern:     `formData\.get\s*\(|FormData.*\.get\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "FormData value",
			NodeTypes:   []string{"call_expression"},
		},
	}
}

// NewMatcher creates a new JavaScript source matcher
func NewMatcher() *Matcher {
	return &Matcher{
		BaseMatcher: common.NewBaseMatcher("javascript", getDefinitions("javascript")),
	}
}

// TypeScriptMatcher matches TypeScript user input sources (same as JavaScript)
type TypeScriptMatcher struct {
	*common.BaseMatcher
}

// NewTypeScriptMatcher creates a new TypeScript source matcher
func NewTypeScriptMatcher() *TypeScriptMatcher {
	return &TypeScriptMatcher{
		BaseMatcher: common.NewBaseMatcher("typescript", getDefinitions("typescript")),
	}
}
