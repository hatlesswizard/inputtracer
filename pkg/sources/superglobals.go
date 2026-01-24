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
var PHPSuperglobals = []PHPSuperglobal{
	{"$_GET", SourceHTTPGet, []InputLabel{LabelHTTPGet, LabelUserInput}, "HTTP GET parameters"},
	{"$_POST", SourceHTTPPost, []InputLabel{LabelHTTPPost, LabelUserInput}, "HTTP POST parameters"},
	{"$_REQUEST", SourceHTTPRequest, []InputLabel{LabelHTTPGet, LabelHTTPPost, LabelUserInput}, "Combined GET/POST/COOKIE"},
	{"$_COOKIE", SourceHTTPCookie, []InputLabel{LabelHTTPCookie, LabelUserInput}, "HTTP cookies"},
	{"$_SERVER", SourceHTTPHeader, []InputLabel{LabelHTTPHeader, LabelUserInput}, "Server and request info"},
	{"$_FILES", SourceHTTPFile, []InputLabel{LabelFile, LabelUserInput}, "Uploaded files"},
	{"$_ENV", SourceEnvVar, []InputLabel{LabelEnvironment}, "Environment variables"},
	{"$_SESSION", SourceSession, []InputLabel{LabelUserInput}, "Session data"},
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
