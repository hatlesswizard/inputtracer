// Package sources - labels.go provides centralized SourceType definitions
// Re-exports from common package for backwards compatibility
package sources

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// SourceType represents the semantic type of an input source
// Re-exported from common package
type SourceType = common.SourceType

const (
	SourceHTTPGet     = common.SourceHTTPGet     // Query string parameters
	SourceHTTPPost    = common.SourceHTTPPost    // POST form data
	SourceHTTPBody    = common.SourceHTTPBody    // Raw request body
	SourceHTTPJSON    = common.SourceHTTPJSON    // JSON request body
	SourceHTTPHeader  = common.SourceHTTPHeader  // HTTP headers
	SourceHTTPCookie  = common.SourceHTTPCookie  // Cookies
	SourceHTTPPath    = common.SourceHTTPPath    // URL path parameters
	SourceHTTPFile    = common.SourceHTTPFile    // Uploaded files ($_FILES)
	SourceHTTPRequest = common.SourceHTTPRequest // Combined GET/POST ($_REQUEST)
	SourceSession     = common.SourceSession     // Session data ($_SESSION)
	SourceCLIArg      = common.SourceCLIArg      // Command line arguments
	SourceEnvVar      = common.SourceEnvVar      // Environment variables
	SourceStdin       = common.SourceStdin       // Standard input
	SourceFile        = common.SourceFile        // File reads
	SourceDatabase    = common.SourceDatabase    // Database query results
	SourceNetwork     = common.SourceNetwork     // Network/socket reads
	SourceUserInput   = common.SourceUserInput   // Generic user input
	SourceUnknown     = common.SourceUnknown     // Unknown source type
)

// AllSourceTypes returns all valid source types for iteration/validation
var AllSourceTypes = common.AllSourceTypes

// IsValidSourceType checks if a string is a valid SourceType
func IsValidSourceType(s string) bool {
	return common.IsValidSourceType(s)
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
