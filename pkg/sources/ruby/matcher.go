package ruby

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// Matcher matches Ruby user input sources
type Matcher struct {
	*common.BaseMatcher
}

// NewMatcher creates a new Ruby source matcher
func NewMatcher() *Matcher {
	defs := []common.Definition{
		// Rails - params
		{
			Name:         "params[]",
			Pattern:      `params\s*\[`,
			Language:     "ruby",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Rails request parameters",
			NodeTypes:    []string{"element_reference", "call"},
			KeyExtractor: `params\s*\[\s*:?['"]?(\w+)['"]?\s*\]`,
		},
		{
			Name:        "params.permit()",
			Pattern:     `params\.permit\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "Rails strong parameters",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "params.require()",
			Pattern:     `params\.require\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "Rails required parameters",
			NodeTypes:   []string{"call"},
		},

		// Rails - request
		{
			Name:        "request.params",
			Pattern:     `request\.params`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "Rails request params",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "request.body",
			Pattern:     `request\.body`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Rails request body",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "request.raw_post",
			Pattern:     `request\.raw_post`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Rails raw POST body",
			NodeTypes:   []string{"call"},
		},
		{
			Name:         "request.headers[]",
			Pattern:      `request\.headers\s*\[`,
			Language:     "ruby",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "Rails request headers",
			NodeTypes:    []string{"element_reference"},
			KeyExtractor: `request\.headers\s*\[\s*['"]([^'"]+)['"]\s*\]`,
		},
		{
			Name:        "request.env",
			Pattern:     `request\.env`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelEnvironment},
			Description: "Rails request environment",
			NodeTypes:   []string{"call"},
		},

		// Rails - cookies
		{
			Name:         "cookies[]",
			Pattern:      `cookies\s*\[`,
			Language:     "ruby",
			Labels:       []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:  "Rails cookies",
			NodeTypes:    []string{"element_reference"},
			KeyExtractor: `cookies\s*\[\s*:?['"]?(\w+)['"]?\s*\]`,
		},
		{
			Name:        "session[]",
			Pattern:     `session\s*\[`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Rails session",
			NodeTypes:   []string{"element_reference"},
		},

		// Sinatra
		{
			Name:        "sinatra params",
			Pattern:     `params\s*\[`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "Sinatra parameters",
			NodeTypes:   []string{"element_reference"},
		},
		{
			Name:        "request.body.read",
			Pattern:     `request\.body\.read`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Sinatra request body",
			NodeTypes:   []string{"call"},
		},

		// CLI
		{
			Name:        "ARGV",
			Pattern:     `\bARGV\b`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"constant"},
		},
		{
			Name:        "ARGV[]",
			Pattern:     `ARGV\s*\[`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Command line argument access",
			NodeTypes:   []string{"element_reference"},
		},

		// Environment
		{
			Name:         "ENV[]",
			Pattern:      `ENV\s*\[`,
			Language:     "ruby",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "Environment variable",
			NodeTypes:    []string{"element_reference"},
			KeyExtractor: `ENV\s*\[\s*['"]([^'"]+)['"]\s*\]`,
		},
		{
			Name:        "ENV.fetch()",
			Pattern:     `ENV\.fetch\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelEnvironment},
			Description: "Environment variable with default",
			NodeTypes:   []string{"call"},
		},

		// Standard input
		{
			Name:        "gets",
			Pattern:     `\bgets\b`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Get line from stdin",
			NodeTypes:   []string{"identifier", "call"},
		},
		{
			Name:        "readline",
			Pattern:     `\breadline\b`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Read line from stdin",
			NodeTypes:   []string{"identifier", "call"},
		},
		{
			Name:        "readlines",
			Pattern:     `\breadlines\b`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Read all lines from stdin",
			NodeTypes:   []string{"identifier", "call"},
		},
		{
			Name:        "STDIN.read",
			Pattern:     `STDIN\.read`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Read from stdin",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "STDIN.gets",
			Pattern:     `STDIN\.gets`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Get line from stdin",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "$stdin.read",
			Pattern:     `\$stdin\.read`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Read from stdin global",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "ARGF.read",
			Pattern:     `ARGF\.read`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelFile},
			Description: "Read from ARGF (files or stdin)",
			NodeTypes:   []string{"call"},
		},

		// File operations
		{
			Name:        "File.read()",
			Pattern:     `File\.read\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read entire file",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "File.readlines()",
			Pattern:     `File\.readlines\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read file as array of lines",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "IO.read()",
			Pattern:     `IO\.read\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read from IO",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "IO.readlines()",
			Pattern:     `IO\.readlines\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read lines from IO",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "File.open().read",
			Pattern:     `\.read\b`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read from file handle",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "File.open().gets",
			Pattern:     `\.gets\b`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Get line from file handle",
			NodeTypes:   []string{"call"},
		},

		// Network
		{
			Name:        "Net::HTTP.get()",
			Pattern:     `Net::HTTP\.get`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "HTTP GET request",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "HTTParty.get()",
			Pattern:     `HTTParty\.get`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "HTTParty GET request",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "Faraday.get()",
			Pattern:     `Faraday\.get|\.get\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "Faraday HTTP request",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "RestClient.get()",
			Pattern:     `RestClient\.get`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "RestClient GET request",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "open-uri",
			Pattern:     `open\s*\(\s*['"]http`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "open-uri HTTP request",
			NodeTypes:   []string{"call"},
		},

		// JSON parsing
		{
			Name:        "JSON.parse()",
			Pattern:     `JSON\.parse\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "JSON parse",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "JSON.load()",
			Pattern:     `JSON\.load\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "JSON load",
			NodeTypes:   []string{"call"},
		},

		// YAML parsing
		{
			Name:        "YAML.load()",
			Pattern:     `YAML\.load\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "YAML load (unsafe)",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "YAML.safe_load()",
			Pattern:     `YAML\.safe_load\s*\(`,
			Language:    "ruby",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "YAML safe load",
			NodeTypes:   []string{"call"},
		},
	}

	return &Matcher{
		BaseMatcher: common.NewBaseMatcher("ruby", defs),
	}
}
