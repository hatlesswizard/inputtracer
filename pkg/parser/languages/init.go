package languages

import (
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/c"
	"github.com/smacker/go-tree-sitter/cpp"
	"github.com/smacker/go-tree-sitter/csharp"
	"github.com/smacker/go-tree-sitter/golang"
	"github.com/smacker/go-tree-sitter/java"
	"github.com/smacker/go-tree-sitter/javascript"
	"github.com/smacker/go-tree-sitter/php"
	"github.com/smacker/go-tree-sitter/python"
	"github.com/smacker/go-tree-sitter/ruby"
	"github.com/smacker/go-tree-sitter/rust"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// Extension lookup map (lazily initialized)
var (
	extToLanguage     map[string]string
	extToLanguageOnce sync.Once
)

func initExtensionMap() {
	extToLanguage = make(map[string]string)
	for _, lang := range GetAllLanguages() {
		for _, ext := range lang.Extensions {
			extToLanguage[ext] = lang.Name
		}
	}
}

// GetLanguageByExtension returns the language name for a file extension (e.g., ".php" -> "php").
// Returns empty string if extension is not recognized.
func GetLanguageByExtension(ext string) string {
	extToLanguageOnce.Do(initExtensionMap)
	return extToLanguage[ext]
}

// GetExtensionsForLanguage returns all file extensions for a given language name.
// Returns nil if the language is not found.
func GetExtensionsForLanguage(langName string) []string {
	for _, lang := range GetAllLanguages() {
		if lang.Name == langName {
			return lang.Extensions
		}
	}
	return nil
}

// BuildIncludePatterns generates glob patterns for all supported file extensions.
// Returns patterns like ["*.php", "*.php5", "*.js", ...].
func BuildIncludePatterns() []string {
	var patterns []string
	for _, lang := range GetAllLanguages() {
		for _, ext := range lang.Extensions {
			patterns = append(patterns, "*"+ext)
		}
	}
	return patterns
}

// LanguageInfo contains information about a supported language
type LanguageInfo struct {
	Name       string
	Language   *sitter.Language
	Extensions []string
}

// GetAllLanguages returns all supported language parsers
func GetAllLanguages() []LanguageInfo {
	return []LanguageInfo{
		{
			Name:       "php",
			Language:   php.GetLanguage(),
			Extensions: []string{".php", ".php5", ".php7", ".phtml"},
		},
		{
			Name:       "javascript",
			Language:   javascript.GetLanguage(),
			Extensions: []string{".js", ".mjs", ".cjs", ".jsx"},
		},
		{
			Name:       "typescript",
			Language:   typescript.GetLanguage(),
			Extensions: []string{".ts", ".mts", ".cts"},
		},
		{
			Name:       "tsx",
			Language:   tsx.GetLanguage(),
			Extensions: []string{".tsx"},
		},
		{
			Name:       "python",
			Language:   python.GetLanguage(),
			Extensions: []string{".py", ".pyw", ".pyi"},
		},
		{
			Name:       "go",
			Language:   golang.GetLanguage(),
			Extensions: []string{".go"},
		},
		{
			Name:       "java",
			Language:   java.GetLanguage(),
			Extensions: []string{".java"},
		},
		{
			Name:       "c",
			Language:   c.GetLanguage(),
			Extensions: []string{".c", ".h"},
		},
		{
			Name:       "cpp",
			Language:   cpp.GetLanguage(),
			Extensions: []string{".cpp", ".cc", ".cxx", ".hpp", ".hxx", ".h++"},
		},
		{
			Name:       "c_sharp",
			Language:   csharp.GetLanguage(),
			Extensions: []string{".cs"},
		},
		{
			Name:       "ruby",
			Language:   ruby.GetLanguage(),
			Extensions: []string{".rb", ".rake", ".gemspec"},
		},
		{
			Name:       "rust",
			Language:   rust.GetLanguage(),
			Extensions: []string{".rs"},
		},
	}
}

// RegisterAll registers all language parsers with a service
type ParserRegistrar interface {
	RegisterLanguage(name string, lang *sitter.Language)
}

// RegisterAllLanguages registers all supported languages with the given registrar
func RegisterAllLanguages(registrar ParserRegistrar) {
	for _, lang := range GetAllLanguages() {
		registrar.RegisterLanguage(lang.Name, lang.Language)
	}
}
