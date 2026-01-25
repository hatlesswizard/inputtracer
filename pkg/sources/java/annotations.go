// Package java - annotations.go provides Java annotation to source type mappings
// This centralizes all annotation-based input source detection for Java frameworks
package java

import "github.com/hatlesswizard/inputtracer/pkg/sources/common"

// AnnotationMapping maps an annotation name to its source type and metadata
type AnnotationMapping struct {
	SourceType  common.SourceType
	Framework   string
	Description string
	Confidence  float64
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
		Confidence:  0.95,
	},
	"PathVariable": {
		SourceType:  common.SourceHTTPPath,
		Framework:   "spring",
		Description: "Spring URL path variable",
		Confidence:  0.95,
	},
	"RequestBody": {
		SourceType:  common.SourceHTTPBody,
		Framework:   "spring",
		Description: "Spring request body (JSON/XML)",
		Confidence:  0.95,
	},
	"RequestHeader": {
		SourceType:  common.SourceHTTPHeader,
		Framework:   "spring",
		Description: "Spring HTTP header value",
		Confidence:  0.95,
	},
	"CookieValue": {
		SourceType:  common.SourceHTTPCookie,
		Framework:   "spring",
		Description: "Spring cookie value",
		Confidence:  0.95,
	},
	"ModelAttribute": {
		SourceType:  common.SourceHTTPPost,
		Framework:   "spring",
		Description: "Spring model attribute (form binding)",
		Confidence:  0.95,
	},
	"RequestPart": {
		SourceType:  common.SourceHTTPFile,
		Framework:   "spring",
		Description: "Spring multipart request part",
		Confidence:  0.95,
	},
	"MatrixVariable": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "spring",
		Description: "Spring matrix variable from URL",
		Confidence:  0.95,
	},

	// ============================================================================
	// JAX-RS annotations (Jersey, RESTEasy, Quarkus)
	// ============================================================================
	"QueryParam": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "jax-rs",
		Description: "JAX-RS query parameter",
		Confidence:  0.95,
	},
	"PathParam": {
		SourceType:  common.SourceHTTPPath,
		Framework:   "jax-rs",
		Description: "JAX-RS URL path parameter",
		Confidence:  0.95,
	},
	"FormParam": {
		SourceType:  common.SourceHTTPPost,
		Framework:   "jax-rs",
		Description: "JAX-RS form parameter",
		Confidence:  0.95,
	},
	"HeaderParam": {
		SourceType:  common.SourceHTTPHeader,
		Framework:   "jax-rs",
		Description: "JAX-RS HTTP header parameter",
		Confidence:  0.95,
	},
	"CookieParam": {
		SourceType:  common.SourceHTTPCookie,
		Framework:   "jax-rs",
		Description: "JAX-RS cookie parameter",
		Confidence:  0.95,
	},
	"BeanParam": {
		SourceType:  common.SourceUserInput,
		Framework:   "jax-rs",
		Description: "JAX-RS bean parameter (aggregates multiple sources)",
		Confidence:  0.9,
	},
	"MatrixParam": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "jax-rs",
		Description: "JAX-RS matrix parameter from URL",
		Confidence:  0.95,
	},

	// ============================================================================
	// Micronaut annotations
	// ============================================================================
	"QueryValue": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "micronaut",
		Description: "Micronaut query value",
		Confidence:  0.95,
	},
	"PathValue": {
		SourceType:  common.SourceHTTPPath,
		Framework:   "micronaut",
		Description: "Micronaut path value",
		Confidence:  0.95,
	},
	"Body": {
		SourceType:  common.SourceHTTPBody,
		Framework:   "micronaut",
		Description: "Micronaut request body",
		Confidence:  0.95,
	},
	"Header": {
		SourceType:  common.SourceHTTPHeader,
		Framework:   "micronaut",
		Description: "Micronaut HTTP header",
		Confidence:  0.95,
	},

	// ============================================================================
	// Vert.x annotations
	// ============================================================================
	"Param": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "vertx",
		Description: "Vert.x request parameter",
		Confidence:  0.9,
	},

	// ============================================================================
	// Struts 2 annotations
	// ============================================================================
	"StrutsParameter": {
		SourceType:  common.SourceHTTPGet,
		Framework:   "struts2",
		Description: "Struts 2 action parameter",
		Confidence:  0.9,
	},

	// ============================================================================
	// Play Framework annotations
	// ============================================================================
	"BodyParser": {
		SourceType:  common.SourceHTTPBody,
		Framework:   "play",
		Description: "Play Framework body parser",
		Confidence:  0.9,
	},

	// ============================================================================
	// Dropwizard/Jersey specific (extends JAX-RS)
	// ============================================================================
	"Auth": {
		SourceType:  common.SourceHTTPHeader,
		Framework:   "dropwizard",
		Description: "Dropwizard authentication (often from header)",
		Confidence:  0.85,
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
