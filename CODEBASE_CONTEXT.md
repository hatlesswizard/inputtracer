# InputTracer Codebase Context

## Overview

InputTracer is a semantic user input flow tracer written in Go. It analyzes codebases to trace how user input (HTTP parameters, form data, CLI args, environment variables, etc.) flows through application code using taint analysis and inter-procedural data flow tracking.

**Language:** Go 1.21+
**Dependencies:** tree-sitter (AST parsing), uuid, sqlite3

## Architecture Summary

```
cmd/
├── inputtracer/     # Main CLI tool
├── trace/           # Legacy simpler tracer
└── patchleaks-extract/  # PatchLeaks database extraction

pkg/
├── semantic/        # Primary analysis engine
│   ├── tracer.go          # Core Tracer struct
│   ├── types/types.go     # Universal data structures
│   ├── analyzer/          # Language-specific analyzers (11 languages)
│   ├── discovery/         # Input carrier discovery
│   ├── classifier/        # Snippet classification
│   ├── symbolic/          # Symbolic execution engine
│   ├── index/             # O(1) symbol lookup indexer
│   ├── callgraph/         # Call graph management
│   ├── batch/             # Batch analysis
│   ├── condition/         # Condition extraction
│   └── pathanalysis/      # Path expansion
├── sources/         # Language-specific input source definitions
├── tracer/          # Legacy tracer implementation
├── parser/          # Tree-sitter parser service with caching
├── ast/             # AST extraction utilities
└── output/          # Output formatters (JSON, DOT, Mermaid, HTML)
```

## Core Types

### pkg/semantic/types/types.go

**FlowNodeType** - Types of nodes in data flow graph:
- `NodeSource` - Original input source (e.g., $_GET, req.body)
- `NodeCarrier` - Object/variable holding user data
- `NodeVariable` - Regular variable in flow
- `NodeFunction` - Function/method call
- `NodeProperty` - Object property access
- `NodeParam` - Function parameter
- `NodeReturn` - Return value
- `NodeSink` - Final usage point

**FlowEdgeType** - How data flows between nodes:
- `EdgeAssignment`, `EdgeParameter`, `EdgeReturn`, `EdgeProperty`
- `EdgeArraySet`, `EdgeArrayGet`, `EdgeMethodCall`, `EdgeConstructor`
- `EdgeFramework`, `EdgeConcatenate`, `EdgeDestructure`
- `EdgeIteration`, `EdgeConditional`, `EdgeCall`, `EdgeDataFlow`

**SourceType** - Input source categories:
- `SourceHTTPGet`, `SourceHTTPPost`, `SourceHTTPBody`, `SourceHTTPJSON`
- `SourceHTTPHeader`, `SourceHTTPCookie`, `SourceHTTPPath`
- `SourceCLIArg`, `SourceEnvVar`, `SourceStdin`
- `SourceFile`, `SourceDatabase`, `SourceNetwork`, `SourceUserInput`

**FlowNode** struct:
```go
type FlowNode struct {
    ID         string
    Type       FlowNodeType
    Language   string
    FilePath   string
    Line, Column, EndLine, EndColumn int
    Name       string
    ClassName  string
    MethodName string
    Scope      string
    TypeInfo   *TypeInfo
    SourceType SourceType
    SourceKey  string
    CarrierType string
    Snippet    string
    Metadata   map[string]interface{}
}
```

**FlowMap** struct - Complete analysis result with deduplication:
```go
type FlowMap struct {
    Target    FlowTarget
    Sources   []FlowNode
    Paths     []FlowPath
    Carriers  []FlowNode
    AllNodes  []FlowNode  // Max 10,000 nodes
    AllEdges  []FlowEdge  // Max 20,000 edges
    Usages    []FlowNode
    CarrierChain *CarrierChain
    CallGraph map[string][]string
    Metadata  FlowMapMetadata
    // Internal deduplication maps
    nodeIndex map[string]bool
    edgeIndex map[string]bool
}
```

**SymbolTable** struct:
```go
type SymbolTable struct {
    FilePath   string
    Language   string
    Imports    []ImportInfo
    Classes    map[string]*ClassDef
    Functions  map[string]*FunctionDef
    Variables  map[string]*VariableDef
    Constants  map[string]*ConstantDef
    Namespace  string
    Framework  string
    Metadata   map[string]interface{}
}
```

**TaintChain** struct - Tracks taint propagation:
```go
type TaintChain struct {
    OriginalSource    string
    OriginalType      SourceType
    OriginalFile      string
    OriginalLine      int
    Steps             []TaintStep
    CurrentExpression string
    Depth             int
}
```

**BackwardTraceResult** struct - Backward taint analysis result:
```go
type BackwardTraceResult struct {
    TargetExpression string
    TargetFile       string
    TargetLine       int
    Paths            []BackwardPath
    Sources          []SourceInfo
    AnalyzedFiles    int
    Duration         time.Duration
}
```

## Key Packages

### pkg/sources/registry.go

**InputLabel** constants:
- `LabelHTTPGet`, `LabelHTTPPost`, `LabelHTTPCookie`, `LabelHTTPHeader`, `LabelHTTPBody`
- `LabelCLI`, `LabelEnvironment`, `LabelFile`, `LabelDatabase`, `LabelNetwork`, `LabelUserInput`

**Definition** struct:
```go
type Definition struct {
    Name         string
    Pattern      string       // Regex pattern
    Language     string
    Labels       []InputLabel
    Description  string
    NodeTypes    []string     // Tree-sitter node types
    KeyExtractor string       // Regex for key extraction
}
```

**Match** struct:
```go
type Match struct {
    SourceType string
    Key        string
    Variable   string
    Line, Column, EndLine, EndColumn int
    Snippet    string
    Labels     []InputLabel
}
```

**Matcher** interface:
```go
type Matcher interface {
    Language() string
    FindSources(root *sitter.Node, src []byte) []Match
}
```

**Registry** struct with methods:
- `NewRegistry() *Registry`
- `RegisterMatcher(matcher Matcher)`
- `AddSource(def Definition)`
- `GetMatcher(language string) Matcher`
- `GetSources(language string) []Definition`

**BaseMatcher** struct with methods:
- `NewBaseMatcher(language string, sources []Definition) *BaseMatcher`
- `Language() string`
- `FindSources(root *sitter.Node, src []byte) []Match`
- `traverse(node *sitter.Node, src []byte, callback func(*sitter.Node))`
- `findAssignmentTarget(node *sitter.Node, src []byte) string`

Helper functions:
- `isLikelyVariable(s string, lang string) bool`
- `extractVariableName(s string, lang string) string`
- `truncateSnippet(s string, maxLen int) string`
- `RegisterAll(r *Registry)` - Registers all 11 language matchers

### pkg/semantic/discovery/carrier_map.go

**InputCarrier** struct:
```go
type InputCarrier struct {
    ClassName      string
    PropertyName   string
    MethodName     string
    SourceTypes    []string
    AccessPattern  string   // "array", "method", "property"
    Confidence     float64
    FilePath       string
    Line           int
    PopulatedFrom  string
}
```

**CarrierMap** struct:
```go
type CarrierMap struct {
    CodebasePath string
    Framework    string
    Carriers     []InputCarrier
    Statistics   CarrierStats
    GeneratedAt  time.Time
}
```

**CarrierStats** struct:
```go
type CarrierStats struct {
    FilesAnalyzed     int
    ClassesAnalyzed   int
    MethodsAnalyzed   int
    UniqueCarriers    int
    SuperglobalTraces int
}
```

Functions:
- `BuildCarrierMap(codebasePath string) (*CarrierMap, error)`
- `LoadCarrierMap(path string) (*CarrierMap, error)`
- `(cm *CarrierMap) SaveToFile(path string) error`
- `(cm *CarrierMap) FindCarrier(propName string) *InputCarrier`
- `(cm *CarrierMap) FindCarriersByClass(className string) []InputCarrier`
- `(cm *CarrierMap) FindCarriersBySourceType(sourceType string) []InputCarrier`
- `(cm *CarrierMap) Summary() string`
- `DetectFramework(codebasePath string) string`

### pkg/semantic/classifier/classifier.go

**Classifier** struct:
```go
type Classifier struct {
    carrierMap      *discovery.CarrierMap
    extractor       *extractor.ExpressionExtractor
    carrierPatterns []*carrierPattern
}
```

**ClassificationResult** struct:
```go
type ClassificationResult struct {
    Snippet      string
    HasUserInput bool
    NeedsTracing bool
    Expressions  []ExpressionResult
    Summary      ClassificationSummary
}
```

**ExpressionResult** struct:
```go
type ExpressionResult struct {
    Expression     string
    SourceTypes    []string
    Key            string
    Confidence     float64
    MatchedCarrier string
    NeedsTracing   bool
    IsSuperglobal  bool
    IsEscaped      bool
}
```

**BatchResult** struct:
```go
type BatchResult struct {
    AnalyzedAt       string
    CarrierMapPath   string
    Framework        string
    TotalFindings    int
    TotalSnippets    int
    WithUserInput    int
    WithoutUserInput int
    NeedsTracing     int
    BySourceType     map[string]int
    Findings         []FindingResult
}
```

Functions:
- `NewClassifier(carrierMap *discovery.CarrierMap) *Classifier`
- `NewDirectClassifier() *Classifier` - Superglobals only, no carrier map
- `(c *Classifier) ClassifySnippet(snippet string) *ClassificationResult`
- `(c *Classifier) ClassifyBatch(inputs []BatchInput) *BatchResult`
- `LoadBatchInput(path string) ([]BatchInput, error)`
- `LoadSimpleBatchInput(path string) ([]BatchInput, error)`
- `SaveBatchResult(result *BatchResult, path string) error`

### pkg/semantic/index/indexer.go

**SymbolType** constants:
- `SymbolTypeFunction`, `SymbolTypeMethod`, `SymbolTypeClass`
- `SymbolTypeVariable`, `SymbolTypeConstant`, `SymbolTypeProperty`
- `SymbolTypeParameter`, `SymbolTypeImport`

**Symbol** struct:
```go
type Symbol struct {
    Name       string
    Type       SymbolType
    FilePath   string
    Line       int
    Column     int
    Signature  string
    Parameters []string
    ReturnType string
    ClassName  string
    Visibility string
    IsStatic   bool
    References []Reference
}
```

**Reference** struct:
```go
type Reference struct {
    FilePath string
    Line     int
    Column   int
    RefType  string  // "call", "access", "assignment", etc.
}
```

**Indexer** struct with methods:
- `NewIndexer() *Indexer`
- `AddSymbol(symbol *Symbol)`
- `Search(query string, opts *SearchOptions) []*Symbol`
- `FindDefinition(name string, symbolType SymbolType) *Symbol`
- `FindUsages(name string) []Reference`
- `GetSymbolsInFile(filePath string) []*Symbol`
- `GetClassMembers(className string) []*Symbol`
- `Stats() IndexerStats`

### pkg/semantic/callgraph/manager.go

**NodeType** constants:
- `NodeTypeRegular`, `NodeTypeEntryPoint`, `NodeTypeSink`, `NodeTypeSource`

**Node** struct:
```go
type Node struct {
    ID                string
    Name              string
    FilePath          string
    Line              int
    Type              NodeType
    Language          string
    ClassName         string
    IsPublic          bool
    Signature         string
    DistanceFromEntry int  // -1 if unreachable
    DistanceToSink    int  // -1 if no path
}
```

**Edge** struct:
```go
type Edge struct {
    CallerID      string
    CalleeID      string
    Line, Column  int
    FilePath      string
    ArgumentCount int
    IsConditional bool
    InLoop        bool
    BranchDepth   int
}
```

**Manager** struct with methods:
- `NewManager() *Manager`
- `AddNode(node *Node)`
- `AddEdge(edge *Edge)`
- `GetNode(id string) *Node`
- `GetCallees(callerID string) []*Node`
- `GetCallers(calleeID string) []*Node`
- `ComputeDistanceFromEntryPoints()` - BFS from all entry points
- `ComputeDistanceToSinks()` - Reverse BFS from all sinks
- `GetShortestPath(fromID, toID string) []*Node`
- `GetDistance(fromID, toID string) int` - With LRU caching
- `GetEntryPoints() []*Node`
- `GetSinks() []*Node`
- `GetReachableSinks(fromID string) []*Node`
- `GetAllPathsToSinks(fromID string, maxPaths int) [][]*Node`
- `PriorityScore(nodeID string) float64` - ATLANTIS-inspired prioritization
- `Stats() map[string]int`

Helper functions:
- `MakeNodeID(filePath, funcName string) string`
- `MakeMethodID(filePath, className, methodName string) string`

### pkg/semantic/analyzer/interface.go

**LanguageAnalyzer** interface:
```go
type LanguageAnalyzer interface {
    Language() string
    SupportedExtensions() []string
    BuildSymbolTable(filePath string, source []byte, root *sitter.Node) (*types.SymbolTable, error)
    ResolveImports(symbolTable *types.SymbolTable, basePath string) ([]string, error)
    ExtractClasses(root *sitter.Node, source []byte) ([]*types.ClassDef, error)
    ExtractFunctions(root *sitter.Node, source []byte) ([]*types.FunctionDef, error)
    ExtractAssignments(root *sitter.Node, source []byte, scope string) ([]*types.Assignment, error)
    ExtractCalls(root *sitter.Node, source []byte, scope string) ([]*types.CallSite, error)
    FindInputSources(root *sitter.Node, source []byte) ([]*types.FlowNode, error)
    AnalyzeMethodBody(method *types.MethodDef, source []byte, state *types.AnalysisState) (*MethodFlowAnalysis, error)
    DetectFrameworks(symbolTable *types.SymbolTable, source []byte) ([]string, error)
    GetFrameworkPatterns() []*types.FrameworkPattern
    TraceExpression(target types.FlowTarget, state *types.AnalysisState) (*types.FlowMap, error)
}
```

AST Helper functions:
- `GetNodeText(node *sitter.Node, source []byte) string`
- `FindChildByType(node *sitter.Node, nodeType string) *sitter.Node`
- `FindChildrenByType(node *sitter.Node, nodeType string) []*sitter.Node`
- `FindChildByFieldName(node *sitter.Node, fieldName string) *sitter.Node`
- `TraverseTree(node *sitter.Node, callback func(*sitter.Node) bool)`
- `FindNodesOfType(root *sitter.Node, nodeType string) []*sitter.Node`
- `GetAncestorOfType(node *sitter.Node, nodeType string) *sitter.Node`
- `GetEnclosingFunction(node *sitter.Node, functionTypes []string) *sitter.Node`
- `GetEnclosingClass(node *sitter.Node, classTypes []string) *sitter.Node`
- `NodeLocation(node *sitter.Node, filePath string) types.Location`
- `CreateFlowNode(node *sitter.Node, source []byte, filePath, language string, nodeType types.FlowNodeType) *types.FlowNode`
- `GenerateNodeID(filePath string, node *sitter.Node) string`

### pkg/semantic/tracer/vartracer.go

**VariableTracer** struct:
```go
type VariableTracer struct {
    parser     *sitter.Parser
    carrierMap *discovery.CarrierMap
    codebase   string
}
```

**TraceReport** struct:
```go
type TraceReport struct {
    Variable           string
    Codebase           string
    TotalDefinitions   int
    WithUserInput      int
    WithoutUserInput   int
    Definitions        []VariableTraceResult
}
```

**VariableTraceResult** struct:
```go
type VariableTraceResult struct {
    File           string
    Line           int
    FunctionName   string
    ClassName      string
    HasUserInput   bool
    InputSources   []string
    FlowPath       []string
    MatchedCarrier string
    Assignments    []Assignment
    ParameterInfo  *ParameterTaintInfo
}
```

**ParameterTaintInfo** struct:
```go
type ParameterTaintInfo struct {
    FunctionName   string
    ParameterName  string
    ParameterIndex int
    IsTainted      bool
    Sources        []string
    CallSites      []CallSiteInfo
}
```

Functions:
- `NewVariableTracer(codebase string, carrierMap *discovery.CarrierMap) *VariableTracer`
- `(t *VariableTracer) TraceVariable(varName string) (*TraceReport, error)`
- `(r *TraceReport) Summary() string`

### pkg/parser/service.go

**Service** struct:
```go
type Service struct {
    languages   map[string]*sitter.Language
    cache       *Cache
    mu          sync.RWMutex
    parserPools map[string]*sync.Pool
}
```

**ParseResult** struct:
```go
type ParseResult struct {
    Root     *sitter.Node
    Source   []byte
    Language string
    FilePath string
}
```

Functions:
- `NewService(cacheSize ...int) *Service`
- `(s *Service) RegisterLanguage(name string, lang *sitter.Language)`
- `(s *Service) GetLanguage(name string) *sitter.Language`
- `(s *Service) SupportedLanguages() []string`
- `(s *Service) ParseFile(filePath string) (*ParseResult, error)`
- `(s *Service) ParseWithTree(source []byte, language string) (*sitter.Tree, *sitter.Node, error)`
- `(s *Service) Parse(source []byte, language string) (*sitter.Node, error)`
- `(s *Service) ParseString(source string, language string) (*sitter.Node, error)`
- `(s *Service) DetectLanguage(filePath string) string`
- `(s *Service) IsSupported(filePath string) bool`
- `(s *Service) ClearCache()`
- `(s *Service) CacheStats() (hits, misses int64)`

### pkg/parser/cache.go

**CachedParse** struct:
```go
type CachedParse struct {
    Root   *sitter.Node
    Tree   *sitter.Tree  // For proper cleanup on eviction
    Source []byte
}
```

**Cache** struct - O(1) LRU cache:
```go
type Cache struct {
    maxEntries int
    maxMemory  int64  // Default 32MB
    currentMem int64
    items      map[string]*list.Element
    evictList  *list.List
    mu         sync.RWMutex
    hits, misses int64
}
```

Functions:
- `NewCache(maxEntries int) *Cache`
- `NewCacheWithMemoryLimit(maxEntries int, maxMemory int64) *Cache`
- `(c *Cache) Get(key string) *CachedParse`
- `(c *Cache) Put(key string, data *CachedParse)`
- `(c *Cache) Remove(key string)`
- `(c *Cache) Clear()`
- `(c *Cache) Size() int`
- `(c *Cache) MemoryUsage() int64`
- `(c *Cache) Stats() (hits, misses int64)`

### pkg/tracer/tracer.go (Legacy)

**Tracer** struct:
```go
type Tracer struct {
    config   *Config
    parser   *parser.Service
    sources  *sources.Registry
    ast      *ast.Registry
    mu       sync.Mutex
}
```

**TraceResult** struct:
```go
type TraceResult struct {
    Sources          []*InputSource
    TaintedVariables []*TaintedVariable
    TaintedFunctions []*TaintedFunction
    FlowGraph        *FlowGraph
    Stats            TraceStats
    Errors           []string
}
```

Functions:
- `New(config *Config) *Tracer`
- `DefaultConfig() *Config`
- `(t *Tracer) TraceDirectory(dirPath string) (*TraceResult, error)`
- `(t *Tracer) TraceFile(filePath string) (*TraceResult, error)`
- `(t *Tracer) GetTaintedFunctions(result *TraceResult) []*TaintedFunction`
- `(t *Tracer) GetFlowPaths(result *TraceResult, source *InputSource) []*PropagationPath`
- `(t *Tracer) DoesReceiveInput(result *TraceResult, funcName string) bool`
- `(t *Tracer) GetInputSources(result *TraceResult) []*InputSource`
- `(t *Tracer) GetTaintedVariables(result *TraceResult) []*TaintedVariable`

## Language Support

Analyzers implemented in `pkg/semantic/analyzer/`:
- **PHP** - Superglobals ($_GET, $_POST, $_REQUEST, $_COOKIE, $_SERVER, $_FILES, $_ENV, $_SESSION), PSR-7, Laravel, Symfony
- **JavaScript** - req.body, req.query, req.params, Express.js, Koa, Fastify
- **TypeScript** - Same as JavaScript with type awareness
- **Python** - request.args, request.form, Flask, Django, FastAPI
- **Go** - http.Request, r.URL.Query(), r.FormValue(), Gin, Echo, Chi
- **Java** - HttpServletRequest, Spring @RequestParam, @PathVariable
- **C#** - ASP.NET Request, HttpContext
- **C** - stdin, argv, getenv, fgets, scanf
- **C++** - Same as C plus std::cin, getline
- **Ruby** - params, request, Rack, Rails
- **Rust** - Actix-web, Axum, Rocket extractors

Source matchers in `pkg/sources/`:
- `php.go` - Comprehensive PHP patterns including MyBB, WordPress, Laravel, Symfony
- `javascript.go` - Express, Koa, Fastify, Next.js patterns
- `python.go` - Flask, Django, FastAPI patterns
- `go.go` - net/http, Gin, Echo, Chi, Fiber patterns
- `java.go` - Servlet, Spring MVC, JAX-RS patterns
- `c.go` - Standard input, command-line, environment
- `cpp.go` - C++ specific patterns plus C patterns
- `csharp.go` - ASP.NET Core patterns
- `ruby.go` - Rails, Sinatra, Rack patterns
- `rust.go` - Actix, Axum, Rocket patterns

## CLI Commands

```bash
# Main analysis (forward taint)
./inputtracer [options] <directory>

# Discover input carriers
./inputtracer discover -codebase /path/to/project -o carriers.json

# Classify snippets with carrier map
./inputtracer classify -carriers carriers.json -snippets snippets.json -o results.json

# Classify snippets (superglobals only)
./inputtracer classify-direct -snippets snippets.json -o results.json

# Trace a variable across all definitions
./inputtracer trace-var -var '$search_sql' -codebase /path/to/project

# Backward taint analysis
./inputtracer trace-back -target '$id' -codebase /path/to/project -o sources.json -v

# Symbolic expression tracing
./inputtracer -trace "$mybb->input['action']" ./mybb
```

## Memory Optimizations

1. **LRU File Cache** (`pkg/semantic/symbolic/filecache.go`) - Limits AST/content memory to ~100 files
2. **AST Release After Parsing** - Releases tree-sitter trees after symbol extraction
3. **Worker Pool Pattern** - Reuses parsers across files
4. **O(1) LRU Cache** (`pkg/parser/cache.go`) - Uses `container/list` for efficient eviction
5. **FlowMap Deduplication** - Map-based deduplication prevents duplicate nodes/edges (max 10K nodes, 20K edges)
6. **Regex Caching** - Caches compiled regexes to avoid recompilation
7. **Method Body Release** - Releases method body strings after pattern analysis

## Testing

```bash
go test ./...                              # Run all tests
go test ./pkg/semantic/callgraph/...       # Specific package
go test -v ./pkg/semantic/callgraph/...    # Verbose
```

Test directories:
- `testdata/` - Sample code snippets for each language
- `testapps/` - Real-world applications (mybb, chi, cobra)
- `psalm-repo/` - Psalm PHP codebase for large-scale testing
- `internal/testapps/` - Go test harness

---

*Generated: 2026-01-04*
*Files analyzed: 59 Go source files in pkg/*
*Total lines of code examined: ~15,000+*
