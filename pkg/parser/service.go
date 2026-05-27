package parser

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"

	"github.com/hatlesswizard/inputtracer/pkg/parser/languages"
	"github.com/hatlesswizard/inputtracer/pkg/sources"
)

// ErrUnsupportedLanguage is returned when a file's language is not supported.
var ErrUnsupportedLanguage = errors.New("parser: unsupported file type")

// ErrLanguageNotRegistered is returned when the requested language has no registered grammar.
var ErrLanguageNotRegistered = errors.New("parser: language not registered")

// ErrParserCreationFailed is returned when a new Tree-Sitter parser cannot be allocated.
var ErrParserCreationFailed = errors.New("parser: failed to create parser instance")

// ErrEmptyParseTree is returned when Tree-Sitter returns a nil tree for valid input.
var ErrEmptyParseTree = errors.New("parser: parse produced an empty tree")

// Service provides parsing capabilities for multiple languages
type Service struct {
	languages   map[string]*sitter.Language
	cache       *Cache
	mu          sync.RWMutex
	parserPools map[string]*sync.Pool // Parser pools per language for reuse
}

// ParseResult contains the result of parsing a file
type ParseResult struct {
	Root     *sitter.Node
	Source   []byte
	Language string
	FilePath string
}

// NewService creates a new parser service
func NewService(cacheSize ...int) *Service {
	size := 1000 // Default
	if len(cacheSize) > 0 && cacheSize[0] > 0 {
		size = cacheSize[0]
	}
	s := &Service{
		languages:   make(map[string]*sitter.Language),
		cache:       NewCache(size),
		parserPools: make(map[string]*sync.Pool),
	}
	return s
}

// RegisterLanguage registers a language parser
func (s *Service) RegisterLanguage(name string, lang *sitter.Language) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.languages[name] = lang

	// Create a pool for this language's parsers
	langRef := lang // Capture for closure
	s.parserPools[name] = &sync.Pool{
		New: func() any {
			p := sitter.NewParser()
			if p != nil {
				p.SetLanguage(langRef)
			}
			return p
		},
	}
}

// getParserFromPool gets a parser from the pool for the specified language
func (s *Service) getParserFromPool(language string) *sitter.Parser {
	s.mu.RLock()
	pool := s.parserPools[language]
	s.mu.RUnlock()

	if pool == nil {
		return nil
	}

	parser := pool.Get()
	if parser == nil {
		return nil
	}
	return parser.(*sitter.Parser)
}

// returnParserToPool returns a parser to its pool
func (s *Service) returnParserToPool(language string, parser *sitter.Parser) {
	if parser == nil {
		return
	}
	s.mu.RLock()
	pool := s.parserPools[language]
	s.mu.RUnlock()

	if pool != nil {
		pool.Put(parser)
	}
}

// GetLanguage returns the registered language by name
func (s *Service) GetLanguage(name string) *sitter.Language {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.languages[name]
}

// SupportedLanguages returns all supported language names
func (s *Service) SupportedLanguages() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	langs := make([]string, 0, len(s.languages))
	for name := range s.languages {
		langs = append(langs, name)
	}
	return langs
}

// ParseFile parses a file and returns the parse result
// MEMORY FIX: Now stores tree in cache for proper cleanup on eviction
func (s *Service) ParseFile(filePath string) (*ParseResult, error) {
	// Detect language from file extension
	lang := s.DetectLanguage(filePath)
	if lang == "" {
		return nil, ErrUnsupportedLanguage
	}

	// Check cache first
	if cached := s.cache.Get(filePath); cached != nil {
		return &ParseResult{
			Root:     cached.Root,
			Source:   cached.Source,
			Language: lang,
			FilePath: filePath,
		}, nil
	}

	// Read file
	source, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	// Parse - use ParseWithTree to get tree for proper memory management
	tree, root, err := s.ParseWithTree(source, lang)
	if err != nil {
		return nil, err
	}
	if root == nil {
		return nil, ErrEmptyParseTree
	}

	// Cache result with tree reference for proper cleanup
	s.cache.Put(filePath, &CachedParse{
		Root:   root,
		Tree:   tree, // MEMORY FIX: Store tree for cleanup on eviction
		Source: source,
	})

	return &ParseResult{
		Root:     root,
		Source:   source,
		Language: lang,
		FilePath: filePath,
	}, nil
}

// ParseWithTree parses source code and returns both tree and root node
// MEMORY FIX: Now returns the tree so it can be closed later
func (s *Service) ParseWithTree(source []byte, language string) (*sitter.Tree, *sitter.Node, error) {
	s.mu.RLock()
	lang := s.languages[language]
	s.mu.RUnlock()

	if lang == nil {
		return nil, nil, ErrLanguageNotRegistered
	}

	// Get parser from pool (reuses parsers instead of creating new ones each time)
	parser := s.getParserFromPool(language)
	if parser == nil {
		// Fallback: create a new parser if pool returns nil
		parser = sitter.NewParser()
		if parser == nil {
			return nil, nil, ErrParserCreationFailed
		}
		parser.SetLanguage(lang)
	}
	defer s.returnParserToPool(language, parser)

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, nil, err
	}
	if tree == nil {
		return nil, nil, ErrEmptyParseTree
	}

	return tree, tree.RootNode(), nil
}

// DetectLanguage detects the programming language from file path.
// Uses the centralized extension mappings from pkg/parser/languages.
func (s *Service) DetectLanguage(filePath string) string {
	// Check for unsupported special filenames first using centralized sources
	basename := strings.ToLower(filepath.Base(filePath))
	if sources.IsUnsupportedFilename(basename) {
		return "" // Not supported
	}

	// Check for special filenames with known languages
	if lang := sources.GetSpecialFilenameLanguage(basename); lang != "" {
		return lang
	}

	ext := strings.ToLower(filepath.Ext(filePath))
	return languages.GetLanguageByExtension(ext)
}

// IsSupported checks if a file type is supported
func (s *Service) IsSupported(filePath string) bool {
	lang := s.DetectLanguage(filePath)
	if lang == "" {
		return false
	}
	s.mu.RLock()
	_, exists := s.languages[lang]
	s.mu.RUnlock()
	return exists
}

// ClearCache clears the parser cache
func (s *Service) ClearCache() {
	s.cache.Clear()
}

// CacheStats returns cache statistics
func (s *Service) CacheStats() (hits, misses int64) {
	return s.cache.Stats()
}
