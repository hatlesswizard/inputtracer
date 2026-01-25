package csharp

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// Matcher matches C# user input sources
type Matcher struct {
	*common.BaseMatcher
}

// NewMatcher creates a new C# source matcher
func NewMatcher() *Matcher {
	defs := []common.Definition{
		// ASP.NET Core / MVC
		{
			Name:         "Request.Query",
			Pattern:      `Request\.Query\s*\[`,
			Language:     "c_sharp",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description:  "ASP.NET query string parameter",
			NodeTypes:    []string{"element_access_expression", "member_access_expression"},
			KeyExtractor: `Request\.Query\s*\[\s*"([^"]+)"`,
		},
		{
			Name:         "Request.Form",
			Pattern:      `Request\.Form\s*\[`,
			Language:     "c_sharp",
			Labels:       []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description:  "ASP.NET form data",
			NodeTypes:    []string{"element_access_expression", "member_access_expression"},
			KeyExtractor: `Request\.Form\s*\[\s*"([^"]+)"`,
		},
		{
			Name:         "Request.Headers",
			Pattern:      `Request\.Headers\s*\[`,
			Language:     "c_sharp",
			Labels:       []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description:  "ASP.NET request headers",
			NodeTypes:    []string{"element_access_expression", "member_access_expression"},
			KeyExtractor: `Request\.Headers\s*\[\s*"([^"]+)"`,
		},
		{
			Name:         "Request.Cookies",
			Pattern:      `Request\.Cookies\s*\[`,
			Language:     "c_sharp",
			Labels:       []common.InputLabel{common.LabelHTTPCookie, common.LabelUserInput},
			Description:  "ASP.NET cookies",
			NodeTypes:    []string{"element_access_expression", "member_access_expression"},
			KeyExtractor: `Request\.Cookies\s*\[\s*"([^"]+)"`,
		},
		{
			Name:        "Request.Body",
			Pattern:     `Request\.Body`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "ASP.NET request body",
			NodeTypes:   []string{"member_access_expression"},
		},
		{
			Name:        "Request.QueryString",
			Pattern:     `Request\.QueryString`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "ASP.NET query string",
			NodeTypes:   []string{"member_access_expression"},
		},
		{
			Name:        "Request.Path",
			Pattern:     `Request\.Path`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "ASP.NET request path",
			NodeTypes:   []string{"member_access_expression"},
		},
		{
			Name:        "Request.RouteValues",
			Pattern:     `Request\.RouteValues`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "ASP.NET route values",
			NodeTypes:   []string{"member_access_expression"},
		},

		// ASP.NET WebForms (legacy)
		{
			Name:         "Request.Params",
			Pattern:      `Request\.Params\s*\[`,
			Language:     "c_sharp",
			Labels:       []common.InputLabel{common.LabelHTTPGet, common.LabelHTTPPost, common.LabelUserInput},
			Description:  "ASP.NET combined parameters",
			NodeTypes:    []string{"element_access_expression"},
			KeyExtractor: `Request\.Params\s*\[\s*"([^"]+)"`,
		},
		{
			Name:        "Request.InputStream",
			Pattern:     `Request\.InputStream`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "ASP.NET request input stream",
			NodeTypes:   []string{"member_access_expression"},
		},
		{
			Name:        "Request.Files",
			Pattern:     `Request\.Files`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "ASP.NET uploaded files",
			NodeTypes:   []string{"member_access_expression"},
		},

		// MVC attributes (parameter binding)
		{
			Name:        "[FromQuery]",
			Pattern:     `\[FromQuery\]`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "ASP.NET MVC query binding",
			NodeTypes:   []string{"attribute"},
		},
		{
			Name:        "[FromBody]",
			Pattern:     `\[FromBody\]`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelHTTPBody, common.LabelUserInput},
			Description: "ASP.NET MVC body binding",
			NodeTypes:   []string{"attribute"},
		},
		{
			Name:        "[FromForm]",
			Pattern:     `\[FromForm\]`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelHTTPPost, common.LabelUserInput},
			Description: "ASP.NET MVC form binding",
			NodeTypes:   []string{"attribute"},
		},
		{
			Name:        "[FromHeader]",
			Pattern:     `\[FromHeader\]`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelHTTPHeader, common.LabelUserInput},
			Description: "ASP.NET MVC header binding",
			NodeTypes:   []string{"attribute"},
		},
		{
			Name:        "[FromRoute]",
			Pattern:     `\[FromRoute\]`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "ASP.NET MVC route binding",
			NodeTypes:   []string{"attribute"},
		},

		// CLI
		{
			Name:        "args[]",
			Pattern:     `\bargs\s*\[`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"element_access_expression"},
		},
		{
			Name:        "Environment.GetCommandLineArgs()",
			Pattern:     `Environment\.GetCommandLineArgs\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Get command line arguments",
			NodeTypes:   []string{"invocation_expression"},
		},

		// Environment
		{
			Name:         "Environment.GetEnvironmentVariable()",
			Pattern:      `Environment\.GetEnvironmentVariable\s*\(`,
			Language:     "c_sharp",
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "Get environment variable",
			NodeTypes:    []string{"invocation_expression"},
			KeyExtractor: `Environment\.GetEnvironmentVariable\s*\(\s*"([^"]+)"`,
		},

		// Console input
		{
			Name:        "Console.ReadLine()",
			Pattern:     `Console\.ReadLine\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Console read line",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "Console.Read()",
			Pattern:     `Console\.Read\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Console read character",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "Console.ReadKey()",
			Pattern:     `Console\.ReadKey\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Console read key",
			NodeTypes:   []string{"invocation_expression"},
		},

		// File operations
		{
			Name:        "File.ReadAllText()",
			Pattern:     `File\.ReadAllText\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read all text from file",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "File.ReadAllLines()",
			Pattern:     `File\.ReadAllLines\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read all lines from file",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "File.ReadAllBytes()",
			Pattern:     `File\.ReadAllBytes\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read all bytes from file",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "File.OpenRead()",
			Pattern:     `File\.OpenRead\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Open file for reading",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "StreamReader.ReadLine()",
			Pattern:     `\.ReadLine\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "Read line from stream",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "StreamReader.ReadToEnd()",
			Pattern:     `\.ReadToEnd\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelNetwork},
			Description: "Read stream to end",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "StreamReader.Read()",
			Pattern:     `\.Read\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelFile, common.LabelNetwork},
			Description: "Read from stream",
			NodeTypes:   []string{"invocation_expression"},
		},

		// Network
		{
			Name:        "HttpClient.GetAsync()",
			Pattern:     `\.GetAsync\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "HTTP GET request",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "HttpClient.PostAsync()",
			Pattern:     `\.PostAsync\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "HTTP POST request",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "HttpResponseMessage.Content",
			Pattern:     `\.Content\.ReadAsStringAsync\s*\(|\.Content\.ReadAsByteArrayAsync\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "HTTP response content",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "WebClient.DownloadString()",
			Pattern:     `\.DownloadString\s*\(`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "WebClient download string",
			NodeTypes:   []string{"invocation_expression"},
		},

		// JSON
		{
			Name:        "JsonSerializer.Deserialize()",
			Pattern:     `JsonSerializer\.Deserialize`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "JSON deserialization",
			NodeTypes:   []string{"invocation_expression"},
		},
		{
			Name:        "JsonConvert.DeserializeObject()",
			Pattern:     `JsonConvert\.DeserializeObject`,
			Language:    "c_sharp",
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Newtonsoft JSON deserialization",
			NodeTypes:   []string{"invocation_expression"},
		},
	}

	return &Matcher{
		BaseMatcher: common.NewBaseMatcher("c_sharp", defs),
	}
}
