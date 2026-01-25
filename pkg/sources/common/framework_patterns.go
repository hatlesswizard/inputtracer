// Package common - framework_patterns.go provides framework pattern definitions
// All framework-specific patterns should be defined using these types
package common

// FrameworkPattern defines a framework-specific input source pattern
// This is the centralized definition - all language analyzers should use this
type FrameworkPattern struct {
	ID          string `json:"id"`
	Framework   string `json:"framework"`
	Language    string `json:"language"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// Pattern matching (regex strings)
	ClassPattern    string `json:"class_pattern,omitempty"`    // Regex for class names
	MethodPattern   string `json:"method_pattern,omitempty"`   // Regex for method names
	PropertyPattern string `json:"property_pattern,omitempty"` // Regex for property names
	AccessPattern   string `json:"access_pattern,omitempty"`   // How data is accessed: "array", "method", "property", "superglobal"

	// Source mapping
	SourceType SourceType `json:"source_type"`
	SourceKey  string     `json:"source_key,omitempty"` // How to extract the key

	// Flow information (for carrier tracking)
	CarrierClass    string   `json:"carrier_class,omitempty"`
	CarrierProperty string   `json:"carrier_property,omitempty"`
	PopulatedBy     string   `json:"populated_by,omitempty"`   // Method that populates the carrier
	PopulatedFrom   []string `json:"populated_from,omitempty"` // Original sources (e.g., ["$_GET", "$_POST"])

	// Confidence score (0.0 to 1.0)
	Confidence float64 `json:"confidence"`

	// Tags for categorization
	Tags []string `json:"tags,omitempty"`
}

// FrameworkPatternRegistry manages framework patterns for a language
type FrameworkPatternRegistry struct {
	language     string
	patterns     []*FrameworkPattern
	byID         map[string]*FrameworkPattern
	byFramework  map[string][]*FrameworkPattern
}

// NewFrameworkPatternRegistry creates a new registry for a language
func NewFrameworkPatternRegistry(language string) *FrameworkPatternRegistry {
	return &FrameworkPatternRegistry{
		language:    language,
		patterns:    make([]*FrameworkPattern, 0),
		byID:        make(map[string]*FrameworkPattern),
		byFramework: make(map[string][]*FrameworkPattern),
	}
}

// Register adds a pattern to the registry
func (r *FrameworkPatternRegistry) Register(pattern *FrameworkPattern) {
	if pattern == nil {
		return
	}
	// Ensure language is set
	if pattern.Language == "" {
		pattern.Language = r.language
	}
	r.patterns = append(r.patterns, pattern)
	if pattern.ID != "" {
		r.byID[pattern.ID] = pattern
	}
	if pattern.Framework != "" {
		r.byFramework[pattern.Framework] = append(r.byFramework[pattern.Framework], pattern)
	}
}

// RegisterAll adds multiple patterns to the registry
func (r *FrameworkPatternRegistry) RegisterAll(patterns []*FrameworkPattern) {
	for _, p := range patterns {
		r.Register(p)
	}
}

// GetAll returns all registered patterns
func (r *FrameworkPatternRegistry) GetAll() []*FrameworkPattern {
	return r.patterns
}

// GetByID returns a pattern by its ID
func (r *FrameworkPatternRegistry) GetByID(id string) *FrameworkPattern {
	return r.byID[id]
}

// GetByFramework returns all patterns for a specific framework
func (r *FrameworkPatternRegistry) GetByFramework(framework string) []*FrameworkPattern {
	return r.byFramework[framework]
}

// GetFrameworks returns a list of all registered frameworks
func (r *FrameworkPatternRegistry) GetFrameworks() []string {
	frameworks := make([]string, 0, len(r.byFramework))
	for fw := range r.byFramework {
		frameworks = append(frameworks, fw)
	}
	return frameworks
}

// Count returns the total number of registered patterns
func (r *FrameworkPatternRegistry) Count() int {
	return len(r.patterns)
}

// FrameworkDetector defines file path indicators for framework detection
type FrameworkDetector struct {
	Framework  string   `json:"framework"`
	Indicators []string `json:"indicators"` // File paths that indicate this framework
}

// Global registry of framework detectors
var frameworkDetectors = make(map[string]*FrameworkDetector)

// RegisterFrameworkDetector registers a framework detector
func RegisterFrameworkDetector(detector *FrameworkDetector) {
	if detector != nil && detector.Framework != "" {
		frameworkDetectors[detector.Framework] = detector
	}
}

// GetFrameworkDetector returns a framework detector by name
func GetFrameworkDetector(framework string) *FrameworkDetector {
	return frameworkDetectors[framework]
}

// GetAllFrameworkDetectors returns all registered detectors
func GetAllFrameworkDetectors() map[string]*FrameworkDetector {
	return frameworkDetectors
}
