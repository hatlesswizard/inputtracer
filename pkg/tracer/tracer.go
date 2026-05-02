package tracer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
		Errors: make([]string, 0),
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
	files, err := t.collectFiles(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to collect files: %w", err)
	}

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

// TraceFile analyzes a single file
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
	Error            string
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
		found := false
		for _, l := range t.config.Languages {
			if l == lang {
				found = true
				break
			}
		}
		if !found {
			return fr
		}
	}

	parseResult, err := t.parser.ParseFile(filePath)
	if err != nil {
		fr.Error = fmt.Sprintf("parse error: %v", err)
		return fr
	}

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

	if fr.Error != "" {
		result.Errors = append(result.Errors, fmt.Sprintf("%s: %s", fr.FilePath, fr.Error))
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
// Delegates to runInterproceduralAnalysisImpl in interprocedural_propagation.go.
func (t *Tracer) runInterproceduralAnalysis(result *TraceResult) {
	t.runInterproceduralAnalysisImpl(result)
}

// buildFlowGraph builds the flow graph from the analysis results
func (t *Tracer) buildFlowGraph(result *TraceResult) {
	nodeMap := make(map[string]bool)

	// Add source nodes
	for _, src := range result.Sources {
		nodeID := "src:" + src.ID
		if !nodeMap[nodeID] {
			result.FlowGraph.Nodes = append(result.FlowGraph.Nodes, FlowNode{
				ID:       nodeID,
				Type:     "source",
				Name:     src.Type,
				Location: src.Location,
			})
			nodeMap[nodeID] = true
		}
	}

	// Add variable nodes and edges from sources
	for _, v := range result.TaintedVariables {
		nodeID := "var:" + v.ID
		if !nodeMap[nodeID] {
			result.FlowGraph.Nodes = append(result.FlowGraph.Nodes, FlowNode{
				ID:       nodeID,
				Type:     "variable",
				Name:     v.Name,
				Location: v.Location,
			})
			nodeMap[nodeID] = true
		}

		// Edge from source to variable
		if v.Source != nil {
			result.FlowGraph.Edges = append(result.FlowGraph.Edges, FlowEdge{
				From:     "src:" + v.Source.ID,
				To:       nodeID,
				Type:     "assignment",
				Location: v.Location,
			})
		}
	}

	// Add function nodes and edges
	for _, fn := range result.TaintedFunctions {
		nodeID := "func:" + fn.ID
		if !nodeMap[nodeID] {
			result.FlowGraph.Nodes = append(result.FlowGraph.Nodes, FlowNode{
				ID:   nodeID,
				Type: "function",
				Name: fn.Name,
				Location: Location{
					FilePath: fn.FilePath,
					Line:     fn.Line,
				},
			})
			nodeMap[nodeID] = true
		}

		// Edges from sources to function params
		for _, param := range fn.TaintedParams {
			if param.Source != nil {
				result.FlowGraph.Edges = append(result.FlowGraph.Edges, FlowEdge{
					From: "src:" + param.Source.ID,
					To:   nodeID,
					Type: "call",
					Location: Location{
						FilePath: fn.FilePath,
						Line:     fn.Line,
					},
				})
			}
		}
	}
}

// collectFiles collects all files to analyze from a directory
func (t *Tracer) collectFiles(dirPath string) ([]string, error) {
	var files []string

	// Build effective skip dirs: merge WordPress vendor dirs for PHP projects
	skipDirs := make(map[string]bool, len(t.config.SkipDirs))
	for _, d := range t.config.SkipDirs {
		skipDirs[d] = true
	}

	// Check if PHP is in the language filter (or no filter = all languages)
	phpIncluded := len(t.config.Languages) == 0
	if !phpIncluded {
		for _, l := range t.config.Languages {
			if l == "php" {
				phpIncluded = true
				break
			}
		}
	}

	// Add WordPress-specific vendor dirs when analyzing PHP
	if phpIncluded {
		wpVendorDirs := []string{
			"freemius", "action-scheduler", "redux-core", "redux-framework",
			"cmb2", "starter-content", "starter-templates",
		}
		for _, d := range wpVendorDirs {
			skipDirs[d] = true
		}
	}

	absRoot, _ := filepath.Abs(dirPath)

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories in skip list
		if info.IsDir() {
			if skipDirs[info.Name()] {
				return filepath.SkipDir
			}

			// Skip non-root directories that contain composer.json (third-party packages)
			if phpIncluded {
				absPath, _ := filepath.Abs(path)
				if absPath != absRoot {
					composerPath := filepath.Join(path, "composer.json")
					if _, err := os.Stat(composerPath); err == nil {
						return filepath.SkipDir
					}
				}
			}

			return nil
		}

		// Check if file has a supported extension
		if t.parser.DetectLanguage(path) != "" {
			files = append(files, path)
		}

		return nil
	})

	return files, err
}

// GetTaintedFunctions returns all functions that receive user input
func (t *Tracer) GetTaintedFunctions(result *TraceResult) []*TaintedFunction {
	return result.TaintedFunctions
}

// GetFlowPaths returns all propagation paths from a specific source
func (t *Tracer) GetFlowPaths(result *TraceResult, source *InputSource) []*PropagationPath {
	var paths []*PropagationPath

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
