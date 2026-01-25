// Package main - inference.go provides dynamic source type inference from method names
package main

import "strings"

// InferSourceType determines SourceType from method/property name.
// Uses context-aware approach: 'get' is generic, only specific names map to specific types.
func InferSourceType(name string) string {
	lower := strings.ToLower(name)

	// Exact matches for specific types (context-aware)
	switch lower {
	case "query":
		return "SourceHTTPGet"
	case "post":
		return "SourceHTTPPost"
	case "cookie", "cookies", "hascookie":
		return "SourceHTTPCookie"
	case "header", "headers", "hasheader", "bearertoken":
		return "SourceHTTPHeader"
	case "file", "files", "allfiles", "hasfile":
		return "SourceHTTPFile"
	case "json", "getpayload", "getcontent", "toarray":
		return "SourceHTTPBody"
	case "server":
		return "SourceEnvVar"
	case "session", "oldinput", "flash", "old", "getoldcollection":
		return "SourceSession"
	}

	// Partial matches for fallback
	switch {
	case strings.Contains(lower, "cookie"):
		return "SourceHTTPCookie"
	case strings.Contains(lower, "header"):
		return "SourceHTTPHeader"
	case strings.Contains(lower, "file"):
		return "SourceHTTPFile"
	case strings.Contains(lower, "flash"):
		return "SourceSession"
	default:
		return "SourceUserInput"
	}
}

// InferPopulatedFrom determines the PHP superglobals that populate this source type.
func InferPopulatedFrom(sourceType string) []string {
	switch sourceType {
	case "SourceHTTPGet":
		return []string{"$_GET"}
	case "SourceHTTPPost":
		return []string{"$_POST"}
	case "SourceHTTPCookie":
		return []string{"$_COOKIE"}
	case "SourceHTTPHeader", "SourceEnvVar":
		return []string{"$_SERVER"}
	case "SourceHTTPFile":
		return []string{"$_FILES"}
	case "SourceHTTPBody":
		return []string{}
	case "SourceSession":
		return []string{"$_SESSION"}
	default:
		return []string{"$_GET", "$_POST"}
	}
}

// InferDescription generates a description for the method based on framework and type.
func InferDescription(framework, methodName string, isProperty bool, sourceType string) string {
	var typeDesc string
	switch sourceType {
	case "SourceHTTPGet":
		typeDesc = "query string parameters"
	case "SourceHTTPPost":
		typeDesc = "POST data"
	case "SourceHTTPCookie":
		typeDesc = "cookie values"
	case "SourceHTTPHeader":
		typeDesc = "HTTP headers"
	case "SourceHTTPFile":
		typeDesc = "uploaded files"
	case "SourceHTTPBody":
		typeDesc = "request body"
	case "SourceEnvVar":
		typeDesc = "server variables"
	case "SourceSession":
		typeDesc = "session/flash data"
	default:
		typeDesc = "user input"
	}

	frameworkTitle := strings.Title(framework)
	if isProperty {
		return frameworkTitle + " $request->" + methodName + " contains " + typeDesc
	}
	return frameworkTitle + " $request->" + methodName + "() returns " + typeDesc
}
