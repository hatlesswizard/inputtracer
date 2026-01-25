package c

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// Matcher matches C user input sources
type Matcher struct {
	*common.BaseMatcher
}

// GetDefinitions returns the source definitions for C (also used by C++)
func GetDefinitions(language string) []common.Definition {
	return []common.Definition{
		// CLI arguments
		{
			Name:        "argv",
			Pattern:     `\bargv\s*\[`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Command line arguments",
			NodeTypes:   []string{"subscript_expression", "identifier"},
		},
		{
			Name:        "argc",
			Pattern:     `\bargc\b`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelCLI},
			Description: "Argument count",
			NodeTypes:   []string{"identifier"},
		},

		// Standard input functions
		{
			Name:        "gets()",
			Pattern:     `\bgets\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Get string from stdin (unsafe)",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fgets()",
			Pattern:     `\bfgets\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelFile},
			Description: "Get string from stream",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "scanf()",
			Pattern:     `\bscanf\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Formatted input from stdin",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fscanf()",
			Pattern:     `\bfscanf\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelFile, common.LabelUserInput},
			Description: "Formatted input from file",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "sscanf()",
			Pattern:     `\bsscanf\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Formatted input from string",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "getchar()",
			Pattern:     `\bgetchar\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Get character from stdin",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "getc()",
			Pattern:     `\bgetc\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelFile},
			Description: "Get character from stream",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fgetc()",
			Pattern:     `\bfgetc\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Get character from file",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "getline()",
			Pattern:     `\bgetline\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelFile},
			Description: "Get line from stream",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "getdelim()",
			Pattern:     `\bgetdelim\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput, common.LabelFile},
			Description: "Get delimited string from stream",
			NodeTypes:   []string{"call_expression"},
		},

		// File operations
		{
			Name:        "fread()",
			Pattern:     `\bfread\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Read from file",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "fopen()",
			Pattern:     `\bfopen\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Open file",
			NodeTypes:   []string{"call_expression"},
		},

		// POSIX read functions
		{
			Name:        "read()",
			Pattern:     `\bread\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelFile, common.LabelNetwork, common.LabelUserInput},
			Description: "POSIX read from file descriptor",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "pread()",
			Pattern:     `\bpread\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "POSIX read at offset",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "readv()",
			Pattern:     `\breadv\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelFile, common.LabelNetwork},
			Description: "POSIX scatter read",
			NodeTypes:   []string{"call_expression"},
		},

		// Network input
		{
			Name:        "recv()",
			Pattern:     `\brecv\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "Receive from socket",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "recvfrom()",
			Pattern:     `\brecvfrom\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "Receive from socket with address",
			NodeTypes:   []string{"call_expression"},
		},
		{
			Name:        "recvmsg()",
			Pattern:     `\brecvmsg\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelNetwork},
			Description: "Receive message from socket",
			NodeTypes:   []string{"call_expression"},
		},

		// Environment
		{
			Name:         "getenv()",
			Pattern:      `\bgetenv\s*\(`,
			Language:     language,
			Labels:       []common.InputLabel{common.LabelEnvironment},
			Description:  "Get environment variable",
			NodeTypes:    []string{"call_expression"},
			KeyExtractor: `\bgetenv\s*\(\s*"([^"]+)"`,
		},
		{
			Name:        "secure_getenv()",
			Pattern:     `\bsecure_getenv\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelEnvironment},
			Description: "Secure get environment variable",
			NodeTypes:   []string{"call_expression"},
		},

		// Memory mapped files
		{
			Name:        "mmap()",
			Pattern:     `\bmmap\s*\(`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelFile},
			Description: "Memory map file",
			NodeTypes:   []string{"call_expression"},
		},

		// CGI environment (web)
		{
			Name:        "QUERY_STRING",
			Pattern:     `\bQUERY_STRING\b`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelHTTPGet, common.LabelUserInput},
			Description: "CGI query string",
			NodeTypes:   []string{"identifier"},
		},
		{
			Name:        "CONTENT_LENGTH",
			Pattern:     `\bCONTENT_LENGTH\b`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelHTTPBody},
			Description: "CGI content length",
			NodeTypes:   []string{"identifier"},
		},
		{
			Name:        "HTTP_COOKIE",
			Pattern:     `\bHTTP_COOKIE\b`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelHTTPCookie},
			Description: "CGI HTTP cookie",
			NodeTypes:   []string{"identifier"},
		},
		{
			Name:        "REQUEST_METHOD",
			Pattern:     `\bREQUEST_METHOD\b`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelHTTPHeader},
			Description: "CGI request method",
			NodeTypes:   []string{"identifier"},
		},

		// stdin
		{
			Name:        "stdin",
			Pattern:     `\bstdin\b`,
			Language:    language,
			Labels:      []common.InputLabel{common.LabelUserInput},
			Description: "Standard input stream",
			NodeTypes:   []string{"identifier"},
		},
	}
}

// NewMatcher creates a new C source matcher
func NewMatcher() *Matcher {
	return &Matcher{
		BaseMatcher: common.NewBaseMatcher("c", GetDefinitions("c")),
	}
}
