// Package sources - superglobals.go provides centralized PHP superglobal mappings
// This is the ONLY place PHP superglobal definitions should exist
package sources

// PHPSuperglobal represents a PHP superglobal variable with all its metadata
type PHPSuperglobal struct {
	Name        string       // "$_GET", "$_POST", etc.
	SourceType  SourceType   // Mapped source type
	Labels      []InputLabel // Input categories
	Description string       // Human-readable description
}

// PHPSuperglobals is the canonical list of PHP superglobals
// This replaces scattered arrays in discovery/superglobal.go and elsewhere
//
// STRICT USER INPUT DEFINITION:
// User input is ONLY data that comes from an HTTP REQUEST.
//
// YES - These ARE user input (from HTTP request):
//   - $_GET, $_POST, $_COOKIE, $_REQUEST, $_FILES
//   - $_SERVER keys that contain request data (HTTP_*, REQUEST_URI, etc.)
//   - php://input, file_get_contents('php://input')
//
// NO - These are NOT user input:
//   - $_SESSION (stored server-side, not sent in request)
//   - $_ENV, getenv() (server configuration, not request data)
//   - Database query results (not request data)
//   - File reads from filesystem (not request data)
var PHPSuperglobals = []PHPSuperglobal{
	// HTTP REQUEST DATA - TRUE USER INPUT
	{"$_GET", SourceHTTPGet, []InputLabel{LabelHTTPGet, LabelUserInput}, "HTTP GET parameters (query string)"},
	{"$_POST", SourceHTTPPost, []InputLabel{LabelHTTPPost, LabelUserInput}, "HTTP POST parameters (form data)"},
	{"$_REQUEST", SourceHTTPRequest, []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelHTTPCookie, LabelUserInput}, "Combined GET/POST/COOKIE (request data)"},
	{"$_COOKIE", SourceHTTPCookie, []InputLabel{LabelHTTPCookie, LabelUserInput}, "HTTP cookies (sent with request)"},
	{"$_FILES", SourceHTTPFile, []InputLabel{LabelFile, LabelUserInput}, "HTTP file uploads (multipart request data)"},
	// $_SERVER contains BOTH user-controllable (HTTP_*, REQUEST_URI) and server-config values
	// We mark it as user input because many keys ARE user-controllable (see PHPServerUserKeys)
	{"$_SERVER", SourceHTTPHeader, []InputLabel{LabelHTTPHeader, LabelUserInput}, "Server/request info (SOME keys are user-controllable)"},

	// NOT USER INPUT - Server-side or configuration data
	{"$_ENV", SourceEnvVar, []InputLabel{LabelEnvironment}, "Environment variables (server config, NOT request data)"},
	{"$_SESSION", SourceSession, []InputLabel{}, "Session data (stored server-side, NOT sent in request)"},
}

// SuperglobalToSourceType maps superglobal name to SourceType
// Replaces switch statements in executor.go, classifier.go, etc.
var SuperglobalToSourceType = map[string]SourceType{
	"$_GET":     SourceHTTPGet,
	"$_POST":    SourceHTTPPost,
	"$_REQUEST": SourceHTTPRequest,
	"$_COOKIE":  SourceHTTPCookie,
	"$_SERVER":  SourceHTTPHeader,
	"$_FILES":   SourceHTTPFile,
	"$_ENV":     SourceEnvVar,
	"$_SESSION": SourceSession,
}

// SourceTypeToSuperglobal maps SourceType back to superglobal name (reverse lookup)
var SourceTypeToSuperglobal = map[SourceType]string{
	SourceHTTPGet:     "$_GET",
	SourceHTTPPost:    "$_POST",
	SourceHTTPRequest: "$_REQUEST",
	SourceHTTPCookie:  "$_COOKIE",
	SourceHTTPHeader:  "$_SERVER",
	SourceHTTPFile:    "$_FILES",
	SourceEnvVar:      "$_ENV",
	SourceSession:     "$_SESSION",
}

// SuperglobalShortNames maps superglobal names to short classifier names
// Replaces classifier.superglobalToSourceTypes
var SuperglobalShortNames = map[string]string{
	"$_GET":     "GET",
	"$_POST":    "POST",
	"$_COOKIE":  "COOKIE",
	"$_REQUEST": "REQUEST",
	"$_SERVER":  "SERVER",
	"$_FILES":   "FILES",
	"$_SESSION": "SESSION",
	"$_ENV":     "ENV",
}

// ShortNameToSuperglobal maps short names back to superglobal names (reverse)
var ShortNameToSuperglobal = map[string]string{
	"GET":     "$_GET",
	"POST":    "$_POST",
	"COOKIE":  "$_COOKIE",
	"REQUEST": "$_REQUEST",
	"SERVER":  "$_SERVER",
	"FILES":   "$_FILES",
	"SESSION": "$_SESSION",
	"ENV":     "$_ENV",
}

// SuperglobalNames returns just the names for iteration
// Replaces discovery.PHPSuperglobals array
func SuperglobalNames() []string {
	names := make([]string, len(PHPSuperglobals))
	for i, sg := range PHPSuperglobals {
		names[i] = sg.Name
	}
	return names
}

// GetSuperglobalInfo returns full info for a superglobal name
func GetSuperglobalInfo(name string) *PHPSuperglobal {
	for i := range PHPSuperglobals {
		if PHPSuperglobals[i].Name == name {
			return &PHPSuperglobals[i]
		}
	}
	return nil
}

// GetSuperglobalSourceType returns the SourceType for a superglobal, or SourceUnknown
func GetSuperglobalSourceType(name string) SourceType {
	if st, ok := SuperglobalToSourceType[name]; ok {
		return st
	}
	return SourceUnknown
}

// GetSuperglobalShortName returns the short name for a superglobal (e.g., "$_GET" -> "GET")
func GetSuperglobalShortName(name string) string {
	if sn, ok := SuperglobalShortNames[name]; ok {
		return sn
	}
	return ""
}

// IsSuperglobal checks if a name is a known PHP superglobal
func IsSuperglobal(name string) bool {
	_, ok := SuperglobalToSourceType[name]
	return ok
}

// GetSuperglobalByShortName returns the superglobal name from its short name (e.g., "GET" -> "$_GET")
func GetSuperglobalByShortName(shortName string) string {
	if name, ok := ShortNameToSuperglobal[shortName]; ok {
		return name
	}
	return ""
}

// PHPServerUserKeys are $_SERVER keys that contain USER-CONTROLLABLE data from the HTTP request.
// An attacker can manipulate these values by crafting their HTTP request.
// This list is based on PHP documentation and JetBrains PHP stubs research.
//
// Reference: https://www.php.net/manual/en/reserved.variables.server.php
// Key insight: "All elements of the $_SERVER array whose keys begin with 'HTTP_'
// come from HTTP request headers and are not to be trusted."
var PHPServerUserKeys = map[string]SourceType{
	// URL/Path - User controls the request URL
	"PHP_SELF":     SourceHTTPPath, // Script path (vulnerable to XSS if not encoded)
	"REQUEST_URI":  SourceHTTPPath, // Full request URI including query string
	"QUERY_STRING": SourceHTTPGet,  // Raw query string
	"PATH_INFO":    SourceHTTPPath, // Extra path after script name
	"ORIG_PATH_INFO": SourceHTTPPath, // Original PATH_INFO before rewriting

	// HTTP Method and Content
	"REQUEST_METHOD": SourceHTTPHeader, // GET, POST, etc. (client chooses)
	"CONTENT_TYPE":   SourceHTTPHeader, // Content-Type header (client sets)
	"CONTENT_LENGTH": SourceHTTPHeader, // Content-Length header (client sets)

	// Authentication (user provides credentials)
	"PHP_AUTH_USER":   SourceHTTPHeader, // HTTP Basic Auth username
	"PHP_AUTH_PW":     SourceHTTPHeader, // HTTP Basic Auth password
	"PHP_AUTH_DIGEST": SourceHTTPHeader, // HTTP Digest Auth header
	"AUTH_TYPE":       SourceHTTPHeader, // Authentication type

	// ALL HTTP_* headers are user-controllable (client sends headers)
	"HTTP_HOST":            SourceHTTPHeader, // Host header (can be spoofed)
	"HTTP_USER_AGENT":      SourceHTTPHeader, // User-Agent header
	"HTTP_ACCEPT":          SourceHTTPHeader, // Accept header
	"HTTP_ACCEPT_LANGUAGE": SourceHTTPHeader, // Accept-Language header
	"HTTP_ACCEPT_ENCODING": SourceHTTPHeader, // Accept-Encoding header
	"HTTP_ACCEPT_CHARSET":  SourceHTTPHeader, // Accept-Charset header
	"HTTP_CONNECTION":      SourceHTTPHeader, // Connection header
	"HTTP_REFERER":         SourceHTTPHeader, // Referer header (commonly misspelled)
	"HTTP_COOKIE":          SourceHTTPCookie, // Raw cookie header
	"HTTP_AUTHORIZATION":   SourceHTTPHeader, // Authorization header
	"HTTP_CACHE_CONTROL":   SourceHTTPHeader, // Cache-Control header
	"HTTP_PRAGMA":          SourceHTTPHeader, // Pragma header
	"HTTP_IF_MODIFIED_SINCE": SourceHTTPHeader,
	"HTTP_IF_NONE_MATCH":     SourceHTTPHeader,
	"HTTP_X_FORWARDED_FOR":   SourceHTTPHeader, // Proxy header (spoofable)
	"HTTP_X_FORWARDED_HOST":  SourceHTTPHeader, // Proxy header (spoofable)
	"HTTP_X_FORWARDED_PROTO": SourceHTTPHeader, // Proxy header (spoofable)
	"HTTP_X_REQUESTED_WITH":  SourceHTTPHeader, // AJAX indicator
	"HTTP_X_REAL_IP":         SourceHTTPHeader, // Real IP header (spoofable)
	"HTTP_ORIGIN":            SourceHTTPHeader, // CORS Origin header

	// Remote client info (can be spoofed via proxies)
	"REMOTE_ADDR": SourceNetwork, // Client IP (spoofable via proxy)
	"REMOTE_HOST": SourceNetwork, // Client hostname
	"REMOTE_PORT": SourceNetwork, // Client port
}

// PHPServerConfigKeys are $_SERVER keys that contain SERVER CONFIGURATION data.
// These are NOT user-controllable and should NOT be marked as user input.
var PHPServerConfigKeys = []string{
	"DOCUMENT_ROOT",
	"SCRIPT_FILENAME",
	"SCRIPT_NAME",
	"SERVER_ADDR",
	"SERVER_NAME", // Can reflect HTTP_HOST if misconfigured, but is server config
	"SERVER_PORT",
	"SERVER_PROTOCOL",
	"SERVER_SOFTWARE",
	"SERVER_ADMIN",
	"GATEWAY_INTERFACE",
	"REQUEST_TIME",
	"REQUEST_TIME_FLOAT",
}

// IsServerKeyUserInput returns true if the $_SERVER key contains user-controllable data
func IsServerKeyUserInput(key string) bool {
	_, ok := PHPServerUserKeys[key]
	return ok
}

// GetServerKeySourceType returns the SourceType for a $_SERVER key
func GetServerKeySourceType(key string) SourceType {
	if st, ok := PHPServerUserKeys[key]; ok {
		return st
	}
	return SourceUnknown
}
