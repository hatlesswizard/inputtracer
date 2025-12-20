package sources

// CMatcher matches C user input sources
type CMatcher struct {
	*BaseMatcher
}

// NewCMatcher creates a new C source matcher
func NewCMatcher() *CMatcher {
	sources := []Definition{
		// CLI arguments
		{
			Name:        "argv",
			Pattern:     `\bargv\s*\[`,
			Language:    "c",
			Labels:      []InputLabel{LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"subscript_expression", "identifier"},
		},
		{
			Name:        "argc",
			Pattern:     `\bargc\b`,
			Language:    "c",
			Labels:      []InputLabel{LabelCLI},
			Description: "Argument count",
			NodeTypes:   []string{"identifier"},
		},

		// Standard input functions
		{
			Name:        "gets()",
			Pattern:     `\bgets\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Get string from stdin (unsafe)",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fgets()",
			Pattern:     `\bfgets\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Get string from stream",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "scanf()",
			Pattern:     `\bscanf\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Formatted input from stdin",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fscanf()",
			Pattern:     `\bfscanf\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelFile, LabelUserInput},
			Description: "Formatted input from file",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "sscanf()",
			Pattern:     `\bsscanf\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Formatted input from string",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "getchar()",
			Pattern:     `\bgetchar\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Get character from stdin",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "getc()",
			Pattern:     `\bgetc\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Get character from stream",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fgetc()",
			Pattern:     `\bfgetc\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelFile},
			Description: "Get character from file",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "getline()",
			Pattern:     `\bgetline\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Get line from stream",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "getdelim()",
			Pattern:     `\bgetdelim\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelUserInput, LabelFile},
			Description: "Get delimited string from stream",
			NodeTypes:   []string{"call_expression"},
		},

		// File operations
		{
			Name:        "fread()",
			Pattern:     `\bfread\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelFile},
			Description: "Read from file",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fopen()",
			Pattern:     `\bfopen\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelFile},
			Description: "Open file",
			NodeTypes:   []string{"call_expression"},
		},

		// POSIX read functions
		{
			Name:        "read()",
			Pattern:     `\bread\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelFile, LabelNetwork, LabelUserInput},
			Description: "POSIX read from file descriptor",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "pread()",
			Pattern:     `\bpread\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelFile},
			Description: "POSIX read at offset",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "readv()",
			Pattern:     `\breadv\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelFile, LabelNetwork},
			Description: "POSIX scatter read",
			NodeTypes:   []string{"call_expression"},
		},

		// Network input
		{
			Name:        "recv()",
			Pattern:     `\brecv\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelNetwork},
			Description: "Receive from socket",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "recvfrom()",
			Pattern:     `\brecvfrom\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelNetwork},
			Description: "Receive from socket with address",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "recvmsg()",
			Pattern:     `\brecvmsg\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelNetwork},
			Description: "Receive message from socket",
			NodeTypes:   []string{"call_expression"},
		},

		// Environment
		{
			Name:         "getenv()",
			Pattern:      `\bgetenv\s*\(`,
			Language:     "c",
			Labels:       []InputLabel{LabelEnvironment},
			Description:  "Get environment variable",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `\bgetenv\s*\(\s*"([^"]+)"`,
		},
		{
			Name:        "secure_getenv()",
			Pattern:     `\bsecure_getenv\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelEnvironment},
			Description: "Secure get environment variable",
			NodeTypes:   []string{"call_expression"},
		},

		// Memory mapped files
		{
			Name:        "mmap()",
			Pattern:     `\bmmap\s*\(`,
			Language:    "c",
			Labels:      []InputLabel{LabelFile},
			Description: "Memory map file",
			NodeTypes:   []string{"call_expression"},
		},

		// CGI environment (web)
		{
			Name:        "QUERY_STRING",
			Pattern:     `\bQUERY_STRING\b`,
			Language:    "c",
			Labels:      []InputLabel{LabelHTTPGet, LabelUserInput},
			Description: "CGI query string",
			NodeTypes:   []string{"identifier"},
		},
		{
			Name:        "CONTENT_LENGTH",
			Pattern:     `\bCONTENT_LENGTH\b`,
			Language:    "c",
			Labels:      []InputLabel{LabelHTTPBody},
			Description: "CGI content length",
			NodeTypes:   []string{"identifier"},
		},
		{
			Name:        "HTTP_COOKIE",
			Pattern:     `\bHTTP_COOKIE\b`,
			Language:    "c",
			Labels:      []InputLabel{LabelHTTPCookie},
			Description: "CGI HTTP cookie",
			NodeTypes:   []string{"identifier"},
		},
		{
			Name:        "REQUEST_METHOD",
			Pattern:     `\bREQUEST_METHOD\b`,
			Language:    "c",
			Labels:      []InputLabel{LabelHTTPHeader},
			Description: "CGI request method",
			NodeTypes:   []string{"identifier"},
		},

		// stdin
		{
			Name:        "stdin",
			Pattern:     `\bstdin\b`,
			Language:    "c",
			Labels:      []InputLabel{LabelUserInput},
			Description: "Standard input stream",
			NodeTypes:   []string{"identifier"},
		},
	}

	return &CMatcher{
		BaseMatcher: NewBaseMatcher("c", sources),
	}
}
