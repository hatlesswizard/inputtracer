// Package sources - input_methods.go provides centralized input method definitions
// All framework input methods and patterns should be defined here
package sources

// InputMethodCategory classifies input method types
type InputMethodCategory string

const (
	CategoryHTTP    InputMethodCategory = "http"
	CategoryFile    InputMethodCategory = "file"
	CategoryCommand InputMethodCategory = "command"
	CategoryGeneric InputMethodCategory = "generic"
)

// InputMethod describes a method that returns user input
type InputMethod struct {
	VarPattern  string              // e.g., "mybb", "request", "*" for any
	MethodName  string              // e.g., "get_input", "variable"
	Category    InputMethodCategory // http, database, file, command
	SourceType  SourceType          // Mapped source type
	Framework   string              // e.g., "mybb", "laravel", "generic"
	Description string              // Human-readable description
}

// InputMethods is the canonical list of input-returning methods
// Replaces hardcoded patterns in extractor.go isInputMethod() and isInterestingMethod()
var InputMethods = []InputMethod{
	// MyBB specific patterns
	{VarPattern: "mybb", MethodName: "get_input", Category: CategoryHTTP, SourceType: SourceUserInput, Framework: "mybb", Description: "MyBB input getter"},
	{VarPattern: "mybb", MethodName: "get_input_array", Category: CategoryHTTP, SourceType: SourceUserInput, Framework: "mybb", Description: "MyBB array input getter"},
	{VarPattern: "mybb", MethodName: "input", Category: CategoryHTTP, SourceType: SourceUserInput, Framework: "mybb", Description: "MyBB input property"},
	{VarPattern: "mybb", MethodName: "cookies", Category: CategoryHTTP, SourceType: SourceHTTPCookie, Framework: "mybb", Description: "MyBB cookies property"},

	// phpBB patterns
	{VarPattern: "request", MethodName: "variable", Category: CategoryHTTP, SourceType: SourceUserInput, Framework: "phpbb", Description: "phpBB request variable"},

	// Generic request patterns
	{VarPattern: "request", MethodName: "get", Category: CategoryHTTP, SourceType: SourceHTTPGet, Framework: "generic", Description: "Generic GET getter"},
	{VarPattern: "request", MethodName: "post", Category: CategoryHTTP, SourceType: SourceHTTPPost, Framework: "generic", Description: "Generic POST getter"},
	{VarPattern: "request", MethodName: "cookie", Category: CategoryHTTP, SourceType: SourceHTTPCookie, Framework: "generic", Description: "Generic cookie getter"},
	{VarPattern: "request", MethodName: "server", Category: CategoryHTTP, SourceType: SourceHTTPHeader, Framework: "generic", Description: "Generic server var getter"},
	{VarPattern: "request", MethodName: "header", Category: CategoryHTTP, SourceType: SourceHTTPHeader, Framework: "generic", Description: "Generic header getter"},

	// Generic patterns (any object)
	{VarPattern: "*", MethodName: "get", Category: CategoryHTTP, SourceType: SourceHTTPGet, Framework: "generic", Description: "Generic GET method"},
	{VarPattern: "*", MethodName: "post", Category: CategoryHTTP, SourceType: SourceHTTPPost, Framework: "generic", Description: "Generic POST method"},
	{VarPattern: "*", MethodName: "cookie", Category: CategoryHTTP, SourceType: SourceHTTPCookie, Framework: "generic", Description: "Generic cookie method"},
	{VarPattern: "*", MethodName: "header", Category: CategoryHTTP, SourceType: SourceHTTPHeader, Framework: "generic", Description: "Generic header method"},
	{VarPattern: "*", MethodName: "param", Category: CategoryHTTP, SourceType: SourceUserInput, Framework: "generic", Description: "Generic param method"},
	{VarPattern: "*", MethodName: "input", Category: CategoryHTTP, SourceType: SourceUserInput, Framework: "generic", Description: "Generic input method"},
	{VarPattern: "*", MethodName: "request", Category: CategoryHTTP, SourceType: SourceUserInput, Framework: "generic", Description: "Generic request method"},
	{VarPattern: "*", MethodName: "query", Category: CategoryHTTP, SourceType: SourceHTTPGet, Framework: "generic", Description: "Generic query method"},

	// File operations
	{VarPattern: "*", MethodName: "read", Category: CategoryFile, SourceType: SourceFile, Framework: "generic", Description: "File read"},
	{VarPattern: "*", MethodName: "write", Category: CategoryFile, SourceType: SourceFile, Framework: "generic", Description: "File write"},
	{VarPattern: "*", MethodName: "file_get_contents", Category: CategoryFile, SourceType: SourceFile, Framework: "generic", Description: "Get file contents"},
	{VarPattern: "*", MethodName: "include", Category: CategoryFile, SourceType: SourceFile, Framework: "generic", Description: "File include"},
	{VarPattern: "*", MethodName: "require", Category: CategoryFile, SourceType: SourceFile, Framework: "generic", Description: "File require"},
	{VarPattern: "*", MethodName: "fopen", Category: CategoryFile, SourceType: SourceFile, Framework: "generic", Description: "Open file"},

	// Command execution
	{VarPattern: "*", MethodName: "exec", Category: CategoryCommand, SourceType: SourceUserInput, Framework: "generic", Description: "Command exec"},
	{VarPattern: "*", MethodName: "shell_exec", Category: CategoryCommand, SourceType: SourceUserInput, Framework: "generic", Description: "Shell exec"},
	{VarPattern: "*", MethodName: "system", Category: CategoryCommand, SourceType: SourceUserInput, Framework: "generic", Description: "System call"},
	{VarPattern: "*", MethodName: "passthru", Category: CategoryCommand, SourceType: SourceUserInput, Framework: "generic", Description: "Passthru"},
}

// inputMethodIndex is a lookup map for fast access
var inputMethodIndex map[string]map[string]*InputMethod

func init() {
	// Build the index
	inputMethodIndex = make(map[string]map[string]*InputMethod)
	for i := range InputMethods {
		im := &InputMethods[i]
		if inputMethodIndex[im.VarPattern] == nil {
			inputMethodIndex[im.VarPattern] = make(map[string]*InputMethod)
		}
		inputMethodIndex[im.VarPattern][im.MethodName] = im
	}
}

// IsInputMethod checks if a var.method combination is a known input method
// Replaces isInputMethod() in extractor.go
func IsInputMethod(varName, methodName string) bool {
	// Check exact var match first
	if methods, ok := inputMethodIndex[varName]; ok {
		if _, found := methods[methodName]; found {
			return true
		}
	}
	// Check wildcard patterns
	if methods, ok := inputMethodIndex["*"]; ok {
		if im, found := methods[methodName]; found {
			// For wildcard, only return true for HTTP category (input sources)
			return im.Category == CategoryHTTP
		}
	}
	return false
}

// GetInputMethodInfo returns full info for a var.method combination
func GetInputMethodInfo(varName, methodName string) *InputMethod {
	// Check exact var match first
	if methods, ok := inputMethodIndex[varName]; ok {
		if im, found := methods[methodName]; found {
			return im
		}
	}
	// Check wildcard patterns
	if methods, ok := inputMethodIndex["*"]; ok {
		if im, found := methods[methodName]; found {
			return im
		}
	}
	return nil
}

// IsInterestingMethod checks if a method name is security-relevant
// Note: This library traces INPUT SOURCES only, not sinks
// This function is kept for compatibility but only returns true for file/command operations
func IsInterestingMethod(methodName string) bool {
	if methods, ok := inputMethodIndex["*"]; ok {
		if im, found := methods[methodName]; found {
			return im.Category == CategoryFile ||
				im.Category == CategoryCommand
		}
	}
	return false
}

// GetMethodsByCategory returns all methods in a category
func GetMethodsByCategory(category InputMethodCategory) []InputMethod {
	var result []InputMethod
	for _, im := range InputMethods {
		if im.Category == category {
			result = append(result, im)
		}
	}
	return result
}

// GetMethodsByFramework returns all methods for a framework
func GetMethodsByFramework(framework string) []InputMethod {
	var result []InputMethod
	for _, im := range InputMethods {
		if im.Framework == framework {
			result = append(result, im)
		}
	}
	return result
}

// GetHTTPInputMethods returns all HTTP input methods
func GetHTTPInputMethods() []InputMethod {
	return GetMethodsByCategory(CategoryHTTP)
}


// GetFileMethods returns all file-related methods
func GetFileMethods() []InputMethod {
	return GetMethodsByCategory(CategoryFile)
}

// GetCommandMethods returns all command execution methods
func GetCommandMethods() []InputMethod {
	return GetMethodsByCategory(CategoryCommand)
}

// MethodToSuperglobals maps common method names to their typical superglobal sources
// Used for PHP input method resolution in tracing
var MethodToSuperglobals = map[string][]string{
	"get_input": {"$_GET", "$_POST"},
	"input":     {"$_GET", "$_POST"},
	"get":       {"$_GET"},
	"post":      {"$_POST"},
	"cookie":    {"$_COOKIE"},
	"server":    {"$_SERVER"},
	"file":      {"$_FILES"},
}

// GetSuperglobalsForMethod returns the superglobals typically accessed by a method name
func GetSuperglobalsForMethod(methodName string) []string {
	if sgs, ok := MethodToSuperglobals[methodName]; ok {
		return sgs
	}
	return nil
}
