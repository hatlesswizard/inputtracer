# InputTracer

A multi-language taint-analysis library that traces how **user input** enters source code
and propagates through variables and function calls. It uses [Tree-sitter](https://tree-sitter.github.io/)
for parsing and supports PHP, JavaScript, TypeScript, Python, Go, Java, C, C++, C#, Ruby, and Rust.

> InputTracer traces **input sources** only — where untrusted data enters and how it flows.
> It does not classify vulnerabilities or match dangerous sinks; that analysis is intentionally
> out of scope.

## Features

- Input-source detection across 11 languages (HTTP parameters, cookies, headers, request
  bodies, CLI args, environment variables, file and network reads, etc.)
- Framework-aware patterns (Laravel, Symfony, WordPress, Express, Django, Flask, Spring,
  Rails, Gin, actix-web, and more)
- Taint propagation through assignments, concatenation, function calls, and returns
- Inter-procedural analysis with a configurable depth limit
- Parallel analysis via a worker pool, with parser pooling and an LRU AST cache
- Output as JSON, a DOT/Mermaid flow graph, or a summary report

## Requirements

- Go 1.23 or newer

## Installation

```bash
go get github.com/hatlesswizard/inputtracer
```

Or clone and build:

```bash
git clone https://github.com/hatlesswizard/inputtracer.git
cd inputtracer
go build ./...
```

## Library usage

```go
package main

import (
    "fmt"

    "github.com/hatlesswizard/inputtracer/pkg/tracer"
    "github.com/hatlesswizard/inputtracer/pkg/output"
)

func main() {
    config := tracer.DefaultConfig()
    config.MaxDepth = 5 // inter-procedural analysis depth

    t := tracer.New(config)

    result, err := t.TraceDirectory("./path/to/code")
    if err != nil {
        panic(err)
    }

    // Inspect results
    fmt.Printf("Found %d input sources\n", len(result.Sources))

    // Export as JSON
    jsonOut, _ := output.NewJSONExporter(true).Export(result)
    fmt.Println(jsonOut)

    // Or as a DOT flow graph
    dot := output.ExportDOT(result.FlowGraph)
    fmt.Println(dot)
}
```

`TraceFile` is available for single-file analysis. Configuration options include the
languages to scan, worker count, skip directories, and custom input-source definitions —
see `tracer.DefaultConfig()`.

## Supported input labels

`HTTP_GET`, `HTTP_POST`, `HTTP_COOKIE`, `HTTP_HEADER`, `HTTP_BODY`, `CLI_ARG`, `ENV_VAR`,
`FILE_READ`, `DATABASE`, `NETWORK`

## Building & testing

```bash
make build          # build cmd/ and pkg/
make test           # run package tests
make generate       # regenerate framework patterns into pkg/sources/php/
go test ./...       # run all tests
go test -race ./...
```

(`make` targets wrap the equivalent `go` commands.)

## License

Licensed under the [GNU General Public License v3.0](LICENSE).
