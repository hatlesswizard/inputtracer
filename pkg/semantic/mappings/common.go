package mappings

import (
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
)

// CGIEnvVars contains CGI environment variable mappings shared between C and C++
var CGIEnvVars = map[string]types.SourceType{
	"QUERY_STRING":    types.SourceHTTPGet,
	"REQUEST_METHOD":  types.SourceHTTPHeader,
	"CONTENT_TYPE":    types.SourceHTTPHeader,
	"CONTENT_LENGTH":  types.SourceHTTPBody,
	"HTTP_COOKIE":     types.SourceHTTPCookie,
	"HTTP_HOST":       types.SourceHTTPHeader,
	"HTTP_USER_AGENT": types.SourceHTTPHeader,
	"HTTP_REFERER":    types.SourceHTTPHeader,
	"HTTP_ACCEPT":     types.SourceHTTPHeader,
	"PATH_INFO":       types.SourceHTTPPath,
	"PATH_TRANSLATED": types.SourceHTTPPath,
	"SCRIPT_NAME":     types.SourceHTTPPath,
	"REQUEST_URI":     types.SourceHTTPPath,
	"REMOTE_ADDR":     types.SourceNetwork,
	"REMOTE_HOST":     types.SourceNetwork,
	"SERVER_NAME":     types.SourceHTTPHeader,
	"SERVER_PORT":     types.SourceHTTPHeader,
	"HTTPS":           types.SourceHTTPHeader,
}

// StandardCInputFunctions contains C standard library input functions shared between C and C++
var StandardCInputFunctions = map[string]types.SourceType{
	// Standard input functions
	"gets":     types.SourceStdin,
	"fgets":    types.SourceFile,
	"scanf":    types.SourceStdin,
	"fscanf":   types.SourceFile,
	"sscanf":   types.SourceUserInput,
	"getchar":  types.SourceStdin,
	"getc":     types.SourceFile,
	"fgetc":    types.SourceFile,
	"getline":  types.SourceStdin,
	"getdelim": types.SourceFile,

	// POSIX file/socket read functions
	"read":   types.SourceFile,
	"pread":  types.SourceFile,
	"readv":  types.SourceFile,
	"preadv": types.SourceFile,
	"fread":  types.SourceFile,

	// Network input functions
	"recv":     types.SourceNetwork,
	"recvfrom": types.SourceNetwork,
	"recvmsg":  types.SourceNetwork,
	"recvmmsg": types.SourceNetwork,

	// Environment variables
	"getenv":        types.SourceEnvVar,
	"secure_getenv": types.SourceEnvVar,

	// Memory-mapped files
	"mmap": types.SourceFile,

	// File operations
	"fopen":  types.SourceFile,
	"open":   types.SourceFile,
	"fdopen": types.SourceFile,
}
