// Package common - source_types.go provides centralized SourceType definitions.
// SourceType is now a type alias pointing to the canonical definition in pkg/sources/core.
package common

import "github.com/hatlesswizard/inputtracer/pkg/sources/core"

// SourceType represents the semantic type of an input source.
// This is a type alias — the canonical definition lives in pkg/sources/core.
type SourceType = core.SourceType

// Re-export SourceType constants from core for backward compatibility.
const (
	SourceHTTPGet     = core.SourceHTTPGet     // Query string parameters
	SourceHTTPPost    = core.SourceHTTPPost    // POST form data
	SourceHTTPBody    = core.SourceHTTPBody    // Raw request body
	SourceHTTPJSON    = core.SourceHTTPJSON    // JSON request body
	SourceHTTPHeader  = core.SourceHTTPHeader  // HTTP headers
	SourceHTTPCookie  = core.SourceHTTPCookie  // Cookies
	SourceHTTPPath    = core.SourceHTTPPath    // URL path parameters
	SourceHTTPFile    = core.SourceHTTPFile    // Uploaded files ($_FILES)
	SourceHTTPRequest = core.SourceHTTPRequest // Combined GET/POST ($_REQUEST)
	SourceSession     = core.SourceSession     // Session data ($_SESSION)
	SourceCLIArg      = core.SourceCLIArg      // Command line arguments
	SourceEnvVar      = core.SourceEnvVar      // Environment variables
	SourceStdin       = core.SourceStdin       // Standard input
	SourceFile        = core.SourceFile        // File reads
	SourceDatabase    = core.SourceDatabase    // Database query results
	SourceNetwork     = core.SourceNetwork     // Network/socket reads
	SourceUserInput   = core.SourceUserInput   // Generic user input
	SourceUnknown     = core.SourceUnknown     // Unknown source type
)

// AllSourceTypes lists all valid source types for iteration/validation.
var AllSourceTypes = core.AllSourceTypes

// IsValidSourceType checks if a string is a valid SourceType.
func IsValidSourceType(s string) bool {
	return core.IsValidSourceType(s)
}
