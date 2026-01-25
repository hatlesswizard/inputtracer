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
go build ./...           # Build all packages
go test ./...            # Run all tests
go test ./pkg/tracer     # Test specific package
go test -v ./...         # Verbose test output
go test -race ./...      # Run with race detector
```

## Architecture

InputTracer is a multi-language taint analysis library that tracks how user input flows through code.

```
┌─────────────────────────────────────────────────────────────┐
│  tracer.New(config) → TraceDirectory(path) → TraceResult   │
└─────────────────────────────────────────────────────────────┘
                            │
        ┌───────────────────┼───────────────────┐
        ▼                   ▼                   ▼
┌──────────────┐    ┌──────────────┐    ┌──────────────┐
│    Parser    │    │   Sources    │    │     AST      │
│ pkg/parser/  │    │ pkg/sources/ │    │   pkg/ast/   │
│              │    │              │    │              │
│ Tree-Sitter  │    │ Input source │    │ Language-    │
│ multi-lang   │    │ matchers per │    │ agnostic     │
│ with pooling │    │ language     │    │ extraction   │
└──────────────┘    └──────────────┘    └──────────────┘
        │                   │                   │
        └───────────────────┼───────────────────┘
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  Tracer (pkg/tracer/)                                       │
│  • Parallel worker pool (config.Workers)                    │
│  • Per-file analysis: sources → assignments → calls         │
│  • Taint propagation tracking                               │
│  • Inter-procedural analysis                                │
│  • Flow graph construction                                  │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│  Output (pkg/output/)                                       │
│  JSON export, Mermaid/DOT graph generation                  │
└─────────────────────────────────────────────────────────────┘
```

### Package Responsibilities

| Package | Purpose |
|---------|---------|
| `pkg/tracer/` | Main orchestrator: Tracer struct, TraceDirectory(), types, propagation logic |
| `pkg/parser/` | Tree-Sitter parsing with language detection, parser pooling, file caching |
| `pkg/sources/` | Language-specific input source matchers (HTTP params, CLI, env vars, etc.) |
| `pkg/semantic/` | Deep analysis: extractors, classifiers, call graphs, symbolic execution |
| `pkg/ast/` | Language-agnostic AST extraction registry |
| `pkg/output/` | Result serialization: JSON, Mermaid, DOT formats |

### Key Entry Points

- `pkg/tracer/tracer.go` - Main `Tracer` struct, `New()`, `TraceDirectory()`
- `pkg/tracer/types.go` - Core types: `TraceResult`, `InputSource`, `TaintedVariable`, `TaintedFunction`, `FlowGraph`
- `pkg/tracer/propagation.go` - Taint propagation through assignments and function calls

## Design Patterns

- **Registry Pattern**: Language-specific implementations (sources, AST extractors, analyzers) registered at init
- **Worker Pool**: Parallel file analysis via goroutines with configurable `config.Workers`
- **Parser Pooling**: `sync.Pool` reuses expensive Tree-Sitter parser instances
- **Deduplication**: Map-based tracking prevents duplicate sources/variables/functions

## Supported Languages

PHP, JavaScript, TypeScript, Python, Go, Java, C, C++, C#, Ruby, Rust

## Adding New Language Support

1. Register parser in `pkg/parser/languages/`
2. Create source matchers in `pkg/sources/{lang}.go` (implement input detection patterns)
3. Add AST extractor in `pkg/ast/` (assignment/call extraction)
4. Add semantic analyzer in `pkg/semantic/analyzer/{lang}/` (optional, for deep analysis)

## Configuration

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

## Input Labels

The library categorizes input sources by type: `HTTP_GET`, `HTTP_POST`, `HTTP_COOKIE`, `HTTP_HEADER`, `HTTP_BODY`, `CLI_ARG`, `ENV_VAR`, `FILE_READ`, `DATABASE`, `NETWORK`

---

## Proactive Plugin Agent Usage

**Use these agents automatically when the situation applies - don't wait to be asked.**

### Development Workflow Agents
- `superpowers:brainstorming` → Before any new feature/functionality
- `superpowers:test-driven-development` → Before writing implementation code
- `superpowers:systematic-debugging` → When encountering bugs or test failures
- `superpowers:writing-plans` → For multi-step tasks with requirements
- `superpowers:verification-before-completion` → Before claiming work is done
- `feature-dev:code-reviewer` → After implementing features
- `superpowers:code-reviewer` → After completing major project steps

### Analysis Agents
- `static-code-analyzer` → When reviewing for hardcoded patterns
- `performance-analyzer` → For threading/performance analysis
- `memory-optimizer` → For memory optimization opportunities
- `dead-code-eliminator` → When cleaning up after refactors

### Exploration Agents
- `Explore` (Task tool) → When searching/understanding codebase
- `feature-dev:code-explorer` → For deep feature analysis
- `feature-dev:code-architect` → When designing architectures

---

## ENFORCED WORKFLOW

**See `~/.claude/CLAUDE.md` for detailed workflow system documentation** (state machine, blocking rules, auto-marking, recovery procedures).

This project follows the global 7-step workflow with these **project-specific notes**:

### InputTracer-Specific Workflow Requirements

| Step | Project-Specific Requirement |
|------|------------------------------|
| **EXECUTE** | Always run `go test ./...` after code changes |
| **MEMORY_CHECK** | Important - Tree-Sitter parsers can be memory-intensive. Use `memory-optimizer` agent. |
| **SIMPLIFY** | Focus on `pkg/tracer/` and `pkg/sources/` packages |
| **INTERACTIVE_TEST** | **Library has no UI** - use test suite verification instead of Playwright |

### Testing Verification (Libraries)

Since InputTracer is a library without a web UI, verify via comprehensive testing:

```bash
go test ./...            # All tests pass
go test -race ./...      # No race conditions
go test -bench=. ./...   # Performance acceptable
go build ./...           # Builds successfully
```

When all tests pass, use `superpowers:verification-before-completion` skill to mark INTERACTIVE_TEST complete.

### Debugging Workflow Issues

If workflow steps aren't auto-marking, check the debug log:

```bash
tail -f ~/.claude/logs/workflow-debug.log
```

**Plan file detection**: The PLAN step searches for `.md` files (modified within 10 min) in:
- `~/.claude/plans/`
- `./docs/plans/` (this project uses this)
- `./plans/`

If PLAN doesn't mark after invoking `superpowers:writing-plans`, ensure your plan file is in one of these directories.

### Quick Start

Use `/enforced-implementation` to run the complete workflow:
```
/enforced-implementation Add feature X
```
