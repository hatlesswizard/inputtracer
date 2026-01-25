package golang

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// Matcher matches Go user input sources
type Matcher struct {
	*common.BaseMatcher
}

// NewMatcher creates a new Go source matcher
func NewMatcher() *Matcher {
	defs := []common.Definition{
		// net/http
		{
			Name:        "r.URL.Query()",
			Pattern:     `\.URL\.Query\s*\(\s*\)`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "HTTP URL query parameters",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:         "r.FormValue()",
			Pattern:      `\.FormValue\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description:  "HTTP form value (GET or POST)",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `\.FormValue\s*\(\s*"([^"]+)"`,
		},
		{
			Name:         "r.PostFormValue()",
			Pattern:      `\.PostFormValue\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "HTTP POST form value",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `\.PostFormValue\s*\(\s*"([^"]+)"`,
		},
		{
			Name:        "r.Body",
			Pattern:     `\.\s*Body\b`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "HTTP request body",
			NodeTypes:   []string{"selector_expression"},
		},
		{
			Name:         "r.Header.Get()",
			Pattern:      `\.Header\.Get\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "HTTP header value",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `\.Header\.Get\s*\(\s*"([^"]+)"`,
		},
		{
			Name:         "r.Cookie()",
			Pattern:      `\.Cookie\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:  "HTTP cookie value",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `\.Cookie\s*\(\s*"([^"]+)"`,
		},
		{
			Name:         "r.PathValue()",
			Pattern:      `\.PathValue\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "HTTP path parameter (Go 1.22+)",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `\.PathValue\s*\(\s*"([^"]+)"`,
		},
		{
			Name:        "r.Form",
			Pattern:     `\.Form\b`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description: "Parsed form data",
			NodeTypes:   []string{"selector_expression"},
		},
		{
			Name:        "r.PostForm",
			Pattern:     `\.PostForm\b`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "Parsed POST form data",
			NodeTypes:   []string{"selector_expression"},
		},
		{
			Name:        "r.MultipartForm",
			Pattern:     `\.MultipartForm\b`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelFile, common.LabelUserInput},
			Description: "Multipart form data",
			NodeTypes:   []string{"selector_expression"},
		},

		// Gin framework
		{
			Name:         "c.Query()",
			Pattern:      `c\.Query\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Gin query parameter",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `c\.Query\s*\(\s*"([^"]+)"`,
		},
		{
			Name:         "c.DefaultQuery()",
			Pattern:      `c\.DefaultQuery\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Gin query with default",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `c\.DefaultQuery\s*\(\s*"([^"]+)"`,
		},
		{
			Name:         "c.PostForm()",
			Pattern:      `c\.PostForm\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Gin POST form value",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `c\.PostForm\s*\(\s*"([^"]+)"`,
		},
		{
			Name:         "c.Param()",
			Pattern:      `c\.Param\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Gin URL parameter",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `c\.Param\s*\(\s*"([^"]+)"`,
		},
		{
			Name:        "c.BindJSON()",
			Pattern:     `c\.BindJSON\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Gin JSON binding",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "c.ShouldBindJSON()",
			Pattern:     `c\.ShouldBindJSON\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Gin JSON binding",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:         "c.GetHeader()",
			Pattern:      `c\.GetHeader\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "Gin header value",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `c\.GetHeader\s*\(\s*"([^"]+)"`,
		},

		// Echo framework
		{
			Name:         "c.QueryParam()",
			Pattern:      `c\.QueryParam\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "Echo query parameter",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `c\.QueryParam\s*\(\s*"([^"]+)"`,
		},
		{
			Name:         "c.FormValue()",
			Pattern:      `c\.FormValue\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "Echo form value",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `c\.FormValue\s*\(\s*"([^"]+)"`,
		},

		// Fiber framework
		{
			Name:        "c.Params()",
			Pattern:     `c\.Params\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "Fiber URL parameters",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "c.BodyParser()",
			Pattern:     `c\.BodyParser\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "Fiber body parser",
			NodeTypes:   []string{"call_expression"},
		},

		// CLI
		{
			Name:        "os.Args",
			Pattern:     `os\.Args`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"selector_expression"},
		},
		{
			Name:        "flag.String()",
			Pattern:     `flag\.String\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Flag string argument",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "flag.Int()",
			Pattern:     `flag\.Int\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Flag int argument",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "flag.Bool()",
			Pattern:     `flag\.Bool\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Flag bool argument",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "flag.Arg()",
			Pattern:     `flag\.Arg\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Flag positional argument",
			NodeTypes:   []string{"call_expression"},
		},

		// Environment
		{
			Name:         "os.Getenv()",
			Pattern:      `os\.Getenv\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "Environment variable",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `os\.Getenv\s*\(\s*"([^"]+)"`,
		},
		{
			Name:         "os.LookupEnv()",
			Pattern:      `os\.LookupEnv\s*\(`,
			Language:     "go",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "Environment variable lookup",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `os\.LookupEnv\s*\(\s*"([^"]+)"`,
		},

		// File operations
		{
			Name:        "os.ReadFile()",
			Pattern:     `os\.ReadFile\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read entire file",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "ioutil.ReadFile()",
			Pattern:     `ioutil\.ReadFile\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read entire file (deprecated)",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "io.ReadAll()",
			Pattern:     `io\.ReadAll\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelNetwork},
			Description: "Read all from reader",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "bufio.Scanner",
			Pattern:     `bufio\.NewScanner\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "Buffered scanner",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "bufio.Reader.ReadString()",
			Pattern:     `\.ReadString\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "Read string from buffer",
			NodeTypes:   []string{"call_expression"},
		},

		// JSON decoding
		{
			Name:        "json.Unmarshal()",
			Pattern:     `json\.Unmarshal\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "JSON unmarshal",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "json.NewDecoder().Decode()",
			Pattern:     `\.Decode\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "JSON decode",
			NodeTypes:   []string{"call_expression"},
		},

		// Standard input
		{
			Name:        "os.Stdin",
			Pattern:     `os\.Stdin`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Standard input",
			NodeTypes:   []string{"selector_expression"},
		},
		{
			Name:        "fmt.Scan()",
			Pattern:     `fmt\.Scan\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Formatted scan from stdin",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fmt.Scanf()",
			Pattern:     `fmt\.Scanf\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Formatted scan from stdin",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fmt.Scanln()",
			Pattern:     `fmt\.Scanln\s*\(`,
			Language:    "go",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Scan line from stdin",
			NodeTypes:   []string{"call_expression"},
		},
	}

	return &Matcher{
		BaseMatcher: common.NewBaseMatcher("go", defs),
	}
}
