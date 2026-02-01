# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Critical: Library Purpose & Constraints

**This library traces INPUT SOURCES ONLY. It does NOT identify security vulnerabilities.**

When writing code for this library:
- **NEVER** create, add, or use sink patterns to identify security issues
- **NEVER** create product/framework-specific code inside the core library packages
- **NEVER** add security vulnerability detection, sink matching, or "dangerous function" lists
- **ALWAYS** create new framework/product patterns inside `pkg/sources/{language}/` directory - nowhere else
- If asked to create cases for a specific framework/product, create it in the language-specific subdirectory:
  - WordPress (PHP) → `pkg/sources/php/wordpress.go`
  - Laravel (PHP) → `pkg/sources/php/laravel.go`
  - Django (Python) → `pkg/sources/python/django.go`
  - Express (JS) → `pkg/sources/javascript/express.go`
  - Spring (Java) → `pkg/sources/java/spring.go`

The library's sole purpose is to trace where user input enters code and how it propagates through variables and function calls. Security analysis (identifying what happens to that input) is intentionally out of scope.

## Build & Test Commands

```bash
# Build
make build                          # Build cmd/ and pkg/
go build ./...                      # Build everything

# Test
make test                           # Run pkg/ tests
go test ./...                       # Run all tests
go test ./pkg/tracer                # Test specific package
go test -run TestMatchesVariable ./pkg/tracer  # Run single test
go test -v ./pkg/sources/...        # Verbose, sources only
go test -race ./...                 # Race detector
go test -bench=. ./...              # Benchmarks

# Code generation
make generate                       # Generate framework patterns → pkg/sources/php/
go run ./cmd/genpatterns -o pkg/sources/php/              # All frameworks
go run ./cmd/genpatterns -framework laravel -o pkg/sources/php/  # Single framework

# Full pipeline
make all                            # generate → build → test
```

## Architecture

InputTracer is a multi-language taint analysis library (Go 1.21) that tracks how user input flows through code. It uses Tree-Sitter for multi-language parsing and supports PHP, JavaScript, TypeScript, Python, Go, Java, C, C++, C#, Ruby, and Rust.

### Analysis Pipeline

```
1. Directory Walk     → Collect files, skip configured dirs, detect language by extension
2. Tree-Sitter Parse  → Parse each file using pooled parsers (sync.Pool), cache AST (LRU, 1000 entries)
3. Parallel Analysis  → Worker pool (config.Workers goroutines) processes files concurrently:
   a. Source Detection  → Match input patterns (superglobals, request methods, env vars, etc.)
   b. Assignment Track  → Extract assignments from AST, track taint through variable assignments
   c. Call Analysis     → Find function calls with tainted arguments
4. Inter-procedural   → Build call graph, track taint across function boundaries (up to MaxDepth=5)
5. Flow Graph         → Construct nodes (sources, variables, functions) and edges (data flow)
6. Output             → Export as JSON, DOT graph, or summary report
```

### Package Responsibilities

| Package | Purpose |
|---------|---------|
| `pkg/tracer/` | Main orchestrator: `Tracer` struct, `TraceDirectory()`, `TraceFile()`, propagation, scope management |
| `pkg/parser/` | Tree-Sitter parsing with language detection, parser pooling (`sync.Pool`), LRU file cache |
| `pkg/sources/` | Input source detection: registry, matchers, type definitions (see structure below) |
| `pkg/semantic/` | Deep analysis: call graphs, symbolic execution, path analysis, classifiers |
| `pkg/ast/` | Language-agnostic AST extraction: assignments, function calls, expression matching |
| `pkg/output/` | Result serialization: JSON, DOT graph, summary report |
| `cmd/genpatterns/` | CLI tool: fetches framework source from GitHub, generates Go pattern definitions |

### Key Entry Points

- `pkg/tracer/tracer.go` - `Tracer` struct, `New()`, `TraceDirectory()`, worker pool orchestration
- `pkg/tracer/types.go` - Core types: `TraceResult`, `InputSource`, `TaintedVariable`, `TaintedFunction`, `FlowGraph`, `AnalysisState`
- `pkg/tracer/propagation.go` - `TaintPropagator`: taint flow through assignments, calls, returns, concatenation
- `pkg/tracer/scope.go` - `ScopeManager`: variable visibility with hierarchical scope stack
- `pkg/tracer/interprocedural.go` - Cross-function taint tracking with depth limiting

### Type Re-export Chain

Types are defined once in inner packages and re-exported outward for API ergonomics:

```
pkg/sources/core/types.go       → Canonical: SourceType, InputLabel, InputPattern, MatchResult
pkg/sources/common/types.go     → Definition, Match, BaseMatcher, InputLabel (re-exported from core)
pkg/sources/constants/           → PropagationStepType, ScopeType, FlowType, etc.
pkg/sources/types.go            → Re-exports common + core types at pkg/sources level
pkg/tracer/types.go             → Re-exports InputLabel, PropagationStepType from constants
```

When adding new shared types, define them in `pkg/sources/core/` or `pkg/sources/constants/` and re-export through the chain.

### Sources Package Structure

```
pkg/sources/
├── core/               # Canonical type definitions (SourceType, InputLabel, InputPattern)
├── common/             # Shared types (Definition, Match, BaseMatcher) + regex/framework patterns
├── constants/          # Enum-like constants (PropagationStepType, ScopeType, FlowType, etc.)
├── patterns/           # Variable boundary patterns, symbolic patterns, condition patterns
├── frameworks/         # Framework auto-detection logic
├── php/                # PHP: superglobals, laravel, symfony, wordpress, database, taint patterns
├── javascript/         # JS: express, fastify, koa, nestjs + generic patterns
├── typescript/         # TS-specific patterns
├── python/             # Python: django, flask + generic patterns
├── golang/             # Go: net/http, gin, echo, fiber frameworks
├── java/               # Java: Spring, servlets, annotations
├── ruby/               # Ruby: Rails, Sinatra frameworks
├── c/                  # C: stdin, argv, env, socket input
├── cpp/                # C++: cin, getline, socket input
├── csharp/             # C#: ASP.NET, MVC frameworks
├── rust/               # Rust: actix-web, rocket, axum frameworks
├── registry.go         # Central matcher registry (language → Matcher)
├── input_methods.go    # Generic input method patterns (request.get, etc.)
├── superglobals.go     # PHP superglobal definitions
├── defaults.go         # Default config values (skip dirs, max depths, cache sizes)
└── types.go            # Re-exports from common + core
```

Each language subdirectory has a `matcher.go` implementing the `Matcher` interface (`Language()` + `FindSources()`), plus optional framework-specific files.

### Semantic Package Subpackages

| Subpackage | Purpose |
|-----------|---------|
| `analyzer/{lang}/` | Language-specific deep analyzers (PHP, JS, Python, Go, Java, Ruby, C, C++, C#, Rust) |
| `callgraph/` | Function call graph construction and traversal |
| `symbolic/` | Symbolic execution engine with file cache |
| `pathanalysis/` | Path-sensitive taint analysis |
| `condition/` | Condition extraction for path sensitivity |
| `discovery/` | Dynamic source discovery (superglobals, carrier maps, taint inference) |
| `extractor/` | AST feature extraction |
| `classifier/` | Code pattern classification |
| `index/` | Code indexing for fast lookup |
| `tracer/` | Semantic-level variable tracing |
| `batch/` | Batch analysis orchestration |
| `types/` | Shared semantic analysis types |

### Variable Boundary Matching

`pkg/sources/patterns/variable_patterns.go` handles language-aware variable matching to avoid false positives:

- **PHP/Ruby**: `(?:^|[^a-zA-Z0-9_$@])$varName\b` — handles `$`/`@` sigil prefixes
- **Standard** (Go, Java, JS, etc.): `\bvarName\b` — word boundary matching

This prevents matching `$input` inside `$input_value` or `@user` inside `@username`.

## Design Patterns

- **Registry Pattern**: `sources.Registry` maps language → `Matcher`; `ast.Registry` maps language → `Extractor`. Both populated via `init()`.
- **Worker Pool**: Configurable goroutines (`config.Workers`, default `runtime.NumCPU()`) process files from a channel.
- **Parser Pooling**: `sync.Pool` reuses expensive Tree-Sitter parser instances per language.
- **Deduplication**: `FullAnalysisState` uses maps (`sourcesMap`, `taintedVarsMap`, `taintedFuncsMap`) for O(1) dedup alongside slices for ordered output.
- **Scope Stack**: `AnalysisState.ScopeStack` manages nested scopes (global → file → module → class → function → block) with variable shadowing via upward lookup.

## Adding New Language Support

1. Register parser in `pkg/parser/languages/` (add Tree-Sitter grammar, file extensions)
2. Create `pkg/sources/{lang}/matcher.go` implementing `sources.Matcher` interface
3. Add AST extractor in `pkg/ast/` implementing `ast.Extractor` interface
4. Optionally add `pkg/semantic/analyzer/{lang}/analyzer.go` for deep analysis

## Adding Framework Support (Existing Language)

1. Create `pkg/sources/{lang}/{framework}.go` with framework-specific `Definition` patterns
2. Register patterns in the language's `matcher.go` init or constructor
3. For PHP frameworks with fetchable source: add to `cmd/genpatterns/frameworks.go` for auto-generation

## Configuration

```go
config := &tracer.Config{
    Languages:       []string{},        // Empty = all supported
    MaxDepth:        5,                  // Inter-procedural analysis depth
    Workers:         runtime.NumCPU(),   // Parallel workers
    CustomSources:   []sources.Definition{},
    SkipDirs:        []string{".git", "node_modules", "vendor"},
    IncludePatterns: []string{},
}
```

## Input Labels

The library categorizes input sources: `HTTP_GET`, `HTTP_POST`, `HTTP_COOKIE`, `HTTP_HEADER`, `HTTP_BODY`, `CLI_ARG`, `ENV_VAR`, `FILE_READ`, `DATABASE`, `NETWORK`

## Test Infrastructure

- `testdata/` — Language-specific test fixture files and expected outputs
- `testapps/` — Real-world applications for integration testing (MyBB, DVWA, php-webgoat, adminer, parsedown, Chi/Cobra Go apps)
- Tests use table-driven patterns (`[]struct{name, input, want}` with `t.Run`)

## Dependencies

- `github.com/smacker/go-tree-sitter` — Multi-language parsing (core dependency)
- `github.com/google/uuid` — Unique ID generation for sources/variables/functions
- `github.com/mattn/go-sqlite3` — SQLite support (optional)
