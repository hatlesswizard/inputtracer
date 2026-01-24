// Package mappings provides centralized input source mappings for all supported languages.
// This package consolidates function-to-source-type mappings that were previously scattered
// across individual language analyzers.
package mappings

import (
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
)

// FrameworkTypeInfo holds information about framework types that carry user input
type FrameworkTypeInfo struct {
	Framework  string
	SourceType types.SourceType
}

// LanguageMappings holds all input source mappings for a single language
type LanguageMappings struct {
	// Language identifier (e.g., "go", "php", "javascript")
	Language string

	// InputFunctions maps function/method names to source types
	// Examples: "getenv" -> SourceEnvVar, "fgets" -> SourceFile
	InputFunctions map[string]types.SourceType

	// InputSources maps property/variable access patterns to source types
	// Examples: "req.body" -> SourceHTTPBody, "process.env" -> SourceEnvVar
	InputSources map[string]types.SourceType

	// Superglobals maps superglobal variables to source types (PHP specific)
	// Examples: "$_GET" -> SourceHTTPGet, "$_POST" -> SourceHTTPPost
	Superglobals map[string]types.SourceType

	// DBFetchFunctions maps database fetch function names (PHP specific)
	// Used to identify functions that return user-controlled data from DB
	DBFetchFunctions map[string]bool

	// GlobalSources maps browser global sources (JS specific)
	// Examples: "location.href" -> SourceHTTPGet, "document.cookie" -> SourceHTTPCookie
	GlobalSources map[string]types.SourceType

	// DOMSources maps DOM property sources (JS specific)
	// Examples: "value" -> SourceUserInput, "innerHTML" -> SourceUserInput
	DOMSources map[string]types.SourceType

	// NodeSources maps Node.js-specific sources (JS specific)
	// Examples: "process.argv" -> SourceCLIArg, "process.env" -> SourceEnvVar
	NodeSources map[string]types.SourceType

	// CGIEnvVars maps CGI environment variables to source types (C/C++ specific)
	// Examples: "QUERY_STRING" -> SourceHTTPGet, "HTTP_COOKIE" -> SourceHTTPCookie
	CGIEnvVars map[string]types.SourceType

	// QtInputMethods maps Qt widget methods to source types (C++ specific)
	// Examples: "text" -> SourceUserInput, "toPlainText" -> SourceUserInput
	QtInputMethods map[string]types.SourceType

	// FrameworkTypes maps framework type names to their info (C++ specific)
	// Examples: "QNetworkReply" -> {Framework: "qt", SourceType: SourceNetwork}
	FrameworkTypes map[string]FrameworkTypeInfo

	// MethodInputs maps method names that return user input (C++ specific)
	// Examples: "body" -> SourceHTTPBody, "getParameter" -> SourceHTTPGet
	MethodInputs map[string]types.SourceType

	// Annotations maps annotation/decorator names to source types (Java/C# specific)
	// Examples: "RequestParam" -> SourceHTTPGet, "RequestBody" -> SourceHTTPBody
	Annotations map[string]types.SourceType

	// InputMethods maps input method names to source types (Java specific)
	// Examples: "getParameter" -> SourceHTTPGet, "getHeader" -> SourceHTTPHeader
	InputMethods map[string]types.SourceType
}

// Registry holds all language mappings keyed by language name
var Registry = make(map[string]*LanguageMappings)

// GetMappings returns the mappings for a specific language, or nil if not found
func GetMappings(language string) *LanguageMappings {
	return Registry[language]
}

// GetInputFunctionsMap returns InputFunctions, never nil
func (lm *LanguageMappings) GetInputFunctionsMap() map[string]types.SourceType {
	if lm.InputFunctions == nil {
		return make(map[string]types.SourceType)
	}
	return lm.InputFunctions
}

// GetInputSourcesMap returns InputSources, never nil
func (lm *LanguageMappings) GetInputSourcesMap() map[string]types.SourceType {
	if lm.InputSources == nil {
		return make(map[string]types.SourceType)
	}
	return lm.InputSources
}

// GetSuperglobalsMap returns Superglobals, never nil
func (lm *LanguageMappings) GetSuperglobalsMap() map[string]types.SourceType {
	if lm.Superglobals == nil {
		return make(map[string]types.SourceType)
	}
	return lm.Superglobals
}

// GetDBFetchFunctionsMap returns DBFetchFunctions, never nil
func (lm *LanguageMappings) GetDBFetchFunctionsMap() map[string]bool {
	if lm.DBFetchFunctions == nil {
		return make(map[string]bool)
	}
	return lm.DBFetchFunctions
}

// GetGlobalSourcesMap returns GlobalSources, never nil
func (lm *LanguageMappings) GetGlobalSourcesMap() map[string]types.SourceType {
	if lm.GlobalSources == nil {
		return make(map[string]types.SourceType)
	}
	return lm.GlobalSources
}

// GetDOMSourcesMap returns DOMSources, never nil
func (lm *LanguageMappings) GetDOMSourcesMap() map[string]types.SourceType {
	if lm.DOMSources == nil {
		return make(map[string]types.SourceType)
	}
	return lm.DOMSources
}

// GetNodeSourcesMap returns NodeSources, never nil
func (lm *LanguageMappings) GetNodeSourcesMap() map[string]types.SourceType {
	if lm.NodeSources == nil {
		return make(map[string]types.SourceType)
	}
	return lm.NodeSources
}

// GetCGIEnvVarsMap returns CGIEnvVars, never nil
func (lm *LanguageMappings) GetCGIEnvVarsMap() map[string]types.SourceType {
	if lm.CGIEnvVars == nil {
		return make(map[string]types.SourceType)
	}
	return lm.CGIEnvVars
}

// GetQtInputMethodsMap returns QtInputMethods, never nil
func (lm *LanguageMappings) GetQtInputMethodsMap() map[string]types.SourceType {
	if lm.QtInputMethods == nil {
		return make(map[string]types.SourceType)
	}
	return lm.QtInputMethods
}

// GetFrameworkTypesMap returns FrameworkTypes, never nil
func (lm *LanguageMappings) GetFrameworkTypesMap() map[string]FrameworkTypeInfo {
	if lm.FrameworkTypes == nil {
		return make(map[string]FrameworkTypeInfo)
	}
	return lm.FrameworkTypes
}

// GetMethodInputsMap returns MethodInputs, never nil
func (lm *LanguageMappings) GetMethodInputsMap() map[string]types.SourceType {
	if lm.MethodInputs == nil {
		return make(map[string]types.SourceType)
	}
	return lm.MethodInputs
}

// GetAnnotationsMap returns Annotations, never nil
func (lm *LanguageMappings) GetAnnotationsMap() map[string]types.SourceType {
	if lm.Annotations == nil {
		return make(map[string]types.SourceType)
	}
	return lm.Annotations
}

// GetInputMethodsMap returns InputMethods, never nil
func (lm *LanguageMappings) GetInputMethodsMap() map[string]types.SourceType {
	if lm.InputMethods == nil {
		return make(map[string]types.SourceType)
	}
	return lm.InputMethods
}

// MergeMaps combines multiple source type maps into one
// Later maps override earlier ones for duplicate keys
func MergeMaps(maps ...map[string]types.SourceType) map[string]types.SourceType {
	result := make(map[string]types.SourceType)
	for _, m := range maps {
		for k, v := range m {
			result[k] = v
		}
	}
	return result
}
