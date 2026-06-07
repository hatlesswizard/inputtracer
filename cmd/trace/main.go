package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/hatlesswizard/inputtracer/pkg/tracer"
)

// protocolSource is one detected source in the findings/PROTOCOL.md format.
type protocolSource struct {
	File  string `json:"file"`
	Line  int    `json:"line"`
	Code  string `json:"code"`
	Label string `json:"label"`
	Type  string `json:"type"`
}

// protocolResult matches findings/tracer/{repo-name}.json in PROTOCOL.md.
type protocolResult struct {
	Repo                 string           `json:"repo"`
	TracerVersion        string           `json:"tracer_version"`
	TotalSourcesDetected int              `json:"total_sources_detected"`
	Sources              []protocolSource `json:"sources"`
	RunErrors            []string         `json:"run_errors"`
	RunDurationSeconds   float64          `json:"run_duration_seconds"`
}

func main() {
	var (
		repoName = flag.String("repo", "", "repo identifier for output (e.g. owner/name); defaults to dir base name")
		version  = flag.String("version", "", "tracer version/commit to record in output")
		raw      = flag.Bool("raw", false, "emit the full raw TraceResult JSON instead of PROTOCOL format")
		workers  = flag.Int("workers", 0, "parallel workers (0 = NumCPU; use 1 to avoid tree-sitter CGO concurrency crashes on large repos)")
	)
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: trace [flags] <directory>\n\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}
	dir := flag.Arg(0)

	if info, err := os.Stat(dir); err != nil || !info.IsDir() {
		fmt.Fprintf(os.Stderr, "error: %q is not a directory\n", dir)
		os.Exit(1)
	}

	cfg := tracer.DefaultConfig()
	cfg.Languages = []string{"php"}
	if *workers > 0 {
		cfg.Workers = *workers
	} else {
		cfg.Workers = runtime.NumCPU()
	}

	t := tracer.New(cfg)

	start := time.Now()
	result, err := t.TraceDirectory(dir)
	elapsed := time.Since(start).Seconds()

	if err != nil {
		// Fatal failure: emit a protocol result carrying the error so the
		// driver always gets parseable JSON.
		out := protocolResult{
			Repo:               resolveRepo(*repoName, dir),
			TracerVersion:      *version,
			Sources:            []protocolSource{},
			RunErrors:          []string{err.Error()},
			RunDurationSeconds: elapsed,
		}
		emit(out)
		os.Exit(2)
	}

	if *raw {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(result); err != nil {
			fmt.Fprintf(os.Stderr, "error encoding result: %v\n", err)
			os.Exit(1)
		}
		return
	}

	out := protocolResult{
		Repo:               resolveRepo(*repoName, dir),
		TracerVersion:      *version,
		Sources:            toProtocolSources(result, dir),
		RunErrors:          errStrings(result.Errors),
		RunDurationSeconds: elapsed,
	}
	out.TotalSourcesDetected = len(out.Sources)
	emit(out)
}

func emit(out protocolResult) {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "error encoding result: %v\n", err)
		os.Exit(1)
	}
}

func resolveRepo(repoName, dir string) string {
	if repoName != "" {
		return repoName
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return filepath.Base(dir)
	}
	return filepath.Base(abs)
}

func toProtocolSources(result *tracer.TraceResult, root string) []protocolSource {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		absRoot = root
	}

	sources := make([]protocolSource, 0, len(result.Sources))
	for _, s := range result.Sources {
		if s == nil {
			continue
		}
		label := ""
		if len(s.Labels) > 0 {
			label = string(s.Labels[0])
		}
		code := s.Location.Snippet
		if code == "" {
			code = s.Type
			if s.Key != "" {
				code = fmt.Sprintf("%s['%s']", s.Type, s.Key)
			}
		}
		sources = append(sources, protocolSource{
			File:  relPath(absRoot, s.Location.FilePath),
			Line:  s.Location.Line,
			Code:  strings.TrimSpace(code),
			Label: label,
			Type:  s.Type,
		})
	}
	return sources
}

func relPath(absRoot, filePath string) string {
	abs, err := filepath.Abs(filePath)
	if err != nil {
		return filePath
	}
	rel, err := filepath.Rel(absRoot, abs)
	if err != nil {
		return filePath
	}
	return filepath.ToSlash(rel)
}

func errStrings(errs []error) []string {
	out := make([]string, 0, len(errs))
	for _, e := range errs {
		if e != nil {
			out = append(out, e.Error())
		}
	}
	return out
}
