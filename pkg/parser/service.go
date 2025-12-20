package parser

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
)

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
		New: func() interface{} {
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
		return nil, nil // Unsupported file type
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
		return nil, nil
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
		return nil, nil, nil // Language not registered
	}

	// Get parser from pool (reuses parsers instead of creating new ones each time)
	parser := s.getParserFromPool(language)
	if parser == nil {
		// Fallback: create a new parser if pool returns nil
		parser = sitter.NewParser()
		if parser == nil {
			return nil, nil, nil
		}
		parser.SetLanguage(lang)
	}
	defer s.returnParserToPool(language, parser)

	tree, err := parser.ParseCtx(context.Background(), nil, source)
	if err != nil {
		return nil, nil, err
	}
	if tree == nil {
		return nil, nil, nil
	}

	return tree, tree.RootNode(), nil
}

// Parse parses source code with the specified language
// NOTE: Caller is responsible for closing the tree when done if not caching
func (s *Service) Parse(source []byte, language string) (*sitter.Node, error) {
	tree, root, err := s.ParseWithTree(source, language)
	if err != nil {
		return nil, err
	}
	// WARNING: Tree is not closed here - this can cause memory leaks if not cached
	// Prefer using ParseWithTree and managing tree lifecycle explicitly
	_ = tree // Intentionally not closing - caller must cache or close
	return root, nil
}

// ParseString parses source code string with the specified language
func (s *Service) ParseString(source string, language string) (*sitter.Node, error) {
	return s.Parse([]byte(source), language)
}

// DetectLanguage detects the programming language from file path
func (s *Service) DetectLanguage(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))

	// Check for special filenames first
	basename := strings.ToLower(filepath.Base(filePath))
	switch basename {
	case "makefile", "gnumakefile":
		return "" // Not supported yet
	}

	switch ext {
	case ".php", ".php5", ".php7", ".phtml":
		return "php"
	case ".js", ".mjs", ".cjs":
		return "javascript"
	case ".ts", ".mts", ".cts":
		return "typescript"
	case ".tsx":
		return "tsx"
	case ".jsx":
		return "javascript"
	case ".py", ".pyw", ".pyi":
		return "python"
	case ".go":
		return "go"
	case ".java":
		return "java"
	case ".c", ".h":
		return "c"
	case ".cpp", ".cc", ".cxx", ".hpp", ".hxx", ".h++":
		return "cpp"
	case ".cs":
		return "c_sharp"
	case ".rb", ".rake", ".gemspec":
		return "ruby"
	case ".rs":
		return "rust"
	default:
		return ""
	}
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
