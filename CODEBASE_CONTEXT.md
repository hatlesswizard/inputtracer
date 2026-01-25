# InputTracer Complete Codebase Documentation
Generated: 2026-01-25 (refreshed)
Project Type: Go Library
Total Lines of Code: 40,717 (Go)
Total Files Analyzed: 95+ core library files

---

## Table of Contents
1. [Project Overview](#1-project-overview)
2. [Directory Structure](#2-complete-directory-structure)
3. [Core Types](#3-core-types)
4. [Key Interfaces](#4-key-interfaces)
5. [Source Files Documentation](#5-source-files-documentation)
6. [Constants](#6-constants)
7. [API Usage](#7-api-usage)
8. [Framework Detection](#8-framework-detection)
9. [Configuration](#9-configuration)
10. [Design Patterns](#10-design-patterns)
11. [Function Index](#11-function-index)
12. [Dependencies](#12-dependencies)
13. [Build & Test](#13-build--test)

---

## 1. Project Overview

### 1.1 Purpose
**InputTracer** is a multi-language taint analysis library that tracks how user input flows through code. It uses Tree-Sitter for parsing and supports 12 programming languages.

**CRITICAL CONSTRAINT**: This library traces INPUT SOURCES ONLY. It does NOT identify security vulnerabilities, sinks, or dangerous functions.

### 1.2 Key Capabilities
- **Source Detection**: Identifies HTTP parameters, CLI args, environment variables, file reads, database results
- **Taint Propagation**: Tracks data flow through assignments and function calls
- **Inter-procedural Analysis**: Follows data across function boundaries
- **Framework Detection**: Auto-detects PHP, JS, Python, Java, Go, Ruby, Rust, C# frameworks
- **Flow Graph Generation**: Outputs DOT/Mermaid/JSON/HTML visualizations

### 1.3 Technology Stack
- **Primary Language**: Go 1.23
- **Parser**: Tree-Sitter (via go-tree-sitter bindings)
- **Dependencies**:
  - `github.com/smacker/go-tree-sitter` - Multi-language AST parsing
  - `github.com/google/uuid` - Unique ID generation
  - `github.com/mattn/go-sqlite3` - SQLite support (optional)

### 1.4 Supported Languages
| Language | Extensions | Parser |
|----------|------------|--------|
| PHP | .php | tree-sitter-php |
| JavaScript | .js, .mjs, .cjs | tree-sitter-javascript |
| TypeScript | .ts, .tsx | tree-sitter-typescript |
| Python | .py | tree-sitter-python |
| Go | .go | tree-sitter-go |
| Java | .java | tree-sitter-java |
| C | .c, .h | tree-sitter-c |
| C++ | .cpp, .cc, .hpp | tree-sitter-cpp |
| C# | .cs | tree-sitter-c-sharp |
| Ruby | .rb | tree-sitter-ruby |
| Rust | .rs | tree-sitter-rust |

### 1.5 Entry Point Flow
```
tracer.New(config)
  -> RegisterAllLanguages()
  -> sources.RegisterAll()
  -> ast.RegisterAll()
  -> Tracer{}

tracer.TraceDirectory(path)
  -> collectFiles()
  -> parallel workers: analyzeFile()
    -> ParseFile()
    -> FindSources()
    -> ExtractAssignments() / Track propagation
    -> ExtractCalls() / Find tainted function calls
  -> mergeFileResult()
  -> runInterproceduralAnalysis()
  -> buildFlowGraph()
  -> TraceResult{}
```

---

## 2. Complete Directory Structure

```
inputtracer/
├── go.mod                           # Module: github.com/hatlesswizard/inputtracer
├── go.sum                           # Dependency checksums
├── CLAUDE.md                        # Claude Code guidance
├── CODEBASE_CONTEXT.md              # This file
│
├── pkg/
│   ├── tracer/                      # Main tracer orchestrator
│   │   ├── tracer.go               # Tracer struct, New(), TraceDirectory(), TraceFile()
│   │   ├── types.go                # Core types: TraceResult, InputSource, TaintedVariable
│   │   ├── propagation.go          # TaintPropagator, language-specific assignment extraction
│   │   ├── interprocedural.go      # InterproceduralAnalyzer, cross-function analysis
│   │   └── scope.go                # ScopeManager, ScopeType definitions
│   │
│   ├── parser/                      # Tree-Sitter parsing service
│   │   ├── service.go              # Service struct with parser pools, ParseFile()
│   │   ├── cache.go                # LRU cache with memory limits (32MB default)
│   │   └── languages/
│   │       └── init.go             # Language registration, extension mappings
│   │
│   ├── sources/                     # Input source definitions (CENTRALIZED)
│   │   ├── registry.go             # Matcher registry, RegisterAll()
│   │   ├── types.go                # Definition, Match, BaseMatcher, Matcher interface
│   │   ├── labels.go               # SourceType constants (re-exports from common)
│   │   ├── input_methods.go        # InputMethod struct, framework patterns
│   │   ├── superglobals.go         # PHP superglobal definitions
│   │   ├── mappings.go             # LanguageMappings for input functions
│   │   ├── ast_patterns.go         # ASTNodeTypes for language constructs
│   │   ├── graph_styles.go         # NodeStyle, EdgeStyle for visualization
│   │   ├── defaults.go             # DefaultSkipDirs, DefaultMaxDepth=5
│   │   ├── special_files.go        # Framework detection by file paths
│   │   │
│   │   ├── common/                  # Shared types for all languages
│   │   │   ├── types.go            # InputLabel, Definition, Match, BaseMatcher
│   │   │   ├── source_types.go     # SourceType enum (18 types)
│   │   │   ├── framework_patterns.go # FrameworkPattern, FrameworkPatternRegistry
│   │   │   └── regex_patterns.go   # Pre-compiled regex patterns (cached)
│   │   │
│   │   ├── frameworks/              # Framework detection utilities
│   │   │   └── detection.go        # DetectFramework(), FrameworkIndicator
│   │   │
│   │   ├── php/                     # PHP-specific patterns
│   │   │   ├── matcher.go          # PHP source matcher (superglobals, PSR-7)
│   │   │   ├── patterns.go         # Centralized PHP regex patterns (432 lines)
│   │   │   ├── laravel.go          # Laravel-specific patterns (generated, 541 lines)
│   │   │   └── symfony.go          # Symfony-specific patterns (generated, 288 lines)
│   │   │
│   │   ├── javascript/              # JavaScript-specific patterns
│   │   │   ├── matcher.go          # JS source matcher
│   │   │   ├── frameworks.go       # JS framework patterns
│   │   │   ├── express.go          # Express.js patterns
│   │   │   ├── koa.go              # Koa patterns
│   │   │   ├── fastify.go          # Fastify patterns
│   │   │   └── nestjs.go           # NestJS patterns
│   │   │
│   │   ├── python/                  # Python-specific patterns
│   │   │   ├── matcher.go          # Python source matcher
│   │   │   └── frameworks.go       # Django, Flask, FastAPI patterns
│   │   │
│   │   ├── golang/                  # Go-specific patterns
│   │   │   ├── matcher.go          # Go source matcher
│   │   │   └── frameworks.go       # Gin, Echo, Fiber patterns
│   │   │
│   │   ├── java/                    # Java-specific patterns
│   │   │   ├── matcher.go          # Java source matcher
│   │   │   ├── frameworks.go       # Spring, Servlet patterns
│   │   │   └── annotations.go      # Spring annotation patterns
│   │   │
│   │   ├── ruby/                    # Ruby-specific patterns
│   │   │   ├── matcher.go          # Ruby source matcher
│   │   │   └── frameworks.go       # Rails, Sinatra patterns
│   │   │
│   │   ├── rust/                    # Rust-specific patterns
│   │   │   ├── matcher.go          # Rust source matcher
│   │   │   └── frameworks.go       # Actix, Rocket, Axum patterns
│   │   │
│   │   ├── c/                       # C-specific patterns
│   │   │   ├── matcher.go          # C source matcher
│   │   │   └── input_patterns.go   # C input patterns (stdin, argv, getenv)
│   │   │
│   │   ├── cpp/                     # C++-specific patterns
│   │   │   ├── matcher.go          # C++ source matcher
│   │   │   └── frameworks.go       # Qt, POCO patterns
│   │   │
│   │   └── csharp/                  # C#-specific patterns
│   │       ├── matcher.go          # C# source matcher
│   │       └── frameworks.go       # ASP.NET patterns
│   │
│   ├── ast/                         # Language-agnostic AST extraction
│   │   ├── extractor.go            # Extractor interface, BaseExtractor
│   │   └── register.go             # RegisterAll(), language-specific node types
│   │
│   ├── output/                      # Result serialization
│   │   ├── json.go                 # JSONExporter, SummaryReport, GenerateSummary()
│   │   └── graph.go                # GraphExporter for DOT/Mermaid, PathFinder
│   │
│   └── semantic/                    # Deep semantic analysis
│       ├── tracer.go               # Full semantic Tracer (2498 lines)
│       ├── output.go               # ToJSON(), ToDOT(), ToMermaid(), ToHTML()
│       │
│       ├── types/
│       │   └── types.go            # FlowNode, FlowEdge, FlowMap, SymbolTable (1100 lines)
│       │
│       ├── analyzer/
│       │   ├── interface.go        # LanguageAnalyzer interface, Registry
│       │   ├── php/analyzer.go
│       │   ├── javascript/analyzer.go
│       │   ├── typescript/analyzer.go
│       │   ├── python/analyzer.go
│       │   ├── golang/analyzer.go
│       │   ├── java/analyzer.go
│       │   ├── ruby/analyzer.go
│       │   ├── rust/analyzer.go
│       │   ├── c/analyzer.go
│       │   ├── cpp/analyzer.go
│       │   └── csharp/analyzer.go
│       │
│       ├── discovery/               # Source discovery
│       │   ├── taint.go
│       │   ├── carrier_map.go      # Carrier map for input objects
│       │   └── superglobal.go
│       │
│       ├── classifier/
│       │   └── classifier.go       # Input source classifier
│       │
│       ├── extractor/
│       │   └── extractor.go        # Expression extractor
│       │
│       ├── condition/
│       │   └── extractor.go        # Condition extractor
│       │
│       ├── batch/
│       │   └── analyzer.go         # Batch analyzer
│       │
│       ├── pathanalysis/
│       │   └── expander.go         # Path expansion
│       │
│       ├── index/
│       │   ├── indexer.go          # Code indexer
│       │   └── indexer_test.go
│       │
│       ├── symbolic/
│       │   ├── executor.go         # Symbolic execution
│       │   └── filecache.go        # File cache for symbolic
│       │
│       ├── callgraph/
│       │   └── manager.go          # Call graph manager
│       │
│       └── tracer/
│           └── vartracer.go        # Variable tracer
│
└── testapps/                        # Test applications (various languages)
    ├── php/                        # PHP test apps (DVWA, phpMyAdmin, etc.)
    ├── javascript/                 # JS test apps (Express, Fastify, etc.)
    ├── typescript/                 # TS test apps
    ├── python/                     # Python test apps (Django, Flask, etc.)
    ├── go/                         # Go test apps (Gin, Echo, Fiber, etc.)
    ├── java/                       # Java test apps
    ├── c/                          # C test apps (Redis, OpenSSL, etc.)
    ├── cpp/                        # C++ test apps
    ├── csharp/                     # C# test apps
    ├── ruby/                       # Ruby test apps
    ├── rust/                       # Rust test apps
    └── mybb/                       # MyBB forum software
```

---

## 3. Core Types

### 3.1 tracer.Config (pkg/tracer/tracer.go:19)
```go
type Config struct {
    Languages       []string            // Languages to analyze (empty = all)
    MaxDepth        int                 // Inter-procedural analysis depth (default: 5)
    Workers         int                 // Parallel workers (default: NumCPU)
    CustomSources   []sources.Definition // Custom source definitions
    SkipDirs        []string            // Directories to skip
    IncludePatterns []string            // File patterns to include
}
```

### 3.2 tracer.InputSource (pkg/tracer/types.go:36)
```go
type InputSource struct {
    ID       string       // Unique identifier
    Type     string       // e.g., "$_GET", "req.body", "argv"
    Key      string       // e.g., "username" in $_GET['username']
    Location Location     // Code location
    Labels   []InputLabel // Categories
    Language string       // Source language
}
```

### 3.3 tracer.TaintedVariable (pkg/tracer/types.go:46)
```go
type TaintedVariable struct {
    ID       string       // Unique identifier
    Name     string       // Variable name
    Scope    string       // Function/class scope
    Source   *InputSource // Original input source
    Location Location     // Code location
    Depth    int          // How many assignments from original source
    Language string       // Language
}
```

### 3.4 tracer.TaintedFunction (pkg/tracer/types.go:65)
```go
type TaintedFunction struct {
    ID              string            // Unique identifier
    Name            string            // Function name
    FilePath        string            // File path
    Line            int               // Line number
    Language        string            // Language
    TaintedParams   []TaintedParam    // Parameters receiving input
    ReceivesThrough []PropagationPath // Propagation paths
}
```

### 3.5 tracer.TraceResult (pkg/tracer/types.go:138)
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

### 3.6 common.SourceType (pkg/sources/common/source_types.go)
```go
const (
    SourceHTTPGet     SourceType = "http_get"
    SourceHTTPPost    SourceType = "http_post"
    SourceHTTPBody    SourceType = "http_body"
    SourceHTTPJSON    SourceType = "http_json"
    SourceHTTPHeader  SourceType = "http_header"
    SourceHTTPCookie  SourceType = "http_cookie"
    SourceHTTPPath    SourceType = "http_path"
    SourceHTTPFile    SourceType = "http_file"
    SourceHTTPRequest SourceType = "http_request"
    SourceSession     SourceType = "session"
    SourceCLIArg      SourceType = "cli_arg"
    SourceEnvVar      SourceType = "env_var"
    SourceStdin       SourceType = "stdin"
    SourceFile        SourceType = "file"
    SourceDatabase    SourceType = "database"
    SourceNetwork     SourceType = "network"
    SourceUserInput   SourceType = "user_input"
    SourceUnknown     SourceType = "unknown"
)
```

### 3.7 common.Definition (pkg/sources/common/types.go:28)
```go
type Definition struct {
    Name         string       // e.g., "$_GET", "req.body"
    Pattern      string       // Regex pattern to match
    Language     string       // Target language
    Labels       []InputLabel // Categories
    Description  string       // Human-readable description
    NodeTypes    []string     // Tree-sitter node types to match
    KeyExtractor string       // Regex to extract key
}
```

### 3.8 common.Match (pkg/sources/common/types.go:39)
```go
type Match struct {
    SourceType string       // e.g., "$_GET", "req.body"
    Key        string       // e.g., "username" in $_GET['username']
    Variable   string       // Variable name if assigned
    Line       int
    Column     int
    EndLine    int
    EndColumn  int
    Snippet    string
    Labels     []InputLabel
}
```

---

## 4. Key Interfaces

### 4.1 Matcher (pkg/sources/common/types.go:52)
```go
type Matcher interface {
    Language() string
    FindSources(root *sitter.Node, src []byte) []Match
}
```
**Implemented By:** PHPMatcher, JSMatcher, PythonMatcher, GoMatcher, JavaMatcher, CMatcher, CPPMatcher, CSharpMatcher, RubyMatcher, RustMatcher

### 4.2 LanguageAnalyzer (pkg/semantic/analyzer/interface.go:12)
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

### 4.3 Extractor (pkg/ast/extractor.go)
```go
type Extractor interface {
    Language() string
    ExtractAssignments(root *sitter.Node, src []byte) []Assignment
    ExtractCalls(root *sitter.Node, src []byte) []FunctionCall
    ExpressionContains(node *sitter.Node, varName string, src []byte) bool
}
```

---

## 5. Source Files Documentation

### 5.1 pkg/tracer/tracer.go (625 lines)
**Purpose:** Main tracer entry point and orchestration

| Function | Signature | Description |
|----------|-----------|-------------|
| DefaultConfig | `func DefaultConfig() *Config` | Returns sensible default config |
| New | `func New(config *Config) *Tracer` | Creates new Tracer with config |
| TraceDirectory | `func (t *Tracer) TraceDirectory(dirPath string) (*TraceResult, error)` | Analyzes entire directory |
| TraceFile | `func (t *Tracer) TraceFile(filePath string) (*TraceResult, error)` | Analyzes single file |
| analyzeFile | `func (t *Tracer) analyzeFile(filePath string) *fileResult` | Per-file analysis |
| mergeFileResult | `func (t *Tracer) mergeFileResult(result *TraceResult, fr *fileResult)` | Merges file results |
| buildFlowGraph | `func (t *Tracer) buildFlowGraph(result *TraceResult)` | Builds flow graph |
| collectFiles | `func (t *Tracer) collectFiles(dirPath string) ([]string, error)` | Collects files to analyze |
| GetTaintedFunctions | `func (t *Tracer) GetTaintedFunctions(result *TraceResult) []*TaintedFunction` | Returns tainted functions |
| GetFlowPaths | `func (t *Tracer) GetFlowPaths(result *TraceResult, source *InputSource) []*PropagationPath` | Returns propagation paths |
| DoesReceiveInput | `func (t *Tracer) DoesReceiveInput(result *TraceResult, funcName string) bool` | Checks if function receives input |

### 5.2 pkg/parser/service.go (276 lines)
**Purpose:** Multi-language parsing service with pooling

| Function | Signature | Description |
|----------|-----------|-------------|
| NewService | `func NewService(cacheSize ...int) *Service` | Creates new parser service |
| RegisterLanguage | `func (s *Service) RegisterLanguage(name string, lang *sitter.Language)` | Registers language parser |
| ParseFile | `func (s *Service) ParseFile(filePath string) (*ParseResult, error)` | Parses file with caching |
| Parse | `func (s *Service) Parse(source []byte, language string) (*sitter.Node, error)` | Parses source code |
| DetectLanguage | `func (s *Service) DetectLanguage(filePath string) string` | Detects language from extension |

### 5.3 pkg/parser/cache.go (201 lines)
**Purpose:** LRU cache with memory limits (32MB default)

| Function | Signature | Description |
|----------|-----------|-------------|
| NewCache | `func NewCache(maxEntries int) *Cache` | Creates cache with 32MB default |
| Get | `func (c *Cache) Get(key string) *CachedParse` | Gets cached result O(1) |
| Put | `func (c *Cache) Put(key string, data *CachedParse)` | Adds to cache O(1) |
| Clear | `func (c *Cache) Clear()` | Clears all entries (calls tree.Close()) |
| MemoryUsage | `func (c *Cache) MemoryUsage() int64` | Returns memory estimate |

### 5.4 pkg/sources/input_methods.go (190 lines)
**Purpose:** Framework input method definitions

| Type | Description |
|------|-------------|
| InputMethodCategory | Classifies input methods (http, file, command, generic) |
| InputMethod | Describes a method that returns user input |
| InputMethods | Canonical list of input-returning methods |

| Function | Signature | Description |
|----------|-----------|-------------|
| IsInputMethod | `func IsInputMethod(varName, methodName string) bool` | Checks if var.method is known input |
| GetInputMethodInfo | `func GetInputMethodInfo(varName, methodName string) *InputMethod` | Returns full info for input method |
| IsInterestingMethod | `func IsInterestingMethod(methodName string) bool` | Checks if method is security-relevant |
| GetMethodsByCategory | `func GetMethodsByCategory(category InputMethodCategory) []InputMethod` | Returns methods by category |
| GetMethodsByFramework | `func GetMethodsByFramework(framework string) []InputMethod` | Returns methods by framework |

### 5.5 pkg/sources/frameworks/detection.go (385 lines)
**Purpose:** Framework detection utilities

| Function | Signature | Description |
|----------|-----------|-------------|
| DetectFramework | `func DetectFramework(codebasePath string) string` | Detects framework by file indicators |
| DetectFrameworkByLanguage | `func DetectFrameworkByLanguage(codebasePath string, language string) string` | Detects framework for specific language |
| GetFrameworkIndicators | `func GetFrameworkIndicators(framework string) []string` | Returns indicators for framework |
| GetFrameworksForLanguage | `func GetFrameworksForLanguage(language string) []string` | Returns known frameworks for language |

---

## 6. Constants

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

### 6.4 Defaults (pkg/sources/defaults.go)
```go
var DefaultSkipDirs = []string{
    ".git", ".svn", ".hg",
    "node_modules", "vendor", "bower_components",
    ".idea", ".vscode",
    "__pycache__", ".pytest_cache",
    "build", "dist", "target",
}

const (
    DefaultMaxDepth      = 5
    DefaultCacheSize     = 1000
    DefaultSnippetLength = 100
)
```

---

## 7. API Usage

### 7.1 Basic Usage
```go
import "github.com/hatlesswizard/inputtracer/pkg/tracer"

// Create tracer with defaults
t := tracer.New(nil)

// Or with custom config
config := &tracer.Config{
    Languages: []string{"php", "javascript"},
    MaxDepth:  5,
    Workers:   4,
    SkipDirs:  []string{".git", "vendor", "node_modules"},
}
t := tracer.New(config)

// Trace a directory
result, err := t.TraceDirectory("/path/to/project")
if err != nil {
    log.Fatal(err)
}

// Access results
for _, source := range result.Sources {
    fmt.Printf("Input: %s at %s:%d\n", source.Type, source.Location.FilePath, source.Location.Line)
}

for _, tv := range result.TaintedVariables {
    fmt.Printf("Tainted: %s (from %s)\n", tv.Name, tv.Source.Type)
}

for _, tf := range result.TaintedFunctions {
    fmt.Printf("Function %s receives input\n", tf.Name)
}

// Export to JSON
json, _ := result.ToJSON()
```

### 7.2 Semantic Analysis
```go
import "github.com/hatlesswizard/inputtracer/pkg/semantic"

// Create semantic tracer
st := semantic.New(&semantic.Config{
    Languages:     []string{"php"},
    MaxDepth:      10,
    Workers:       runtime.NumCPU(),
    CacheSize:     1000,
    SnippetLength: 100,
})

// Trace backward from an expression
result, err := st.TraceBackward("/path/to/file.php", 42, 10, "$data")

// Output formats
dot := result.ToDOT()
mermaid := result.ToMermaid()
html := result.ToHTML()
```

---

## 8. Framework Detection

### 8.1 PHP Frameworks
| Framework | Indicators |
|-----------|------------|
| Laravel | `artisan`, `bootstrap/app.php` |
| Symfony | `symfony.lock`, `config/bundles.php`, `src/Kernel.php` |

### 8.2 JavaScript Frameworks
| Framework | Indicators |
|-----------|------------|
| Express | `node_modules/express` |
| Next.js | `next.config.js`, `next.config.mjs` |
| NestJS | `nest-cli.json`, `src/main.ts` |
| Koa | `node_modules/koa` |
| Fastify | `node_modules/fastify` |

### 8.3 Python Frameworks
| Framework | Indicators |
|-----------|------------|
| Django | `manage.py`, `settings.py`, `urls.py` |
| Flask | `app.py`, `wsgi.py` |
| FastAPI | `main.py` |

### 8.4 Other Frameworks
| Language | Frameworks |
|----------|------------|
| Java | Spring, Spring Boot |
| Go | Gin, Echo |
| C# | ASP.NET Core, ASP.NET MVC |
| Ruby | Rails, Sinatra, Hanami, Padrino |
| Rust | Actix-web, Rocket, Axum |
| C++ | Qt, POCO |

---

## 9. Configuration

### 9.1 Tracer Config
```go
config := &tracer.Config{
    Languages:       []string{},        // Empty = all supported
    MaxDepth:        5,                  // Inter-procedural analysis depth
    Workers:         runtime.NumCPU(),  // Parallel workers
    CustomSources:   []sources.Definition{},
    SkipDirs:        []string{".git", "node_modules", "vendor"},
    IncludePatterns: []string{},
}
```

### 9.2 Semantic Config
```go
config := &semantic.Config{
    Languages:     []string{},
    MaxDepth:      10,
    Workers:       runtime.NumCPU(),
    CacheSize:     1000,
    SnippetLength: 100,
}
```

---

## 10. Design Patterns

### 10.1 Registry Pattern
Language-specific implementations are registered at `init()`:
- `sources.RegisterAll(registry)`
- `ast.RegisterAll(registry)`
- `languages.RegisterAllLanguages(service)`

### 10.2 Worker Pool
Parallel file analysis via goroutines:
```go
for i := 0; i < t.config.Workers; i++ {
    go func() {
        for filePath := range fileChan {
            resultChan <- t.analyzeFile(filePath)
        }
    }()
}
```

### 10.3 Parser Pooling
`sync.Pool` reuses expensive Tree-Sitter parser instances per language.

### 10.4 LRU Cache
Parser service uses an LRU cache (32MB default) with proper `tree.Close()` on eviction.

### 10.5 Map Pre-allocation
Analysis state uses pre-sized maps (128-256 capacity) for O(1) deduplication.

---

## 11. Function Index

### Main Entry Points
| Function | Package | Description |
|----------|---------|-------------|
| `New` | tracer | Creates new Tracer with config |
| `TraceDirectory` | tracer | Analyzes entire directory |
| `TraceFile` | tracer | Analyzes single file |
| `NewService` | parser | Creates parser service |
| `ParseFile` | parser | Parses file with caching |
| `DetectLanguage` | parser | Detects language from extension |
| `NewRegistry` | sources | Creates source registry |
| `RegisterAll` | sources | Registers all language matchers |
| `FindSources` | sources | Finds sources in AST |
| `DetectFramework` | frameworks | Detects framework by indicators |

### Analysis Functions
| Function | Package | Description |
|----------|---------|-------------|
| `ExtractAssignments` | ast | Extracts all assignments from AST |
| `ExtractCalls` | ast | Extracts all function calls from AST |
| `ExpressionContains` | ast | Checks if expression contains variable |
| `BuildSymbolTable` | semantic/analyzer | Builds symbol table for file |
| `TraceExpression` | semantic/analyzer | Traces expression to sources |

### Output Functions
| Function | Package | Description |
|----------|---------|-------------|
| `ToJSON` | tracer | Exports TraceResult to JSON |
| `ExportDOT` | output | Exports flow graph to Graphviz DOT |
| `ExportMermaid` | output | Exports flow graph to Mermaid |
| `GenerateSummary` | output | Generates summary report |

---

## 12. Dependencies

```
github.com/google/uuid v1.6.0
github.com/smacker/go-tree-sitter v0.0.0-20240827094217-dd81d9e9be82
github.com/mattn/go-sqlite3 v1.14.24
```

---

## 13. Build & Test

```bash
go build ./...           # Build all packages
go test ./...            # Run all tests
go test ./pkg/tracer     # Test specific package
go test -v ./...         # Verbose test output
go test -race ./...      # Run with race detector
```

---

## Adding New Framework Support

Create framework patterns in the language-specific subdirectory:
- `pkg/sources/php/wordpress.go`
- `pkg/sources/php/laravel.go`
- `pkg/sources/javascript/express.go`
- `pkg/sources/python/django.go`
- etc.

**NEVER** add framework-specific code to core library packages.

---

## File Count Verification

| Directory | Files |
|-----------|-------|
| pkg/tracer/ | 5 |
| pkg/parser/ | 3 |
| pkg/sources/ (root) | 10 |
| pkg/sources/common/ | 4 |
| pkg/sources/constants/ | 3 |
| pkg/sources/php/ | 4 (matcher.go, patterns.go, laravel.go, symfony.go) |
| pkg/sources/javascript/ | 2 |
| pkg/sources/python/ | 2 |
| pkg/sources/golang/ | 2 |
| pkg/sources/java/ | 2 |
| pkg/sources/ruby/ | 2 |
| pkg/sources/rust/ | 2 |
| pkg/sources/c/ | 2 |
| pkg/sources/cpp/ | 2 |
| pkg/sources/csharp/ | 2 |
| pkg/ast/ | 2 |
| pkg/output/ | 2 |
| pkg/semantic/ | 25+ |
| **Total** | ~75+ |

---

---

## 14. Semantic Types Reference

### 14.1 FlowNode & FlowEdge Types (`pkg/semantic/types/types.go`)

```go
// FlowNodeType - Type of node in data flow graph
const (
    NodeSource   FlowNodeType = "source"    // Input source
    NodeCarrier  FlowNodeType = "carrier"   // Object carrying input
    NodeVariable FlowNodeType = "variable"  // Variable
    NodeFunction FlowNodeType = "function"  // Function
    NodeProperty FlowNodeType = "property"  // Object property
    NodeParam    FlowNodeType = "param"     // Function parameter
    NodeReturn   FlowNodeType = "return"    // Return value
)

// FlowEdgeType - How data flows between nodes
const (
    EdgeAssignment  FlowEdgeType = "assignment"   // Direct assignment
    EdgeParameter   FlowEdgeType = "parameter"    // Param passing
    EdgeReturn      FlowEdgeType = "return"       // Return value
    EdgeProperty    FlowEdgeType = "property"     // Property access
    EdgeArraySet    FlowEdgeType = "array_set"    // Array element set
    EdgeArrayGet    FlowEdgeType = "array_get"    // Array element get
    EdgeMethodCall  FlowEdgeType = "method_call"  // Method invocation
    EdgeConstructor FlowEdgeType = "constructor"  // Object construction
    EdgeFramework   FlowEdgeType = "framework"    // Framework-specific
    EdgeConcatenate FlowEdgeType = "concatenate"  // String concat
    EdgeDestructure FlowEdgeType = "destructure"  // Destructuring
    EdgeIteration   FlowEdgeType = "iteration"    // Loop iteration
    EdgeConditional FlowEdgeType = "conditional"  // Conditional branch
    EdgeCall        FlowEdgeType = "call"         // Function call
    EdgeDataFlow    FlowEdgeType = "data_flow"    // Generic data flow
)
```

### 14.2 TaintChain (`pkg/semantic/types/types.go:665`)

```go
// TaintChain tracks complete propagation path of tainted data
type TaintChain struct {
    OriginalSource   string     // e.g., "$_GET['id']"
    OriginalType     SourceType // e.g., "http_get"
    OriginalFile     string
    OriginalLine     int
    Steps            []TaintStep
    CurrentExpression string    // What the taint looks like now
    Depth            int        // How many hops from source
}

// TaintStep represents one step in the taint propagation chain
type TaintStep struct {
    StepType    string // "assignment", "parameter", "return", "property", "method_call"
    Expression  string // The code at this step
    FilePath    string
    Line        int
    Description string // Human-readable description
}

// Usage:
chain := NewTaintChain("$_GET['id']", "http_get", "/path/to/file.php", 10)
chain.AddStep("assignment", "$userId = $_GET['id']", file, 11, "Assigned to $userId")
chain.AddStep("parameter", "processUser($userId)", file, 20, "Passed to processUser()")
cloned := chain.Clone() // For branching flows
```

### 14.3 FlowMap with O(1) Deduplication

```go
// FlowMap - Complete flow analysis result with O(1) dedup
type FlowMap struct {
    Sources   []FlowNode
    Paths     []FlowPath
    Carriers  []FlowNode
    AllNodes  []FlowNode
    AllEdges  []FlowEdge
    Usages    []FlowNode
    CallGraph map[string][]string

    // Internal O(1) deduplication
    nodeIndex map[string]bool // nodeID -> exists
    edgeIndex map[string]bool // edgeKey -> exists

    // Configurable limits
    maxNodes int // Default: 10000
    maxEdges int // Default: 20000
}

// Memory-safe operations
fm := NewFlowMap() // Uses default limits
fm := NewFlowMapWithLimits(5000, 10000) // Custom limits

added := fm.AddNode(node)  // Returns false if duplicate or at capacity
added := fm.AddEdge(edge)  // Returns false if duplicate or at capacity
exists := fm.HasNode(id)   // O(1) lookup
exists := fm.HasEdge(from, to, edgeType) // O(1) lookup
```

### 14.4 Symbol Table (`pkg/semantic/types/types.go:373`)

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

// Memory optimization - call after analysis
st.ReleaseBodySources() // Frees method body strings
```

---

## 15. Symbolic Execution Patterns

### 15.1 Expression Patterns (`pkg/sources/patterns/symbolic_patterns.go`)

```go
// SuperglobalAccessPattern - $_GET['key'], $_POST["name"]
SuperglobalAccessPattern = regexp.MustCompile(`^\$_(GET|POST|COOKIE|REQUEST|SERVER|FILES|ENV|SESSION)\[['"]?(\w+)['"]?\]$`)

// PropertyAccessPattern - $var->property['key']
PropertyAccessPattern = regexp.MustCompile(`^\$(\w+)->(\w+)(?:\[['"]?(\w+)['"]?\])?$`)

// ChainPropertyWithKeyPattern - ->property['key']
ChainPropertyWithKeyPattern = regexp.MustCompile(`^->(\w+)\[['"]?(\w+)['"]?\]`)

// FunctionCallPattern - functionName(args)
FunctionCallPattern = regexp.MustCompile(`^(\w+)\(([^)]*)\)$`)

// Dynamic pattern builders
BuildVariableAssignPattern(varName) // $varname = something;
BuildPropertyExternalAssignPattern(varName, propName) // $var->prop = something;
BuildPropertyArrayExternalAssignPattern(varName, propName) // $var->prop['key'] = something;
```

### 15.2 PHP Taint Patterns (`pkg/sources/php/taint_patterns.go`)

```go
TaintPatterns.ThisArrayPattern      // $this->prop[$key] = ...
TaintPatterns.DynamicPropPattern    // $this->$key = $val
TaintPatterns.ReturnThisPattern     // return $this->prop
TaintPatterns.SuperglobalKeyPattern // $_GET['key']
TaintPatterns.LoopVariablePattern   // foreach($x as $key => $value)
TaintPatterns.ForeachValueOnlyPattern // foreach($x as $value)
```

---

## 16. Memory Optimization Strategies

### 16.1 LRU Cache with Proper Cleanup (`pkg/parser/cache.go`)
- Memory limit: 32MB default
- O(1) get/put operations
- **CRITICAL**: Calls `tree.Close()` on eviction to free Tree-sitter memory

### 16.2 Symbolic File Cache (`pkg/semantic/symbolic/filecache.go`)
- Memory limit: 64MB default
- Lazy loading of file content and AST
- Automatic tree cleanup on eviction

### 16.3 Assignment Caching (`pkg/semantic/tracer.go`)
- Caches extracted assignments (tiny) instead of full ASTs (5-10x source size)
- Reduces memory footprint dramatically

### 16.4 FlowMap Limits
```go
DefaultMaxFlowNodes = 10000 // Prevents unbounded node growth
DefaultMaxFlowEdges = 20000 // Prevents unbounded edge growth
```

### 16.5 Body Source Release
```go
// After analysis, free large strings
symbolTable.ReleaseBodySources()
classDef.ReleaseBodySources()
```

---

## 17. Scope Management (`pkg/tracer/scope.go`)

```go
type ScopeManager struct {
    scopes    []*Scope
    current   *Scope
    variables map[string][]*ScopedVariable // name -> definitions
}

type ScopedVariable struct {
    Name      string
    Scope     *Scope
    Tainted   bool
    Source    *InputSource
    Depth     int
    Location  Location
    Shadowing *ScopedVariable // Previous definition being shadowed
}

// Key operations
sm := NewScopeManager()
scope := sm.EnterScope(ScopeFunction, "myFunc", loc) // Enter new scope
sm.ExitScope() // Return to parent scope

sv := sm.DeclareVariable("$var", true, source, depth, loc) // Declare tainted var
sv := sm.LookupVariable("$var") // Look up in scope chain
isTainted := sm.IsTainted("$var") // Check if tainted in current scope

sm.MarkTainted("$var", source, depth) // Mark existing var as tainted
tainted := sm.GetAllTaintedInScope() // Get all visible tainted vars
chain := sm.GetScopeChain() // Get scope chain (current to global)
name := sm.GetScopeQualifiedName() // e.g., "MyClass.myMethod"

sm.Reset() // Reset to initial state
clone := sm.Clone() // Clone for parallel analysis
```

---

## 18. Call Graph Management (`pkg/semantic/callgraph/manager.go`)

```go
type Manager struct {
    nodes     map[string]*CallNode
    edges     map[string]map[string]*CallEdge
    backEdges map[string]map[string]*CallEdge // For backward traversal
}

// Key operations
mgr := NewManager()
mgr.AddNode("myFunc", &CallNodeInfo{...})
mgr.AddEdge("caller", "callee", &CallEdgeInfo{...})

// Distance computation (BFS from entry points)
distances := mgr.ComputeDistanceFromEntryPoints()

// Path finding (with caching)
path := mgr.GetShortestPath("fromFunc", "toFunc")

// Bidirectional lookup
callers := mgr.GetCallers("funcName")  // Who calls this?
callees := mgr.GetCallees("funcName")  // What does this call?
```

---

## 19. Condition Extraction (`pkg/semantic/condition/extractor.go`)

```go
type KeyCondition struct {
    Type      string // "isset", "empty", "array_key_exists", etc.
    ArrayExpr string // The array expression
    KeyExpr   string // The key expression
}

type ConditionPath struct {
    Conditions []KeyCondition
    Reachable  bool
}

// Extract conditions from code
conditions := ExtractFromCode(source)

// Classify condition type
condType := classifyCondition(node, source)
```

---

## 20. Path Analysis (`pkg/semantic/pathanalysis/expander.go`)

```go
type PathExpander struct {
    callGraph   *callgraph.Manager
    symbolTable map[string]*types.SymbolTable
    maxDepth    int
}

type PathNode struct {
    FunctionName string
    FilePath     string
    Line         int
    CallSite     *types.CallSite
}

type ExecutionPath struct {
    Nodes []PathNode
    Depth int
}

// Expand paths between functions
expander := NewPathExpander(callGraph, symbolTables, maxDepth)
paths := expander.ExpandPaths(startFunc, endFunc)
```

---

## 21. Code Indexer (`pkg/semantic/index/indexer.go`)

```go
type Indexer struct {
    symbolsByName     map[string][]*Symbol
    symbolsByFile     map[string][]*Symbol
    symbolsByFunction map[string][]*Symbol
    symbolsByClass    map[string][]*Symbol
    symbolsByMethod   map[string][]*Symbol
}

// O(1) symbol lookups
indexer := NewIndexer()
indexer.AddSymbol(symbol)

symbols := indexer.FindByName("functionName")
symbols := indexer.FindByFile("/path/to/file.php")
symbols := indexer.FindByClass("ClassName")
symbols := indexer.FindByMethod("ClassName.methodName")
```

---

*Last updated: 2026-01-25 (refreshed via /use-context)*

---

## 22. PHP Centralized Patterns Reference

### 22.1 Universal Input Detection Patterns (`pkg/sources/php/patterns.go`)

```go
// InputMethodPattern - Methods that ALWAYS indicate user input
// Matches: input, getInput, getPost, getQuery, getCookie, getHeader, etc.
InputMethodPattern = regexp.MustCompile(`(?i)^(get_?)?(input|var|variable|query_?params?|parsed_?body|cookie_?params?|server_?params?|uploaded_?files?|headers?|all)$|^(get_?)?(post|cookie|param)s?$`)

// InputPropertyPattern - Properties that hold user input
// Matches: input, request, params, query, cookies, headers, body, data, args, etc.
InputPropertyPattern = regexp.MustCompile(`(?i)^(input|request|params?|query|cookies?|headers?|body|data|args?|post|get|files?|server|attributes?|payload)s?$`)

// InputObjectPattern - Objects that carry user input
InputObjectPattern = regexp.MustCompile(`(?i)(request|input|req|params?|http|ctx|context|getRequest\(\)|getApplication\(\))`)

// ExcludeMethodPattern - Methods to EXCLUDE (false positive prevention)
ExcludeMethodPattern = regexp.MustCompile(`(?i)^(getData|getBody|getContent|fetch|find|load|read)$`)
```

### 22.2 SQL Embedded Expression Patterns

```go
// SQLCurlyBracePattern - {$var->prop['key']} in SQL strings
SQLCurlyBracePattern = regexp.MustCompile(`\{\s*\$(\w+)->(\w+)\s*\[\s*['""]([^'""]+)['"]\s*\]\s*\}`)

// SQLSimpleCurlyPattern - {$var->prop} without array access
SQLSimpleCurlyPattern = regexp.MustCompile(`\{\s*\$(\w+)->(\w+)\s*\}`)

// SQLNoCurlyPattern - "$var->prop['key']" direct interpolation
SQLNoCurlyPattern = regexp.MustCompile(`"\s*[^"]*\$(\w+)->(\w+)\s*\[\s*['""]([^'""]+)['"]\s*\]`)
```

### 22.3 Helper Functions

```go
// Detection helpers
IsInputMethod(methodName string) bool
IsInputProperty(propName string) bool
IsInputObject(objName string) bool
IsExcludedMethod(methodName string) bool
IsContextDependentMethod(methodName string) bool
MatchesInputCarrier(objName, propOrMethodName string, isMethod bool) bool

// Extraction helpers
ExtractSQLEmbeddedExpressions(line string) []SQLEmbeddedMatch
ExtractConcatenatedExpressions(line string) []ConcatMatch
ExtractEscapedExpressions(line string) []EscapeMatch
ContainsSuperglobal(text string) (bool, string)

// Dynamic pattern builders
BuildPropertyAssignLoopPattern(propertyName, keyVar, valVar string) *regexp.Regexp
BuildDirectAssignPattern(propertyName string) *regexp.Regexp
BuildThisPropertyAssignPattern(paramName string) *regexp.Regexp
BuildReturnPropertyPattern(propertyName string) *regexp.Regexp
BuildMethodCallPattern(methodName string) *regexp.Regexp
GetTypeHintPatterns(varName string) []*regexp.Regexp
```

---

## 23. Constants Reference (`pkg/sources/constants/`)

### 23.1 Input Labels (`input_labels.go`)
| Constant | Value | Description |
|----------|-------|-------------|
| `SourceHTTPGet` | `HTTP_GET` | GET query parameters |
| `SourceHTTPPost` | `HTTP_POST` | POST form data |
| `SourceHTTPCookie` | `HTTP_COOKIE` | Cookie values |
| `SourceHTTPHeader` | `HTTP_HEADER` | HTTP headers |
| `SourceHTTPBody` | `HTTP_BODY` | Raw request body |
| `SourceHTTPFile` | `HTTP_FILE` | Uploaded files |
| `SourceHTTPPath` | `HTTP_PATH` | URL path parameters |
| `SourceCLIArg` | `CLI_ARG` | Command line arguments |
| `SourceEnvVar` | `ENV_VAR` | Environment variables |
| `SourceFileRead` | `FILE_READ` | File system reads |
| `SourceDatabase` | `DATABASE` | Database queries |
| `SourceNetwork` | `NETWORK` | Network socket data |
| `SourceSession` | `SESSION` | Session data |
| `SourceUserInput` | `USER_INPUT` | Generic user input |

### 23.2 Propagation Types (`propagation_types.go`)
| Constant | Value | Description |
|----------|-------|-------------|
| `PropagationDirect` | `direct` | Direct variable assignment |
| `PropagationFunction` | `function` | Through function call |
| `PropagationReturn` | `return` | Function return value |
| `PropagationParameter` | `parameter` | Function parameter |
| `PropagationProperty` | `property` | Object property access |
| `PropagationArray` | `array` | Array element access |

### 23.3 Scope Types (`scope_types.go`)
| Constant | Value | Description |
|----------|-------|-------------|
| `ScopeGlobal` | `global` | Global scope |
| `ScopeFunction` | `function` | Function scope |
| `ScopeMethod` | `method` | Method scope |
| `ScopeClass` | `class` | Class scope |
| `ScopeBlock` | `block` | Block scope |
| `ScopeClosure` | `closure` | Closure scope |

---

## 24. AST Node Types Reference (`pkg/sources/ast_patterns.go`)

### 24.1 Universal Node Types (All Languages)
```go
UniversalASTNodeTypes = ASTNodeTypes{
    FunctionTypes: []string{
        "function_definition", "function_declaration", "method_definition",
        "method_declaration", "function_item", "arrow_function",
        "function_expression", "lambda", "def", "fn_item",
    },
    ScopeTypes: []string{
        "function_definition", "function_declaration", "method_definition",
        "method_declaration", "class_definition", "class_declaration",
        "module", "program", "source_file",
    },
    AssignmentTypes: []string{
        "assignment_expression", "assignment_statement", "augmented_assignment",
        "variable_declarator", "short_var_declaration",
    },
    CallTypes: []string{
        "call_expression", "function_call_expression", "member_call_expression",
        "method_invocation",
    },
    IdentifierTypes: []string{
        "identifier", "variable_name", "name", "property_identifier",
        "attribute", "constant",
    },
}
```

### 24.2 Language-Specific Node Types
Supported languages with custom AST mappings:
- PHP, JavaScript, TypeScript, TSX, Python, Go, Java, C, C++, C#, Ruby, Rust

### 24.3 Helper Functions
```go
IsFunctionNode(nodeType string) bool
IsFunctionNodeForLanguage(nodeType, language string) bool
IsScopeNode(nodeType string) bool
IsScopeNodeForLanguage(nodeType, language string) bool
IsAssignmentNode(nodeType string) bool
IsAssignmentNodeForLanguage(nodeType, language string) bool
IsCallNode(nodeType string) bool
IsCallNodeForLanguage(nodeType, language string) bool
IsIdentifierNode(nodeType string) bool
IsIdentifierNodeForLanguage(nodeType, language string) bool
GetASTNodeTypesForLanguage(language string) ASTNodeTypes
```
