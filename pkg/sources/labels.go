// Package sources - labels.go provides centralized SourceType definitions
// These should be the ONLY source type definitions in the entire codebase
package sources

// SourceType represents the semantic type of an input source
// Consolidated from pkg/semantic/types/types.go
type SourceType string

const (
	SourceHTTPGet     SourceType = "http_get"     // Query string parameters
	SourceHTTPPost    SourceType = "http_post"    // POST form data
	SourceHTTPBody    SourceType = "http_body"    // Raw request body
	SourceHTTPJSON    SourceType = "http_json"    // JSON request body
	SourceHTTPHeader  SourceType = "http_header"  // HTTP headers
	SourceHTTPCookie  SourceType = "http_cookie"  // Cookies
	SourceHTTPPath    SourceType = "http_path"    // URL path parameters
	SourceHTTPFile    SourceType = "http_file"    // Uploaded files ($_FILES)
	SourceHTTPRequest SourceType = "http_request" // Combined GET/POST ($_REQUEST)
	SourceSession     SourceType = "session"      // Session data ($_SESSION)
	SourceCLIArg      SourceType = "cli_arg"      // Command line arguments
	SourceEnvVar      SourceType = "env_var"      // Environment variables
	SourceStdin       SourceType = "stdin"        // Standard input
	SourceFile        SourceType = "file"         // File reads
	SourceDatabase    SourceType = "database"     // Database query results
	SourceNetwork     SourceType = "network"      // Network/socket reads
	SourceUserInput   SourceType = "user_input"   // Generic user input
	SourceUnknown     SourceType = "unknown"      // Unknown source type
)

// AllSourceTypes returns all valid source types for iteration/validation
var AllSourceTypes = []SourceType{
	SourceHTTPGet, SourceHTTPPost, SourceHTTPBody, SourceHTTPJSON,
	SourceHTTPHeader, SourceHTTPCookie, SourceHTTPPath, SourceHTTPFile,
	SourceHTTPRequest, SourceSession, SourceCLIArg, SourceEnvVar,
	SourceStdin, SourceFile, SourceDatabase, SourceNetwork, SourceUserInput,
}

// IsValidSourceType checks if a string is a valid SourceType
func IsValidSourceType(s string) bool {
	for _, st := range AllSourceTypes {
		if string(st) == s {
			return true
		}
	}
	return false
}

// LabelToSourceType maps InputLabel to SourceType for conversion
var LabelToSourceType = map[InputLabel]SourceType{
	LabelHTTPGet:     SourceHTTPGet,
	LabelHTTPPost:    SourceHTTPPost,
	LabelHTTPCookie:  SourceHTTPCookie,
	LabelHTTPHeader:  SourceHTTPHeader,
	LabelHTTPBody:    SourceHTTPBody,
	LabelCLI:         SourceCLIArg,
	LabelEnvironment: SourceEnvVar,
	LabelFile:        SourceFile,
	LabelDatabase:    SourceDatabase,
	LabelNetwork:     SourceNetwork,
	LabelUserInput:   SourceUserInput,
}

// SourceTypeToLabel maps SourceType back to InputLabel (reverse lookup)
var SourceTypeToLabel = map[SourceType]InputLabel{
	SourceHTTPGet:    LabelHTTPGet,
	SourceHTTPPost:   LabelHTTPPost,
	SourceHTTPCookie: LabelHTTPCookie,
	SourceHTTPHeader: LabelHTTPHeader,
	SourceHTTPBody:   LabelHTTPBody,
	SourceCLIArg:     LabelCLI,
	SourceEnvVar:     LabelEnvironment,
	SourceFile:       LabelFile,
	SourceDatabase:   LabelDatabase,
	SourceNetwork:    LabelNetwork,
	SourceUserInput:  LabelUserInput,
}
