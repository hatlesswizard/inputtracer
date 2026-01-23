// Package semantic provides a complete semantic input tracer
// that analyzes codebases to trace user input flow with full
// cross-file, inter-procedural analysis.
package semantic

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/hatlesswizard/inputtracer/pkg/parser"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"

	// Tree-sitter language bindings
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/typescript"

	// Import language analyzers to register them
	_ "github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer/c"
	_ "github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer/cpp"
	_ "github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer/csharp"
	_ "github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer/golang"
	_ "github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer/java"
	_ "github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer/javascript"
	_ "github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer/php"
	_ "github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer/python"
	_ "github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer/ruby"
	_ "github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer/rust"
	_ "github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer/typescript"
)

// Config configures the semantic tracer
type Config struct {
	// Languages to analyze (empty = auto-detect all)
	Languages []string

	// MaxDepth for inter-procedural analysis
	MaxDepth int

	// Workers for parallel analysis
	Workers int

	// FollowImports enables cross-file analysis
	FollowImports bool

	// Verbose enables detailed logging
	Verbose bool

	// IncludePatterns for file filtering (glob patterns)
	IncludePatterns []string

	// ExcludePatterns for file filtering (glob patterns)
	ExcludePatterns []string

	// MaxMemoryMB is the maximum memory usage in MB (0 = use default 100MB)
	// Applied to all modes to prevent OOM on large codebases
	MaxMemoryMB int

	// MaxFileSizeBytes is the maximum file size to parse (0 = unlimited)
	MaxFileSizeBytes int64

	// MaxFiles is the maximum number of files to parse (0 = unlimited)
	MaxFiles int
}

// getMemoryUsageMB returns current memory usage in MB (allocated heap memory)
func getMemoryUsageMB() uint64 {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)
	// Use Alloc (currently allocated heap) not Sys (total memory from OS)
	// Sys includes memory the Go runtime reserves but hasn't released to OS,
	// which causes false memory limit triggers when running inside larger applications
	return m.Alloc / 1024 / 1024
}

// DefaultConfig returns sensible defaults
func DefaultConfig() *Config {
	return &Config{
		Languages:        []string{}, // Auto-detect
		MaxDepth:         10,
		Workers:          runtime.NumCPU(),
		FollowImports:    true,
		Verbose:          false,
		MaxFileSizeBytes: 5 * 1024 * 1024, // 5MB - skip very large files (ASTs are ~10x source size)
		IncludePatterns: []string{
			// PHP
			"*.php", "*.php5", "*.php7", "*.phtml", "*.inc",
			// JavaScript/TypeScript
			"*.js", "*.jsx", "*.mjs", "*.cjs",
			"*.ts", "*.tsx", "*.mts", "*.cts",
			// Python
			"*.py", "*.pyw", "*.pyi",
			// Go
			"*.go",
			// Java
			"*.java",
			// C/C++
			"*.c", "*.h",
			"*.cpp", "*.cc", "*.cxx", "*.hpp", "*.hxx", "*.h++",
			// C#
			"*.cs",
			// Ruby
			"*.rb", "*.rake", "*.gemspec",
			// Rust
			"*.rs",
		},
		ExcludePatterns: []string{
			"**/node_modules/**", "**/vendor/**", "**/.git/**",
			"**/dist/**", "**/build/**", "**/__pycache__/**",
			"**/target/**", "**/bin/**", "**/obj/**",
		},
	}
}

// Tracer is the main semantic input tracer
type Tracer struct {
	config *Config

	// Parsers for each language
	parsers map[string]*sitter.Parser

	// Parser service for on-demand AST access (uses LRU cache)
	parserService *parser.Service

	// Cached data
	files       map[string]*FileInfo
	symbolTable *types.SymbolTable
	mu          sync.RWMutex

	// Statistics
	stats *TraceStats
}

// FileInfo holds information about a parsed file
// Optimized to not retain AST and file content in memory after parsing
type FileInfo struct {
	Path        string
	Language    string
	SymbolTable *types.SymbolTable
	Sources     []*types.FlowNode
	Assignments []*types.Assignment // Cached assignments for flow tracing (avoids re-parsing)
	Calls       []*types.CallSite   // Cached calls for flow tracing (avoids re-parsing)
	Root        *sitter.Node        // Only populated during parsing, released after
	Content     []byte              // Only populated if NeedsReparse is false
	ParseTime   time.Duration
	Error       error
	// NeedsReparse indicates the file needs re-parsing for deeper analysis
	// (AST was released to save memory)
	NeedsReparse bool
}

// TraceStats holds tracing statistics
type TraceStats struct {
	FilesScanned     int
	FilesParsed      int
	FilesSkipped     int
	ParseErrors      int
	SourcesFound     int
	FlowsTraced      int
	CrossFileFlows   int
	TotalDuration    time.Duration
	ParseDuration    time.Duration
	AnalysisDuration time.Duration
	ByLanguage       map[string]*LanguageStats
}

// LanguageStats holds per-language statistics
type LanguageStats struct {
	Files        int
	Sources      int
	Flows        int
	ParseErrors  int
	ParseTime    time.Duration
	AnalysisTime time.Duration
}

// TraceResult is the complete result of semantic tracing
type TraceResult struct {
	// All discovered input sources
	Sources []*types.FlowNode

	// Complete flow map
	FlowMap *types.FlowMap

	// Per-file information
	Files map[string]*FileInfo

	// Global symbol table (merged from all files)
	GlobalSymbolTable *types.SymbolTable

	// Per-file symbol tables (for symbolic execution)
	SymbolTable map[string]*types.SymbolTable

	// Statistics
	Stats *TraceStats
}

// TraceContext provides per-trace-invocation isolation for thread safety
// Each TraceBackward() call gets its own context with:
// - Own parser instances (not shared → thread-safe)
// - Cached assignments ONLY (extracted once per file, reused in recursion)
// - NO AST caching (ASTs are huge, assignments are tiny)
// - Released on completion (memory-efficient)
type TraceContext struct {
	phpParser        *sitter.Parser
	jsParser         *sitter.Parser
	assignmentsCache map[string][]*types.Assignment // ONLY cache assignments, NOT ASTs
	mu               sync.RWMutex
}

// newTraceContext creates a new trace context with its own parsers
func newTraceContext() *TraceContext {
	phpParser := sitter.NewParser()
	phpParser.SetLanguage(php.GetLanguage())
	jsParser := sitter.NewParser()
	jsParser.SetLanguage(javascript.GetLanguage())

	return &TraceContext{
		phpParser:        phpParser,
		jsParser:         jsParser,
		assignmentsCache: make(map[string][]*types.Assignment, 64), // Only cache assignments, NOT ASTs
	}
}

// Close releases all resources held by the context
func (ctx *TraceContext) Close() {
	ctx.mu.Lock()
	defer ctx.mu.Unlock()
	// No AST cache to clean up - we only cache assignments (tiny)
	ctx.assignmentsCache = nil
}

// getParser returns the parser for a language
func (ctx *TraceContext) getParser(language string) *sitter.Parser {
	switch language {
	case "php":
		return ctx.phpParser
	case "javascript":
		return ctx.jsParser
	default:
		return nil
	}
}

// getAssignmentsDirectly parses a file, extracts assignments, and IMMEDIATELY discards the AST
// This is memory-efficient: ASTs are huge (5-10x source size), assignments are tiny
func (ctx *TraceContext) getAssignmentsDirectly(filePath string, language string) []*types.Assignment {
	// Check cache first (fast path)
	ctx.mu.RLock()
	if cached, ok := ctx.assignmentsCache[filePath]; ok {
		ctx.mu.RUnlock()
		return cached
	}
	ctx.mu.RUnlock()

	// Cache miss: parse → extract → discard AST
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil
	}

	parser := ctx.getParser(language)
	if parser == nil {
		return nil
	}

	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil || tree == nil {
		return nil
	}

	root := tree.RootNode()

	// Extract assignments
	langAnalyzer := analyzer.DefaultRegistry.Get(language)
	if langAnalyzer == nil {
		tree.Close() // Don't leak memory
		return nil
	}
	assignments, _ := langAnalyzer.ExtractAssignments(root, content, "")

	// CRITICAL: Close the tree immediately to release AST memory
	// Assignments are copied strings, safe to use after tree.Close()
	tree.Close()

	// Cache ONLY assignments (tiny compared to AST)
	ctx.mu.Lock()
	// Double-check pattern
	if existing, ok := ctx.assignmentsCache[filePath]; ok {
		ctx.mu.Unlock()
		return existing
	}
	ctx.assignmentsCache[filePath] = assignments
	ctx.mu.Unlock()

	return assignments
}

// New creates a new semantic tracer
func New(config *Config) *Tracer {
	if config == nil {
		config = DefaultConfig()
	}

	// MEMORY FIX: Set strict memory limit for all modes to prevent OOM
	// Target: stay under 200MB for any analysis (library requirement)
	// Note: runtime.MemStats.Sys underreports actual memory usage by ~30-50%
	// so we use a lower internal limit to stay under 200MB external
	if config.MaxMemoryMB == 0 {
		config.MaxMemoryMB = 120 // 120MB internal -> ~180MB external
	}

	// Keep MaxDepth at default 10 for thorough analysis
	if config.MaxDepth == 0 {
		config.MaxDepth = 10
	}

	// Create parser service with LRU cache for on-demand AST access
	// Small cache to limit memory usage
	cacheSize := 5
	parserSvc := parser.NewService(cacheSize)
	parserSvc.RegisterLanguage("php", php.GetLanguage())
	parserSvc.RegisterLanguage("javascript", javascript.GetLanguage())

	t := &Tracer{
		config:        config,
		parsers:       make(map[string]*sitter.Parser),
		parserService: parserSvc,
		files:         make(map[string]*FileInfo),
		symbolTable: &types.SymbolTable{
			Classes:   make(map[string]*types.ClassDef),
			Functions: make(map[string]*types.FunctionDef),
		},
		stats: &TraceStats{
			ByLanguage: make(map[string]*LanguageStats),
		},
	}

	// Initialize parsers for all languages
	t.initParsers()

	return t
}

// initParsers initializes tree-sitter parsers for available languages
func (t *Tracer) initParsers() {
	languages := map[string]*sitter.Language{
		// PHP
		"php": php.GetLanguage(),
		// JavaScript/TypeScript
		"javascript": javascript.GetLanguage(),
		"typescript": typescript.GetLanguage(),
		// Python
		"python": python.GetLanguage(),
		// Go
		"go": golang.GetLanguage(),
		// Java
		"java": java.GetLanguage(),
		// C/C++
		"c":   c.GetLanguage(),
		"cpp": cpp.GetLanguage(),
		// C#
		"c_sharp": csharp.GetLanguage(),
		// Ruby
		"ruby": ruby.GetLanguage(),
		// Rust
		"rust": rust.GetLanguage(),
	}

	for name, lang := range languages {
		parser := sitter.NewParser()
		parser.SetLanguage(lang)
		t.parsers[name] = parser
	}

	// Also register with parser service for on-demand parsing
	if t.parserService != nil {
		for name, lang := range languages {
			t.parserService.RegisterLanguage(name, lang)
		}
	}
}

// Close releases all resources held by the Tracer
// MEMORY FIX: Call this after analysis to free memory
func (t *Tracer) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Clear parser cache to release AST memory
	if t.parserService != nil {
		t.parserService.ClearCache()
	}

	// Clear file info map
	t.files = make(map[string]*FileInfo)

	// Clear symbol tables
	t.symbolTable = &types.SymbolTable{
		Classes:   make(map[string]*types.ClassDef),
		Functions: make(map[string]*types.FunctionDef),
	}
}

// ParseOnly parses files and builds symbol tables without flow analysis (fast mode for symbolic tracing)
func (t *Tracer) ParseOnly(path string) (*TraceResult, error) {
	startTime := time.Now()

	// Phase 1: Discover files
	if t.config.Verbose {
		fmt.Printf("[Phase 1] Discovering files in %s\n", path)
	}
	files, err := t.discoverFiles(path)
	if err != nil {
		return nil, fmt.Errorf("file discovery failed: %w", err)
	}
	t.stats.FilesScanned = len(files)

	// Apply MaxFiles limit if configured
	maxFiles := t.config.MaxFiles
	if maxFiles > 0 && len(files) > maxFiles {
		if t.config.Verbose {
			fmt.Printf("  Found %d files, limiting to %d\n", len(files), maxFiles)
		}
		files = files[:maxFiles]
	} else if t.config.Verbose {
		fmt.Printf("  Found %d files to analyze\n", len(files))
	}

	// Phase 2: Parse all files in parallel
	if t.config.Verbose {
		fmt.Printf("[Phase 2] Parsing files (workers: %d)\n", t.config.Workers)
	}
	parseStart := time.Now()
	t.parseFiles(files)
	t.stats.ParseDuration = time.Since(parseStart)

	if t.config.Verbose {
		fmt.Printf("  Parsed %d files (%d errors) in %v\n",
			t.stats.FilesParsed, t.stats.ParseErrors, t.stats.ParseDuration)
	}

	// Phase 3: Build global symbol table
	if t.config.Verbose {
		fmt.Printf("[Phase 3] Building global symbol table\n")
	}
	t.buildGlobalSymbolTable()

	if t.config.Verbose {
		fmt.Printf("  Classes: %d, Functions: %d\n",
			len(t.symbolTable.Classes),
			len(t.symbolTable.Functions))
	}

	// MEMORY FIX: Release body sources after symbol table is built
	// This frees large strings that are no longer needed
	t.releaseBodySources()

	t.stats.TotalDuration = time.Since(startTime)

	if t.config.Verbose {
		fmt.Printf("\nParsing complete in %v\n", t.stats.TotalDuration)
	}

	// Build per-file symbol table map
	perFileSymbolTables := make(map[string]*types.SymbolTable)
	for filePath, fileInfo := range t.files {
		if fileInfo.SymbolTable != nil {
			perFileSymbolTables[filePath] = fileInfo.SymbolTable
		}
	}

	return &TraceResult{
		Sources:           nil,
		FlowMap:           &types.FlowMap{},
		Files:             t.files,
		GlobalSymbolTable: t.symbolTable,
		SymbolTable:       perFileSymbolTables,
		Stats:             t.stats,
	}, nil
}

// TraceDirectory performs semantic tracing on a directory
func (t *Tracer) TraceDirectory(path string) (*TraceResult, error) {
	startTime := time.Now()

	// Phase 1: Discover and filter files
	if t.config.Verbose {
		fmt.Printf("[Phase 1] Discovering files in %s\n", path)
	}
	files, err := t.discoverFiles(path)
	if err != nil {
		return nil, fmt.Errorf("file discovery failed: %w", err)
	}
	t.stats.FilesScanned = len(files)

	// Apply MaxFiles limit if configured
	maxFiles := t.config.MaxFiles
	if maxFiles > 0 && len(files) > maxFiles {
		if t.config.Verbose {
			fmt.Printf("  Found %d files, limiting to %d\n", len(files), maxFiles)
		}
		files = files[:maxFiles]
	} else if t.config.Verbose {
		fmt.Printf("  Found %d files to analyze\n", len(files))
	}

	// Phase 2: Parse all files in parallel
	if t.config.Verbose {
		fmt.Printf("[Phase 2] Parsing files (workers: %d)\n", t.config.Workers)
	}
	parseStart := time.Now()
	t.parseFiles(files)
	t.stats.ParseDuration = time.Since(parseStart)

	if t.config.Verbose {
		fmt.Printf("  Parsed %d files (%d errors) in %v\n",
			t.stats.FilesParsed, t.stats.ParseErrors, t.stats.ParseDuration)
	}

	// Phase 3: Build global symbol table
	if t.config.Verbose {
		fmt.Printf("[Phase 3] Building global symbol table\n")
	}
	t.buildGlobalSymbolTable()

	if t.config.Verbose {
		fmt.Printf("  Classes: %d, Functions: %d\n",
			len(t.symbolTable.Classes),
			len(t.symbolTable.Functions))
	}

	// MEMORY FIX: Release per-file symbol tables to reduce memory pressure
	// The global symbol table now has all needed info
	t.releasePerFileSymbolTables()

	// Phase 4: Collect all input sources
	if t.config.Verbose {
		fmt.Printf("[Phase 4] Collecting input sources\n")
	}
	sources := t.collectSources()
	t.stats.SourcesFound = len(sources)

	if t.config.Verbose {
		fmt.Printf("  Found %d input sources\n", len(sources))
	}

	// Phase 5: Cross-file flow analysis
	if t.config.Verbose {
		fmt.Printf("[Phase 5] Cross-file flow analysis\n")
	}
	analysisStart := time.Now()
	flowMap := t.traceAllFlows(sources, path)
	t.stats.AnalysisDuration = time.Since(analysisStart)

	if t.config.Verbose {
		fmt.Printf("  Traced %d flows (%d cross-file) in %v\n",
			t.stats.FlowsTraced, t.stats.CrossFileFlows, t.stats.AnalysisDuration)
	}

	// MEMORY FIX: Release body sources after flow analysis is complete
	// This frees large strings that are no longer needed
	t.releaseBodySources()

	t.stats.TotalDuration = time.Since(startTime)

	if t.config.Verbose {
		fmt.Printf("\nAnalysis complete in %v\n", t.stats.TotalDuration)
		t.printSummary()
	}

	// Build per-file symbol table map
	perFileSymbolTables := make(map[string]*types.SymbolTable)
	for filePath, fileInfo := range t.files {
		if fileInfo.SymbolTable != nil {
			perFileSymbolTables[filePath] = fileInfo.SymbolTable
		}
	}

	return &TraceResult{
		Sources:           sources,
		FlowMap:           flowMap,
		Files:             t.files,
		GlobalSymbolTable: t.symbolTable,
		SymbolTable:       perFileSymbolTables,
		Stats:             t.stats,
	}, nil
}

// TraceFile performs semantic tracing on a single file
func (t *Tracer) TraceFile(path string) (*TraceResult, error) {
	return t.TraceDirectory(filepath.Dir(path))
}

// TraceBackwardBatch performs backward taint analysis for MULTIPLE target expressions in a SINGLE pass
// This is CRITICAL for performance: instead of N × files reads (for N variables),
// we do a single pass through all files, checking all variables at once.
// PERF: Shares TraceContext and assignment cache across all variables
func (t *Tracer) TraceBackwardBatch(targets []string, codebasePath string) (*types.BatchTraceResult, error) {
	startTime := time.Now()

	if len(targets) == 0 {
		return &types.BatchTraceResult{
			HasUserInput:   false,
			PerVariable:    make(map[string]*types.BackwardTraceResult),
			TotalDuration:  time.Since(startTime),
			AnalyzedFiles:  0,
			VariablesFound: 0,
		}, nil
	}

	// First parse the codebase if not already done
	if len(t.files) == 0 {
		_, err := t.ParseOnly(codebasePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse codebase: %w", err)
		}
	}

	result := &types.BatchTraceResult{
		HasUserInput:   false,
		PerVariable:    make(map[string]*types.BackwardTraceResult),
		TotalDuration:  0,
		AnalyzedFiles:  len(t.files),
		VariablesFound: 0,
	}

	// Initialize per-variable results
	for _, target := range targets {
		result.PerVariable[target] = &types.BackwardTraceResult{
			TargetExpression: target,
			Paths:            make([]types.BackwardPath, 0),
			Sources:          make([]types.SourceInfo, 0),
			AnalyzedFiles:    len(t.files),
		}
	}

	// Clean target variable names - build lookup map
	targetVars := make(map[string]string) // cleaned -> original
	for _, target := range targets {
		targetVar := strings.TrimSpace(target)
		if strings.HasPrefix(targetVar, "$") {
			targetVar = targetVar[1:] // Remove $ prefix for matching
		}
		targetVars[targetVar] = target
	}

	// Collect all file paths for processing
	t.mu.RLock()
	filePaths := make([]string, 0, len(t.files))
	for filePath := range t.files {
		filePaths = append(filePaths, filePath)
	}
	t.mu.RUnlock()

	// CRITICAL: Create ONE shared TraceContext for ALL variables
	// This is the key optimization - the assignment cache is shared!
	ctx := newTraceContext()
	defer ctx.Close()

	// Global dedup map for sources
	seenSources := make(map[string]map[string]bool) // variable -> sourceKey -> seen
	for _, target := range targets {
		seenSources[target] = make(map[string]bool)
	}

	// Process ALL files in a SINGLE pass
	for _, filePath := range filePaths {
		// Get file info
		t.mu.RLock()
		fileInfo := t.files[filePath]
		t.mu.RUnlock()
		if fileInfo == nil {
			continue
		}

		// Get cached assignments (parses → extracts → immediately discards AST)
		// This is the expensive operation that's now SHARED across all variables
		assignments := ctx.getAssignmentsDirectly(filePath, fileInfo.Language)
		if assignments == nil {
			continue
		}

		// Check ALL target variables against these assignments in ONE loop
		for _, assign := range assignments {
			assignTarget := strings.TrimPrefix(assign.Target, "$")

			// Check if this assignment is to ANY of our target variables
			originalTarget, isTarget := targetVars[assignTarget]
			if !isTarget {
				continue
			}

			// Found an assignment to one of our targets
			varResult := result.PerVariable[originalTarget]

			path := types.BackwardPath{
				Steps:      make([]types.BackwardStep, 0),
				Confidence: 0.8,
				CrossFile:  false,
			}

			path.Steps = append(path.Steps, types.BackwardStep{
				StepNumber:  1,
				Expression:  fmt.Sprintf("$%s = %s", assignTarget, assign.Source),
				FilePath:    filePath,
				Line:        assign.Line,
				StepType:    "assignment",
				Description: fmt.Sprintf("$%s assigned from %s", assignTarget, assign.Source),
			})

			// Check if the source is a superglobal
			sourceInfo := t.identifySource(assign.Source, filePath, assign.Line)
			if sourceInfo != nil {
				path.Source = *sourceInfo
				path.Steps = append([]types.BackwardStep{{
					StepNumber:  0,
					Expression:  sourceInfo.Expression,
					FilePath:    sourceInfo.FilePath,
					Line:        sourceInfo.Line,
					StepType:    "source",
					Description: fmt.Sprintf("Input source: %s (%s)", sourceInfo.Expression, sourceInfo.Type),
				}}, path.Steps...)

				varResult.Paths = append(varResult.Paths, path)

				sourceKey := fmt.Sprintf("%s:%s", sourceInfo.Type, sourceInfo.Expression)
				if !seenSources[originalTarget][sourceKey] {
					seenSources[originalTarget][sourceKey] = true
					varResult.Sources = append(varResult.Sources, *sourceInfo)
				}

				result.HasUserInput = true
				result.VariablesFound++
			} else if strings.HasPrefix(assign.Source, "$") {
				// The source is another variable - trace recursively WITH SHARED CONTEXT
				innerSources := t.traceBackwardRecursiveWithContext(ctx, assign.Source, filePath, make(map[string]bool), 0)
				for _, innerSource := range innerSources {
					innerPath := types.BackwardPath{
						Source:     innerSource,
						Steps:      make([]types.BackwardStep, 0),
						Confidence: 0.6,
						CrossFile:  innerSource.FilePath != filePath,
					}
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  0,
						Expression:  innerSource.Expression,
						FilePath:    innerSource.FilePath,
						Line:        innerSource.Line,
						StepType:    "source",
						Description: fmt.Sprintf("Input source: %s", innerSource.Expression),
					})
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  1,
						Expression:  assign.Source,
						FilePath:    filePath,
						Line:        assign.Line,
						StepType:    "intermediate",
						Description: fmt.Sprintf("Via %s", assign.Source),
					})
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  2,
						Expression:  fmt.Sprintf("$%s = %s", assignTarget, assign.Source),
						FilePath:    filePath,
						Line:        assign.Line,
						StepType:    "assignment",
						Description: fmt.Sprintf("Assigned to $%s", assignTarget),
					})

					varResult.Paths = append(varResult.Paths, innerPath)

					sourceKey := fmt.Sprintf("%s:%s", innerSource.Type, innerSource.Expression)
					if !seenSources[originalTarget][sourceKey] {
						seenSources[originalTarget][sourceKey] = true
						varResult.Sources = append(varResult.Sources, innerSource)
					}

					result.HasUserInput = true
					result.VariablesFound++
				}
			}
		}
	}

	// Set durations for all per-variable results
	totalDuration := time.Since(startTime)
	for _, varResult := range result.PerVariable {
		varResult.Duration = totalDuration
	}
	result.TotalDuration = totalDuration

	return result, nil
}

// TraceBackward performs backward taint analysis from a target expression (GAP 2)
// This traces from a target variable/expression back to its input sources
func (t *Tracer) TraceBackward(target string, codebasePath string) (*types.BackwardTraceResult, error) {
	startTime := time.Now()

	// First parse the codebase if not already done
	if len(t.files) == 0 {
		_, err := t.ParseOnly(codebasePath)
		if err != nil {
			return nil, fmt.Errorf("failed to parse codebase: %w", err)
		}
	}

	result := &types.BackwardTraceResult{
		TargetExpression: target,
		Paths:            make([]types.BackwardPath, 0),
		Sources:          make([]types.SourceInfo, 0),
		AnalyzedFiles:    len(t.files),
	}

	// Clean target variable name
	targetVar := strings.TrimSpace(target)
	if strings.HasPrefix(targetVar, "$") {
		targetVar = targetVar[1:] // Remove $ prefix for matching
	}

	// Collect all file paths for parallel processing (with lock)
	t.mu.RLock()
	filePaths := make([]string, 0, len(t.files))
	for filePath := range t.files {
		filePaths = append(filePaths, filePath)
	}
	t.mu.RUnlock()

	// If few files, process sequentially with single context
	if len(filePaths) <= 4 {
		ctx := newTraceContext()
		defer ctx.Close()

		seenSources := make(map[string]bool)
		for _, filePath := range filePaths {
			paths, sources := t.traceBackwardInFileWithContext(ctx, filePath, targetVar)
			result.Paths = append(result.Paths, paths...)
			for _, src := range sources {
				sourceKey := fmt.Sprintf("%s:%s", src.Type, src.Expression)
				if !seenSources[sourceKey] {
					seenSources[sourceKey] = true
					result.Sources = append(result.Sources, src)
				}
			}
		}
		result.Duration = time.Since(startTime)
		return result, nil
	}

	// Parallel processing with worker pool
	// MEMORY FIX: Limit workers to prevent memory explosion from parallel AST parsing
	numWorkers := t.config.Workers
	if numWorkers <= 0 {
		numWorkers = 4 // Limit default to 4 workers to control memory
	}
	if numWorkers > 8 {
		numWorkers = 8 // Cap at 8 workers max
	}
	if numWorkers > len(filePaths) {
		numWorkers = len(filePaths)
	}

	// Create file path channel
	pathChan := make(chan string, len(filePaths))
	for _, fp := range filePaths {
		pathChan <- fp
	}
	close(pathChan)

	// Worker results
	type workerResult struct {
		paths   []types.BackwardPath
		sources []types.SourceInfo
	}
	results := make(chan workerResult, numWorkers)

	// Start workers - each with its own TraceContext
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each worker gets its own context (thread-safe, caches AST within worker)
			ctx := newTraceContext()
			defer ctx.Close()

			localPaths := make([]types.BackwardPath, 0, 16)
			localSources := make([]types.SourceInfo, 0, 8)

			for filePath := range pathChan {
				paths, sources := t.traceBackwardInFileWithContext(ctx, filePath, targetVar)
				localPaths = append(localPaths, paths...)
				localSources = append(localSources, sources...)
			}

			results <- workerResult{localPaths, localSources}
		}()
	}

	// Close results when workers done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Merge results
	seenSources := make(map[string]bool)
	for wr := range results {
		result.Paths = append(result.Paths, wr.paths...)
		for _, src := range wr.sources {
			sourceKey := fmt.Sprintf("%s:%s", src.Type, src.Expression)
			if !seenSources[sourceKey] {
				seenSources[sourceKey] = true
				result.Sources = append(result.Sources, src)
			}
		}
	}

	result.Duration = time.Since(startTime)
	return result, nil
}

// traceBackwardInFileWithContext processes a single file for backward tracing using a TraceContext
func (t *Tracer) traceBackwardInFileWithContext(ctx *TraceContext, filePath string, targetVar string) ([]types.BackwardPath, []types.SourceInfo) {
	paths := make([]types.BackwardPath, 0)
	sources := make([]types.SourceInfo, 0)

	// Lock for reading t.files to get file metadata
	t.mu.RLock()
	fileInfo := t.files[filePath]
	t.mu.RUnlock()

	if fileInfo == nil {
		return paths, sources
	}

	// Get cached assignments (parses → extracts → immediately discards AST)
	// This is memory-efficient: only assignments are cached, not ASTs
	assignments := ctx.getAssignmentsDirectly(filePath, fileInfo.Language)
	if assignments == nil {
		return paths, sources
	}

	for _, assign := range assignments {
		// Check if this assignment is to our target variable
		assignTarget := strings.TrimPrefix(assign.Target, "$")
		if assignTarget != targetVar {
			continue
		}

		// Found an assignment to target - trace backward from source
		path := types.BackwardPath{
			Steps:      make([]types.BackwardStep, 0),
			Confidence: 0.8,
			CrossFile:  false,
		}

		// Add the assignment as a step
		path.Steps = append(path.Steps, types.BackwardStep{
			StepNumber:  1,
			Expression:  fmt.Sprintf("$%s = %s", targetVar, assign.Source),
			FilePath:    filePath,
			Line:        assign.Line,
			StepType:    "assignment",
			Description: fmt.Sprintf("$%s assigned from %s", targetVar, assign.Source),
		})

		// Check if the source is a superglobal
		sourceInfo := t.identifySource(assign.Source, filePath, assign.Line)
		if sourceInfo != nil {
			path.Source = *sourceInfo
			path.Steps = append([]types.BackwardStep{{
				StepNumber:  0,
				Expression:  sourceInfo.Expression,
				FilePath:    sourceInfo.FilePath,
				Line:        sourceInfo.Line,
				StepType:    "source",
				Description: fmt.Sprintf("Input source: %s (%s)", sourceInfo.Expression, sourceInfo.Type),
			}}, path.Steps...)

			paths = append(paths, path)
			sources = append(sources, *sourceInfo)
		} else {
			// The source might be another variable - trace recursively
			if strings.HasPrefix(assign.Source, "$") {
				innerSources := t.traceBackwardRecursiveWithContext(ctx, assign.Source, filePath, make(map[string]bool), 0)
				for _, innerSource := range innerSources {
					innerPath := types.BackwardPath{
						Source:     innerSource,
						Steps:      make([]types.BackwardStep, 0),
						Confidence: 0.6,
						CrossFile:  innerSource.FilePath != filePath,
					}
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  0,
						Expression:  innerSource.Expression,
						FilePath:    innerSource.FilePath,
						Line:        innerSource.Line,
						StepType:    "source",
						Description: fmt.Sprintf("Input source: %s", innerSource.Expression),
					})
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  1,
						Expression:  assign.Source,
						FilePath:    filePath,
						Line:        assign.Line,
						StepType:    "intermediate",
						Description: fmt.Sprintf("Via %s", assign.Source),
					})
					innerPath.Steps = append(innerPath.Steps, types.BackwardStep{
						StepNumber:  2,
						Expression:  fmt.Sprintf("$%s = %s", targetVar, assign.Source),
						FilePath:    filePath,
						Line:        assign.Line,
						StepType:    "assignment",
						Description: fmt.Sprintf("Assigned to $%s", targetVar),
					})

					paths = append(paths, innerPath)
					sources = append(sources, innerSource)
				}
			}
		}
	}

	return paths, sources
}

// traceBackwardRecursiveWithContext recursively traces backward with caching and early termination
func (t *Tracer) traceBackwardRecursiveWithContext(ctx *TraceContext, varExpr string, startFile string, visited map[string]bool, depth int) []types.SourceInfo {
	if depth > t.config.MaxDepth {
		return nil
	}

	// Prevent infinite loops
	visitKey := fmt.Sprintf("%s:%s", startFile, varExpr)
	if visited[visitKey] {
		return nil
	}
	visited[visitKey] = true

	var sources []types.SourceInfo
	varName := strings.TrimPrefix(strings.TrimSpace(varExpr), "$")

	// OPTIMIZATION 1: Search current file FIRST (most common case)
	if found := t.searchFileForVar(ctx, startFile, varName, visited, depth, &sources); found {
		return sources // Early termination - found source!
	}

	// OPTIMIZATION 2: Only search other files if not found in current
	// Use read lock, NO map copying
	t.mu.RLock()
	filePaths := make([]string, 0, len(t.files))
	for fp := range t.files {
		if fp != startFile { // Skip already-searched file
			filePaths = append(filePaths, fp)
		}
	}
	t.mu.RUnlock()

	// OPTIMIZATION 3: Search other files with early termination
	for _, filePath := range filePaths {
		if found := t.searchFileForVar(ctx, filePath, varName, visited, depth, &sources); found {
			return sources // Early termination - found source!
		}
	}

	return sources
}

// searchFileForVar searches a single file for variable assignments
// Returns true if a source was found (for early termination)
func (t *Tracer) searchFileForVar(ctx *TraceContext, filePath string, varName string, visited map[string]bool, depth int, sources *[]types.SourceInfo) bool {
	// Get file language
	t.mu.RLock()
	fileInfo := t.files[filePath]
	t.mu.RUnlock()
	if fileInfo == nil {
		return false
	}

	// Get cached assignments (parses → extracts → immediately discards AST)
	// This is memory-efficient: only assignments are cached, not ASTs
	assignments := ctx.getAssignmentsDirectly(filePath, fileInfo.Language)

	for _, assign := range assignments {
		if strings.TrimPrefix(assign.Target, "$") != varName {
			continue
		}

		// Check if source is user input
		if sourceInfo := t.identifySource(assign.Source, filePath, assign.Line); sourceInfo != nil {
			*sources = append(*sources, *sourceInfo)
			return true // FOUND! Early termination
		}

		// Recurse if source is another variable
		if strings.HasPrefix(assign.Source, "$") {
			innerSources := t.traceBackwardRecursiveWithContext(ctx, assign.Source, filePath, visited, depth+1)
			if len(innerSources) > 0 {
				*sources = append(*sources, innerSources...)
				return true // FOUND! Early termination
			}
		}
	}

	return false
}

// identifySource checks if an expression is an input source and returns its info
func (t *Tracer) identifySource(expr string, filePath string, line int) *types.SourceInfo {
	expr = strings.TrimSpace(expr)

	// Check PHP superglobals
	superglobals := map[string]types.SourceType{
		"$_GET":     types.SourceHTTPGet,
		"$_POST":    types.SourceHTTPPost,
		"$_REQUEST": types.SourceType("http_request"),
		"$_COOKIE":  types.SourceHTTPCookie,
		"$_SERVER":  types.SourceHTTPHeader,
		"$_FILES":   types.SourceHTTPBody,
		"$_ENV":     types.SourceEnvVar,
		"$_SESSION": types.SourceType("session"),
	}

	for sg, sourceType := range superglobals {
		if strings.Contains(expr, sg) {
			return &types.SourceInfo{
				Type:       sourceType,
				Expression: expr,
				FilePath:   filePath,
				Line:       line,
				Confidence: 1.0,
			}
		}
	}

	// Check for input functions
	inputFuncs := []string{
		"file_get_contents", "fgets", "fread", "fgetc",
		"getenv", "getallheaders", "apache_request_headers",
	}
	for _, fn := range inputFuncs {
		if strings.Contains(expr, fn+"(") {
			return &types.SourceInfo{
				Type:       types.SourceFile,
				Expression: expr,
				FilePath:   filePath,
				Line:       line,
				Confidence: 0.9,
			}
		}
	}

	// =====================================================
	// UNIVERSAL PHP FRAMEWORK PATTERNS
	// These detect user input across ALL PHP frameworks
	// =====================================================

	// Object property array access patterns (e.g., $mybb->input['key'], $request->data['key'])
	// These are universal patterns used by many PHP frameworks
	inputPropertyPatterns := []string{
		"->input[",    // MyBB, generic
		"->data[",     // Generic data array
		"->request[",  // Symfony, generic
		"->params[",   // Generic params
		"->cookies[",  // Cookie arrays
		"->query[",    // Symfony query bag
		"->post[",     // POST data arrays
		"->get[",      // GET data arrays
		"->files[",    // File uploads
		"->server[",   // Server vars
		"->headers[",  // Headers
		"->attributes[", // PSR-7 attributes
		"->payload[",  // API payloads
		"->args[",     // Arguments
	}
	for _, pattern := range inputPropertyPatterns {
		if strings.Contains(expr, pattern) {
			return &types.SourceInfo{
				Type:       types.SourceUserInput,
				Expression: expr,
				FilePath:   filePath,
				Line:       line,
				Confidence: 0.95,
			}
		}
	}

	// Method call patterns for input getters (e.g., $mybb->get_input('key'), $request->input('key'))
	// These are universal method names used across PHP frameworks
	inputMethodPatterns := []string{
		"->get_input(",     // MyBB
		"->getInput(",      // CamelCase variant
		"->get_var(",       // phpBB, generic
		"->getVar(",        // CamelCase variant
		"->variable(",      // phpBB
		"->input(",         // Laravel
		"->query(",         // Symfony query
		"->post(",          // POST getter
		"->cookie(",        // Cookie getter
		"->header(",        // Header getter
		"->file(",          // File getter
		"->get(",           // Generic getter
		"->all(",           // Get all input (Laravel)
		// PSR-7 methods
		"->getQueryParams(",
		"->getParsedBody(",
		"->getCookieParams(",
		"->getUploadedFiles(",
		"->getServerParams(",
		"->getHeaders(",
		"->getHeader(",
		"->getHeaderLine(",
		"->getAttribute(",
		// Database fetch (can contain user data)
		"->fetch_array(",
		"->fetch_assoc(",
		"->fetch_row(",
		"->fetch_object(",
		"->fetch(",
		"->fetchAll(",
		"->fetchColumn(",
	}
	for _, pattern := range inputMethodPatterns {
		if strings.Contains(expr, pattern) {
			sourceType := types.SourceUserInput
			confidence := 0.9
			// Higher confidence for explicit input methods
			if strings.Contains(pattern, "input") || strings.Contains(pattern, "Input") ||
				strings.Contains(pattern, "Query") || strings.Contains(pattern, "Body") ||
				strings.Contains(pattern, "Cookie") || strings.Contains(pattern, "Header") {
				confidence = 0.95
			}
			// Database results have lower confidence
			if strings.Contains(pattern, "fetch") {
				sourceType = types.SourceDatabase
				confidence = 0.7
			}
			return &types.SourceInfo{
				Type:       sourceType,
				Expression: expr,
				FilePath:   filePath,
				Line:       line,
				Confidence: confidence,
			}
		}
	}

	// Deserialization functions (receive potentially tainted data)
	deserializeFuncs := []string{
		"unserialize(",
		"json_decode(",
		"simplexml_load_string(",
		"yaml_parse(",
	}
	for _, fn := range deserializeFuncs {
		if strings.Contains(expr, fn) {
			return &types.SourceInfo{
				Type:       types.SourceUserInput,
				Expression: expr,
				FilePath:   filePath,
				Line:       line,
				Confidence: 0.85,
			}
		}
	}

	// cURL responses (external data)
	if strings.Contains(expr, "curl_exec(") || strings.Contains(expr, "curl_multi_getcontent(") {
		return &types.SourceInfo{
			Type:       types.SourceType("network"),
			Expression: expr,
			FilePath:   filePath,
			Line:       line,
			Confidence: 0.8,
		}
	}

	return nil
}

// discoverFiles finds all relevant source files
func (t *Tracer) discoverFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			// Check exclude patterns for directories
			rel, _ := filepath.Rel(root, path)
			for _, pattern := range t.config.ExcludePatterns {
				if matched, _ := doubleStarMatch(pattern, rel); matched {
					return filepath.SkipDir
				}
			}
			return nil
		}

		// Check exclude patterns
		rel, _ := filepath.Rel(root, path)
		for _, pattern := range t.config.ExcludePatterns {
			if matched, _ := doubleStarMatch(pattern, rel); matched {
				return nil
			}
		}

		// Check include patterns
		for _, pattern := range t.config.IncludePatterns {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				// Check language filter
				lang := detectLanguage(path)
				if len(t.config.Languages) == 0 || contains(t.config.Languages, lang) {
					files = append(files, path)
				}
				return nil
			}
		}

		return nil
	})

	return files, err
}

// parseFiles parses all files using an optimized worker pool pattern
// This reuses parsers across files within each worker to reduce allocations
func (t *Tracer) parseFiles(files []string) {
	// MEMORY FIX: Limit workers to prevent memory explosion from parallel AST parsing
	// Single worker is most memory-efficient (sequential parsing)
	numWorkers := 1 // Sequential processing for memory safety

	// Create file channel
	fileChan := make(chan string, len(files))
	for _, f := range files {
		fileChan <- f
	}
	close(fileChan)

	// Memory limit tracking
	var memoryExceeded bool
	var memCheckMu sync.Mutex
	filesProcessed := 0
	gcInterval := 25 // GC every 25 files for tighter memory control

	// Start fixed number of workers - each reuses its parser
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Each worker gets its own parsers (reused across all files it processes)
			// We create parsers lazily and cache them per worker
			parsers := make(map[string]*sitter.Parser)

			for path := range fileChan {
				// Check if memory limit exceeded
				memCheckMu.Lock()
				if memoryExceeded {
					memCheckMu.Unlock()
					continue // Skip remaining files
				}
				filesProcessed++
				localCount := filesProcessed
				memCheckMu.Unlock()

				lang := detectLanguage(path)
				if lang == "" {
					continue
				}

				// Get or create parser for this language
				parser, ok := parsers[lang]
				if !ok {
					parser = createParser(lang)
					if parser == nil {
						continue
					}
					parsers[lang] = parser
				}

				t.parseFileWithParser(path, lang, parser)

				// Periodic memory check and GC (enabled for all modes when memory limit is set)
				if t.config.MaxMemoryMB > 0 && localCount%gcInterval == 0 {
					runtime.GC()
					memMB := getMemoryUsageMB()
					maxMB := uint64(t.config.MaxMemoryMB)
					if memMB > maxMB {
						memCheckMu.Lock()
						memoryExceeded = true
						memCheckMu.Unlock()
						if t.config.Verbose {
							fmt.Printf("  [Memory] Limit exceeded (%d MB > %d MB) after %d files - stopping\n",
								memMB, maxMB, localCount)
						}
					} else if t.config.Verbose {
						fmt.Printf("  [Memory] After %d files: %d MB\n", localCount, memMB)
					}
				}
			}
		}()
	}

	wg.Wait()
}

// createParser creates a new parser for a language
func createParser(lang string) *sitter.Parser {
	parser := sitter.NewParser()
	switch lang {
	case "php":
		parser.SetLanguage(php.GetLanguage())
	case "javascript":
		parser.SetLanguage(javascript.GetLanguage())
	case "typescript":
		parser.SetLanguage(typescript.GetLanguage())
	case "python":
		parser.SetLanguage(python.GetLanguage())
	case "go":
		parser.SetLanguage(golang.GetLanguage())
	case "java":
		parser.SetLanguage(java.GetLanguage())
	case "c":
		parser.SetLanguage(c.GetLanguage())
	case "cpp":
		parser.SetLanguage(cpp.GetLanguage())
	case "c_sharp":
		parser.SetLanguage(csharp.GetLanguage())
	case "ruby":
		parser.SetLanguage(ruby.GetLanguage())
	case "rust":
		parser.SetLanguage(rust.GetLanguage())
	default:
		return nil
	}
	return parser
}

// parseFileWithParser parses a single file using the provided parser
// Optimized to release AST memory after extracting symbol table and sources
func (t *Tracer) parseFileWithParser(path string, lang string, parser *sitter.Parser) {
	startTime := time.Now()

	// Get analyzer
	langAnalyzer := analyzer.DefaultRegistry.Get(lang)
	if langAnalyzer == nil {
		return
	}

	// Check file size limit (skip giant files)
	maxFileSize := t.config.MaxFileSizeBytes
	if maxFileSize > 0 {
		fileInfo, err := os.Stat(path)
		if err == nil && fileInfo.Size() > maxFileSize {
			t.mu.Lock()
			t.files[path] = &FileInfo{
				Path:     path,
				Language: lang,
				Error:    fmt.Errorf("file too large: %d bytes (limit: %d)", fileInfo.Size(), maxFileSize),
			}
			t.stats.FilesSkipped++
			t.mu.Unlock()
			if t.config.Verbose {
				fmt.Printf("  Skipping large file: %s (%d MB)\n", path, fileInfo.Size()/1024/1024)
			}
			return
		}
	}

	// Read file content
	content, err := os.ReadFile(path)
	if err != nil {
		t.mu.Lock()
		t.files[path] = &FileInfo{
			Path:     path,
			Language: lang,
			Error:    err,
		}
		t.stats.ParseErrors++
		t.mu.Unlock()
		return
	}

	// Parse with tree-sitter
	tree, err := parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		t.mu.Lock()
		t.files[path] = &FileInfo{
			Path:     path,
			Language: lang,
			Error:    err,
		}
		t.stats.ParseErrors++
		t.mu.Unlock()
		return
	}

	// Get root node before we close the tree
	root := tree.RootNode()

	// Build symbol table (extract all needed info while AST is available)
	symbolTable, err := langAnalyzer.BuildSymbolTable(path, content, root)
	if err != nil {
		// On error, still release the tree
		tree.Close()
		t.mu.Lock()
		t.files[path] = &FileInfo{
			Path:         path,
			Language:     lang,
			Error:        err,
			NeedsReparse: true,
		}
		t.stats.ParseErrors++
		t.mu.Unlock()
		return
	}

	// Find input sources (extract while AST is available)
	sources, err := langAnalyzer.FindInputSources(root, content)
	if err != nil {
		sources = []*types.FlowNode{} // Continue with empty sources on error
	}

	// Update file paths in sources
	for _, src := range sources {
		src.FilePath = path
		if src.ID == "" || !strings.Contains(src.ID, path) {
			src.ID = fmt.Sprintf("%s:%d:%d", path, src.Line, src.Column)
		}
	}

	// MEMORY FIX: Extract assignments and calls NOW to avoid re-parsing during flow tracing
	// This caches lightweight data structures instead of re-creating heavy ASTs later
	var assignments []*types.Assignment
	var calls []*types.CallSite
	if len(sources) > 0 { // Only extract if we found sources (optimization)
		assignments, _ = langAnalyzer.ExtractAssignments(root, content, "")
		calls, _ = langAnalyzer.ExtractCalls(root, content, "")
	}

	// MEMORY OPTIMIZATION: Close the tree to release AST memory
	// We've extracted all needed info into symbolTable, sources, assignments, and calls
	tree.Close()

	parseTime := time.Since(startTime)

	t.mu.Lock()
	t.files[path] = &FileInfo{
		Path:         path,
		Language:     lang,
		SymbolTable:  symbolTable,
		Sources:      sources,
		Assignments:  assignments, // Cached for flow tracing
		Calls:        calls,       // Cached for flow tracing
		Root:         nil,         // Don't retain AST - saves ~10x file size in memory
		Content:      nil,         // Don't retain content - can re-read if needed
		ParseTime:    parseTime,
		NeedsReparse: true, // Mark that AST was released
	}
	t.stats.FilesParsed++

	// Update language stats
	if t.stats.ByLanguage[lang] == nil {
		t.stats.ByLanguage[lang] = &LanguageStats{}
	}
	t.stats.ByLanguage[lang].Files++
	t.stats.ByLanguage[lang].Sources += len(sources)
	t.stats.ByLanguage[lang].ParseTime += parseTime
	t.mu.Unlock()
}

// buildGlobalSymbolTable merges all file symbol tables
func (t *Tracer) buildGlobalSymbolTable() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for filePath, fileInfo := range t.files {
		if fileInfo.SymbolTable == nil {
			continue
		}

		st := fileInfo.SymbolTable

		// Merge classes
		for name, class := range st.Classes {
			key := filePath + "::" + name
			t.symbolTable.Classes[key] = class
			// Also add short name for lookup
			if t.symbolTable.Classes[name] == nil {
				t.symbolTable.Classes[name] = class
			}
		}

		// Merge functions
		for name, fn := range st.Functions {
			key := filePath + "::" + name
			t.symbolTable.Functions[key] = fn
			if t.symbolTable.Functions[name] == nil {
				t.symbolTable.Functions[name] = fn
			}
		}
	}
}

// releaseBodySources releases all body source strings from symbol tables
// MEMORY FIX: Call this after analysis is complete to free large string memory
func (t *Tracer) releaseBodySources() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, fileInfo := range t.files {
		if fileInfo.SymbolTable != nil {
			fileInfo.SymbolTable.ReleaseBodySources()
		}
	}

	// Also release from global symbol table
	for _, class := range t.symbolTable.Classes {
		class.ReleaseBodySources()
	}
	for _, fn := range t.symbolTable.Functions {
		fn.BodySource = ""
	}
}

// releasePerFileSymbolTables releases per-file symbol tables after global table is built
// MEMORY FIX: Frees significant memory by clearing redundant per-file tables
// Call this after buildGlobalSymbolTable() and before flow analysis if memory constrained
func (t *Tracer) releasePerFileSymbolTables() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, fileInfo := range t.files {
		// Keep Sources but release the full SymbolTable
		if fileInfo.SymbolTable != nil {
			// Release body sources first
			fileInfo.SymbolTable.ReleaseBodySources()
			// Clear the per-file tables (global table has what we need)
			fileInfo.SymbolTable.Variables = nil
			fileInfo.SymbolTable.Constants = nil
			fileInfo.SymbolTable.Imports = nil
			// Keep Classes and Functions in global table only
			fileInfo.SymbolTable.Classes = nil
			fileInfo.SymbolTable.Functions = nil
		}
	}
}

// collectSources collects all input sources from all files
func (t *Tracer) collectSources() []*types.FlowNode {
	t.mu.RLock()
	defer t.mu.RUnlock()

	var sources []*types.FlowNode
	for _, fileInfo := range t.files {
		sources = append(sources, fileInfo.Sources...)
	}
	return sources
}

// traceAllFlows traces flows from all sources using parallel workers
func (t *Tracer) traceAllFlows(sources []*types.FlowNode, rootPath string) *types.FlowMap {
	// Use NewFlowMap for O(1) deduplication support
	flowMap := types.NewFlowMap()

	// Add all sources as nodes using O(1) AddNode
	for _, source := range sources {
		flowMap.AddNode(*source)
	}

	// MEMORY FIX: Limit number of sources to trace for memory safety
	// Each source traced requires file re-parsing which consumes memory
	maxSources := 200
	if len(sources) > maxSources {
		if t.config.Verbose {
			fmt.Printf("  Limiting flow analysis to %d sources (of %d) for memory safety\n", maxSources, len(sources))
		}
		sources = sources[:maxSources]
	}

	// Run GC before flow tracing to start with clean slate
	runtime.GC()

	// If few sources, trace sequentially to avoid overhead
	if len(sources) <= 2 {
		for _, source := range sources {
			t.traceSource(source, flowMap, rootPath)
		}
		return flowMap
	}

	// MEMORY FIX: Single worker for flow tracing to minimize memory usage
	// Each source trace uses cached assignments so parallel benefit is minimal
	numWorkers := 1 // Sequential for memory safety

	// Create source channel
	sourceChan := make(chan *types.FlowNode, len(sources))
	for _, s := range sources {
		sourceChan <- s
	}
	close(sourceChan)

	// Worker results channel
	type workerResult struct {
		nodes []types.FlowNode
		edges []types.FlowEdge
	}
	results := make(chan workerResult, numWorkers)

	// Mutex for protecting flowMap writes (for stats and intermediate writes)
	var flowMu sync.Mutex

	// Memory tracking for flow tracing
	var memoryExceeded bool
	var memCheckMu sync.Mutex
	sourcesProcessed := 0
	memCheckInterval := 20 // Check memory every 20 sources

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			// Worker-local storage for results
			localNodes := make([]types.FlowNode, 0, 64)
			localEdges := make([]types.FlowEdge, 0, 128)

			// Create worker-local flowMap for collection
			localFlowMap := types.NewFlowMap()

			for source := range sourceChan {
				// Check if memory limit exceeded
				memCheckMu.Lock()
				if memoryExceeded {
					memCheckMu.Unlock()
					continue // Skip remaining sources
				}
				sourcesProcessed++
				localCount := sourcesProcessed
				memCheckMu.Unlock()

				// Trace into local flowMap
				t.traceSourceParallel(source, localFlowMap, rootPath, &flowMu)

				// Periodic memory check
				if t.config.MaxMemoryMB > 0 && localCount%memCheckInterval == 0 {
					runtime.GC()
					memMB := getMemoryUsageMB()
					maxMB := uint64(t.config.MaxMemoryMB)
					if memMB > maxMB {
						memCheckMu.Lock()
						memoryExceeded = true
						memCheckMu.Unlock()
						if t.config.Verbose {
							fmt.Printf("  [Memory] Flow tracing stopped at %d MB (limit: %d MB)\n", memMB, maxMB)
						}
					}
				}
			}

			// Collect local results
			localNodes = append(localNodes, localFlowMap.AllNodes...)
			localEdges = append(localEdges, localFlowMap.AllEdges...)

			results <- workerResult{localNodes, localEdges}
		}()
	}

	// Close results channel when workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Merge results (dedup happens automatically via AddNode/AddEdge)
	for result := range results {
		for _, n := range result.nodes {
			flowMap.AddNode(n)
		}
		for _, e := range result.edges {
			flowMap.AddEdge(e)
		}
	}

	return flowMap
}

// traceSourceParallel is a thread-safe version of traceSource for parallel execution
// MEMORY FIX: Uses cached assignments and calls to avoid re-parsing files
func (t *Tracer) traceSourceParallel(source *types.FlowNode, flowMap *types.FlowMap, rootPath string, statsMu *sync.Mutex) {
	// Get file info with lock to prevent concurrent map access
	t.mu.RLock()
	fileInfo := t.files[source.FilePath]
	t.mu.RUnlock()

	if fileInfo == nil {
		return
	}

	// Get analyzer (read-only, safe)
	langAnalyzer := analyzer.DefaultRegistry.Get(fileInfo.Language)
	if langAnalyzer == nil {
		return
	}

	// Create initial taint chain for this source
	initialChain := types.NewTaintChain(
		source.Snippet,
		string(source.SourceType),
		source.FilePath,
		source.Line,
	)

	// MEMORY FIX: Use cached assignments instead of re-parsing the file
	// Assignments were extracted during initial parsing to avoid memory explosion
	assignments := fileInfo.Assignments
	if assignments == nil {
		return // No assignments cached (file had no sources during parsing)
	}

	// Find assignments that use this source
	for _, assign := range assignments {
		if assign.IsTainted && containsSourceName(assign.Source, source.Name) {
			varNode := types.FlowNode{
				ID:         fmt.Sprintf("%s:%d:%d", source.FilePath, assign.Line, assign.Column),
				Type:       types.NodeVariable,
				Language:   fileInfo.Language,
				FilePath:   source.FilePath,
				Line:       assign.Line,
				Column:     assign.Column,
				Name:       assign.Target,
				Snippet:    fmt.Sprintf("%s = %s", assign.Target, assign.Source),
				SourceType: source.SourceType,
			}
			flowMap.AddNode(varNode)

			edge := types.FlowEdge{
				From:        source.ID,
				To:          varNode.ID,
				Type:        types.EdgeAssignment,
				Description: "assigned to",
			}
			flowMap.AddEdge(edge)

			// Update stats with lock
			statsMu.Lock()
			t.stats.FlowsTraced++
			statsMu.Unlock()

			// Clone and extend taint chain for this assignment
			varChain := initialChain.Clone()
			varChain.AddStep("assignment", assign.Target, source.FilePath, assign.Line,
				fmt.Sprintf("%s assigned from %s", assign.Target, source.Name))

			// Recursively trace this variable with taint chain
			t.traceVariableWithChain(&varNode, varChain, flowMap, rootPath, fileInfo, langAnalyzer, 1)
		}
	}

	// MEMORY FIX: Use cached calls instead of re-parsing
	calls := fileInfo.Calls
	if calls == nil {
		return // No calls cached
	}

	for _, call := range calls {
		if call.HasTaintedArgs {
			for _, argIdx := range call.TaintedArgIndices {
				if argIdx < len(call.Arguments) {
					arg := call.Arguments[argIdx]
					if containsSourceName(arg.Value, source.Name) {
						t.traceCall(source, call, flowMap, rootPath, 1)
					}
				}
			}
		}
	}
}

// traceSource traces flows from a single source
// MEMORY FIX: Uses cached assignments and calls to avoid re-parsing files
func (t *Tracer) traceSource(source *types.FlowNode, flowMap *types.FlowMap, rootPath string) {
	// Get file info with lock to prevent concurrent map access
	t.mu.RLock()
	fileInfo := t.files[source.FilePath]
	t.mu.RUnlock()

	if fileInfo == nil {
		return
	}

	// Get analyzer
	langAnalyzer := analyzer.DefaultRegistry.Get(fileInfo.Language)
	if langAnalyzer == nil {
		return
	}

	// GAP 5: Create initial taint chain for this source
	initialChain := types.NewTaintChain(
		source.Snippet,
		string(source.SourceType),
		source.FilePath,
		source.Line,
	)

	// MEMORY FIX: Use cached assignments instead of re-parsing
	assignments := fileInfo.Assignments
	if assignments == nil {
		return // No assignments cached
	}

	// Find assignments that use this source
	for _, assign := range assignments {
		if assign.IsTainted && containsSourceName(assign.Source, source.Name) {
			// Create node for the assigned variable
			varNode := types.FlowNode{
				ID:         fmt.Sprintf("%s:%d:%d", source.FilePath, assign.Line, assign.Column),
				Type:       types.NodeVariable,
				Language:   fileInfo.Language,
				FilePath:   source.FilePath,
				Line:       assign.Line,
				Column:     assign.Column,
				Name:       assign.Target,
				Snippet:    fmt.Sprintf("%s = %s", assign.Target, assign.Source),
				SourceType: source.SourceType,
			}
			flowMap.AddNode(varNode)

			// Create edge from source to variable
			edge := types.FlowEdge{
				From:        source.ID,
				To:          varNode.ID,
				Type:        types.EdgeAssignment,
				Description: "assigned to",
			}
			flowMap.AddEdge(edge)
			t.stats.FlowsTraced++

			// GAP 5: Clone and extend taint chain for this assignment
			varChain := initialChain.Clone()
			varChain.AddStep("assignment", assign.Target, source.FilePath, assign.Line,
				fmt.Sprintf("%s assigned from %s", assign.Target, source.Name))

			// Recursively trace this variable with taint chain
			t.traceVariableWithChain(&varNode, varChain, flowMap, rootPath, fileInfo, langAnalyzer, 1)
		}
	}

	// MEMORY FIX: Use cached calls instead of re-parsing
	calls := fileInfo.Calls
	if calls == nil {
		return // No calls cached
	}

	for _, call := range calls {
		if call.HasTaintedArgs {
			for _, argIdx := range call.TaintedArgIndices {
				if argIdx < len(call.Arguments) {
					arg := call.Arguments[argIdx]
					if containsSourceName(arg.Value, source.Name) {
						t.traceCall(source, call, flowMap, rootPath, 1)
					}
				}
			}
		}
	}
}

// traceVariable traces flows from a tainted variable
// MEMORY FIX: Uses cached assignments and calls to avoid re-parsing files
func (t *Tracer) traceVariable(varNode *types.FlowNode, flowMap *types.FlowMap, rootPath string, fileInfo *FileInfo, langAnalyzer analyzer.LanguageAnalyzer, depth int) {
	if depth > t.config.MaxDepth {
		return
	}

	// MEMORY FIX: Use cached assignments instead of re-parsing
	assignments := fileInfo.Assignments
	if assignments == nil {
		return
	}

	for _, assign := range assignments {
		if assign.Line > varNode.Line && containsSourceName(assign.Source, varNode.Name) {
			// Create node for new variable
			newVarNode := types.FlowNode{
				ID:         fmt.Sprintf("%s:%d:%d", varNode.FilePath, assign.Line, assign.Column),
				Type:       types.NodeVariable,
				Language:   fileInfo.Language,
				FilePath:   varNode.FilePath,
				Line:       assign.Line,
				Column:     assign.Column,
				Name:       assign.Target,
				Snippet:    fmt.Sprintf("%s = %s", assign.Target, assign.Source),
				SourceType: varNode.SourceType,
			}

			// Use O(1) AddNode with built-in deduplication
			if flowMap.AddNode(newVarNode) {
				edge := types.FlowEdge{
					From:        varNode.ID,
					To:          newVarNode.ID,
					Type:        types.EdgeAssignment,
					Description: "assigned to",
				}
				flowMap.AddEdge(edge)
				t.stats.FlowsTraced++

				// Recursively trace
				t.traceVariable(&newVarNode, flowMap, rootPath, fileInfo, langAnalyzer, depth+1)
			}
		}
	}

	// MEMORY FIX: Use cached calls instead of re-parsing
	calls := fileInfo.Calls
	if calls == nil {
		return
	}

	for _, call := range calls {
		if call.Line > varNode.Line {
			for i, arg := range call.Arguments {
				if containsSourceName(arg.Value, varNode.Name) {
					// Create copy with taint info for this specific call
					callCopy := *call
					callCopy.HasTaintedArgs = true
					callCopy.TaintedArgIndices = []int{i}
					t.traceCall(varNode, &callCopy, flowMap, rootPath, depth)
				}
			}
		}
	}
}

// traceVariableWithChain traces flows from a tainted variable with full taint chain tracking (GAP 5)
// MEMORY FIX: Uses cached assignments and calls to avoid re-parsing files
func (t *Tracer) traceVariableWithChain(varNode *types.FlowNode, chain *types.TaintChain, flowMap *types.FlowMap, rootPath string, fileInfo *FileInfo, langAnalyzer analyzer.LanguageAnalyzer, depth int) {
	if depth > t.config.MaxDepth {
		return
	}

	// MEMORY FIX: Use cached assignments instead of re-parsing
	assignments := fileInfo.Assignments
	if assignments == nil {
		return
	}

	for _, assign := range assignments {
		if assign.Line > varNode.Line && containsSourceName(assign.Source, varNode.Name) {
			// Create node for new variable
			newVarNode := types.FlowNode{
				ID:         fmt.Sprintf("%s:%d:%d", varNode.FilePath, assign.Line, assign.Column),
				Type:       types.NodeVariable,
				Language:   fileInfo.Language,
				FilePath:   varNode.FilePath,
				Line:       assign.Line,
				Column:     assign.Column,
				Name:       assign.Target,
				Snippet:    fmt.Sprintf("%s = %s", assign.Target, assign.Source),
				SourceType: varNode.SourceType,
			}

			// Use O(1) AddNode with built-in deduplication
			if flowMap.AddNode(newVarNode) {
				edge := types.FlowEdge{
					From:        varNode.ID,
					To:          newVarNode.ID,
					Type:        types.EdgeAssignment,
					Description: "assigned to",
				}
				flowMap.AddEdge(edge)
				t.stats.FlowsTraced++

				// GAP 5: Clone and extend taint chain
				newChain := chain.Clone()
				newChain.AddStep("assignment", assign.Target, varNode.FilePath, assign.Line,
					fmt.Sprintf("%s assigned from %s", assign.Target, varNode.Name))

				// Recursively trace with chain
				t.traceVariableWithChain(&newVarNode, newChain, flowMap, rootPath, fileInfo, langAnalyzer, depth+1)
			}
		}
	}

	// MEMORY FIX: Use cached calls instead of re-parsing
	calls := fileInfo.Calls
	if calls == nil {
		return
	}

	for _, call := range calls {
		if call.Line > varNode.Line {
			for i, arg := range call.Arguments {
				if containsSourceName(arg.Value, varNode.Name) {
					// Create copy with taint info and chain for this specific call
					callCopy := *call
					callCopy.HasTaintedArgs = true
					callCopy.TaintedArgIndices = []int{i}

					// GAP 5: Attach taint chain to the argument
					if i < len(callCopy.Arguments) {
						argChain := chain.Clone()
						argChain.AddStep("parameter", fmt.Sprintf("arg[%d] = %s", i, arg.Value),
							varNode.FilePath, call.Line,
							fmt.Sprintf("passed as argument %d to %s", i, call.FunctionName))
						callCopy.Arguments[i].TaintChain = argChain
					}

					t.traceCallWithChain(varNode, &callCopy, chain, flowMap, rootPath, depth)
				}
			}
		}
	}
}

// traceCallWithChain traces a function call with tainted argument and chain (GAP 5)
func (t *Tracer) traceCallWithChain(source *types.FlowNode, call *types.CallSite, chain *types.TaintChain, flowMap *types.FlowMap, rootPath string, depth int) {
	if depth > t.config.MaxDepth {
		return
	}

	// Create node for the function call
	callNode := types.FlowNode{
		ID:         fmt.Sprintf("%s:%d:%d:call", source.FilePath, call.Line, call.Column),
		Type:       types.NodeSink,
		Language:   source.Language,
		FilePath:   source.FilePath,
		Line:       call.Line,
		Column:     call.Column,
		Name:       call.FunctionName,
		Snippet:    call.FunctionName + "()",
		SourceType: source.SourceType,
	}

	// Use O(1) AddNode with built-in deduplication
	if flowMap.AddNode(callNode) {
		argStr := "arg"
		if len(call.TaintedArgIndices) > 0 {
			argStr = fmt.Sprintf("arg[%d]", call.TaintedArgIndices[0])
		}

		edge := types.FlowEdge{
			From:        source.ID,
			To:          callNode.ID,
			Type:        types.EdgeCall,
			Description: argStr,
		}
		flowMap.AddEdge(edge)
		t.stats.FlowsTraced++
	}

	// If cross-file tracing is enabled, find the function definition and trace into it with chain
	if t.config.FollowImports {
		t.traceIntoFunctionWithChain(&callNode, call, chain, flowMap, rootPath, depth+1)
	}
}

// traceCall traces a function call with tainted argument
func (t *Tracer) traceCall(source *types.FlowNode, call *types.CallSite, flowMap *types.FlowMap, rootPath string, depth int) {
	if depth > t.config.MaxDepth {
		return
	}

	// Create node for the function call
	callNode := types.FlowNode{
		ID:         fmt.Sprintf("%s:%d:%d:call", source.FilePath, call.Line, call.Column),
		Type:       types.NodeSink,
		Language:   source.Language,
		FilePath:   source.FilePath,
		Line:       call.Line,
		Column:     call.Column,
		Name:       call.FunctionName,
		Snippet:    call.FunctionName + "()",
		SourceType: source.SourceType,
	}

	// Use O(1) AddNode with built-in deduplication
	if flowMap.AddNode(callNode) {
		argStr := "arg"
		if len(call.TaintedArgIndices) > 0 {
			argStr = fmt.Sprintf("arg[%d]", call.TaintedArgIndices[0])
		}

		edge := types.FlowEdge{
			From:        source.ID,
			To:          callNode.ID,
			Type:        types.EdgeCall,
			Description: argStr,
		}
		flowMap.AddEdge(edge)
		t.stats.FlowsTraced++
	}

	// If cross-file tracing is enabled, find the function definition and trace into it
	if t.config.FollowImports {
		t.traceIntoFunction(&callNode, call, flowMap, rootPath, depth+1)
	}
}

// traceIntoFunction traces execution into a called function
func (t *Tracer) traceIntoFunction(callNode *types.FlowNode, call *types.CallSite, flowMap *types.FlowMap, rootPath string, depth int) {
	if depth > t.config.MaxDepth {
		return
	}

	// Find function definition in global symbol table
	t.mu.RLock()
	var funcDef *types.FunctionDef
	var funcFile string

	// Try different name patterns
	funcNames := []string{call.FunctionName, call.MethodName}
	if call.ClassName != "" {
		funcNames = append(funcNames, call.ClassName+"::"+call.MethodName)
	}

	for _, name := range funcNames {
		if fn, ok := t.symbolTable.Functions[name]; ok {
			funcDef = fn
			funcFile = fn.FilePath
			break
		}
		// Also search with file prefix
		for key, fn := range t.symbolTable.Functions {
			if strings.HasSuffix(key, "::"+name) {
				funcDef = fn
				funcFile = fn.FilePath
				break
			}
		}
	}
	t.mu.RUnlock()

	if funcDef == nil {
		return
	}

	// Create node for the function definition
	funcNode := types.FlowNode{
		ID:       fmt.Sprintf("%s:%d:func", funcFile, funcDef.Line),
		Type:     types.NodeFunction,
		Language: callNode.Language,
		FilePath: funcFile,
		Line:     funcDef.Line,
		Name:     funcDef.Name,
		Snippet:  funcDef.Name + "()",
	}

	// Use O(1) AddNode with built-in deduplication
	if flowMap.AddNode(funcNode) {
		edge := types.FlowEdge{
			From:        callNode.ID,
			To:          funcNode.ID,
			Type:        types.EdgeCall,
			Description: "calls",
		}
		flowMap.AddEdge(edge)
		t.stats.FlowsTraced++

		if callNode.FilePath != funcFile {
			t.stats.CrossFileFlows++
		}
	}

	// If function has parameters and we have tainted args, trace parameter
	if len(call.TaintedArgIndices) > 0 && len(funcDef.Parameters) > 0 {
		for _, argIdx := range call.TaintedArgIndices {
			if argIdx < len(funcDef.Parameters) {
				param := funcDef.Parameters[argIdx]

				paramNode := types.FlowNode{
					ID:       fmt.Sprintf("%s:%d:param:%s", funcFile, funcDef.Line, param.Name),
					Type:     types.NodeVariable,
					Language: callNode.Language,
					FilePath: funcFile,
					Line:     funcDef.Line,
					Name:     param.Name,
					Snippet:  fmt.Sprintf("param $%s", param.Name),
				}

				// Use O(1) AddNode with built-in deduplication
				if flowMap.AddNode(paramNode) {
					edge := types.FlowEdge{
						From:        funcNode.ID,
						To:          paramNode.ID,
						Type:        types.EdgeDataFlow,
						Description: "param",
					}
					flowMap.AddEdge(edge)
					t.stats.FlowsTraced++

					// Continue tracing inside the function
					t.mu.RLock()
					fileInfo := t.files[funcFile]
					t.mu.RUnlock()

					if fileInfo != nil {
						langAnalyzer := analyzer.DefaultRegistry.Get(fileInfo.Language)
						if langAnalyzer != nil {
							t.traceVariable(&paramNode, flowMap, rootPath, fileInfo, langAnalyzer, depth)
						}
					}
				}
			}
		}
	}
}

// traceIntoFunctionWithChain traces execution into a called function with taint chain (GAP 5)
func (t *Tracer) traceIntoFunctionWithChain(callNode *types.FlowNode, call *types.CallSite, chain *types.TaintChain, flowMap *types.FlowMap, rootPath string, depth int) {
	if depth > t.config.MaxDepth {
		return
	}

	// Find function definition in global symbol table
	t.mu.RLock()
	var funcDef *types.FunctionDef
	var funcFile string

	// Try different name patterns
	funcNames := []string{call.FunctionName, call.MethodName}
	if call.ClassName != "" {
		funcNames = append(funcNames, call.ClassName+"::"+call.MethodName)
	}

	for _, name := range funcNames {
		if fn, ok := t.symbolTable.Functions[name]; ok {
			funcDef = fn
			funcFile = fn.FilePath
			break
		}
		// Also search with file prefix
		for key, fn := range t.symbolTable.Functions {
			if strings.HasSuffix(key, "::"+name) {
				funcDef = fn
				funcFile = fn.FilePath
				break
			}
		}
	}
	t.mu.RUnlock()

	if funcDef == nil {
		return
	}

	// Create node for the function definition
	funcNode := types.FlowNode{
		ID:       fmt.Sprintf("%s:%d:func", funcFile, funcDef.Line),
		Type:     types.NodeFunction,
		Language: callNode.Language,
		FilePath: funcFile,
		Line:     funcDef.Line,
		Name:     funcDef.Name,
		Snippet:  funcDef.Name + "()",
	}

	// Use O(1) AddNode with built-in deduplication
	if flowMap.AddNode(funcNode) {
		edge := types.FlowEdge{
			From:        callNode.ID,
			To:          funcNode.ID,
			Type:        types.EdgeCall,
			Description: "calls",
		}
		flowMap.AddEdge(edge)
		t.stats.FlowsTraced++

		if callNode.FilePath != funcFile {
			t.stats.CrossFileFlows++
		}
	}

	// If function has parameters and we have tainted args, trace parameter with chain
	if len(call.TaintedArgIndices) > 0 && len(funcDef.Parameters) > 0 {
		for _, argIdx := range call.TaintedArgIndices {
			if argIdx < len(funcDef.Parameters) {
				param := funcDef.Parameters[argIdx]

				paramNode := types.FlowNode{
					ID:       fmt.Sprintf("%s:%d:param:%s", funcFile, funcDef.Line, param.Name),
					Type:     types.NodeParam,
					Language: callNode.Language,
					FilePath: funcFile,
					Line:     funcDef.Line,
					Name:     param.Name,
					Snippet:  fmt.Sprintf("param $%s", param.Name),
				}

				// Use O(1) AddNode with built-in deduplication
				if flowMap.AddNode(paramNode) {
					edge := types.FlowEdge{
						From:        funcNode.ID,
						To:          paramNode.ID,
						Type:        types.EdgeParameter,
						Description: fmt.Sprintf("param[%d]", argIdx),
					}
					flowMap.AddEdge(edge)
					t.stats.FlowsTraced++

					// GAP 5: Get taint chain from argument if available, otherwise clone
					var paramChain *types.TaintChain
					if argIdx < len(call.Arguments) && call.Arguments[argIdx].TaintChain != nil {
						paramChain = call.Arguments[argIdx].TaintChain.Clone()
					} else {
						paramChain = chain.Clone()
					}
					paramChain.AddStep("parameter", param.Name, funcFile, funcDef.Line,
						fmt.Sprintf("received as parameter %s in %s", param.Name, funcDef.Name))

					// Continue tracing inside the function with chain
					t.mu.RLock()
					fileInfo := t.files[funcFile]
					t.mu.RUnlock()

					if fileInfo != nil {
						langAnalyzer := analyzer.DefaultRegistry.Get(fileInfo.Language)
						if langAnalyzer != nil {
							t.traceVariableWithChain(&paramNode, paramChain, flowMap, rootPath, fileInfo, langAnalyzer, depth)
						}
					}
				}
			}
		}
	}
}

// printSummary prints analysis summary
func (t *Tracer) printSummary() {
	fmt.Println("\n=== Analysis Summary ===")
	fmt.Printf("Files scanned: %d\n", t.stats.FilesScanned)
	if t.stats.FilesSkipped > 0 {
		fmt.Printf("Files parsed: %d (%d errors, %d skipped for size)\n", t.stats.FilesParsed, t.stats.ParseErrors, t.stats.FilesSkipped)
	} else {
		fmt.Printf("Files parsed: %d (%d errors)\n", t.stats.FilesParsed, t.stats.ParseErrors)
	}
	fmt.Printf("Input sources found: %d\n", t.stats.SourcesFound)
	fmt.Printf("Flows traced: %d (%d cross-file)\n", t.stats.FlowsTraced, t.stats.CrossFileFlows)
	fmt.Printf("\nBy language:\n")

	for lang, stats := range t.stats.ByLanguage {
		fmt.Printf("  %s: %d files, %d sources\n", lang, stats.Files, stats.Sources)
	}
}

// Helper functions

// containsSourceName checks if an expression contains a source reference
func containsSourceName(expr, sourceName string) bool {
	return strings.Contains(expr, sourceName)
}

// detectLanguage detects programming language from file extension
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".php", ".php5", ".php7", ".phtml", ".inc":
		return "php"
	case ".js", ".jsx", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".mts", ".cts":
		return "typescript"
	case ".tsx":
		return "typescript" // TSX uses typescript parser
	case ".py", ".pyw", ".pyi":
		return "python"
	case ".go":
		return "go"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp", ".hxx", ".h++":
		return "cpp"
	case ".cs":
		return "c_sharp"
	case ".rb", ".rake", ".gemspec":
		return "ruby"
	case ".rs":
		return "rust"
	default:
		return ""
	}
}

// contains checks if a string slice contains a string
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// doubleStarMatch matches glob patterns with ** for any subdirectory
func doubleStarMatch(pattern, name string) (bool, error) {
	if strings.Contains(pattern, "**") {
		parts := strings.Split(pattern, "**")
		if len(parts) == 2 {
			prefix := strings.TrimSuffix(parts[0], "/")
			suffix := strings.TrimPrefix(parts[1], "/")

			if prefix != "" && !strings.HasPrefix(name, prefix) {
				return false, nil
			}
			if suffix != "" && !strings.HasSuffix(name, suffix) {
				return false, nil
			}

			// For patterns like "**/node_modules/**"
			middle := strings.Trim(pattern, "*/*")
			if strings.Contains(name, middle) {
				return true, nil
			}

			return filepath.Match(suffix, filepath.Base(name))
		}
	}
	return filepath.Match(pattern, name)
}

// Output methods

// ToJSON outputs the result as JSON
func (r *TraceResult) ToJSON() (string, error) {
	return ToJSON(r)
}

// ToDOT outputs the result as GraphViz DOT
func (r *TraceResult) ToDOT() string {
	return ToDOT(r)
}

// ToMermaid outputs the result as Mermaid diagram
func (r *TraceResult) ToMermaid() string {
	return ToMermaid(r)
}

// ToHTML outputs the result as interactive HTML
func (r *TraceResult) ToHTML() string {
	return ToHTML(r)
}

// Query methods

// GetSourcesByType returns sources filtered by type
func (r *TraceResult) GetSourcesByType(sourceType types.SourceType) []*types.FlowNode {
	var result []*types.FlowNode
	for _, source := range r.Sources {
		if source.SourceType == sourceType {
			result = append(result, source)
		}
	}
	return result
}

// GetSourcesByFile returns sources in a specific file
func (r *TraceResult) GetSourcesByFile(filePath string) []*types.FlowNode {
	var result []*types.FlowNode
	for _, source := range r.Sources {
		if source.FilePath == filePath {
			result = append(result, source)
		}
	}
	return result
}

// HasInputAtFunction checks if a function receives user input
func (r *TraceResult) HasInputAtFunction(funcName string) bool {
	for _, node := range r.FlowMap.AllNodes {
		if (node.Type == types.NodeFunction || node.Type == types.NodeSink) && strings.Contains(node.Name, funcName) {
			// Check if any edge leads to this function
			for _, edge := range r.FlowMap.AllEdges {
				if edge.To == node.ID {
					return true
				}
			}
		}
	}
	return false
}
