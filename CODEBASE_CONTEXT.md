# InputTracer Complete Codebase Documentation
Generated: 2026-01-24
Project Type: Go Library
Total Files Analyzed: 54 core library files

---

## Table of Contents
1. [Project Overview](#1-project-overview)
2. [Directory Structure](#2-complete-directory-structure)
3. [Source Files Documentation](#3-source-files---complete-documentation)
4. [Interfaces](#4-all-interfaces)
5. [Types/Structs](#5-all-typesstructs)
6. [Constants](#6-all-constants)
7. [Data Flow](#7-data-flow-diagrams)
8. [Function Index](#8-function-index-alphabetical)
9. [File Index](#9-file-index)
10. [Dependencies](#10-dependency-graph)

---

## 1. Project Overview

### 1.1 Purpose
InputTracer is a multi-language static analysis library for **taint analysis**. It identifies where user input enters an application (sources), tracks how that input propagates through variables and function calls (taint propagation), and reports which functions receive user-controlled data.

Key capabilities:
- **Source Detection**: Identifies HTTP parameters, CLI args, environment variables, file reads, database results
- **Taint Propagation**: Tracks data flow through assignments and function calls
- **Inter-procedural Analysis**: Follows data across function boundaries
- **Flow Graph Generation**: Outputs DOT/Mermaid/JSON visualizations

### 1.2 Technology Stack
- **Primary Language**: Go 1.21
- **Parser**: Tree-Sitter (via go-tree-sitter bindings)
- **Dependencies**:
  - `github.com/smacker/go-tree-sitter` - Multi-language AST parsing
  - `github.com/google/uuid` - Unique ID generation
  - `github.com/mattn/go-sqlite3` - SQLite support (optional)

### 1.3 Supported Languages
PHP, JavaScript, TypeScript, TSX, Python, Go, Java, C, C++, C#, Ruby, Rust

### 1.4 Entry Point Flow
```
tracer.New(config)
  → RegisterAllLanguages()
  → sources.RegisterAll()
  → ast.RegisterAll()
  → Tracer{}

tracer.TraceDirectory(path)
  → collectFiles()
  → parallel workers: analyzeFile()
    → ParseFile()
    → FindSources()
    → ExtractAssignments() / Track propagation
    → ExtractCalls() / Find tainted function calls
  → mergeFileResult()
  → runInterproceduralAnalysis()
  → buildFlowGraph()
  → TraceResult{}
```

---

## 2. Complete Directory Structure

```
/inputtracer/
├── go.mod                           # Go module definition
├── go.sum                           # Dependency checksums
├── CLAUDE.md                        # Claude Code guidance
├── CODEBASE_CONTEXT.md              # This file
│
├── pkg/
│   ├── tracer/                      # Core tracer (entry point)
│   │   ├── tracer.go               # Main Tracer struct, TraceDirectory()
│   │   ├── types.go                # Core data structures
│   │   ├── propagation.go          # Taint propagation logic
│   │   ├── scope.go                # Variable scope management
│   │   └── interprocedural.go      # Cross-function analysis
│   │
│   ├── parser/                      # Multi-language parsing
│   │   ├── service.go              # Parser service with pooling
│   │   ├── cache.go                # LRU cache with memory limits
│   │   └── languages/
│   │       └── init.go             # Language registration
│   │
│   ├── sources/                     # Input source detection
│   │   ├── registry.go             # Source matcher registry
│   │   ├── php.go                  # PHP source patterns
│   │   ├── javascript.go           # JS/TS source patterns
│   │   ├── python.go               # Python source patterns
│   │   ├── go.go                   # Go source patterns
│   │   ├── java.go                 # Java source patterns
│   │   ├── c.go                    # C source patterns
│   │   ├── cpp.go                  # C++ source patterns
│   │   ├── csharp.go               # C# source patterns
│   │   ├── ruby.go                 # Ruby source patterns
│   │   └── rust.go                 # Rust source patterns
│   │
│   ├── ast/                         # AST extraction
│   │   ├── extractor.go            # Base extractor + Registry
│   │   └── register.go             # Language-specific registration
│   │
│   ├── output/                      # Result export
│   │   ├── json.go                 # JSON exporter
│   │   └── graph.go                # DOT/Mermaid graph export
│   │
│   └── semantic/                    # Advanced semantic analysis
│       ├── tracer.go               # Semantic tracer
│       ├── output.go               # Output utilities
│       ├── types/
│       │   └── types.go            # FlowNode, FlowEdge, FlowMap, SymbolTable
│       ├── analyzer/
│       │   ├── interface.go        # Analyzer interface
│       │   ├── php/analyzer.go
│       │   ├── javascript/analyzer.go
│       │   ├── typescript/analyzer.go
│       │   ├── python/analyzer.go
│       │   ├── golang/analyzer.go
│       │   ├── java/analyzer.go
│       │   ├── c/analyzer.go
│       │   ├── cpp/analyzer.go
│       │   ├── csharp/analyzer.go
│       │   ├── ruby/analyzer.go
│       │   └── rust/analyzer.go
│       ├── discovery/              # Source discovery
│       │   ├── taint.go
│       │   ├── carrier_map.go
│       │   └── superglobal.go
│       ├── classifier/
│       │   └── classifier.go
│       ├── extractor/
│       │   └── extractor.go
│       ├── condition/
│       │   ├── extractor.go
│       │   └── extractor_test.go
│       ├── batch/
│       │   └── analyzer.go
│       ├── pathanalysis/
│       │   ├── expander.go
│       │   └── expander_test.go
│       ├── index/
│       │   ├── indexer.go
│       │   └── indexer_test.go
│       ├── symbolic/
│       │   ├── executor.go
│       │   └── filecache.go
│       ├── callgraph/
│       │   ├── manager.go
│       │   └── manager_test.go
│       └── tracer/
│           └── vartracer.go
│
├── testapps/                        # Test applications (real projects)
│   ├── php/                        # PHP test apps (DVWA, phpMyAdmin, etc.)
│   ├── javascript/                 # JS test apps (Express, Fastify, etc.)
│   ├── typescript/                 # TS test apps
│   ├── python/                     # Python test apps (Django, Flask, etc.)
│   ├── go/                         # Go test apps (Gin, Echo, Fiber, etc.)
│   ├── java/                       # Java test apps
│   ├── c/                          # C test apps (Redis, OpenSSL, etc.)
│   ├── cpp/                        # C++ test apps
│   ├── csharp/                     # C# test apps
│   ├── ruby/                       # Ruby test apps
│   ├── rust/                       # Rust test apps
│   └── mybb/                       # MyBB forum software
│
└── testdata/                        # Test fixtures
```

### 2.1 Directory Purposes
| Directory | Purpose | Key Files |
|-----------|---------|-----------|
| `pkg/tracer/` | Main entry point, orchestration | `tracer.go`, `types.go` |
| `pkg/parser/` | Tree-Sitter parsing service | `service.go`, `cache.go` |
| `pkg/sources/` | Input source pattern matching | `registry.go`, `php.go`, `javascript.go` |
| `pkg/ast/` | Language-agnostic AST extraction | `extractor.go`, `register.go` |
| `pkg/output/` | Export to JSON/DOT/Mermaid | `json.go`, `graph.go` |
| `pkg/semantic/` | Advanced semantic analysis | `tracer.go`, `types/types.go` |
| `testapps/` | Real-world test applications | Various framework examples |

---

## 3. Source Files - Complete Documentation

### 3.1 pkg/tracer/tracer.go
**Location:** `pkg/tracer/tracer.go`
**Purpose:** Main tracer entry point and orchestration
**Lines:** 625

#### Imports
```go
import (
    "fmt", "os", "path/filepath", "runtime", "sync", "time"
    "github.com/google/uuid"
    "github.com/hatlesswizard/inputtracer/pkg/ast"
    "github.com/hatlesswizard/inputtracer/pkg/parser"
    "github.com/hatlesswizard/inputtracer/pkg/parser/languages"
    "github.com/hatlesswizard/inputtracer/pkg/sources"
)
```

#### Functions
| Function | Signature | Lines | Description |
|----------|-----------|-------|-------------|
| DefaultConfig | `func DefaultConfig() *Config` | 40-48 | Returns sensible default config |
| New | `func New(config *Config) *Tracer` | 59-90 | Creates new Tracer with config |
| TraceDirectory | `func (t *Tracer) TraceDirectory(dirPath string) (*TraceResult, error)` | 92-164 | Analyzes entire directory |
| TraceFile | `func (t *Tracer) TraceFile(filePath string) (*TraceResult, error)` | 166-198 | Analyzes single file |
| analyzeFile | `func (t *Tracer) analyzeFile(filePath string) *fileResult` | 211-403 | Per-file analysis |
| mergeFileResult | `func (t *Tracer) mergeFileResult(result *TraceResult, fr *fileResult)` | 405-423 | Merges file results |
| runInterproceduralAnalysis | `func (t *Tracer) runInterproceduralAnalysis(result *TraceResult)` | 425-444 | Cross-function analysis |
| buildFlowGraph | `func (t *Tracer) buildFlowGraph(result *TraceResult)` | 446-519 | Builds flow graph |
| collectFiles | `func (t *Tracer) collectFiles(dirPath string) ([]string, error)` | 521-549 | Collects files to analyze |
| GetTaintedFunctions | `func (t *Tracer) GetTaintedFunctions(result *TraceResult) []*TaintedFunction` | 551-554 | Returns tainted functions |
| GetFlowPaths | `func (t *Tracer) GetFlowPaths(result *TraceResult, source *InputSource) []*PropagationPath` | 556-603 | Returns propagation paths |
| DoesReceiveInput | `func (t *Tracer) DoesReceiveInput(result *TraceResult, funcName string) bool` | 605-614 | Checks if function receives input |
| GetInputSources | `func (t *Tracer) GetInputSources(result *TraceResult) []*InputSource` | 616-619 | Returns all input sources |
| GetTaintedVariables | `func (t *Tracer) GetTaintedVariables(result *TraceResult) []*TaintedVariable` | 621-624 | Returns tainted variables |

#### Types
| Name | Type | Purpose |
|------|------|---------|
| Config | struct | Tracer configuration |
| Tracer | struct | Main tracer instance |
| fileResult | struct | Per-file analysis result |

---

### 3.2 pkg/tracer/types.go
**Location:** `pkg/tracer/types.go`
**Purpose:** Core data structures for trace results
**Lines:** 482

#### Types
| Name | Type | Fields | Purpose |
|------|------|--------|---------|
| InputLabel | string | - | Input category enum |
| Location | struct | FilePath, Line, Column, EndLine, EndColumn, Snippet | Code location |
| InputSource | struct | ID, Type, Key, Location, Labels, Language | User input entry point |
| TaintedVariable | struct | ID, Name, Scope, Source, Location, Depth, Language | Variable holding user input |
| TaintedParam | struct | Index, Name, Source, Path | Function parameter with taint |
| TaintedFunction | struct | ID, Name, FilePath, Line, Language, TaintedParams, ReceivesThrough | Function receiving user input |
| PropagationStepType | string | - | Step type enum |
| PropagationStep | struct | Type, Variable, Function, Location | One step in propagation |
| PropagationPath | struct | Source, Steps, Destination | Complete flow path |
| FlowNode | struct | ID, Type, Name, Location | Graph node |
| FlowEdge | struct | From, To, Type, Location | Graph edge |
| FlowGraph | struct | Nodes, Edges | Complete flow graph |
| TraceStats | struct | FilesAnalyzed, SourcesFound, etc. | Analysis statistics |
| TraceResult | struct | Sources, TaintedVariables, TaintedFunctions, FlowGraph, Stats, Errors | Complete analysis result |
| Scope | struct | ID, Type, Name, Parent, Children, Variables, StartLine, EndLine | Variable scope |
| AnalysisState | struct | CurrentScope, ScopeStack, TaintedValues, FunctionSummaries, VisitedFunctions | Analysis state |
| FullAnalysisState | struct | embedded AnalysisState + dedup maps | Optimized analysis state |
| ParameterInfo | struct | Index, Name, Type | Function parameter info |
| FunctionSummary | struct | Name, FilePath, Language, Parameters, ParamsToReturn, etc. | Function taint summary |

---

### 3.3 pkg/tracer/propagation.go
**Location:** `pkg/tracer/propagation.go`
**Purpose:** Taint propagation through code
**Lines:** 524

#### Functions
| Function | Signature | Lines | Description |
|----------|-----------|-------|-------------|
| getOrCompileRegex | `func getOrCompileRegex(pattern string) *regexp.Regexp` | 15-22 | Cached regex compilation |
| NewTaintPropagator | `func NewTaintPropagator(state *FullAnalysisState, language string) *TaintPropagator` | 30-36 | Creates new propagator |
| PropagateFromAssignment | `func (tp *TaintPropagator) PropagateFromAssignment(...)` | 38-68 | Propagates taint from assignment |
| PropagateFromFunctionCall | `func (tp *TaintPropagator) PropagateFromFunctionCall(...)` | 70-109 | Propagates through function calls |
| PropagateFromReturn | `func (tp *TaintPropagator) PropagateFromReturn(...)` | 111-141 | Propagates from return statements |
| checkTainted | `func (tp *TaintPropagator) checkTainted(...) *TaintInfo` | 149-168 | Checks if value is tainted |
| matchesVariable | `func (tp *TaintPropagator) matchesVariable(value, varName string) bool` | 170-187 | Variable reference matching |
| extractAssignmentParts | `func (tp *TaintPropagator) extractAssignmentParts(...) (target, value string)` | 196-231 | Extracts assignment LHS/RHS |
| extractPHPAssignment | `func (tp *TaintPropagator) extractPHPAssignment(...) (string, string)` | 233-251 | PHP-specific extraction |
| extractJSAssignment | `func (tp *TaintPropagator) extractJSAssignment(...) (string, string)` | 253-274 | JS-specific extraction |
| extractPythonAssignment | `func (tp *TaintPropagator) extractPythonAssignment(...) (string, string)` | 276-295 | Python-specific extraction |
| extractGoAssignment | `func (tp *TaintPropagator) extractGoAssignment(...) (string, string)` | 297-316 | Go-specific extraction |
| extractFunctionName | `func (tp *TaintPropagator) extractFunctionName(...) string` | 375-393 | Extracts function name from call |
| extractArguments | `func (tp *TaintPropagator) extractArguments(...) []Argument` | 395-424 | Extracts function arguments |
| nodeToLocation | `func nodeToLocation(node *sitter.Node, src []byte, filePath string) Location` | 504-523 | Converts node to Location |

---

### 3.4 pkg/tracer/scope.go
**Location:** `pkg/tracer/scope.go`
**Purpose:** Variable scope management
**Lines:** 289

#### Types
| Name | Type | Purpose |
|------|------|---------|
| ScopeType | string | Scope type enum (global, file, module, class, function, block) |
| ScopeManager | struct | Manages variable scopes during analysis |
| ScopedVariable | struct | Variable within a specific scope |

#### Functions
| Function | Signature | Description |
|----------|-----------|-------------|
| NewScopeManager | `func NewScopeManager() *ScopeManager` | Creates new scope manager |
| EnterScope | `func (sm *ScopeManager) EnterScope(...) *Scope` | Creates and enters new scope |
| ExitScope | `func (sm *ScopeManager) ExitScope() *Scope` | Exits current scope |
| CurrentScope | `func (sm *ScopeManager) CurrentScope() *Scope` | Returns current scope |
| DeclareVariable | `func (sm *ScopeManager) DeclareVariable(...) *ScopedVariable` | Declares variable in current scope |
| LookupVariable | `func (sm *ScopeManager) LookupVariable(name string) *ScopedVariable` | Looks up variable respecting scope |
| IsTainted | `func (sm *ScopeManager) IsTainted(name string) bool` | Checks if variable is tainted |
| MarkTainted | `func (sm *ScopeManager) MarkTainted(...)` | Marks variable as tainted |
| GetAllTaintedInScope | `func (sm *ScopeManager) GetAllTaintedInScope() []*ScopedVariable` | Gets all tainted vars in scope |
| Clone | `func (sm *ScopeManager) Clone() *ScopeManager` | Clones scope manager for parallel analysis |

---

### 3.5 pkg/tracer/interprocedural.go
**Location:** `pkg/tracer/interprocedural.go`
**Purpose:** Cross-function taint analysis
**Lines:** 479

#### Types
| Name | Type | Purpose |
|------|------|---------|
| InterproceduralAnalyzer | struct | Handles cross-function taint analysis |

#### Functions
| Function | Signature | Description |
|----------|-----------|-------------|
| NewInterproceduralAnalyzer | `func NewInterproceduralAnalyzer(...) *InterproceduralAnalyzer` | Creates new analyzer |
| BuildFunctionSummary | `func (ipa *InterproceduralAnalyzer) BuildFunctionSummary(...) *FunctionSummary` | Builds function taint summary |
| extractFunctionName | `func (ipa *InterproceduralAnalyzer) extractFunctionName(...) string` | Extracts function name |
| extractParameters | `func (ipa *InterproceduralAnalyzer) extractParameters(...) []ParameterInfo` | Extracts function parameters |
| analyzeFlowWithinFunction | `func (ipa *InterproceduralAnalyzer) analyzeFlowWithinFunction(...)` | Analyzes data flow in function |
| PropagateInterproceduralTaint | `func (ipa *InterproceduralAnalyzer) PropagateInterproceduralTaint(...)` | Propagates taint across calls |
| GetCallGraph | `func (ipa *InterproceduralAnalyzer) GetCallGraph() map[string][]string` | Returns call graph |
| GetFunctionSummary | `func (ipa *InterproceduralAnalyzer) GetFunctionSummary(name string) *FunctionSummary` | Gets function summary |

---

### 3.6 pkg/parser/service.go
**Location:** `pkg/parser/service.go`
**Purpose:** Multi-language parsing service with pooling
**Lines:** 276

#### Types
| Name | Type | Purpose |
|------|------|---------|
| Service | struct | Parser service with caching |
| ParseResult | struct | Result of parsing a file |

#### Functions
| Function | Signature | Description |
|----------|-----------|-------------|
| NewService | `func NewService(cacheSize ...int) *Service` | Creates new parser service |
| RegisterLanguage | `func (s *Service) RegisterLanguage(name string, lang *sitter.Language)` | Registers language parser |
| getParserFromPool | `func (s *Service) getParserFromPool(language string) *sitter.Parser` | Gets parser from pool |
| returnParserToPool | `func (s *Service) returnParserToPool(language string, parser *sitter.Parser)` | Returns parser to pool |
| GetLanguage | `func (s *Service) GetLanguage(name string) *sitter.Language` | Gets registered language |
| SupportedLanguages | `func (s *Service) SupportedLanguages() []string` | Lists supported languages |
| ParseFile | `func (s *Service) ParseFile(filePath string) (*ParseResult, error)` | Parses file with caching |
| ParseWithTree | `func (s *Service) ParseWithTree(source []byte, language string) (*sitter.Tree, *sitter.Node, error)` | Parses and returns tree |
| Parse | `func (s *Service) Parse(source []byte, language string) (*sitter.Node, error)` | Parses source code |
| DetectLanguage | `func (s *Service) DetectLanguage(filePath string) string` | Detects language from extension |
| IsSupported | `func (s *Service) IsSupported(filePath string) bool` | Checks if file type supported |
| ClearCache | `func (s *Service) ClearCache()` | Clears parser cache |

---

### 3.7 pkg/parser/cache.go
**Location:** `pkg/parser/cache.go`
**Purpose:** LRU cache with memory limits
**Lines:** 201

#### Types
| Name | Type | Purpose |
|------|------|---------|
| CachedParse | struct | Cached parse result with tree reference |
| Cache | struct | LRU cache with O(1) operations |
| cacheEntry | struct | Internal cache entry |

#### Functions
| Function | Signature | Description |
|----------|-----------|-------------|
| NewCache | `func NewCache(maxEntries int) *Cache` | Creates cache with 32MB default |
| NewCacheWithMemoryLimit | `func NewCacheWithMemoryLimit(maxEntries int, maxMemory int64) *Cache` | Creates cache with custom limit |
| Get | `func (c *Cache) Get(key string) *CachedParse` | Gets cached result O(1) |
| Put | `func (c *Cache) Put(key string, data *CachedParse)` | Adds to cache O(1) |
| evictOldest | `func (c *Cache) evictOldest()` | Evicts LRU entry |
| Remove | `func (c *Cache) Remove(key string)` | Removes entry |
| Clear | `func (c *Cache) Clear()` | Clears all entries |
| Size | `func (c *Cache) Size() int` | Returns entry count |
| MemoryUsage | `func (c *Cache) MemoryUsage() int64` | Returns memory estimate |
| Stats | `func (c *Cache) Stats() (hits, misses int64)` | Returns hit/miss stats |

---

### 3.8 pkg/sources/registry.go
**Location:** `pkg/sources/registry.go`
**Purpose:** Source matcher registry
**Lines:** 306

#### Types
| Name | Type | Purpose |
|------|------|---------|
| Definition | struct | Input source definition (pattern, labels, node types) |
| Match | struct | Matched source in code |
| Registry | struct | Manages all source matchers |
| BaseMatcher | struct | Common functionality for source matching |

#### Functions
| Function | Signature | Description |
|----------|-----------|-------------|
| NewRegistry | `func NewRegistry() *Registry` | Creates new registry |
| RegisterMatcher | `func (r *Registry) RegisterMatcher(matcher Matcher)` | Registers matcher |
| AddSource | `func (r *Registry) AddSource(def Definition)` | Adds source definition |
| GetMatcher | `func (r *Registry) GetMatcher(language string) Matcher` | Gets matcher for language |
| NewBaseMatcher | `func NewBaseMatcher(language string, sources []Definition) *BaseMatcher` | Creates base matcher |
| FindSources | `func (m *BaseMatcher) FindSources(root *sitter.Node, src []byte) []Match` | Finds sources in AST |
| RegisterAll | `func RegisterAll(r *Registry)` | Registers all language matchers |

---

### 3.9 pkg/ast/extractor.go
**Location:** `pkg/ast/extractor.go`
**Purpose:** Base AST extractor and registry
**Lines:** 312

#### Types
| Name | Type | Purpose |
|------|------|---------|
| Assignment | struct | Assignment operation (LHS, RHS, Scope, Location) |
| CallArgument | struct | Function call argument |
| FunctionCall | struct | Function call (Name, Arguments, Location) |
| Registry | struct | Manages AST extractors |
| BaseExtractor | struct | Common extraction functionality |

#### Functions
| Function | Signature | Description |
|----------|-----------|-------------|
| NewRegistry | `func NewRegistry() *Registry` | Creates new registry |
| Register | `func (r *Registry) Register(extractor Extractor)` | Registers extractor |
| GetExtractor | `func (r *Registry) GetExtractor(language string) Extractor` | Gets extractor |
| NewBaseExtractor | `func NewBaseExtractor(...) *BaseExtractor` | Creates base extractor |
| ExtractAssignments | `func (e *BaseExtractor) ExtractAssignments(...) []Assignment` | Extracts all assignments |
| ExtractCalls | `func (e *BaseExtractor) ExtractCalls(...) []FunctionCall` | Extracts all calls |
| ExpressionContains | `func (e *BaseExtractor) ExpressionContains(...) bool` | Checks if expr contains var |

---

### 3.10 pkg/output/json.go
**Location:** `pkg/output/json.go`
**Purpose:** JSON export functionality
**Lines:** 182

#### Types
| Name | Type | Purpose |
|------|------|---------|
| JSONExporter | struct | JSON exporter with pretty-print |
| SummaryReport | struct | Summary statistics |
| FileStatistic | struct | Per-file statistics |

#### Functions
| Function | Signature | Description |
|----------|-----------|-------------|
| NewJSONExporter | `func NewJSONExporter(prettyPrint bool) *JSONExporter` | Creates JSON exporter |
| Export | `func (e *JSONExporter) Export(result *tracer.TraceResult) (string, error)` | Exports to JSON string |
| ExportToWriter | `func (e *JSONExporter) ExportToWriter(...) error` | Exports to io.Writer |
| ExportToFile | `func (e *JSONExporter) ExportToFile(...) error` | Exports to file |
| GenerateSummary | `func GenerateSummary(result *tracer.TraceResult) *SummaryReport` | Generates summary report |
| ExportSummary | `func (e *JSONExporter) ExportSummary(...) (string, error)` | Exports just summary |

---

### 3.11 pkg/output/graph.go
**Location:** `pkg/output/graph.go`
**Purpose:** DOT/Mermaid graph export
**Lines:** 272

#### Types
| Name | Type | Purpose |
|------|------|---------|
| GraphExporter | struct | Graph exporter |
| PathFinder | struct | Path finding in flow graph |

#### Functions
| Function | Signature | Description |
|----------|-----------|-------------|
| NewGraphExporter | `func NewGraphExporter() *GraphExporter` | Creates graph exporter |
| ExportDOT | `func (e *GraphExporter) ExportDOT(graph *tracer.FlowGraph) string` | Exports to Graphviz DOT |
| ExportMermaid | `func (e *GraphExporter) ExportMermaid(graph *tracer.FlowGraph) string` | Exports to Mermaid |
| ExportJSON | `func (e *GraphExporter) ExportJSON(...) (string, error)` | Exports graph as JSON |
| NewPathFinder | `func NewPathFinder(graph *tracer.FlowGraph, maxDepth int) *PathFinder` | Creates path finder |
| FindAllPaths | `func (pf *PathFinder) FindAllPaths(sourceID string) [][]string` | Finds all paths from source |
| FindPathsToFunction | `func (pf *PathFinder) FindPathsToFunction(funcID string) [][]string` | Finds paths to function |

---

## 4. All Interfaces

### 4.1 Matcher (pkg/sources/registry.go:52)
```go
type Matcher interface {
    Language() string
    FindSources(root *sitter.Node, src []byte) []Match
}
```
**Implemented By:** PHPMatcher, JavaScriptMatcher, TypeScriptMatcher, PythonMatcher, GoMatcher, JavaMatcher, CMatcher, CPPMatcher, CSharpMatcher, RubyMatcher, RustMatcher

### 4.2 Extractor (pkg/ast/extractor.go:42)
```go
type Extractor interface {
    Language() string
    ExtractAssignments(root *sitter.Node, src []byte) []Assignment
    ExtractCalls(root *sitter.Node, src []byte) []FunctionCall
    ExpressionContains(node *sitter.Node, varName string, src []byte) bool
}
```
**Implemented By:** BaseExtractor (all 11 languages use this)

### 4.3 ParserRegistrar (pkg/parser/languages/init.go:93)
```go
type ParserRegistrar interface {
    RegisterLanguage(name string, lang *sitter.Language)
}
```
**Implemented By:** parser.Service

---

## 5. All Types/Structs

### 5.1 Config (pkg/tracer/tracer.go:19)
```go
type Config struct {
    Languages       []string            // Languages to analyze (empty = all)
    MaxDepth        int                 // Inter-procedural analysis depth
    Workers         int                 // Parallel workers
    CustomSources   []sources.Definition // Custom source definitions
    SkipDirs        []string            // Directories to skip
    IncludePatterns []string            // File patterns to include
}
```

### 5.2 Tracer (pkg/tracer/tracer.go:50)
```go
type Tracer struct {
    config   *Config
    parser   *parser.Service
    sources  *sources.Registry
    ast      *ast.Registry
    mu       sync.Mutex
}
```

### 5.3 TraceResult (pkg/tracer/types.go:137)
```go
type TraceResult struct {
    Sources          []*InputSource     // All discovered input sources
    TaintedVariables []*TaintedVariable // Variables holding user input
    TaintedFunctions []*TaintedFunction // Functions receiving user input
    FlowGraph        *FlowGraph         // Complete flow graph
    Stats            TraceStats         // Statistics
    Errors           []string           // Errors encountered
}
```

### 5.4 InputSource (pkg/tracer/types.go:35)
```go
type InputSource struct {
    ID       string       // Unique identifier
    Type     string       // e.g., "$_GET", "req.body"
    Key      string       // e.g., "username" in $_GET['username']
    Location Location     // Code location
    Labels   []InputLabel // Categories
    Language string       // Source language
}
```

### 5.5 FlowMap (pkg/semantic/types/types.go:142)
```go
type FlowMap struct {
    Target       FlowTarget           // Target expression
    Sources      []FlowNode           // Ultimate sources
    Paths        []FlowPath           // Complete paths
    Carriers     []FlowNode           // Intermediate carriers
    AllNodes     []FlowNode           // All nodes
    AllEdges     []FlowEdge           // All edges
    Usages       []FlowNode           // Usage locations
    CarrierChain *CarrierChain        // Carrier chain
    CallGraph    map[string][]string  // Relevant call graph
    Metadata     FlowMapMetadata      // Analysis metadata
    nodeIndex    map[string]bool      // Deduplication
    edgeIndex    map[string]bool      // Deduplication
}
```

---

## 6. All Constants

### 6.1 Input Labels (pkg/tracer/types.go:11-23)
```go
const (
    LabelHTTPGet     InputLabel = "http_get"
    LabelHTTPPost    InputLabel = "http_post"
    LabelHTTPCookie  InputLabel = "http_cookie"
    LabelHTTPHeader  InputLabel = "http_header"
    LabelHTTPBody    InputLabel = "http_body"
    LabelCLI         InputLabel = "cli"
    LabelEnvironment InputLabel = "environment"
    LabelFile        InputLabel = "file"
    LabelDatabase    InputLabel = "database"
    LabelNetwork     InputLabel = "network"
    LabelUserInput   InputLabel = "user_input"
)
```

### 6.2 Propagation Step Types (pkg/tracer/types.go:78-86)
```go
const (
    StepAssignment    PropagationStepType = "assignment"
    StepParameterPass PropagationStepType = "parameter_pass"
    StepReturn        PropagationStepType = "return"
    StepConcatenation PropagationStepType = "concatenation"
    StepArrayAccess   PropagationStepType = "array_access"
    StepObjectAccess  PropagationStepType = "object_access"
    StepDestructure   PropagationStepType = "destructure"
)
```

### 6.3 Scope Types (pkg/tracer/scope.go:8-18)
```go
const (
    ScopeGlobal   ScopeType = "global"
    ScopeFile     ScopeType = "file"
    ScopeModule   ScopeType = "module"
    ScopeClass    ScopeType = "class"
    ScopeFunction ScopeType = "function"
    ScopeBlock    ScopeType = "block"
)
```

### 6.4 Flow Node/Edge Types (pkg/semantic/types/types.go:17-49)
```go
const (
    NodeSource, NodeCarrier, NodeVariable, NodeFunction,
    NodeProperty, NodeParam, NodeReturn, NodeSink FlowNodeType

    EdgeAssignment, EdgeParameter, EdgeReturn, EdgeProperty,
    EdgeArraySet, EdgeArrayGet, EdgeMethodCall, EdgeConstructor,
    EdgeFramework, EdgeConcatenate, EdgeDestructure, EdgeIteration,
    EdgeConditional, EdgeCall, EdgeDataFlow FlowEdgeType
)
```

---

## 7. Data Flow Diagrams

### 7.1 Main Data Flow
```
[User Code]
    │
    ▼
┌────────────────────────────────────────────┐
│  TraceDirectory(path)                      │
│  ├── collectFiles()                        │
│  │   └── Walk directory, filter by ext     │
│  └── Parallel Workers:                     │
│      └── analyzeFile()                     │
└────────────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────────────┐
│  analyzeFile(filePath)                     │
│  ├── DetectLanguage()                      │
│  ├── ParseFile() → AST                     │
│  │   └── Tree-Sitter parsing              │
│  ├── FindSources() → []Match              │
│  │   └── Pattern matching on AST          │
│  ├── ExtractAssignments()                  │
│  │   └── Track taint propagation          │
│  └── ExtractCalls()                        │
│      └── Find tainted function args       │
└────────────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────────────┐
│  mergeFileResult()                         │
│  └── Aggregate all file results           │
└────────────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────────────┐
│  runInterproceduralAnalysis()              │
│  └── Cross-function taint propagation     │
└────────────────────────────────────────────┘
    │
    ▼
┌────────────────────────────────────────────┐
│  buildFlowGraph()                          │
│  └── Create nodes/edges for visualization │
└────────────────────────────────────────────┘
    │
    ▼
[TraceResult]
  ├── Sources
  ├── TaintedVariables
  ├── TaintedFunctions
  └── FlowGraph
```

### 7.2 Source Detection Flow
```
Source Pattern (e.g., "$_GET\[")
    │
    ▼
┌────────────────────────┐
│  BaseMatcher.traverse  │
│  └── Walk AST          │
└────────────────────────┘
    │
    ▼
┌────────────────────────┐
│  Pattern Match?        │
│  ├── Check NodeType    │
│  └── Check Regex       │
└────────────────────────┘
    │ Yes
    ▼
┌────────────────────────┐
│  Extract Key           │
│  └── KeyExtractor      │
└────────────────────────┘
    │
    ▼
┌────────────────────────┐
│  Find Assignment       │
│  └── Walk up to parent │
└────────────────────────┘
    │
    ▼
[Match{SourceType, Key, Variable, Labels}]
```

---

## 8. Function Index (Alphabetical)

| Function | Package | File:Line | Signature |
|----------|---------|-----------|-----------|
| AddNode | semantic/types | types.go:199 | `func (fm *FlowMap) AddNode(node FlowNode) bool` |
| AddPropagationStep | tracer | types.go:370 | `func (s *FullAnalysisState) AddPropagationStep(...)` |
| AddSource | sources | registry.go:80 | `func (r *Registry) AddSource(def Definition)` |
| AddSource | tracer | types.go:321 | `func (s *FullAnalysisState) AddSource(source *InputSource)` |
| AddTaintedFunction | tracer | types.go:345 | `func (s *FullAnalysisState) AddTaintedFunction(tf *TaintedFunction)` |
| AddTaintedVariable | tracer | types.go:329 | `func (s *FullAnalysisState) AddTaintedVariable(tv *TaintedVariable)` |
| BuildFlowGraph | tracer | types.go:399 | `func (s *FullAnalysisState) BuildFlowGraph() *FlowGraph` |
| BuildFunctionSummary | tracer | interprocedural.go:32 | `func (ipa *InterproceduralAnalyzer) BuildFunctionSummary(...) *FunctionSummary` |
| Clear | parser | cache.go:156 | `func (c *Cache) Clear()` |
| ClearCache | parser | service.go:267 | `func (s *Service) ClearCache()` |
| Clone | tracer | scope.go:271 | `func (sm *ScopeManager) Clone() *ScopeManager` |
| collectFiles | tracer | tracer.go:521 | `func (t *Tracer) collectFiles(dirPath string) ([]string, error)` |
| DeclareVariable | tracer | scope.go:95 | `func (sm *ScopeManager) DeclareVariable(...) *ScopedVariable` |
| DefaultConfig | tracer | tracer.go:39 | `func DefaultConfig() *Config` |
| DetectLanguage | parser | service.go:212 | `func (s *Service) DetectLanguage(filePath string) string` |
| DoesReceiveInput | tracer | tracer.go:606 | `func (t *Tracer) DoesReceiveInput(result *TraceResult, funcName string) bool` |
| EnterScope | tracer | scope.go:56 | `func (sm *ScopeManager) EnterScope(...) *Scope` |
| EnterScope | tracer | types.go:230 | `func (s *AnalysisState) EnterScope(...) *Scope` |
| ExitScope | tracer | scope.go:77 | `func (sm *ScopeManager) ExitScope() *Scope` |
| Export | output | json.go:25 | `func (e *JSONExporter) Export(result *tracer.TraceResult) (string, error)` |
| ExportDOT | output | graph.go:19 | `func (e *GraphExporter) ExportDOT(graph *tracer.FlowGraph) string` |
| ExportMermaid | output | graph.go:103 | `func (e *GraphExporter) ExportMermaid(graph *tracer.FlowGraph) string` |
| ExtractAssignments | ast | extractor.go:100 | `func (e *BaseExtractor) ExtractAssignments(...) []Assignment` |
| ExtractCalls | ast | extractor.go:117 | `func (e *BaseExtractor) ExtractCalls(...) []FunctionCall` |
| ExpressionContains | ast | extractor.go:134 | `func (e *BaseExtractor) ExpressionContains(...) bool` |
| FindAllPaths | output | graph.go:207 | `func (pf *PathFinder) FindAllPaths(sourceID string) [][]string` |
| FindSources | sources | registry.go:120 | `func (m *BaseMatcher) FindSources(root *sitter.Node, src []byte) []Match` |
| GenerateSummary | output | json.go:83 | `func GenerateSummary(result *tracer.TraceResult) *SummaryReport` |
| Get | parser | cache.go:68 | `func (c *Cache) Get(key string) *CachedParse` |
| GetAllSummaries | tracer | interprocedural.go:422 | `func (ipa *InterproceduralAnalyzer) GetAllSummaries() map[string]*FunctionSummary` |
| GetAllTaintedInScope | tracer | scope.go:177 | `func (sm *ScopeManager) GetAllTaintedInScope() []*ScopedVariable` |
| GetCallGraph | tracer | interprocedural.go:402 | `func (ipa *InterproceduralAnalyzer) GetCallGraph() map[string][]string` |
| GetExtractor | ast | extractor.go:70 | `func (r *Registry) GetExtractor(language string) Extractor` |
| GetFlowPaths | tracer | tracer.go:556 | `func (t *Tracer) GetFlowPaths(...) []*PropagationPath` |
| GetFunctionSummary | tracer | interprocedural.go:415 | `func (ipa *InterproceduralAnalyzer) GetFunctionSummary(name string) *FunctionSummary` |
| GetInputSources | tracer | tracer.go:616 | `func (t *Tracer) GetInputSources(result *TraceResult) []*InputSource` |
| GetLanguage | parser | service.go:93 | `func (s *Service) GetLanguage(name string) *sitter.Language` |
| GetMatcher | sources | registry.go:87 | `func (r *Registry) GetMatcher(language string) Matcher` |
| GetTaintedFunctions | tracer | tracer.go:551 | `func (t *Tracer) GetTaintedFunctions(result *TraceResult) []*TaintedFunction` |
| GetTaintedVariables | tracer | tracer.go:621 | `func (t *Tracer) GetTaintedVariables(result *TraceResult) []*TaintedVariable` |
| IsTainted | tracer | scope.go:149 | `func (sm *ScopeManager) IsTainted(name string) bool` |
| LookupVariable | tracer | scope.go:120 | `func (sm *ScopeManager) LookupVariable(name string) *ScopedVariable` |
| LookupVariable | tracer | types.go:259 | `func (s *AnalysisState) LookupVariable(name string) (*TaintedVariable, bool)` |
| MarkTainted | tracer | scope.go:164 | `func (sm *ScopeManager) MarkTainted(...)` |
| MemoryUsage | parser | cache.go:182 | `func (c *Cache) MemoryUsage() int64` |
| New | tracer | tracer.go:59 | `func New(config *Config) *Tracer` |
| NewAnalysisState | tracer | types.go:213 | `func NewAnalysisState() *AnalysisState` |
| NewBaseExtractor | ast | extractor.go:85 | `func NewBaseExtractor(...) *BaseExtractor` |
| NewBaseMatcher | sources | registry.go:107 | `func NewBaseMatcher(language string, sources []Definition) *BaseMatcher` |
| NewCache | parser | cache.go:46 | `func NewCache(maxEntries int) *Cache` |
| NewFlowMap | semantic/types | types.go:180 | `func NewFlowMap() *FlowMap` |
| NewFullAnalysisState | tracer | types.go:305 | `func NewFullAnalysisState() *FullAnalysisState` |
| NewGraphExporter | output | graph.go:14 | `func NewGraphExporter() *GraphExporter` |
| NewInterproceduralAnalyzer | tracer | interprocedural.go:21 | `func NewInterproceduralAnalyzer(...) *InterproceduralAnalyzer` |
| NewJavaScriptMatcher | sources | javascript.go:8 | `func NewJavaScriptMatcher() *JavaScriptMatcher` |
| NewJSONExporter | output | json.go:17 | `func NewJSONExporter(prettyPrint bool) *JSONExporter` |
| NewPathFinder | output | graph.go:191 | `func NewPathFinder(graph *tracer.FlowGraph, maxDepth int) *PathFinder` |
| NewPHPMatcher | sources | php.go:8 | `func NewPHPMatcher() *PHPMatcher` |
| NewRegistry | ast | extractor.go:56 | `func NewRegistry() *Registry` |
| NewRegistry | sources | registry.go:65 | `func NewRegistry() *Registry` |
| NewScopeManager | tracer | scope.go:39 | `func NewScopeManager() *ScopeManager` |
| NewService | parser | service.go:29 | `func NewService(cacheSize ...int) *Service` |
| NewTaintPropagator | tracer | propagation.go:30 | `func NewTaintPropagator(...) *TaintPropagator` |
| Parse | parser | service.go:195 | `func (s *Service) Parse(source []byte, language string) (*sitter.Node, error)` |
| ParseFile | parser | service.go:112 | `func (s *Service) ParseFile(filePath string) (*ParseResult, error)` |
| ParseWithTree | parser | service.go:161 | `func (s *Service) ParseWithTree(...) (*sitter.Tree, *sitter.Node, error)` |
| PropagateFromAssignment | tracer | propagation.go:38 | `func (tp *TaintPropagator) PropagateFromAssignment(...)` |
| PropagateFromFunctionCall | tracer | propagation.go:70 | `func (tp *TaintPropagator) PropagateFromFunctionCall(...)` |
| PropagateFromReturn | tracer | propagation.go:111 | `func (tp *TaintPropagator) PropagateFromReturn(...)` |
| PropagateInterproceduralTaint | tracer | interprocedural.go:276 | `func (ipa *InterproceduralAnalyzer) PropagateInterproceduralTaint(...)` |
| Put | parser | cache.go:84 | `func (c *Cache) Put(key string, data *CachedParse)` |
| Register | ast | extractor.go:63 | `func (r *Registry) Register(extractor Extractor)` |
| RegisterAll | ast | register.go:4 | `func RegisterAll(r *Registry)` |
| RegisterAll | sources | registry.go:271 | `func RegisterAll(r *Registry)` |
| RegisterAllLanguages | parser/languages | init.go:97 | `func RegisterAllLanguages(registrar ParserRegistrar)` |
| RegisterLanguage | parser | service.go:43 | `func (s *Service) RegisterLanguage(name string, lang *sitter.Language)` |
| RegisterMatcher | sources | registry.go:73 | `func (r *Registry) RegisterMatcher(matcher Matcher)` |
| Reset | tracer | scope.go:253 | `func (sm *ScopeManager) Reset()` |
| SetTainted | tracer | types.go:276 | `func (s *AnalysisState) SetTainted(name string, tainted *TaintedVariable)` |
| Size | parser | cache.go:175 | `func (c *Cache) Size() int` |
| Stats | parser | cache.go:189 | `func (c *Cache) Stats() (hits, misses int64)` |
| SupportedLanguages | parser | service.go:100 | `func (s *Service) SupportedLanguages() []string` |
| ToJSON | tracer | types.go:158 | `func (r *TraceResult) ToJSON() (string, error)` |
| TraceDirectory | tracer | tracer.go:92 | `func (t *Tracer) TraceDirectory(dirPath string) (*TraceResult, error)` |
| TraceFile | tracer | tracer.go:166 | `func (t *Tracer) TraceFile(filePath string) (*TraceResult, error)` |

---

## 9. File Index

| File Path | Lines | Functions | Types | Purpose |
|-----------|-------|-----------|-------|---------|
| pkg/tracer/tracer.go | 625 | 14 | 3 | Main tracer entry point |
| pkg/tracer/types.go | 482 | 15 | 18 | Core data structures |
| pkg/tracer/propagation.go | 524 | 20 | 2 | Taint propagation |
| pkg/tracer/scope.go | 289 | 14 | 3 | Scope management |
| pkg/tracer/interprocedural.go | 479 | 15 | 1 | Cross-function analysis |
| pkg/parser/service.go | 276 | 15 | 2 | Parser service |
| pkg/parser/cache.go | 201 | 12 | 3 | LRU cache |
| pkg/parser/languages/init.go | 103 | 2 | 1 | Language registration |
| pkg/sources/registry.go | 306 | 12 | 4 | Source matcher registry |
| pkg/sources/php.go | 605 | 1 | 1 | PHP source patterns |
| pkg/sources/javascript.go | 252 | 2 | 2 | JS/TS source patterns |
| pkg/sources/python.go | ~200 | 1 | 1 | Python source patterns |
| pkg/sources/go.go | ~150 | 1 | 1 | Go source patterns |
| pkg/sources/java.go | ~180 | 1 | 1 | Java source patterns |
| pkg/sources/c.go | ~120 | 1 | 1 | C source patterns |
| pkg/sources/cpp.go | ~150 | 1 | 1 | C++ source patterns |
| pkg/sources/csharp.go | ~150 | 1 | 1 | C# source patterns |
| pkg/sources/ruby.go | ~120 | 1 | 1 | Ruby source patterns |
| pkg/sources/rust.go | ~120 | 1 | 1 | Rust source patterns |
| pkg/ast/extractor.go | 312 | 12 | 5 | AST extraction |
| pkg/ast/register.go | 89 | 1 | 0 | Language registration |
| pkg/output/json.go | 182 | 7 | 3 | JSON export |
| pkg/output/graph.go | 272 | 10 | 3 | Graph export |
| pkg/semantic/types/types.go | 1026 | 25+ | 20+ | Semantic types |
| pkg/semantic/tracer.go | ~800 | 20+ | 5 | Semantic tracer |

---

## 10. Dependency Graph

### 10.1 Internal Import Graph
```
pkg/tracer/
├── imports pkg/parser/
├── imports pkg/parser/languages/
├── imports pkg/sources/
└── imports pkg/ast/

pkg/output/
└── imports pkg/tracer/

pkg/semantic/
├── imports pkg/parser/
├── imports pkg/semantic/analyzer/
├── imports pkg/semantic/types/
└── imports tree-sitter language bindings

pkg/sources/
└── imports tree-sitter (sitter.Node)

pkg/ast/
└── imports tree-sitter (sitter.Node)
```

### 10.2 External Dependencies
| Package | Version | Purpose |
|---------|---------|---------|
| github.com/smacker/go-tree-sitter | v0.0.0-20240827 | Multi-language AST parsing |
| github.com/google/uuid | v1.6.0 | Unique ID generation |
| github.com/mattn/go-sqlite3 | v1.14.32 | SQLite support (optional) |

---

## Verification

```
Files Read:
- [x] All tracer files: 5 files
- [x] All parser files: 3 files
- [x] All sources files: 11 files
- [x] All ast files: 2 files
- [x] All output files: 2 files
- [x] Semantic types: 1 file
- [x] Semantic tracer: 1 file
- [x] go.mod: 1 file

TOTAL CORE FILES READ: 26+ files
```
