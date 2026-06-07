package tracer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/hatlesswizard/inputtracer/pkg/ast"
	"github.com/hatlesswizard/inputtracer/pkg/parser"
	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/sources"
)

// Config configures the tracer
type Config struct {
	// Languages to analyze (empty = all supported)
	Languages []string

	// Maximum inter-procedural analysis depth
	MaxDepth int

	// Number of parallel workers
	Workers int

	// Custom source definitions (in addition to built-in)
	CustomSources []sources.Definition

	// Skip directories matching these patterns
	SkipDirs []string

	// Include only files matching these patterns (empty = all)
	IncludePatterns []string

	// Verbose enables diagnostic logging to stdout during analysis
	Verbose bool
}

// DefaultConfig returns sensible defaults using centralized sources
func DefaultConfig() *Config {
	return &Config{
		Languages:       []string{}, // All supported
		MaxDepth:        sources.DefaultMaxDepth,
		Workers:         runtime.NumCPU(),
		SkipDirs:        sources.DefaultSkipDirs,
		IncludePatterns: []string{},
	}
}

// Tracer is the main entry point for input tracing
type Tracer struct {
	config  *Config
	parser  *parser.Service
	sources *sources.Registry
	ast     *ast.Registry
	mu      sync.Mutex
}

// Validate checks that the config has valid values.
func (c *Config) Validate() error {
	if c.MaxDepth <= 0 {
		return fmt.Errorf("MaxDepth must be > 0, got %d", c.MaxDepth)
	}
	if c.Workers <= 0 {
		return fmt.Errorf("Workers must be > 0, got %d", c.Workers)
	}
	return nil
}

// newEmptyResult returns a zero-value TraceResult with all slices initialized.
func newEmptyResult() *TraceResult {
	return &TraceResult{
		Sources:          make([]*InputSource, 0),
		TaintedVariables: make([]*TaintedVariable, 0),
		TaintedFunctions: make([]*TaintedFunction, 0),
		FlowGraph: &FlowGraph{
			Nodes: make([]FlowNode, 0),
			Edges: make([]FlowEdge, 0),
		},
		Stats:  TraceStats{ByLanguage: make(map[string]int)},
		Errors: make([]error, 0),
	}
}

// New creates a new Tracer with the given configuration
func New(config *Config) *Tracer {
	if config == nil {
		config = DefaultConfig()
	}
	if err := config.Validate(); err != nil {
		panic("inputtracer: invalid config: " + err.Error())
	}

	// Initialize parser service
	parserSvc := parser.NewService()

	// Register all language parsers
	languages.RegisterAllLanguages(parserSvc)

	// Initialize source registry with all language sources
	sourceReg := sources.NewRegistry()
	sources.RegisterAll(sourceReg)

	// Register custom sources if provided
	for _, src := range config.CustomSources {
		sourceReg.AddSource(src)
	}

	// Initialize AST registry
	astReg := ast.NewRegistry()
	ast.RegisterAll(astReg)

	return &Tracer{
		config:  config,
		parser:  parserSvc,
		sources: sourceReg,
		ast:     astReg,
	}
}

// TraceDirectory analyzes a directory and returns all input flow information
func (t *Tracer) TraceDirectory(dirPath string) (*TraceResult, error) {
	startTime := time.Now()

	result := newEmptyResult()

	// Collect all files to analyze
	files, walkErrs, err := t.collectFiles(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to collect files: %w", err)
	}
	result.Errors = append(result.Errors, walkErrs...)

	// Create work channels
	fileChan := make(chan string, len(files))
	resultChan := make(chan *fileResult, len(files))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < t.config.Workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filePath := range fileChan {
				fr := t.analyzeFile(filePath)
				resultChan <- fr
			}
		}()
	}

	// Send files to workers
	for _, f := range files {
		fileChan <- f
	}
	close(fileChan)

	// Wait for workers to finish
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results
	for fr := range resultChan {
		t.mergeFileResult(result, fr)
	}

	// Phase 2: Inter-procedural analysis
	t.runInterproceduralAnalysis(result)

	// Build flow graph
	t.buildFlowGraph(result)

	// Finalize stats
	result.Stats.SourcesFound = len(result.Sources)
	result.Stats.TaintedVarsFound = len(result.TaintedVariables)
	result.Stats.TaintedFuncsFound = len(result.TaintedFunctions)
	result.Stats.AnalysisDuration = time.Since(startTime)
	result.Stats.DurationMs = result.Stats.AnalysisDuration.Milliseconds()

	return result, nil
}

// TraceFile analyzes a single source file and returns all input flow
// information found within it. Unlike TraceDirectory it does NOT walk the
// filesystem or run inter-procedural analysis across multiple files — taint
// propagation is limited to what can be observed within filePath alone.
//
// The returned TraceResult follows the same schema as TraceDirectory so callers
// can use the same output/reporting code for both entry points.
//
// filePath must be an absolute or relative path to a regular file. If the file
// cannot be parsed (unsupported language, I/O error, etc.) the error is
// recorded in TraceResult.Errors rather than returned as the function error; a
// non-nil function error is only returned for truly unexpected failures.
func (t *Tracer) TraceFile(filePath string) (*TraceResult, error) {
	startTime := time.Now()

	result := newEmptyResult()

	fr := t.analyzeFile(filePath)
	t.mergeFileResult(result, fr)

	// Build flow graph
	t.buildFlowGraph(result)

	// Finalize stats
	result.Stats.SourcesFound = len(result.Sources)
	result.Stats.TaintedVarsFound = len(result.TaintedVariables)
	result.Stats.TaintedFuncsFound = len(result.TaintedFunctions)
	result.Stats.AnalysisDuration = time.Since(startTime)
	result.Stats.DurationMs = result.Stats.AnalysisDuration.Milliseconds()

	return result, nil
}

// fileResult holds the result of analyzing a single file
type fileResult struct {
	FilePath         string
	Language         string
	Sources          []*InputSource
	TaintedVariables []*TaintedVariable
	TaintedFunctions []*TaintedFunction
	Paths            []PropagationPath
	Error            error
}

// analyzeFile analyzes a single file and returns its results
func (t *Tracer) analyzeFile(filePath string) *fileResult {
	fr := &fileResult{
		FilePath:         filePath,
		Sources:          make([]*InputSource, 0),
		TaintedVariables: make([]*TaintedVariable, 0),
		TaintedFunctions: make([]*TaintedFunction, 0),
		Paths:            make([]PropagationPath, 0),
	}

	lang := t.parser.DetectLanguage(filePath)
	if lang == "" {
		return fr
	}
	fr.Language = lang

	if len(t.config.Languages) > 0 {
		if !slices.Contains(t.config.Languages, lang) {
			return fr
		}
	}

	parseResult, err := t.parser.ParseFile(filePath)
	if err != nil {
		fr.Error = fmt.Errorf("parse error in %s: %w", filePath, err)
		return fr
	}
	// The tree-sitter nodes traversed below point into parseResult.Tree's C
	// memory, which the binding frees via a finalizer once the tree is
	// unreachable. Keep the tree alive until all traversal completes so a
	// concurrent cache eviction can't let it be finalized mid-traversal.
	defer runtime.KeepAlive(parseResult.Tree)

	sourceMatcher := t.sources.GetMatcher(lang)
	if sourceMatcher == nil {
		return fr
	}

	astExtractor := t.ast.GetExtractor(lang)
	if astExtractor == nil {
		return fr
	}

	srcs, taintedFromSources, state := detectSources(sourceMatcher, parseResult, filePath, lang)
	fr.Sources = srcs
	fr.TaintedVariables = taintedFromSources

	newVars, paths := trackAssignments(astExtractor, parseResult, filePath, lang, state)
	fr.TaintedVariables = append(fr.TaintedVariables, newVars...)
	fr.Paths = paths

	fr.TaintedFunctions = analyzeTaintedCalls(astExtractor, parseResult, filePath, lang, state)

	return fr
}

// detectSources finds all input sources in the parsed file and initializes analysis state.
func detectSources(matcher sources.Matcher, parseResult *parser.ParseResult, filePath, lang string) ([]*InputSource, []*TaintedVariable, *AnalysisState) {
	srcs := make([]*InputSource, 0)
	tainted := make([]*TaintedVariable, 0)
	state := NewAnalysisState()

	for _, match := range matcher.FindSources(parseResult.Root, parseResult.Source) {
		labels := make([]InputLabel, len(match.Labels))
		for i, l := range match.Labels {
			labels[i] = InputLabel(l)
		}

		src := &InputSource{
			ID:   uuid.New().String(),
			Type: match.SourceType,
			Key:  match.Key,
			Location: Location{
				FilePath:  filePath,
				Line:      match.Line,
				Column:    match.Column,
				EndLine:   match.EndLine,
				EndColumn: match.EndColumn,
				Snippet:   match.Snippet,
			},
			Labels:   labels,
			Language: lang,
		}
		srcs = append(srcs, src)

		if match.Variable != "" {
			tv := &TaintedVariable{
				ID:       uuid.New().String(),
				Name:     match.Variable,
				Scope:    "file",
				Source:   src,
				Location: src.Location,
				Depth:    0,
				Language: lang,
			}
			tainted = append(tainted, tv)
			state.SetTainted(match.Variable, tv)
		}
	}

	return srcs, tainted, state
}

// trackAssignments propagates taint through variable assignments in the parsed file.
func trackAssignments(extractor ast.Extractor, parseResult *parser.ParseResult, filePath, lang string, state *AnalysisState) ([]*TaintedVariable, []PropagationPath) {
	var newVars []*TaintedVariable
	var paths []PropagationPath

	for _, assign := range extractor.ExtractAssignments(parseResult.Root, parseResult.Source) {
		// Record simple variable aliases (e.g. $r = $request).
		// A "simple" RHS is a bare variable name with no property access,
		// subscript, method call, or operators.
		if rhsText := strings.TrimSpace(assign.RHSText); rhsText != "" && isBareVariable(rhsText) {
			lhs := stripVarSigil(assign.LHS)
			rhs := stripVarSigil(rhsText)
			if lhs != "" && rhs != "" && lhs != rhs {
				state.Aliases[lhs] = rhs
			}
		}

		for varName, tainted := range state.TaintedValues {
			if extractor.ExpressionContains(assign.RHS, varName, parseResult.Source) {
				loc := Location{
					FilePath:  filePath,
					Line:      assign.Line,
					Column:    assign.Column,
					EndLine:   assign.EndLine,
					EndColumn: assign.EndColumn,
					Snippet:   assign.Snippet,
				}
				newTainted := &TaintedVariable{
					ID:       uuid.New().String(),
					Name:     assign.LHS,
					Scope:    assign.Scope,
					Source:   tainted.Source,
					Location: loc,
					Depth:    tainted.Depth + 1,
					Language: lang,
				}
				newVars = append(newVars, newTainted)
				state.SetTainted(assign.LHS, newTainted)

				paths = append(paths, PropagationPath{
					Source: tainted.Source,
					Steps: []PropagationStep{
						{
							Type:     StepAssignment,
							Variable: assign.LHS,
							Location: loc,
						},
					},
					Destination: loc,
				})
			}
		}
	}

	return newVars, paths
}

// isBareVariable reports whether s looks like a bare variable name (possibly
// with a $ or @ sigil prefix) and contains no operators, property access,
// subscript, or function call syntax.
func isBareVariable(s string) bool {
	if s == "" {
		return false
	}
	for _, ch := range s {
		switch ch {
		case '.', '[', ']', '(', ')', '{', '}', '+', '-', '*', '/', '&', '|', '!', '=', '<', '>', ',', ':', '?', ' ', '\t':
			return false
		}
	}
	// After ruling out operators, must start with a letter, underscore, $, or @
	first := rune(s[0])
	return first == '$' || first == '@' || first == '_' || (first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z')
}

// stripVarSigil removes a leading $ or @ from a variable name so that alias
// keys are stored without language-specific sigil prefixes.
func stripVarSigil(name string) string {
	if len(name) > 1 && (name[0] == '$' || name[0] == '@') {
		return name[1:]
	}
	return name
}

// analyzeTaintedCalls finds function calls that receive tainted arguments.
func analyzeTaintedCalls(extractor ast.Extractor, parseResult *parser.ParseResult, filePath, lang string, state *AnalysisState) []*TaintedFunction {
	var funcs []*TaintedFunction

	for _, call := range extractor.ExtractCalls(parseResult.Root, parseResult.Source) {
		var taintedParams []TaintedParam

		for i, arg := range call.Arguments {
			for varName, tainted := range state.TaintedValues {
				if extractor.ExpressionContains(arg.Node, varName, parseResult.Source) {
					taintedParams = append(taintedParams, TaintedParam{
						Index:  i,
						Name:   arg.Name,
						Source: tainted.Source,
						Path: &PropagationPath{
							Source: tainted.Source,
							Steps: []PropagationStep{
								{
									Type:     StepParameterPass,
									Variable: varName,
									Function: call.Name,
									Location: Location{
										FilePath:  filePath,
										Line:      call.Line,
										Column:    call.Column,
										EndLine:   call.EndLine,
										EndColumn: call.EndColumn,
									},
								},
							},
							Destination: Location{
								FilePath: filePath,
								Line:     call.Line,
								Column:   call.Column,
							},
						},
					})
					break
				}
			}
		}

		if len(taintedParams) > 0 {
			funcs = append(funcs, &TaintedFunction{
				ID:            uuid.New().String(),
				Name:          call.Name,
				FilePath:      filePath,
				Line:          call.Line,
				Language:      lang,
				TaintedParams: taintedParams,
			})
		}
	}

	return funcs
}

// mergeFileResult merges a file result into the main result
func (t *Tracer) mergeFileResult(result *TraceResult, fr *fileResult) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if fr.Error != nil {
		result.Errors = append(result.Errors, fr.Error)
	}

	if fr.Language != "" {
		result.Stats.FilesAnalyzed++
		result.Stats.ByLanguage[fr.Language]++
	}

	result.Sources = append(result.Sources, fr.Sources...)
	result.TaintedVariables = append(result.TaintedVariables, fr.TaintedVariables...)
	result.TaintedFunctions = append(result.TaintedFunctions, fr.TaintedFunctions...)
	result.Stats.PropagationPaths += len(fr.Paths)
}

// runInterproceduralAnalysis performs cross-function taint analysis.
func (t *Tracer) runInterproceduralAnalysis(result *TraceResult) {
	maxDepth := t.config.MaxDepth
	if maxDepth <= 0 {
		maxDepth = defaultInterproceduralDepth
	}
	state := NewFullAnalysisState()
	analyzer := NewInterproceduralAnalyzer(state, maxDepth, t.parser)
	analyzer.RunAnalysis(result)
}

// buildFlowGraph delegates to FullAnalysisState.BuildFlowGraph, which is the
// single canonical implementation. It constructs a temporary FullAnalysisState
// from the result slices so that the owning type performs the actual graph
// construction, then writes the result back into result.FlowGraph.
func (t *Tracer) buildFlowGraph(result *TraceResult) {
	state := NewFullAnalysisState()
	for _, src := range result.Sources {
		state.AddSource(src)
	}
	for _, tv := range result.TaintedVariables {
		state.AddTaintedVariable(tv)
	}
	for _, tf := range result.TaintedFunctions {
		state.AddTaintedFunction(tf)
	}
	result.FlowGraph = state.BuildFlowGraph()
}

// collectFiles collects all files to analyze from a directory.
// It returns the list of files, any non-fatal walk errors (e.g. permission
// denied), and a fatal error that aborted the walk (if any).
func (t *Tracer) collectFiles(dirPath string) ([]string, []error, error) {
	var files []string

	// Build effective skip dirs: merge WordPress vendor dirs for PHP projects
	skipDirs := make(map[string]bool, len(t.config.SkipDirs))
	for _, d := range t.config.SkipDirs {
		skipDirs[d] = true
	}

	// Check if PHP is in the language filter (or no filter = all languages)
	phpIncluded := len(t.config.Languages) == 0 || slices.Contains(t.config.Languages, "php")

	// Add WordPress-specific vendor dirs when analyzing PHP.
	// These are defined in sources.WordPressVendorDirs so that
	// product-specific knowledge stays in the sources layer. Callers
	// that do not want this behaviour can clear Config.SkipDirs and
	// rebuild it themselves.
	if phpIncluded {
		for _, d := range sources.WordPressVendorDirs {
			skipDirs[d] = true
		}
	}

	var walkErrors []error

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			if os.IsPermission(err) {
				// Non-fatal: record and continue walking.
				walkErrors = append(walkErrors, fmt.Errorf("permission denied: %s", path))
				return nil
			}
			// Fatal I/O error: abort the walk.
			return err
		}

		// Skip directories in skip list
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}

			return nil
		}

		// Check if file has a supported extension
		if t.parser.DetectLanguage(path) != "" {
			files = append(files, path)
		}

		return nil
	})

	return files, walkErrors, err
}

// GetTaintedFunctions returns all functions that receive user input
func (t *Tracer) GetTaintedFunctions(result *TraceResult) []*TaintedFunction {
	return result.TaintedFunctions
}

// GetFlowPaths returns all propagation paths from a specific source
func (t *Tracer) GetFlowPaths(result *TraceResult, source *InputSource) []*PropagationPath {
	paths := make([]*PropagationPath, 0)

	// Find all variables that came from this source
	for _, v := range result.TaintedVariables {
		if v.Source != nil && v.Source.ID == source.ID {
			paths = append(paths, &PropagationPath{
				Source: source,
				Steps: []PropagationStep{
					{
						Type:     StepAssignment,
						Variable: v.Name,
						Location: v.Location,
					},
				},
				Destination: v.Location,
			})
		}
	}

	// Find all functions that received this source
	for _, fn := range result.TaintedFunctions {
		for _, param := range fn.TaintedParams {
			if param.Source != nil && param.Source.ID == source.ID {
				paths = append(paths, &PropagationPath{
					Source: source,
					Steps: []PropagationStep{
						{
							Type:     StepParameterPass,
							Variable: param.Name,
							Function: fn.Name,
							Location: Location{
								FilePath: fn.FilePath,
								Line:     fn.Line,
							},
						},
					},
					Destination: Location{
						FilePath: fn.FilePath,
						Line:     fn.Line,
					},
				})
			}
		}
	}

	return paths
}

// DoesReceiveInput checks if a specific function receives user input
func (t *Tracer) DoesReceiveInput(result *TraceResult, funcName string) bool {
	for _, fn := range result.TaintedFunctions {
		if fn.Name == funcName && len(fn.TaintedParams) > 0 {
			return true
		}
	}
	return false
}

// GetInputSources returns all input sources found
func (t *Tracer) GetInputSources(result *TraceResult) []*InputSource {
	return result.Sources
}

// GetTaintedVariables returns all variables that hold user input
func (t *Tracer) GetTaintedVariables(result *TraceResult) []*TaintedVariable {
	return result.TaintedVariables
}
