// Package main - exclusions.go defines methods to exclude from pattern generation
package main

// ExcludedMethods contains method names that don't return user input.
// These are utility methods, setters, or boolean checks without data access.
var ExcludedMethods = map[string]bool{
	// Boolean checks (don't return data)
	"has":           true,
	"hasAny":        true,
	"filled":        true,
	"isNotFilled":   true,
	"anyFilled":     true,
	"missing":       true,
	"isEmptyString": true,
	"exists":        true,
	"hasSession":    true,
	"isSecure":      true,
	"ajax":          true,
	"pjax":          true,
	"prefetch":      true,
	"wantsJson":     true,
	"acceptsAnyContentType": true,
	"acceptsJson":   true,
	"acceptsHtml":   true,
	"prefers":       true,

	// Request metadata (not user-controlled input)
	"method":       true,
	"path":         true,
	"decodedPath":  true,
	"url":          true,
	"fullUrl":      true,
	"fullUrlWithQuery": true,
	"fullUrlWithoutQuery": true,
	"root":         true,
	"route":        true,
	"routeIs":      true,
	"is":           true,
	"segment":      true,
	"segments":     true,
	"ip":           true,
	"ips":          true,
	"userAgent":    true,
	"fingerprint":  true,
	"host":         true,
	"httpHost":     true,
	"schemeAndHttpHost": true,

	// Setters/mutators
	"merge":           true,
	"mergeIfMissing":  true,
	"replace":         true,
	"set":             true,
	"add":             true,
	"remove":          true,
	"offsetSet":       true,
	"offsetUnset":     true,
	"offsetExists":    true,

	// Internal/magic methods
	"count":      true,
	"getIterator": true,
	"keys":       true, // Returns keys, not values

	// Symfony-specific non-input
	"getSession":       true,
	"hasPreviousSession": true,
	"isMethodSafe":     true,
	"isMethodIdempotent": true,
	"isMethodCacheable": true,
	"getProtocolVersion": true,
	"getContentType": true,
	"getContentTypeFormat": true,
	"getDefaultLocale": true,
	"getLocale":        true,
	"setLocale":        true,
	"getFormat":        true,
	"setFormat":        true,
	"getMimeType":      true,
	"getMimeTypes":     true,
	"isXmlHttpRequest": true,
	"preferSafeContent": true,
	"isFromTrustedProxy": true,
}

// IsExcluded returns true if the method should not generate a pattern.
func IsExcluded(methodName string) bool {
	return ExcludedMethods[methodName]
}
