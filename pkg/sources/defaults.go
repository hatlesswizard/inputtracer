// Package sources - defaults.go provides centralized default configuration values
// All default values should be defined here and referenced from other packages
package sources

import "strings"

// DefaultSkipDirs contains directories that should be skipped during analysis
// Replaces hardcoded array in tracer.go DefaultConfig()
var DefaultSkipDirs = []string{
	".git",
	"node_modules",
	"vendor",
	"__pycache__",
	".venv",
	"venv",
	"target",
	"build",
	"dist",
	".idea",
	".vscode",
	".cache",
}

// WordPressVendorDirs lists WordPress-ecosystem bundled third-party directories
// that should be excluded from analysis. Callers that know they are analyzing a
// WordPress project can append these to Config.SkipDirs.
var WordPressVendorDirs = []string{
	"freemius",
	"action-scheduler",
	"redux-core",
	"redux-framework",
	"cmb2",
	"starter-content",
	"starter-templates",
}

// LanguageSkipDirs provides language-specific skip directories
var LanguageSkipDirs = map[string][]string{
	"common":     {".git", ".svn", ".hg"},
	"javascript": {"node_modules", "bower_components", "dist", "build"},
	"python":     {"__pycache__", ".venv", "venv", "env", ".tox", ".pytest_cache"},
	"go":         {"vendor"},
	"rust":       {"target"},
	"java":       {"target", "build", "bin", "out"},
	"c_sharp":    {"bin", "obj", "packages"},
	"ruby":       {"vendor", ".bundle"},
	"php":        {"vendor", "cache", "tests", "tmp", "storage"},
	"ide":        {".idea", ".vscode", ".vs"},
}

// DefaultMaxDepth is the default inter-procedural analysis depth.
// 5 hops covers most real-world call chains without exponential blowup.
const DefaultMaxDepth = 5

// DefaultSymbolicMaxDepth is used for symbolic execution tracing.
// Higher than MaxDepth because property-chain resolution needs more hops.
const DefaultSymbolicMaxDepth = 10

// DefaultPathMaxDepth limits path expansion to prevent combinatorial explosion.
const DefaultPathMaxDepth = 50

// DefaultCacheSize is the default parser cache size (number of AST entries).
const DefaultCacheSize = 1000

// DefaultCacheMemoryLimit is the memory ceiling for LRU caches.
// 32 MB prevents RSS spikes when analysing large repos.
const DefaultCacheMemoryLimit = 32 * 1024 * 1024

// DefaultFileCacheSize is the number of parsed files kept in the symbolic executor cache.
const DefaultFileCacheSize = 100

// DefaultMaxFlowNodes caps graph size to prevent unbounded memory growth.
const DefaultMaxFlowNodes = 10000

// DefaultMaxFlowEdges caps graph size to prevent unbounded memory growth.
const DefaultMaxFlowEdges = 20000

// Pre-allocation hints for slices
const (
	InitialCallStackCapacity = 32
	InitialFlowsCapacity     = 64
	InitialSourcesCapacity   = 16
	InitialVariablesCapacity = 32
)

// DefaultSnippetLength is the default maximum length for code snippets
const DefaultSnippetLength = 100

// DefaultTopFilesCount is the default number of "most tainted files" to return
const DefaultTopFilesCount = 10

// GetSkipDirsForLanguages returns combined skip directories for specified languages
func GetSkipDirsForLanguages(languages []string) []string {
	dirs := make(map[string]bool)

	// Always include common directories
	for _, d := range LanguageSkipDirs["common"] {
		dirs[d] = true
	}

	// Add IDE directories
	for _, d := range LanguageSkipDirs["ide"] {
		dirs[d] = true
	}

	// Add language-specific directories
	for _, lang := range languages {
		if langDirs, ok := LanguageSkipDirs[lang]; ok {
			for _, d := range langDirs {
				dirs[d] = true
			}
		}
	}

	// Convert to slice
	result := make([]string, 0, len(dirs))
	for d := range dirs {
		result = append(result, d)
	}
	return result
}

// PHPDiscoverySkipDirs returns directories to skip during PHP discovery
// These include vendor, cache, tests, and VCS directories
var PHPDiscoverySkipDirs = []string{
	"/vendor/",
	"/cache/",
	"/tests/",
	"/.git/",
	"/tmp/",
	"/storage/",
	"/freemius/",
	"/action-scheduler/",
	"/redux-core/",
	"/redux-framework/",
	"/cmb2/",
	"/starter-content/",
	"/starter-templates/",
}

// ShouldSkipPHPPath checks if a path should be skipped during PHP discovery
func ShouldSkipPHPPath(path string) bool {
	for _, skip := range PHPDiscoverySkipDirs {
		if strings.Contains(path, skip) {
			return true
		}
	}
	return false
}
