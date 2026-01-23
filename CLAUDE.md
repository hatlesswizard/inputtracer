# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

InputTracer is a semantic user input flow tracer written in Go. It analyzes codebases to trace how user input (HTTP parameters, form data, CLI args, etc.) flows through application code using taint analysis and inter-procedural data flow tracking.

## Build Commands

Go 1.21 or later is required. CGO must be enabled (default) for tree-sitter bindings.

```bash
# Build the main CLI tool
go build -o inputtracer ./cmd/inputtracer

# Build auxiliary tools
go build -o patchleaks-extract ./cmd/patchleaks-extract

# Build all commands
go build ./cmd/...
```

## CLI Usage

```bash
# Discover input carriers from a PHP codebase
./inputtracer discover -codebase /path/to/project -o carriers.json

# Classify code snippets using discovered carriers
./inputtracer classify -carriers carriers.json -snippets snippets.json -o results.json

# Classify snippets using superglobals only (no carrier map)
./inputtracer classify-direct -snippets snippets.json -o results.json

# Trace a variable across all definitions
./inputtracer trace-var -var '$search_sql' -codebase /path/to/project

# Backward taint analysis from a target expression
./inputtracer trace-back -target '$id' -codebase /path/to/project
./inputtracer trace-back -target '$search_sql' -codebase /path/to/project -o sources.json -v

# Full directory analysis (forward taint)
./inputtracer -format=json -o flows.json ./myproject

# Symbolic expression tracing
./inputtracer -trace "$mybb->input['action']" ./mybb

# Extract snippets from PatchLeaks database for batch analysis
./patchleaks-extract -db analysis.db -analysis <id> -codebase /path/to/project -output snippets.json
```

## Architecture

### Core Packages

- **`pkg/semantic/`** - Main semantic tracer orchestrating analysis (used by `cmd/inputtracer`)
  - `tracer.go` - Core `Tracer` struct with `TraceDirectory()`, `ParseOnly()`, `TraceBackward()`
  - `types/types.go` - Universal data structures (`FlowNode`, `FlowEdge`, `SourceType`)
  - `analyzer/` - Language-specific analyzers implementing symbol table extraction and input source detection
  - `discovery/` - Input carrier discovery and superglobal taint tracing
  - `classifier/` - Snippet classification using carrier maps
  - `symbolic/` - Symbolic execution engine for deep expression tracing
  - `tracer/vartracer.go` - Variable tracing across definitions

- **`pkg/semantic/index/`** - Code indexer with O(1) symbol lookup
  - Signature-based matching (partial name, parameter patterns)
  - Cross-file reference resolution with LRU caching

- **`pkg/semantic/callgraph/`** - Call graph construction and traversal

- **`pkg/semantic/condition/`** - Condition extraction from control flow

- **`pkg/semantic/pathanalysis/`** - Path expansion for flow analysis

- **`pkg/sources/`** - Language-specific input source definitions
  - `registry.go` - Source matcher registry and base matcher
  - Language files (`php.go`, `javascript.go`, etc.) - Define superglobals and input patterns

- **`pkg/tracer/`** - Legacy data flow propagation library (public API preserved)
  - `tracer.go` - Basic flow tracing
  - `interprocedural.go` - Cross-function flow analysis
  - `propagation.go` - Taint propagation rules
  - Note: This is a simpler, older implementation. The main CLI uses `pkg/semantic/` instead.

- **`pkg/output/`** - Output formatters (JSON, GraphViz DOT, Mermaid, HTML)

### Language Support

Analyzers implemented in `pkg/semantic/analyzer/` for all 11 languages:
- **PHP** - Superglobals ($_GET, $_POST, etc.), framework carriers
- **JavaScript** - req.body, req.query, Express.js patterns
- **TypeScript** - Same as JavaScript with type awareness
- **Python** - request.args, Flask/Django patterns
- **Go** - http.Request, Gin/Echo patterns
- **Java** - HttpServletRequest, Spring patterns
- **C#** - ASP.NET Request patterns
- **C/C++** - stdin, argv, getenv patterns
- **Ruby** - params, Rack/Rails patterns
- **Rust** - Actix/Axum request patterns

### Data Flow Model

1. **Sources** - Input entry points (superglobals, framework methods)
2. **Carriers** - Objects/variables that hold user data
3. **Edges** - Assignment, parameter passing, return values, property access
4. **Sinks** - Final usage points (SQL queries, exec calls, etc.)

### Analysis Approaches

- **Forward Taint Analysis** (`TraceDirectory`) - Traces from sources to sinks
- **Backward Taint Analysis** (`TraceBackward`, `trace-back` command) - Traces from a target expression back to its input sources
- **Variable Tracing** (`trace-var` command) - Finds all definitions of a variable and determines which have user input

### Key Types

```go
// FlowNode represents a node in the data flow graph
type FlowNode struct {
    Type       FlowNodeType  // source, carrier, variable, function, sink
    SourceType SourceType    // http_get, http_post, cli_arg, etc.
    FilePath   string
    Line       int
    Name       string
}

// FlowEdge represents data flow between nodes
type FlowEdge struct {
    From, To string
    Type     FlowEdgeType  // assignment, parameter, return, etc.
}
```

### Analysis Pipeline

1. **File Discovery** - Glob patterns to find source files
2. **Parallel Parsing** - Tree-sitter parses files concurrently
3. **Symbol Table Building** - Extract classes, functions, variables per file
4. **Source Detection** - Find input sources using language-specific patterns
5. **Flow Tracing** - Inter-procedural taint propagation from sources
6. **Output Generation** - JSON, DOT, Mermaid, or HTML visualization

## Dependencies

- `github.com/smacker/go-tree-sitter` - Tree-sitter bindings for AST parsing
- `github.com/google/uuid` - UUID generation
- `github.com/mattn/go-sqlite3` - SQLite driver (used by patchleaks-extract)

## Testing

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/semantic/callgraph/...

# Run with verbose output
go test -v ./pkg/semantic/callgraph/...
```

Test files are co-located with implementation (e.g., `manager_test.go` alongside `manager.go`).

### Test Data

- `testdata/` - Sample code snippets for each language to test input source detection
- `testapps/` - Downloaded real-world applications (mybb) for integration testing
- `psalm-repo/` - Psalm PHP codebase for large-scale testing

Note: Unit tests are minimal; the project relies primarily on integration testing against real codebases.

## Memory Optimizations

The codebase includes several memory optimizations to handle large codebases:

1. **LRU File Cache** (`pkg/semantic/symbolic/filecache.go`) - Limits AST/content memory to ~100 files
2. **AST Release After Parsing** (`pkg/semantic/tracer.go`) - Releases tree-sitter trees after symbol extraction
3. **Worker Pool Pattern** - Reuses parsers across files instead of creating new ones
4. **O(1) LRU Cache** (`pkg/parser/cache.go`) - Uses `container/list` for efficient eviction
5. **FlowMap Deduplication** (`pkg/semantic/types/types.go`) - Map-based deduplication prevents duplicate nodes/edges
6. **Regex Caching** (`pkg/tracer/propagation.go`) - Caches compiled regexes to avoid recompilation
7. **Method Body Release** (`pkg/semantic/discovery/taint.go`) - Releases method body strings after pattern analysis

For memory-constrained environments, use `NewExecutionEngineWithCacheSize(n)` to control cache size.
