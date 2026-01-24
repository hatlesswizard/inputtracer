// Package classifier provides snippet classification using carrier maps
package classifier

import (
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/hatlesswizard/inputtracer/pkg/semantic/discovery"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/extractor"
	"github.com/hatlesswizard/inputtracer/pkg/sources"
)

// Classifier classifies code snippets for user input
type Classifier struct {
	carrierMap     *discovery.CarrierMap
	extractor      *extractor.ExpressionExtractor
	carrierPatterns []*carrierPattern
}

// carrierPattern is a compiled pattern for matching carriers
type carrierPattern struct {
	carrier  *discovery.InputCarrier
	pattern  *regexp.Regexp
	isMethod bool
}

// ClassificationResult contains the complete analysis of a snippet
type ClassificationResult struct {
	Snippet        string              `json:"snippet"`
	HasUserInput   bool                `json:"has_user_input"`
	NeedsTracing   bool                `json:"needs_tracing"`
	Expressions    []ExpressionResult  `json:"expressions"`
	Summary        ClassificationSummary `json:"summary"`
}

// ExpressionResult contains the classification of a single expression
type ExpressionResult struct {
	Expression     string   `json:"expression"`
	SourceTypes    []string `json:"source_types,omitempty"`
	Key            string   `json:"key,omitempty"`
	Confidence     float64  `json:"confidence"`
	MatchedCarrier string   `json:"matched_carrier,omitempty"`
	NeedsTracing   bool     `json:"needs_tracing"`
	IsSuperglobal  bool     `json:"is_superglobal"`
	IsEscaped      bool     `json:"is_escaped,omitempty"`
}

// ClassificationSummary provides statistics about the classification
type ClassificationSummary struct {
	TotalExpressions  int      `json:"total_expressions"`
	UserInputCount    int      `json:"user_input_count"`
	NeedsTracingCount int      `json:"needs_tracing_count"`
	SourceTypesSeen   []string `json:"source_types_seen"`
}

// BatchInput represents input for batch classification
type BatchInput struct {
	Finding  string   `json:"finding"`
	Snippets []string `json:"snippets"`
}

// BatchResult contains results for batch classification
type BatchResult struct {
	AnalyzedAt       string          `json:"analyzed_at"`
	CarrierMapPath   string          `json:"carrier_map_path,omitempty"`
	Framework        string          `json:"framework,omitempty"`
	TotalFindings    int             `json:"total_findings"`
	TotalSnippets    int             `json:"total_snippets"`
	WithUserInput    int             `json:"with_user_input"`
	WithoutUserInput int             `json:"without_user_input"`
	NeedsTracing     int             `json:"needs_tracing"`
	BySourceType     map[string]int  `json:"by_source_type"`
	Findings         []FindingResult `json:"findings"`
}

// FindingResult contains results for a single finding
type FindingResult struct {
	FindingID     string                 `json:"finding_id"`
	Filename      string                 `json:"filename,omitempty"`
	TotalSnippets int                    `json:"total_snippets"`
	WithUserInput int                    `json:"with_user_input"`
	Snippets      []ClassificationResult `json:"snippets"`
}

// NewClassifier creates a classifier with a carrier map
func NewClassifier(carrierMap *discovery.CarrierMap) *Classifier {
	c := &Classifier{
		carrierMap: carrierMap,
		extractor:  extractor.New(),
	}

	if carrierMap != nil {
		c.compileCarrierPatterns()
	}

	return c
}

// NewDirectClassifier creates a classifier for superglobals-only (no carrier map)
func NewDirectClassifier() *Classifier {
	return &Classifier{
		carrierMap: nil,
		extractor:  extractor.New(),
	}
}

// compileCarrierPatterns compiles regex patterns for each carrier
func (c *Classifier) compileCarrierPatterns() {
	if c.carrierMap == nil {
		return
	}

	for i := range c.carrierMap.Carriers {
		carrier := &c.carrierMap.Carriers[i]
		var pattern *regexp.Regexp
		isMethod := false

		if carrier.PropertyName != "" {
			// Property pattern: $var->property['key']
			// Match any variable name, specific property
			if carrier.AccessPattern == "array" {
				pattern = regexp.MustCompile(`\$\w+->\s*` + regexp.QuoteMeta(carrier.PropertyName) + `\s*\[`)
			} else {
				pattern = regexp.MustCompile(`\$\w+->\s*` + regexp.QuoteMeta(carrier.PropertyName) + `(?:\s*\[|$)`)
			}
		} else if carrier.MethodName != "" {
			// Method pattern: $var->method('key')
			isMethod = true
			pattern = regexp.MustCompile(`\$\w+->\s*` + regexp.QuoteMeta(carrier.MethodName) + `\s*\(`)
		}

		if pattern != nil {
			c.carrierPatterns = append(c.carrierPatterns, &carrierPattern{
				carrier:  carrier,
				pattern:  pattern,
				isMethod: isMethod,
			})
		}
	}
}

// ClassifySnippet analyzes a single code snippet
func (c *Classifier) ClassifySnippet(snippet string) *ClassificationResult {
	result := &ClassificationResult{
		Snippet:     snippet,
		Expressions: make([]ExpressionResult, 0),
		Summary: ClassificationSummary{
			SourceTypesSeen: make([]string, 0),
		},
	}

	// Extract all expressions from the snippet
	expressions := c.extractor.ExtractFromSnippet(snippet)

	sourceTypesSet := make(map[string]bool)

	for _, expr := range expressions {
		exprResult := c.classifyExpression(expr)
		result.Expressions = append(result.Expressions, exprResult)

		result.Summary.TotalExpressions++

		if exprResult.IsSuperglobal || len(exprResult.SourceTypes) > 0 {
			result.HasUserInput = true
			result.Summary.UserInputCount++
			for _, st := range exprResult.SourceTypes {
				sourceTypesSet[st] = true
			}
		}

		if exprResult.NeedsTracing {
			result.NeedsTracing = true
			result.Summary.NeedsTracingCount++
		}
	}

	// Convert source types set to slice
	for st := range sourceTypesSet {
		result.Summary.SourceTypesSeen = append(result.Summary.SourceTypesSeen, st)
	}

	return result
}

// classifyExpression classifies a single expression
func (c *Classifier) classifyExpression(expr extractor.ExtractedExpression) ExpressionResult {
	result := ExpressionResult{
		Expression:  expr.Expression,
		Key:         expr.Key,
		IsEscaped:   expr.IsEscaped,
		Confidence:  0,
		SourceTypes: make([]string, 0),
	}

	// Check if it's a direct superglobal
	if expr.Type == "superglobal" {
		result.IsSuperglobal = true
		result.Confidence = 1.0
		result.SourceTypes = superglobalToSourceTypes(expr.PropertyName) // PropertyName stores the superglobal type
		return result
	}

	// If we have a carrier map, check against carriers
	if c.carrierMap != nil {
		for _, cp := range c.carrierPatterns {
			if cp.pattern.MatchString(expr.Expression) {
				result.SourceTypes = cp.carrier.SourceTypes
				result.Confidence = cp.carrier.Confidence
				result.MatchedCarrier = cp.carrier.ClassName + "." + cp.carrier.PropertyName
				if cp.carrier.MethodName != "" {
					result.MatchedCarrier = cp.carrier.ClassName + "." + cp.carrier.MethodName + "()"
				}
				return result
			}
		}
	}

	// No carrier map or no match - mark as needs tracing
	// But only if it looks like it could be user input (property access or method call)
	if expr.Type == "property_access" || expr.Type == "method_call" ||
		expr.Type == "sql_embedded" || expr.Type == "concatenated" || expr.Type == "escaped" {
		result.NeedsTracing = true
		result.Confidence = 0
	}

	return result
}

// superglobalToSourceTypes maps superglobal names to source types using centralized mappings
func superglobalToSourceTypes(sgType string) []string {
	// Use centralized reverse mapping
	superglobalName := sources.GetSuperglobalByShortName(sgType)
	if superglobalName != "" {
		return []string{superglobalName}
	}
	return []string{"$_" + sgType}
}

// ClassifyBatch analyzes multiple findings with snippets
func (c *Classifier) ClassifyBatch(inputs []BatchInput) *BatchResult {
	result := &BatchResult{
		AnalyzedAt:   time.Now().Format(time.RFC3339),
		BySourceType: make(map[string]int),
		Findings:     make([]FindingResult, 0, len(inputs)),
	}

	if c.carrierMap != nil {
		result.CarrierMapPath = c.carrierMap.CodebasePath
		result.Framework = c.carrierMap.Framework
	}

	for _, input := range inputs {
		findingResult := c.classifyFinding(input)
		result.Findings = append(result.Findings, findingResult)

		result.TotalFindings++
		result.TotalSnippets += findingResult.TotalSnippets
		result.WithUserInput += findingResult.WithUserInput
		result.WithoutUserInput += findingResult.TotalSnippets - findingResult.WithUserInput

		// Count needs tracing
		for _, sr := range findingResult.Snippets {
			if sr.NeedsTracing && !sr.HasUserInput {
				result.NeedsTracing++
			}
			for _, st := range sr.Summary.SourceTypesSeen {
				result.BySourceType[st]++
			}
		}
	}

	return result
}

// classifyFinding classifies all snippets in a finding
func (c *Classifier) classifyFinding(input BatchInput) FindingResult {
	result := FindingResult{
		FindingID:     input.Finding,
		Filename:      extractFilename(input.Finding),
		TotalSnippets: len(input.Snippets),
		Snippets:      make([]ClassificationResult, 0, len(input.Snippets)),
	}

	for _, snippet := range input.Snippets {
		sr := c.ClassifySnippet(snippet)
		result.Snippets = append(result.Snippets, *sr)

		if sr.HasUserInput {
			result.WithUserInput++
		}
	}

	return result
}

// extractFilename extracts the filename from a finding ID
func extractFilename(finding string) string {
	// Finding format: CWE-89_finding_8_admin/modules/forum/announcements.php
	parts := strings.Split(finding, "_finding_")
	if len(parts) >= 2 {
		// Get everything after the number
		subParts := strings.SplitN(parts[1], "_", 2)
		if len(subParts) >= 2 {
			return subParts[1]
		}
	}
	return finding
}

// LoadBatchInput loads batch input from a JSON file (PatchLeaks format)
func LoadBatchInput(path string) ([]BatchInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var inputs []BatchInput
	if err := json.Unmarshal(data, &inputs); err != nil {
		return nil, err
	}

	return inputs, nil
}

// LoadSimpleBatchInput loads batch input from a simple JSON array of strings
func LoadSimpleBatchInput(path string) ([]BatchInput, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var snippets []string
	if err := json.Unmarshal(data, &snippets); err != nil {
		return nil, err
	}

	// Convert to batch input format
	inputs := make([]BatchInput, 0, len(snippets))
	for i, snippet := range snippets {
		inputs = append(inputs, BatchInput{
			Finding:  string(rune('A' + i%26)) + string(rune('0'+i/26)),
			Snippets: []string{snippet},
		})
	}

	return inputs, nil
}

// SaveBatchResult saves batch results to a JSON file
func SaveBatchResult(result *BatchResult, path string) error {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// Summary returns a human-readable summary of batch results
func (r *BatchResult) Summary() string {
	s := "Classification Summary\n"
	s += "=====================\n"
	s += "Analyzed: " + r.AnalyzedAt + "\n"
	if r.CarrierMapPath != "" {
		s += "Carrier Map: " + r.CarrierMapPath + "\n"
	}
	if r.Framework != "" {
		s += "Framework: " + r.Framework + "\n"
	}
	s += "\nStatistics:\n"
	s += "  - Total Findings: " + itoa(r.TotalFindings) + "\n"
	s += "  - Total Snippets: " + itoa(r.TotalSnippets) + "\n"
	s += "  - With User Input: " + itoa(r.WithUserInput) + "\n"
	s += "  - Without User Input: " + itoa(r.WithoutUserInput) + "\n"
	s += "  - Needs Tracing: " + itoa(r.NeedsTracing) + "\n"

	if len(r.BySourceType) > 0 {
		s += "\nBy Source Type:\n"
		for st, count := range r.BySourceType {
			s += "  - " + st + ": " + itoa(count) + "\n"
		}
	}

	return s
}

// simple int to string without importing strconv
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	if i < 0 {
		return "-" + itoa(-i)
	}
	s := ""
	for i > 0 {
		s = string(rune('0'+i%10)) + s
		i /= 10
	}
	return s
}
