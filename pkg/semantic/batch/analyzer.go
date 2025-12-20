// Package batch provides batch analysis capabilities for analyzing multiple code snippets
package batch

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/hatlesswizard/inputtracer/pkg/semantic"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/extractor"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/symbolic"
)

// SnippetInput represents a code snippet to analyze
type SnippetInput struct {
	ID       string   `json:"id"`
	Filename string   `json:"filename"`
	Context  []string `json:"context"`
}

// BatchInput represents the input for batch analysis
type BatchInput struct {
	CodebasePath string         `json:"codebase_path"`
	Snippets     []SnippetInput `json:"snippets"`
}

// ExpressionResult represents the analysis result for a single expression
type ExpressionResult struct {
	Expression   string   `json:"expression"`
	HasUserInput bool     `json:"has_user_input"`
	InputTypes   []string `json:"input_types,omitempty"`
	TraceSteps   []string `json:"trace_steps,omitempty"`
	TraceError   string   `json:"trace_error,omitempty"`
}

// SnippetResult represents the analysis result for a single snippet
type SnippetResult struct {
	ID           string             `json:"id"`
	Filename     string             `json:"filename"`
	Expressions  []ExpressionResult `json:"expressions"`
	HasAnyInput  bool               `json:"has_any_input"`
	InputSummary []string           `json:"input_summary"`
}

// BatchOutput represents the output of batch analysis
type BatchOutput struct {
	CodebasePath   string          `json:"codebase_path"`
	AnalyzedAt     string          `json:"analyzed_at"`
	TotalSnippets  int             `json:"total_snippets"`
	WithUserInput  int             `json:"with_user_input"`
	TotalExprs     int             `json:"total_expressions"`
	TracedExprs    int             `json:"traced_expressions"`
	InputExprs     int             `json:"input_expressions"`
	Results        []SnippetResult `json:"results"`
}

// BatchAnalyzer performs batch analysis of code snippets
type BatchAnalyzer struct {
	codebasePath string
	engine       *symbolic.ExecutionEngine
	extractor    *extractor.ExpressionExtractor
	traceResult  *semantic.TraceResult
}

// NewBatchAnalyzer creates a new batch analyzer
func NewBatchAnalyzer(codebasePath string) *BatchAnalyzer {
	return &BatchAnalyzer{
		codebasePath: codebasePath,
		extractor:    extractor.New(),
	}
}

// Initialize parses the codebase and builds symbol tables
func (a *BatchAnalyzer) Initialize() error {
	config := semantic.DefaultConfig()
	config.Languages = []string{"php"}
	tracer := semantic.New(config)

	result, err := tracer.ParseOnly(a.codebasePath)
	if err != nil {
		return fmt.Errorf("failed to parse codebase: %w", err)
	}

	a.traceResult = result

	// Create execution engine
	a.engine = symbolic.NewExecutionEngine()

	// Add symbol tables
	for filePath, st := range result.SymbolTable {
		a.engine.AddSymbolTable(filePath, st)
	}

	// Add parsed files
	for filePath, fileInfo := range result.Files {
		if fileInfo.Root != nil && fileInfo.Content != nil {
			a.engine.AddParsedFile(filePath, fileInfo.Root, fileInfo.Content)
		}
	}

	return nil
}

// AnalyzeBatch analyzes all snippets in the batch input
func (a *BatchAnalyzer) AnalyzeBatch(input *BatchInput) (*BatchOutput, error) {
	output := &BatchOutput{
		CodebasePath:  a.codebasePath,
		AnalyzedAt:    time.Now().Format(time.RFC3339),
		TotalSnippets: len(input.Snippets),
		Results:       make([]SnippetResult, 0, len(input.Snippets)),
	}

	inputTypesSeen := make(map[string]bool)

	for _, snippet := range input.Snippets {
		result := a.analyzeSnippet(snippet)
		output.Results = append(output.Results, result)

		output.TotalExprs += len(result.Expressions)

		for _, expr := range result.Expressions {
			if expr.TraceError == "" {
				output.TracedExprs++
			}
			if expr.HasUserInput {
				output.InputExprs++
				for _, t := range expr.InputTypes {
					inputTypesSeen[t] = true
				}
			}
		}

		if result.HasAnyInput {
			output.WithUserInput++
		}
	}

	return output, nil
}

// analyzeSnippet analyzes a single code snippet
func (a *BatchAnalyzer) analyzeSnippet(snippet SnippetInput) SnippetResult {
	result := SnippetResult{
		ID:          snippet.ID,
		Filename:    snippet.Filename,
		Expressions: make([]ExpressionResult, 0),
	}

	// Extract expressions from the snippet
	extracted := a.extractor.ExtractFromDiffContext(snippet.Context)

	inputTypesSeen := make(map[string]bool)

	for _, expr := range extracted {
		exprResult := a.traceExpression(expr.Expression, snippet.Filename)
		result.Expressions = append(result.Expressions, exprResult)

		if exprResult.HasUserInput {
			result.HasAnyInput = true
			for _, t := range exprResult.InputTypes {
				inputTypesSeen[t] = true
			}
		}
	}

	// Build input summary
	for t := range inputTypesSeen {
		result.InputSummary = append(result.InputSummary, t)
	}

	return result
}

// traceExpression traces a single expression using the symbolic execution engine
func (a *BatchAnalyzer) traceExpression(expression, contextFile string) ExpressionResult {
	result := ExpressionResult{
		Expression: expression,
	}

	// Try to trace the expression
	flow, err := a.engine.TracePropertyAccess(expression, contextFile)
	if err != nil {
		result.TraceError = err.Error()
		return result
	}

	// Check if any sources are user input
	for _, source := range flow.Sources {
		switch source.Type {
		case "http_get", "http_post", "http_cookie", "http_header", "http_request", "http_body":
			result.HasUserInput = true
			if !contains(result.InputTypes, source.Type) {
				result.InputTypes = append(result.InputTypes, source.Type)
			}
		}
	}

	// Add trace steps
	for _, step := range flow.Steps {
		result.TraceSteps = append(result.TraceSteps, step.Description)
	}

	return result
}

// contains checks if a slice contains a string
func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// ParseBatchInput parses JSON input into BatchInput
func ParseBatchInput(data []byte) (*BatchInput, error) {
	var input BatchInput
	if err := json.Unmarshal(data, &input); err != nil {
		return nil, fmt.Errorf("failed to parse batch input: %w", err)
	}
	return &input, nil
}
