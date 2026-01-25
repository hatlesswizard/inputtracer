package python

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// Matcher matches Python user input sources
type Matcher struct {
	*common.BaseMatcher
}

// NewMatcher creates a new Python source matcher
func NewMatcher() *Matcher {
	defs := []common.Definition{
		// Flask
		{
			Name:         "request.args",
			Pattern:      `request\.args(?:\.\w+|\[|\.get)`,
			Language:     "python",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Flask GET parameters",
			NodeTypes:    []string{"attribute", "subscript"},
			KeyExtractor: `request\.args\.get\s*\(\s*['"](\w+)['"]|request\.args\[['"](\w+)['"]\]`,
		},
		{
			Name:         "request.form",
			Pattern:      `request\.form(?:\.\w+|\[|\.get)`,
			Language:     "python",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Flask POST form data",
			NodeTypes:    []string{"attribute", "subscript"},
			KeyExtractor: `request\.form\.get\s*\(\s*['"](\w+)['"]|request\.form\[['"](\w+)['"]\]`,
		},
		{
			Name:         "request.values",
			Pattern:      `request\.values(?:\.\w+|\[|\.get)`,
			Language:     "python",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Flask combined GET/POST",
			NodeTypes:    []string{"attribute", "subscript"},
			KeyExtractor: `request\.values\.get\s*\(\s*['"](\w+)['"]|request\.values\[['"](\w+)['"]\]`,
		},
		{
			Name:        "request.json",
			Pattern:     `request\.json`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Flask JSON body",
			NodeTypes:   []string{"attribute"},
		},
		{
			Name:        "request.data",
			Pattern:     `request\.data`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Flask raw body",
			NodeTypes:   []string{"attribute"},
		},
		{
			Name:         "request.files",
			Pattern:      `request\.files(?:\.\w+|\[|\.get)`,
			Language:     "python",
			Labels:       []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description:  "Flask file uploads",
			NodeTypes:    []string{"attribute", "subscript"},
			KeyExtractor: `request\.files\.get\s*\(\s*['"](\w+)['"]|request\.files\[['"](\w+)['"]\]`,
		},
		{
			Name:         "request.headers",
			Pattern:      `request\.headers(?:\.\w+|\[|\.get)`,
			Language:     "python",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "Flask HTTP headers",
			NodeTypes:    []string{"attribute", "subscript"},
			KeyExtractor: `request\.headers\.get\s*\(\s*['"]([^'"]+)['"]|request\.headers\[['"]([^'"]+)['"]\]`,
		},
		{
			Name:         "request.cookies",
			Pattern:      `request\.cookies(?:\.\w+|\[|\.get)`,
			Language:     "python",
			Labels:       []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:  "Flask cookies",
			NodeTypes:    []string{"attribute", "subscript"},
			KeyExtractor: `request\.cookies\.get\s*\(\s*['"](\w+)['"]|request\.cookies\[['"](\w+)['"]\]`,
		},

		// Django
		{
			Name:         "request.GET",
			Pattern:      `request\.GET(?:\.\w+|\[|\.get)`,
			Language:     "python",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Django GET parameters",
			NodeTypes:    []string{"attribute", "subscript"},
			KeyExtractor: `request\.GET\.get\s*\(\s*['"](\w+)['"]|request\.GET\[['"](\w+)['"]\]`,
		},
		{
			Name:         "request.POST",
			Pattern:      `request\.POST(?:\.\w+|\[|\.get)`,
			Language:     "python",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Django POST data",
			NodeTypes:    []string{"attribute", "subscript"},
			KeyExtractor: `request\.POST\.get\s*\(\s*['"](\w+)['"]|request\.POST\[['"](\w+)['"]\]`,
		},
		{
			Name:        "request.body",
			Pattern:     `request\.body`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Django raw body",
			NodeTypes:   []string{"attribute"},
		},

		// FastAPI
		{
			Name:        "Query()",
			Pattern:     `Query\s*\(`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "FastAPI query parameter",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "Body()",
			Pattern:     `Body\s*\(`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "FastAPI body parameter",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "Path()",
			Pattern:     `Path\s*\(`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "FastAPI path parameter",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "Header()",
			Pattern:     `Header\s*\(`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "FastAPI header parameter",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "Cookie()",
			Pattern:     `Cookie\s*\(`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description: "FastAPI cookie parameter",
			NodeTypes:   []string{"call"},
		},

		// Built-in input
		{
			Name:        "input()",
			Pattern:     `\binput\s*\(`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Standard input",
			NodeTypes:   []string{"call"},
		},

		// CLI
		{
			Name:        "sys.argv",
			Pattern:     `sys\.argv`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"attribute", "subscript"},
		},

		// Environment
		{
			Name:         "os.environ",
			Pattern:      `os\.environ(?:\.\w+|\[|\.get)`,
			Language:     "python",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "Environment variables",
			NodeTypes:    []string{"attribute", "subscript"},
			KeyExtractor: `os\.environ\.get\s*\(\s*['"](\w+)['"]|os\.environ\[['"](\w+)['"]\]`,
		},
		{
			Name:        "os.getenv()",
			Pattern:     `os\.getenv\s*\(`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelEnvironment},
			Description: "Get environment variable",
			NodeTypes:   []string{"call"},
		},

		// File operations
		{
			Name:        "open().read()",
			Pattern:     `\.read\s*\(`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "File read",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "open().readline()",
			Pattern:     `\.readline\s*\(`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "File readline",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "open().readlines()",
			Pattern:     `\.readlines\s*\(`,
			Language:    "python",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "File readlines",
			NodeTypes:   []string{"call"},
		},

	}

	return &Matcher{
		BaseMatcher: common.NewBaseMatcher("python", defs),
	}
}
