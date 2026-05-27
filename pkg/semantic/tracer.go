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
	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"

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

	// MaxFlowNodes is the maximum number of nodes in the flow graph (0 = default 10000)
	MaxFlowNodes int

	// MaxFlowEdges is the maximum number of edges in the flow graph (0 = default 20000)
	MaxFlowEdges int
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
		MaxMemoryMB:      120, // 120MB internal limit (~180MB external)
		MaxFileSizeBytes: 5 * 1024 * 1024, // 5MB - skip very large files (ASTs are ~10x source size)
		IncludePatterns:  languages.BuildIncludePatterns(),
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

	// Extracted sub-responsibilities (thin orchestration layer)
	discoverer    *fileDiscoverer
	symbolBuilder *symbolTableBuilder
	flowTracer    *forwardFlowTracer
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

// newTraceContext creates a new trace context with its own parsers.
// Parsers are sourced from the centralised languages.GetAllLanguages() registry
// so that this function never needs direct tree-sitter language imports.
func newTraceContext() *TraceContext {
	// Build a language-name → *sitter.Language index from the registry.
	langMap := make(map[string]*sitter.Language, len(languages.GetAllLanguages()))
	for _, info := range languages.GetAllLanguages() {
		langMap[info.Name] = info.Language
	}

	var phpParser, jsParser *sitter.Parser
	if phpLang, ok := langMap["php"]; ok {
		phpParser = sitter.NewParser()
		phpParser.SetLanguage(phpLang)
	}
	if jsLang, ok := langMap["javascript"]; ok {
		jsParser = sitter.NewParser()
		jsParser.SetLanguage(jsLang)
	}

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

	// Create parser service with LRU cache for on-demand AST access.
	// Small cache to limit memory usage.
	// All supported languages are registered via the centralised registry
	// so that New() never needs direct tree-sitter language imports.
	cacheSize := 5
	parserSvc := parser.NewService(cacheSize)
	languages.RegisterAllLanguages(parserSvc)

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
		discoverer:    newFileDiscoverer(config),
		symbolBuilder: newSymbolTableBuilder(),
	}

	// Initialize parsers for all languages
	t.initParsers()

	// forwardFlowTracer holds references into t; construct it after t is fully
	// initialised so that the shared pointer values are stable.
	t.flowTracer = newForwardFlowTracer(t)

	return t
}

// initParsers initializes tree-sitter parsers for all languages registered in
// the centralised languages.GetAllLanguages() registry.  Adding a new language
// to that registry is the single change required — no edits needed here.
func (t *Tracer) initParsers() {
	for _, info := range languages.GetAllLanguages() {
		p := sitter.NewParser()
		p.SetLanguage(info.Language)
		t.parsers[info.Name] = p
	}

	// The parser service was already populated via languages.RegisterAllLanguages
	// in New(), so no extra registration is required here.
}

// Close releases all resources held by the Tracer
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

// ensureParsed guarantees that t.files is populated for codebasePath.
// It eliminates the ad-hoc "if len(t.files) == 0 { ParseOnly(...) }" blocks
// that were scattered across TraceBackward and TraceBackwardBatch, replacing
// hidden temporal coupling with a single, explicit gate.
func (t *Tracer) ensureParsed(codebasePath string) error {
	t.mu.RLock()
	populated := len(t.files) > 0
	t.mu.RUnlock()
	if populated {
		return nil
	}
	_, err := t.ParseOnly(codebasePath)
	return err
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

	// Release body sources after symbol table is built to free large strings
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

	// Release per-file symbol tables; the global symbol table now has all needed info
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

	// Release body sources after flow analysis is complete to free large strings
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

// discoverFiles finds all relevant source files — delegates to fileDiscoverer.
func (t *Tracer) discoverFiles(root string) ([]string, error) {
	return t.discoverer.discoverFiles(root)
}

// parseFiles parses all files — delegates to fileDiscoverer.
func (t *Tracer) parseFiles(files []string) {
	t.discoverer.parseFiles(t, files)
}

// buildGlobalSymbolTable merges all file symbol tables — delegates to symbolTableBuilder.
func (t *Tracer) buildGlobalSymbolTable() {
	t.symbolBuilder.buildGlobalSymbolTable(t)
}

// releaseBodySources releases body-source strings — delegates to symbolTableBuilder.
func (t *Tracer) releaseBodySources() {
	t.symbolBuilder.releaseBodySources(t)
}

// releasePerFileSymbolTables releases per-file symbol tables — delegates to symbolTableBuilder.
func (t *Tracer) releasePerFileSymbolTables() {
	t.symbolBuilder.releasePerFileSymbolTables(t)
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

// traceAllFlows delegates forward flow tracing to the forwardFlowTracer.
// This keeps Tracer as a thin orchestrator: it coordinates phases but does
// not implement the tracing logic itself.
func (t *Tracer) traceAllFlows(sources []*types.FlowNode, rootPath string) *types.FlowMap {
	return t.flowTracer.traceAllFlows(sources, rootPath)
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
		if node.Type == types.NodeFunction && strings.Contains(node.Name, funcName) {
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
