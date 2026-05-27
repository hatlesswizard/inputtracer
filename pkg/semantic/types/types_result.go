package types

import "time"

// ============================================================================
// Backward Taint Analysis Types
// ============================================================================

// BackwardTraceResult represents the result of backward taint analysis,
// tracing from a target expression back to its input sources.
type BackwardTraceResult struct {
	// Target expression being traced
	TargetExpression string `json:"target_expression"`
	TargetFile       string `json:"target_file"`
	TargetLine       int    `json:"target_line"`

	// All paths from sources to this target
	Paths []BackwardPath `json:"paths"`

	// Summary of all sources found
	Sources []SourceInfo `json:"sources"`

	// Analysis metadata
	AnalyzedFiles int           `json:"analyzed_files"`
	Duration      time.Duration `json:"duration"`
}

// BackwardPath represents one path from a source to the target
type BackwardPath struct {
	// Source information
	Source SourceInfo `json:"source"`

	// Steps from source to target (in forward order for readability)
	Steps []BackwardStep `json:"steps"`

	// Whether path crosses file boundaries
	CrossFile bool `json:"cross_file"`
}

// BackwardStep represents one step in a backward trace path
type BackwardStep struct {
	StepNumber  int    `json:"step_number"`
	Expression  string `json:"expression"` // The code at this step
	FilePath    string `json:"file_path"`
	Line        int    `json:"line"`
	StepType    string `json:"step_type"` // "source", "assignment", "parameter", "return", "property"
	Description string `json:"description"`
}

// BatchTraceResult represents the result of batch backward taint analysis.
// Traces multiple target expressions in a SINGLE pass through the codebase
// for performance: reduces file reads from N*files to just files.
type BatchTraceResult struct {
	// Whether ANY variable traces back to user input
	HasUserInput bool `json:"has_user_input"`

	// Results for each variable traced
	PerVariable map[string]*BackwardTraceResult `json:"per_variable"`

	// Analysis metadata
	TotalDuration  time.Duration `json:"total_duration"`
	AnalyzedFiles  int           `json:"analyzed_files"`
	VariablesFound int           `json:"variables_found"`
}

// SourceInfo provides details about a discovered input source
type SourceInfo struct {
	Type       SourceType `json:"type"`       // http_get, http_post, etc.
	Expression string     `json:"expression"` // e.g., "$_GET['id']"
	FilePath   string     `json:"file_path"`
	Line       int        `json:"line"`
}

// ============================================================================
// Framework Knowledge Types
// ============================================================================

// FrameworkPattern defines a known framework input pattern
type FrameworkPattern struct {
	ID          string `json:"id"`
	Framework   string `json:"framework"`
	Language    string `json:"language"`
	Name        string `json:"name"`
	Description string `json:"description"`

	// Pattern matching
	ClassPattern    string `json:"class_pattern,omitempty"`    // Regex for class names
	MethodPattern   string `json:"method_pattern,omitempty"`   // Regex for method names
	PropertyPattern string `json:"property_pattern,omitempty"` // Regex for property names
	AccessPattern   string `json:"access_pattern,omitempty"`   // How data is accessed

	// Source mapping
	SourceType SourceType `json:"source_type"`
	SourceKey  string     `json:"source_key,omitempty"` // How to extract the key

	// Flow information
	CarrierClass    string   `json:"carrier_class,omitempty"`
	CarrierProperty string   `json:"carrier_property,omitempty"`
	PopulatedBy     string   `json:"populated_by,omitempty"`   // Method that populates
	PopulatedFrom   []string `json:"populated_from,omitempty"` // Original sources
}

// FrameworkPatternData is a plain data struct for importing patterns from pkg/sources,
// avoiding import cycles.
type FrameworkPatternData struct {
	ID              string
	Framework       string
	Language        string
	Name            string
	Description     string
	ClassPattern    string
	MethodPattern   string
	PropertyPattern string
	AccessPattern   string
	SourceType      string
	SourceKey       string
	CarrierClass    string
	CarrierProperty string
	PopulatedBy     string
	PopulatedFrom   []string
}

// ToFrameworkPattern converts a FrameworkPatternData into a FrameworkPattern
func (d *FrameworkPatternData) ToFrameworkPattern() *FrameworkPattern {
	return &FrameworkPattern{
		ID:              d.ID,
		Framework:       d.Framework,
		Language:        d.Language,
		Name:            d.Name,
		Description:     d.Description,
		ClassPattern:    d.ClassPattern,
		MethodPattern:   d.MethodPattern,
		PropertyPattern: d.PropertyPattern,
		AccessPattern:   d.AccessPattern,
		SourceType:      SourceType(d.SourceType),
		SourceKey:       d.SourceKey,
		CarrierClass:    d.CarrierClass,
		CarrierProperty: d.CarrierProperty,
		PopulatedBy:     d.PopulatedBy,
		PopulatedFrom:   d.PopulatedFrom,
	}
}
