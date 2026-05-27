package semantic

// file_discoverer.go — file discovery and parsing responsibility extracted from tracer.go.
//
// fileDiscoverer handles:
//   1. Walking a directory tree to collect relevant source files (discoverFiles)
//   2. Parsing those files in parallel using a worker pool (parseFiles)
//   3. Parsing a single file with a given tree-sitter parser (parseFileWithParser)
//   4. Language detection helpers and glob utilities
//
// The Tracer delegates all of these operations here. Results are written back
// into t.files / t.stats (still owned by *Tracer) so the public API is undisturbed.

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/analyzer"
	"github.com/hatlesswizard/inputtracer/pkg/semantic/types"
	sitter "github.com/smacker/go-tree-sitter"
)

// fileDiscoverer is responsible for finding source files on disk and parsing
// them into FileInfo records that the Tracer can work with.
//
// It does NOT own t.files or t.stats — those stay on *Tracer. Parsed results
// are deposited directly into those maps under the Tracer's mu lock so the
// rest of the tracer pipeline sees them unchanged.
type fileDiscoverer struct {
	config *Config
}

// newFileDiscoverer constructs a fileDiscoverer with the given configuration.
func newFileDiscoverer(config *Config) *fileDiscoverer {
	return &fileDiscoverer{config: config}
}

// discoverFiles walks root and returns all source files that pass include/
// exclude filters and the configured language list.
func (fd *fileDiscoverer) discoverFiles(root string) ([]string, error) {
	var files []string

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}

		if info.IsDir() {
			rel, _ := filepath.Rel(root, path)
			for _, pattern := range fd.config.ExcludePatterns {
				if doubleStarMatch(pattern, rel) {
					return filepath.SkipDir
				}
			}
			return nil
		}

		rel, _ := filepath.Rel(root, path)
		for _, pattern := range fd.config.ExcludePatterns {
			if doubleStarMatch(pattern, rel) {
				return nil
			}
		}

		for _, pattern := range fd.config.IncludePatterns {
			if matched, _ := filepath.Match(pattern, filepath.Base(path)); matched {
				lang := detectLanguage(path)
				if len(fd.config.Languages) == 0 || slices.Contains(fd.config.Languages, lang) {
					files = append(files, path)
				}
				return nil
			}
		}

		return nil
	})

	return files, err
}

// parseFiles parses all files using a parallel worker pool.
// Each worker owns its own per-language parsers (reused across files within
// that worker). Results are deposited into t.files / t.stats under t.mu.
func (fd *fileDiscoverer) parseFiles(t *Tracer, files []string) {
	numWorkers := fd.config.Workers
	if numWorkers <= 0 {
		numWorkers = runtime.NumCPU()
	}

	fileChan := make(chan string, len(files))
	for _, f := range files {
		fileChan <- f
	}
	close(fileChan)

	var memoryExceeded bool
	var memCheckMu sync.Mutex
	filesProcessed := 0
	gcInterval := 25

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			parsers := make(map[string]*sitter.Parser)

			for path := range fileChan {
				memCheckMu.Lock()
				if memoryExceeded {
					memCheckMu.Unlock()
					continue
				}
				filesProcessed++
				localCount := filesProcessed
				memCheckMu.Unlock()

				lang := detectLanguage(path)
				if lang == "" {
					continue
				}

				p, ok := parsers[lang]
				if !ok {
					p = createParser(lang)
					if p == nil {
						continue
					}
					parsers[lang] = p
				}

				fd.parseFileWithParser(t, path, lang, p)

				if fd.config.MaxMemoryMB > 0 && localCount%gcInterval == 0 {
					runtime.GC()
					memMB := getMemoryUsageMB()
					maxMB := uint64(fd.config.MaxMemoryMB)
					if memMB > maxMB {
						memCheckMu.Lock()
						memoryExceeded = true
						memCheckMu.Unlock()
						if fd.config.Verbose {
							fmt.Printf("  [Memory] Limit exceeded (%d MB > %d MB) after %d files - stopping\n",
								memMB, maxMB, localCount)
						}
					} else if fd.config.Verbose {
						fmt.Printf("  [Memory] After %d files: %d MB\n", localCount, memMB)
					}
				}
			}
		}()
	}

	wg.Wait()
}

// parseFileWithParser parses a single file and records results in t.files / t.stats.
// Memory-efficient: AST is released immediately after extracting symbol-table
// data, assignments, and call sites.
func (fd *fileDiscoverer) parseFileWithParser(t *Tracer, path string, lang string, p *sitter.Parser) {
	startTime := time.Now()

	langAnalyzer := analyzer.DefaultRegistry.Get(lang)
	if langAnalyzer == nil {
		return
	}

	// Check file size limit
	maxFileSize := fd.config.MaxFileSizeBytes
	if maxFileSize > 0 {
		fi, err := os.Stat(path)
		if err == nil && fi.Size() > maxFileSize {
			t.mu.Lock()
			t.files[path] = &FileInfo{
				Path:     path,
				Language: lang,
				Error:    fmt.Errorf("file too large: %d bytes (limit: %d)", fi.Size(), maxFileSize),
			}
			t.stats.FilesSkipped++
			t.mu.Unlock()
			if fd.config.Verbose {
				fmt.Printf("  Skipping large file: %s (%d MB)\n", path, fi.Size()/1024/1024)
			}
			return
		}
	}

	content, err := os.ReadFile(path)
	if err != nil {
		t.mu.Lock()
		t.files[path] = &FileInfo{Path: path, Language: lang, Error: err}
		t.stats.ParseErrors++
		t.mu.Unlock()
		return
	}

	tree, err := p.ParseCtx(context.Background(), nil, content)
	if err != nil {
		t.mu.Lock()
		t.files[path] = &FileInfo{Path: path, Language: lang, Error: err}
		t.stats.ParseErrors++
		t.mu.Unlock()
		return
	}

	root := tree.RootNode()

	symbolTable, err := langAnalyzer.BuildSymbolTable(path, content, root)
	if err != nil {
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

	srcs, err := langAnalyzer.FindInputSources(root, content)
	if err != nil {
		srcs = []*types.FlowNode{}
	}

	for _, src := range srcs {
		src.FilePath = path
		if src.ID == "" || !strings.Contains(src.ID, path) {
			src.ID = fmt.Sprintf("%s:%d:%d", path, src.Line, src.Column)
		}
	}

	var assignments []*types.Assignment
	var calls []*types.CallSite
	if len(srcs) > 0 {
		assignments, _ = langAnalyzer.ExtractAssignments(root, content, "")
		calls, _ = langAnalyzer.ExtractCalls(root, content, "")
	}

	// Release AST immediately — all needed info has been extracted above.
	tree.Close()

	parseTime := time.Since(startTime)

	t.mu.Lock()
	t.files[path] = &FileInfo{
		Path:         path,
		Language:     lang,
		SymbolTable:  symbolTable,
		Sources:      srcs,
		Assignments:  assignments,
		Calls:        calls,
		Root:         nil,    // Don't retain AST — saves ~10x file size in memory
		Content:      nil,    // Don't retain content — can re-read if needed
		ParseTime:    parseTime,
		NeedsReparse: true,   // Mark that AST was released
	}
	t.stats.FilesParsed++
	if t.stats.ByLanguage[lang] == nil {
		t.stats.ByLanguage[lang] = &LanguageStats{}
	}
	t.stats.ByLanguage[lang].Files++
	t.stats.ByLanguage[lang].Sources += len(srcs)
	t.stats.ByLanguage[lang].ParseTime += parseTime
	t.mu.Unlock()
}

// createParser creates a new tree-sitter parser for lang by consulting the
// centralised languages.GetAllLanguages() registry.
// Adding a new language requires no changes here — only in the registry.
func createParser(lang string) *sitter.Parser {
	for _, info := range languages.GetAllLanguages() {
		if info.Name == lang {
			p := sitter.NewParser()
			p.SetLanguage(info.Language)
			return p
		}
	}
	return nil
}

// detectLanguage maps a file path to its programming language name using the
// centralised extension registry in pkg/parser/languages.
func detectLanguage(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	lang := languages.GetLanguageByExtension(ext)
	// TSX uses the typescript analyzer for semantic analysis.
	if lang == "tsx" {
		return "typescript"
	}
	return lang
}

// ---- Glob / path-matching helpers ----------------------------------------

// doubleStarMatch reports whether name matches the glob pattern with ** support.
func doubleStarMatch(pattern, name string) bool {
	return matchGlob(pattern, name)
}

// matchGlob splits pattern and name into path segments and delegates to matchParts.
func matchGlob(pattern, name string) bool {
	if pattern == "**" {
		return true
	}
	patParts := strings.Split(pattern, "/")
	nameParts := strings.Split(name, "/")
	return matchParts(patParts, nameParts)
}

// matchParts matches path segment slices, treating "**" as zero-or-more segments.
func matchParts(patParts, nameParts []string) bool {
	for len(patParts) > 0 {
		pat := patParts[0]
		if pat == "**" {
			if len(patParts) == 1 {
				return true
			}
			for i := 0; i <= len(nameParts); i++ {
				if matchParts(patParts[1:], nameParts[i:]) {
					return true
				}
			}
			return false
		}
		if len(nameParts) == 0 {
			return false
		}
		if !singleSegmentMatch(pat, nameParts[0]) {
			return false
		}
		patParts = patParts[1:]
		nameParts = nameParts[1:]
	}
	return len(nameParts) == 0
}

// singleSegmentMatch matches a single path segment using standard glob wildcards.
func singleSegmentMatch(pattern, name string) bool {
	matched, err := filepath.Match(pattern, name)
	return err == nil && matched
}
