package sources

// RubyMatcher matches Ruby user input sources
type RubyMatcher struct {
	*BaseMatcher
}

// NewRubyMatcher creates a new Ruby source matcher
func NewRubyMatcher() *RubyMatcher {
	sources := []Definition{
		// Rails - params
		{
			Name:         "params[]",
			Pattern:      `params\s*\[`,
			Language:     "ruby",
			Labels:       []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelUserInput},
			Description:  "Rails request parameters",
			NodeTypes:    []string{"element_reference", "call"},
			KeyExtractor: `params\s*\[\s*:?['"]?(\w+)['"]?\s*\]`,
		},
		{
			Name:        "params.permit()",
			Pattern:     `params\.permit\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelUserInput},
			Description: "Rails strong parameters",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "params.require()",
			Pattern:     `params\.require\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelUserInput},
			Description: "Rails required parameters",
			NodeTypes:   []string{"call"},
		},

		// Rails - request
		{
			Name:        "request.params",
			Pattern:     `request\.params`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelUserInput},
			Description: "Rails request params",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "request.body",
			Pattern:     `request\.body`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelHTTPBody, LabelUserInput},
			Description: "Rails request body",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "request.raw_post",
			Pattern:     `request\.raw_post`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelHTTPBody, LabelUserInput},
			Description: "Rails raw POST body",
			NodeTypes:   []string{"call"},
		},
		{
			Name:         "request.headers[]",
			Pattern:      `request\.headers\s*\[`,
			Language:     "ruby",
			Labels:       []InputLabel{LabelHTTPHeader, LabelUserInput},
			Description:  "Rails request headers",
			NodeTypes:    []string{"element_reference"},
			KeyExtractor: `request\.headers\s*\[\s*['"]([^'"]+)['"]\s*\]`,
		},
		{
			Name:        "request.env",
			Pattern:     `request\.env`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelHTTPHeader, LabelEnvironment},
			Description: "Rails request environment",
			NodeTypes:   []string{"call"},
		},

		// Rails - cookies
		{
			Name:         "cookies[]",
			Pattern:      `cookies\s*\[`,
			Language:     "ruby",
			Labels:       []InputLabel{LabelHTTPCookie, LabelUserInput},
			Description:  "Rails cookies",
			NodeTypes:    []string{"element_reference"},
			KeyExtractor: `cookies\s*\[\s*:?['"]?(\w+)['"]?\s*\]`,
		},
		{
			Name:        "session[]",
			Pattern:     `session\s*\[`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Rails session",
			NodeTypes:   []string{"element_reference"},
		},

		// Sinatra
		{
			Name:        "sinatra params",
			Pattern:     `params\s*\[`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelUserInput},
			Description: "Sinatra parameters",
			NodeTypes:   []string{"element_reference"},
		},
		{
			Name:        "request.body.read",
			Pattern:     `request\.body\.read`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelHTTPBody, LabelUserInput},
			Description: "Sinatra request body",
			NodeTypes:   []string{"call"},
		},

		// CLI
		{
			Name:        "ARGV",
			Pattern:     `\bARGV\b`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"constant"},
		},
		{
			Name:        "ARGV[]",
			Pattern:     `ARGV\s*\[`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelCLI},
			Description: "Command line argument access",
			NodeTypes:   []string{"element_reference"},
		},

		// Environment
		{
			Name:         "ENV[]",
			Pattern:      `ENV\s*\[`,
			Language:     "ruby",
			Labels:       []InputLabel{LabelEnvironment},
			Description:  "Environment variable",
			NodeTypes:    []string{"element_reference"},
			KeyExtractor: `ENV\s*\[\s*['"]([^'"]+)['"]\s*\]`,
		},
		{
			Name:        "ENV.fetch()",
			Pattern:     `ENV\.fetch\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelEnvironment},
			Description: "Environment variable with default",
			NodeTypes:   []string{"call"},
		},

		// Standard input
		{
			Name:        "gets",
			Pattern:     `\bgets\b`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Get line from stdin",
			NodeTypes:   []string{"identifier", "call"},
		},
		{
			Name:        "readline",
			Pattern:     `\breadline\b`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Read line from stdin",
			NodeTypes:   []string{"identifier", "call"},
		},
		{
			Name:        "readlines",
			Pattern:     `\breadlines\b`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Read all lines from stdin",
			NodeTypes:   []string{"identifier", "call"},
		},
		{
			Name:        "STDIN.read",
			Pattern:     `STDIN\.read`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Read from stdin",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "STDIN.gets",
			Pattern:     `STDIN\.gets`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Get line from stdin",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "$stdin.read",
			Pattern:     `\$stdin\.read`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Read from stdin global",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "ARGF.read",
			Pattern:     `ARGF\.read`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Read from ARGF (files or stdin)",
			NodeTypes:   []string{"call"},
		},

		// File operations
		{
			Name:        "File.read()",
			Pattern:     `File\.read\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelFile},
			Description: "Read entire file",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "File.readlines()",
			Pattern:     `File\.readlines\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelFile},
			Description: "Read file as array of lines",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "IO.read()",
			Pattern:     `IO\.read\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelFile},
			Description: "Read from IO",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "IO.readlines()",
			Pattern:     `IO\.readlines\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelFile},
			Description: "Read lines from IO",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "File.open().read",
			Pattern:     `\.read\b`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelFile},
			Description: "Read from file handle",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "File.open().gets",
			Pattern:     `\.gets\b`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelFile},
			Description: "Get line from file handle",
			NodeTypes:   []string{"call"},
		},

		// Database (ActiveRecord)
		{
			Name:        "Model.find()",
			Pattern:     `\.find\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelDatabase},
			Description: "ActiveRecord find",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "Model.find_by()",
			Pattern:     `\.find_by\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelDatabase},
			Description: "ActiveRecord find by",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "Model.where()",
			Pattern:     `\.where\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelDatabase},
			Description: "ActiveRecord where",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "Model.all",
			Pattern:     `\.all\b`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelDatabase},
			Description: "ActiveRecord all",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "Model.first",
			Pattern:     `\.first\b`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelDatabase},
			Description: "ActiveRecord first",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "Model.last",
			Pattern:     `\.last\b`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelDatabase},
			Description: "ActiveRecord last",
			NodeTypes:   []string{"call"},
		},

		// Network
		{
			Name:        "Net::HTTP.get()",
			Pattern:     `Net::HTTP\.get`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelNetwork},
			Description: "HTTP GET request",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "HTTParty.get()",
			Pattern:     `HTTParty\.get`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelNetwork},
			Description: "HTTParty GET request",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "Faraday.get()",
			Pattern:     `Faraday\.get|\.get\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelNetwork},
			Description: "Faraday HTTP request",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "RestClient.get()",
			Pattern:     `RestClient\.get`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelNetwork},
			Description: "RestClient GET request",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "open-uri",
			Pattern:     `open\s*\(\s*['"]http`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelNetwork},
			Description: "open-uri HTTP request",
			NodeTypes:   []string{"call"},
		},

		// JSON parsing
		{
			Name:        "JSON.parse()",
			Pattern:     `JSON\.parse\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput},
			Description: "JSON parse",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "JSON.load()",
			Pattern:     `JSON\.load\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput},
			Description: "JSON load",
			NodeTypes:   []string{"call"},
		},

		// YAML parsing
		{
			Name:        "YAML.load()",
			Pattern:     `YAML\.load\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput},
			Description: "YAML load (unsafe)",
			NodeTypes:   []string{"call"},
		},
		{
			Name:        "YAML.safe_load()",
			Pattern:     `YAML\.safe_load\s*\(`,
			Language:    "ruby",
			Labels:      []InputLabel{LabelUserInput},
			Description: "YAML safe load",
			NodeTypes:   []string{"call"},
		},
	}

	return &RubyMatcher{
		BaseMatcher: NewBaseMatcher("ruby", sources),
	}
}
