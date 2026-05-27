// Package php provides PHP function patterns for input source detection
package php

import (
	"strings"

	"github.com/hatlesswizard/inputtracer/pkg/sources/common"
)

// =============================================================================
// INPUT FUNCTIONS
// Functions that read data from files, environment, or network
// =============================================================================

// InputFunctions are functions that read external data
// These are sources of potentially untrusted data
var InputFunctions = []string{
	// File reading
	"file_get_contents",
	"fgets",
	"fread",
	"fgetc",
	"fgetss",
	"fgetcsv",
	"file",
	"readfile",
	"stream_get_contents",
	// Environment/headers
	"getenv",
	"getallheaders",
	"apache_request_headers",
	// Other input
	"readline",
	"fscanf",
	"fpassthru",
}

// InputFunctionsMap provides O(1) lookup
var InputFunctionsMap = func() map[string]bool {
	m := make(map[string]bool)
	for _, fn := range InputFunctions {
		m[fn] = true
	}
	return m
}()

// IsInputFunction returns true if the function name is a known input function
func IsInputFunction(funcName string) bool {
	return InputFunctionsMap[funcName]
}

// ContainsInputFunction checks if expression contains any input function call
func ContainsInputFunction(expr string) bool {
	for _, fn := range InputFunctions {
		if strings.Contains(expr, fn+"(") {
			return true
		}
	}
	return false
}

// GetInputFunctionSourceType returns the source type for an input function
func GetInputFunctionSourceType(funcName string) common.SourceType {
	switch funcName {
	case "getenv":
		return common.SourceEnvVar
	case "getallheaders", "apache_request_headers":
		return common.SourceHTTPHeader
	case "readline":
		return common.SourceStdin
	default:
		return common.SourceFile
	}
}

// =============================================================================
// DESERIALIZATION FUNCTIONS
// Functions that deserialize data - potential sources of tainted data
// =============================================================================

// DeserializationFunctions are functions that deserialize external data
// The data being deserialized may come from untrusted sources
var DeserializationFunctions = []string{
	"unserialize",
	"json_decode",
	"simplexml_load_string",
	"simplexml_load_file",
	"yaml_parse",
	"yaml_parse_file",
	"yaml_parse_url",
	"msgpack_unpack",
	"igbinary_unserialize",
	"parse_str", // populates variables from query string
	"mb_parse_str",
}

// DeserializationFunctionsMap provides O(1) lookup
var DeserializationFunctionsMap = func() map[string]bool {
	m := make(map[string]bool)
	for _, fn := range DeserializationFunctions {
		m[fn] = true
	}
	return m
}()

// IsDeserializationFunction returns true if the function deserializes data
func IsDeserializationFunction(funcName string) bool {
	return DeserializationFunctionsMap[funcName]
}

// ContainsDeserializationFunction checks if expression contains deserialization
func ContainsDeserializationFunction(expr string) bool {
	for _, fn := range DeserializationFunctions {
		if strings.Contains(expr, fn+"(") {
			return true
		}
	}
	return false
}

// =============================================================================
// NETWORK FUNCTIONS
// Functions that fetch data from network sources
// =============================================================================

// NetworkFunctions are functions that fetch external network data
var NetworkFunctions = []string{
	// cURL
	"curl_exec",
	"curl_multi_getcontent",
	"curl_multi_exec",
	// File/URL fetching
	"file_get_contents", // Can fetch URLs
	"fopen",             // Can open URLs
	"fsockopen",
	"pfsockopen",
	// HTTP specific
	"http_get",
	"http_post",
	"http_request",
	// Socket
	"socket_read",
	"socket_recv",
	"socket_recvfrom",
	// Stream
	"stream_socket_recvfrom",
	"stream_get_contents",
}

// NetworkFunctionsMap provides O(1) lookup
var NetworkFunctionsMap = func() map[string]bool {
	m := make(map[string]bool)
	for _, fn := range NetworkFunctions {
		m[fn] = true
	}
	return m
}()

// IsNetworkFunction returns true if the function fetches network data
func IsNetworkFunction(funcName string) bool {
	return NetworkFunctionsMap[funcName]
}

// ContainsNetworkFunction checks if expression contains network function call
func ContainsNetworkFunction(expr string) bool {
	for _, fn := range NetworkFunctions {
		if strings.Contains(expr, fn+"(") {
			return true
		}
	}
	return false
}

// ContainsCurlFunction specifically checks for cURL functions
func ContainsCurlFunction(expr string) bool {
	return strings.Contains(expr, "curl_exec(") ||
		strings.Contains(expr, "curl_multi_getcontent(") ||
		strings.Contains(expr, "curl_multi_exec(")
}

// =============================================================================
// COMBINED CHECKS
// =============================================================================

// IsExternalDataFunction checks if a function reads external data (any type)
func IsExternalDataFunction(funcName string) bool {
	return IsInputFunction(funcName) ||
		IsDeserializationFunction(funcName) ||
		IsNetworkFunction(funcName)
}

// IdentifyExternalDataSource identifies the source type from an expression
// Returns the source type and confidence level
func IdentifyExternalDataSource(expr string) (common.SourceType, float64) {
	// Check network functions first (more specific)
	if ContainsCurlFunction(expr) {
		return common.SourceNetwork, 0.8
	}
	if ContainsNetworkFunction(expr) {
		return common.SourceNetwork, 0.75
	}

	// Check deserialization
	if ContainsDeserializationFunction(expr) {
		return common.SourceUserInput, 0.85
	}

	// Check input functions
	for _, fn := range InputFunctions {
		if strings.Contains(expr, fn+"(") {
			return GetInputFunctionSourceType(fn), 0.9
		}
	}

	return common.SourceUnknown, 0.0
}
