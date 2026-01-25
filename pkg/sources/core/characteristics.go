package core

// Characteristic defines what makes something an input source
type Characteristic struct {
	Description string       // Human-readable description
	Indicators  []string     // Method/property names that indicate this type
	Category    SourceType   // Primary category
	Labels      []InputLabel // Labels to apply
	IsUserInput bool         // True if this is user-controlled input
	Reason      string       // Why this categorization
}

// UserInputCharacteristics defines patterns that ARE user input
var UserInputCharacteristics = []Characteristic{
	{
		Description: "HTTP GET query parameters",
		Indicators: []string{
			"query", "queryParams", "query_params", "searchParams",
			"getQueryParam", "getQuery", "query_string", "qs",
			"GET", "$_GET", "request.query", "req.query",
		},
		Category:    SourceHTTPGet,
		Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
		IsUserInput: true,
		Reason:      "Query parameters are directly controlled by the client URL",
	},
	{
		Description: "HTTP POST body data",
		Indicators: []string{
			"body", "postData", "formData", "requestBody", "parsedBody",
			"getBody", "getParsedBody", "input", "post",
			"POST", "$_POST", "request.body", "req.body",
		},
		Category:    SourceHTTPPost,
		Labels:      []InputLabel{LabelHTTPPost, LabelUserInput},
		IsUserInput: true,
		Reason:      "POST body is directly sent by the client",
	},
	{
		Description: "HTTP cookies",
		Indicators: []string{
			"cookies", "cookie", "getCookie", "cookieParams",
			"COOKIE", "$_COOKIE", "request.cookies", "req.cookies",
		},
		Category:    SourceHTTPCookie,
		Labels:      []InputLabel{LabelHTTPCookie, LabelUserInput},
		IsUserInput: true,
		Reason:      "Cookies are sent by the client in the request",
	},
	{
		Description: "HTTP headers",
		Indicators: []string{
			"headers", "header", "getHeader", "getHeaders",
			"SERVER", "$_SERVER", "request.headers", "req.headers",
		},
		Category:    SourceHTTPHeader,
		Labels:      []InputLabel{LabelHTTPHeader, LabelUserInput},
		IsUserInput: true,
		Reason:      "HTTP headers are sent by the client",
	},
	{
		Description: "HTTP file uploads",
		Indicators: []string{
			"files", "uploadedFiles", "getUploadedFiles", "file",
			"FILES", "$_FILES", "request.files", "req.files",
		},
		Category:    SourceHTTPFile,
		Labels:      []InputLabel{LabelFile, LabelUserInput},
		IsUserInput: true,
		Reason:      "Uploaded files are sent by the client",
	},
	{
		Description: "HTTP request (generic)",
		Indicators: []string{
			"REQUEST", "$_REQUEST", "all", "input", "getInput",
		},
		Category:    SourceHTTPRequest,
		Labels:      []InputLabel{LabelUserInput},
		IsUserInput: true,
		Reason:      "Generic request data is client-controlled",
	},
	{
		Description: "Command line arguments",
		Indicators: []string{
			"argv", "args", "arguments", "sys.argv", "os.Args",
			"process.argv", "ARGV", "$argv",
		},
		Category:    SourceCLIArg,
		Labels:      []InputLabel{LabelCLI, LabelUserInput},
		IsUserInput: true,
		Reason:      "CLI arguments are provided by the user",
	},
	{
		Description: "Standard input",
		Indicators: []string{
			"stdin", "STDIN", "sys.stdin", "os.Stdin",
			"process.stdin", "php://input", "readline",
		},
		Category:    SourceStdin,
		Labels:      []InputLabel{LabelUserInput},
		IsUserInput: true,
		Reason:      "Standard input is provided by the user",
	},
}

// NonUserInputCharacteristics defines patterns that are NOT user input
var NonUserInputCharacteristics = []Characteristic{
	{
		Description: "Session data (server-side storage)",
		Indicators: []string{
			"session", "getSession", "sessionData", "SESSION", "$_SESSION",
			"request.session", "req.session",
		},
		Category:    SourceSession,
		Labels:      []InputLabel{}, // NO LabelUserInput!
		IsUserInput: false,
		Reason:      "Session data is stored server-side, NOT sent in the request",
	},
	{
		Description: "Database query results",
		Indicators: []string{
			"fetch", "fetchAll", "fetchOne", "fetchRow", "fetchColumn",
			"query", "execute", "findOne", "findAll", "findBy",
			"get", "first", "all", "find",
		},
		Category:    SourceDatabase,
		Labels:      []InputLabel{LabelDatabase},
		IsUserInput: false,
		Reason:      "Database results may contain user data but are not direct input",
	},
	{
		Description: "File system reads",
		Indicators: []string{
			"readFile", "file_get_contents", "fread", "fgets", "file",
			"readFileSync", "fs.read", "ioutil.ReadFile", "os.ReadFile",
		},
		Category:    SourceFile,
		Labels:      []InputLabel{LabelFile},
		IsUserInput: false,
		Reason:      "File contents are server-side data, not user input",
	},
	{
		Description: "Environment variables",
		Indicators: []string{
			"env", "getenv", "environ", "ENV", "$_ENV",
			"process.env", "os.Getenv", "os.environ",
		},
		Category:    SourceEnvVar,
		Labels:      []InputLabel{LabelEnvironment},
		IsUserInput: false,
		Reason:      "Environment variables are server configuration",
	},
	{
		Description: "Cache/transient data",
		Indicators: []string{
			"cache", "getCache", "cacheGet", "transient", "getTransient",
			"memcache", "redis", "wp_cache_get",
		},
		Category:    SourceDatabase,
		Labels:      []InputLabel{},
		IsUserInput: false,
		Reason:      "Cache data is server-side storage",
	},
}

// GetCharacteristicForIndicator finds the characteristic for an indicator
func GetCharacteristicForIndicator(indicator string) *Characteristic {
	// Check user input characteristics first
	for i := range UserInputCharacteristics {
		for _, ind := range UserInputCharacteristics[i].Indicators {
			if ind == indicator {
				return &UserInputCharacteristics[i]
			}
		}
	}

	// Check non-user input characteristics
	for i := range NonUserInputCharacteristics {
		for _, ind := range NonUserInputCharacteristics[i].Indicators {
			if ind == indicator {
				return &NonUserInputCharacteristics[i]
			}
		}
	}

	return nil
}

// IsUserInputIndicator checks if an indicator represents user input
func IsUserInputIndicator(indicator string) bool {
	char := GetCharacteristicForIndicator(indicator)
	return char != nil && char.IsUserInput
}

// GetCategoryForIndicator returns the SourceType for an indicator
func GetCategoryForIndicator(indicator string) SourceType {
	char := GetCharacteristicForIndicator(indicator)
	if char != nil {
		return char.Category
	}
	return SourceUnknown
}

// GetLabelsForIndicator returns the labels for an indicator
func GetLabelsForIndicator(indicator string) []InputLabel {
	char := GetCharacteristicForIndicator(indicator)
	if char != nil {
		return char.Labels
	}
	return nil
}
