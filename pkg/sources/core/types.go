// Package core provides the centralized type definitions and registry for input detection.
// This is the SINGLE SOURCE OF TRUTH for all input-related types.
package core

// SourceType categorizes the origin of input data
type SourceType string

const (
	SourceHTTPGet     SourceType = "http_get"
	SourceHTTPPost    SourceType = "http_post"
	SourceHTTPBody    SourceType = "http_body"
	SourceHTTPJSON    SourceType = "http_json"
	SourceHTTPHeader  SourceType = "http_header"
	SourceHTTPCookie  SourceType = "http_cookie"
	SourceHTTPPath    SourceType = "http_path"
	SourceHTTPFile    SourceType = "http_file"
	SourceHTTPRequest SourceType = "http_request"
	SourceSession     SourceType = "session"
	SourceCLIArg      SourceType = "cli_arg"
	SourceEnvVar      SourceType = "env_var"
	SourceStdin       SourceType = "stdin"
	SourceFile        SourceType = "file"
	SourceDatabase    SourceType = "database"
	SourceNetwork     SourceType = "network"
	SourceUserInput   SourceType = "user_input"
	SourceUnknown     SourceType = "unknown"
)

// InputLabel provides additional categorization for input sources
type InputLabel string

const (
	LabelHTTPGet     InputLabel = "http_get"
	LabelHTTPPost    InputLabel = "http_post"
	LabelHTTPCookie  InputLabel = "http_cookie"
	LabelHTTPHeader  InputLabel = "http_header"
	LabelHTTPBody    InputLabel = "http_body"
	LabelCLI         InputLabel = "cli"
	LabelEnvironment InputLabel = "environment"
	LabelFile        InputLabel = "file"
	LabelDatabase    InputLabel = "database"
	LabelNetwork     InputLabel = "network"
	LabelUserInput   InputLabel = "user_input"
)

// IsUserInput returns true if this source type represents direct user input
func (s SourceType) IsUserInput() bool {
	switch s {
	case SourceHTTPGet, SourceHTTPPost, SourceHTTPBody, SourceHTTPJSON,
		SourceHTTPCookie, SourceHTTPPath, SourceHTTPFile, SourceHTTPRequest,
		SourceCLIArg, SourceStdin, SourceUserInput:
		return true
	default:
		return false
	}
}

// IsServerSideData returns true if this source type is server-controlled
func (s SourceType) IsServerSideData() bool {
	switch s {
	case SourceSession, SourceDatabase, SourceFile, SourceEnvVar:
		return true
	default:
		return false
	}
}

// LabelToSourceType maps InputLabel to SourceType
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

// SourceTypeToLabel maps SourceType to InputLabel
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
