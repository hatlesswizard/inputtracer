// Package java - annotations.go provides Java annotation to source type mappings
// This centralizes all annotation-based input source detection for Java frameworks
package java

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// AnnotationMapping maps an annotation name to its source type and metadata
type AnnotationMapping struct {
	SourceType  common.SourceType
	Framework   string
	Description string
}

// InputAnnotations maps Java annotation names to their source type mappings
// This is the canonical source for all Java input-related annotations
var InputAnnotations = map[string]AnnotationMapping{
	// ============================================================================
	// Spring MVC annotations
	// ============================================================================
	"RequestParam": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "spring",
		Description: "Spring request parameter from query string or form",
	},
	"PathVariable": {
		SourceType:  common.SourceHTTPPath,
		Framework:   "spring",
		Description: "Spring URL path variable",
	},
	"RequestBody": {
		SourceType:  common.SourceHTTPBody,
		Framework:   "spring",
		Description: "Spring request body (JSON/XML)",
	},
	"RequestHeader": {
		SourceType:  common.SourceHTTPHeader,
		Framework:   "spring",
		Description: "Spring HTTP header value",
	},
	"CookieValue": {
		SourceType:  common.SourceHTTPCookie,
		Framework:   "spring",
		Description: "Spring cookie value",
	},
	"ModelAttribute": {
		SourceType:  common.SourceHTTPPost,
		Framework:   "spring",
		Description: "Spring model attribute (form binding)",
	},
	"RequestPart": {
		SourceType:  common.SourceHTTPFile,
		Framework:   "spring",
		Description: "Spring multipart request part",
	},
	"MatrixVariable": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "spring",
		Description: "Spring matrix variable from URL",
	},

	// ============================================================================
	// JAX-RS annotations (Jersey, RESTEasy, Quarkus)
	// ============================================================================
	"QueryParam": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "jax-rs",
		Description: "JAX-RS query parameter",
	},
	"PathParam": {
		SourceType:  common.SourceHTTPPath,
		Framework:   "jax-rs",
		Description: "JAX-RS URL path parameter",
	},
	"FormParam": {
		SourceType:  common.SourceHTTPPost,
		Framework:   "jax-rs",
		Description: "JAX-RS form parameter",
	},
	"HeaderParam": {
		SourceType:  common.SourceHTTPHeader,
		Framework:   "jax-rs",
		Description: "JAX-RS HTTP header parameter",
	},
	"CookieParam": {
		SourceType:  common.SourceHTTPCookie,
		Framework:   "jax-rs",
		Description: "JAX-RS cookie parameter",
	},
	"BeanParam": {
		SourceType:  common.SourceUserInput,
		Framework:   "jax-rs",
		Description: "JAX-RS bean parameter (aggregates multiple sources)",
	},
	"MatrixParam": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "jax-rs",
		Description: "JAX-RS matrix parameter from URL",
	},

	// ============================================================================
	// Micronaut annotations
	// ============================================================================
	"QueryValue": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "micronaut",
		Description: "Micronaut query value",
	},
	"PathValue": {
		SourceType:  common.SourceHTTPPath,
		Framework:   "micronaut",
		Description: "Micronaut path value",
	},
	"Body": {
		SourceType:  common.SourceHTTPBody,
		Framework:   "micronaut",
		Description: "Micronaut request body",
	},
	"Header": {
		SourceType:  common.SourceHTTPHeader,
		Framework:   "micronaut",
		Description: "Micronaut HTTP header",
	},

	// ============================================================================
	// Vert.x annotations
	// ============================================================================
	"Param": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "vertx",
		Description: "Vert.x request parameter",
	},

	// ============================================================================
	// Struts 2 annotations
	// ============================================================================
	"StrutsParameter": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "struts2",
		Description: "Struts 2 action parameter",
	},

	// ============================================================================
	// Play Framework annotations
	// ============================================================================
	"BodyParser": {
		SourceType:  common.SourceHTTPBody,
		Framework:   "play",
		Description: "Play Framework body parser",
	},

	// ============================================================================
	// Dropwizard/Jersey specific (extends JAX-RS)
	// ============================================================================
	"Auth": {
		SourceType:  common.SourceHTTPHeader,
		Framework:   "dropwizard",
		Description: "Dropwizard authentication (often from header)",
	},
}

// GetAnnotationMapping returns the mapping for a given annotation name
// Returns nil if the annotation is not found
func GetAnnotationMapping(annotation string) *AnnotationMapping {
	if mapping, ok := InputAnnotations[annotation]; ok {
		return &mapping
	}
	return nil
}

// GetSourceTypeForAnnotation returns the SourceType for a given annotation
// Returns SourceUnknown if the annotation is not found
func GetSourceTypeForAnnotation(annotation string) common.SourceType {
	if mapping, ok := InputAnnotations[annotation]; ok {
		return mapping.SourceType
	}
	return common.SourceUnknown
}

// GetAnnotationsByFramework returns all annotations for a specific framework
func GetAnnotationsByFramework(framework string) map[string]AnnotationMapping {
	result := make(map[string]AnnotationMapping)
	for name, mapping := range InputAnnotations {
		if mapping.Framework == framework {
			result[name] = mapping
		}
	}
	return result
}

// GetAnnotationsBySourceType returns all annotations that map to a specific source type
func GetAnnotationsBySourceType(sourceType common.SourceType) []string {
	var result []string
	for name, mapping := range InputAnnotations {
		if mapping.SourceType == sourceType {
			result = append(result, name)
		}
	}
	return result
}

// IsInputAnnotation returns true if the annotation name indicates user input
func IsInputAnnotation(annotation string) bool {
	_, ok := InputAnnotations[annotation]
	return ok
}

// GetAllAnnotationNames returns a list of all known input annotation names
func GetAllAnnotationNames() []string {
	names := make([]string, 0, len(InputAnnotations))
	for name := range InputAnnotations {
		names = append(names, name)
	}
	return names
}

// GetAllFrameworks returns a list of all frameworks with annotation mappings
func GetAllFrameworks() []string {
	frameworks := make(map[string]bool)
	for _, mapping := range InputAnnotations {
		frameworks[mapping.Framework] = true
	}
	result := make([]string, 0, len(frameworks))
	for fw := range frameworks {
		result = append(result, fw)
	}
	return result
}
