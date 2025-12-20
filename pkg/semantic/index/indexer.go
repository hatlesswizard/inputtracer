// Package index provides a unified code indexer with signature-based lookup,
// inspired by ATLANTIS's multi-tier code retrieval approach.
//
// Key features:
// - Fast O(1) lookup of functions, classes, and methods
// - Signature-based matching (partial name, parameter patterns)
// - Cross-file reference resolution
// - Definition-usage linking
// - LRU caching for performance
// - Glob pattern matching for bulk queries
package index

import (
	"container/list"
	"regexp"
	"strings"
	"sync"
)

// SymbolType represents the type of a code symbol
type SymbolType string

const (
	SymbolFunction   SymbolType = "function"
	SymbolMethod     SymbolType = "method"
	SymbolClass      SymbolType = "class"
	SymbolInterface  SymbolType = "interface"
	SymbolVariable   SymbolType = "variable"
	SymbolConstant   SymbolType = "constant"
	SymbolProperty   SymbolType = "property"
	SymbolParameter  SymbolType = "parameter"
)

// Symbol represents an indexed code symbol
type Symbol struct {
	ID          string     `json:"id"`           // Unique identifier
	Name        string     `json:"name"`         // Symbol name
	Type        SymbolType `json:"type"`
	FilePath    string     `json:"file_path"`
	Line        int        `json:"line"`
	Column      int        `json:"column"`
	EndLine     int        `json:"end_line"`
	EndColumn   int        `json:"end_column"`

	// For functions/methods
	Signature   string     `json:"signature,omitempty"`    // Full signature
	Parameters  []string   `json:"parameters,omitempty"`   // Parameter names
	ParamTypes  []string   `json:"param_types,omitempty"`  // Parameter types
	ReturnType  string     `json:"return_type,omitempty"`

	// For methods/properties
	ClassName   string     `json:"class_name,omitempty"`
	Visibility  string     `json:"visibility,omitempty"`   // public/private/protected
	IsStatic    bool       `json:"is_static,omitempty"`

	// Metadata
	Language    string     `json:"language"`
	DocComment  string     `json:"doc_comment,omitempty"`
	Annotations []string   `json:"annotations,omitempty"`

	// References
	References  []Reference `json:"references,omitempty"`
}

// Reference represents a usage of a symbol
type Reference struct {
	FilePath string `json:"file_path"`
	Line     int    `json:"line"`
	Column   int    `json:"column"`
	Context  string `json:"context,omitempty"` // Surrounding code
	RefType  string `json:"ref_type"`          // call, assignment, import, etc.
}

// SearchQuery represents a search query
type SearchQuery struct {
	Name         string      // Exact or partial name
	NamePattern  string      // Regex pattern for name
	Type         SymbolType  // Filter by type
	ClassName    string      // Filter by class
	FilePath     string      // Filter by file
	FilePattern  string      // Glob pattern for files
	Language     string      // Filter by language
	HasParameter string      // Has parameter matching this name
	ReturnType   string      // Filter by return type
	Limit        int         // Max results (0 = unlimited)
}

// SearchResult represents a search result
type SearchResult struct {
	Symbol   *Symbol `json:"symbol"`
	Score    float64 `json:"score"`    // Match quality score
	MatchedBy string `json:"matched_by"` // What matched (name, signature, etc.)
}

// Indexer provides fast symbol lookup and cross-reference
type Indexer struct {
	mu sync.RWMutex

	// Primary indexes
	symbolsByID     map[string]*Symbol              // id -> symbol
	symbolsByName   map[string][]*Symbol            // name -> symbols (multiple files)
	symbolsByFile   map[string][]*Symbol            // filepath -> symbols

	// Secondary indexes for fast lookup
	functionIndex   map[string][]*Symbol            // func name -> functions
	classIndex      map[string]*Symbol              // class name -> class
	methodIndex     map[string]map[string]*Symbol   // class -> method -> symbol

	// Signature index for pattern matching
	signatureIndex  map[string][]*Symbol            // normalized signature -> symbols

	// Reference index
	referenceIndex  map[string][]Reference          // symbolID -> references

	// LRU cache for search results
	searchCache     map[string][]*SearchResult
	searchCacheLRU  *list.List
	searchCacheMap  map[string]*list.Element
	maxCacheSize    int

	// Statistics
	totalSymbols    int
	totalReferences int
}

// NewIndexer creates a new code indexer
func NewIndexer() *Indexer {
	return &Indexer{
		symbolsByID:    make(map[string]*Symbol),
		symbolsByName:  make(map[string][]*Symbol),
		symbolsByFile:  make(map[string][]*Symbol),
		functionIndex:  make(map[string][]*Symbol),
		classIndex:     make(map[string]*Symbol),
		methodIndex:    make(map[string]map[string]*Symbol),
		signatureIndex: make(map[string][]*Symbol),
		referenceIndex: make(map[string][]Reference),
		searchCache:    make(map[string][]*SearchResult),
		searchCacheLRU: list.New(),
		searchCacheMap: make(map[string]*list.Element),
		maxCacheSize:   1000,
	}
}

// AddSymbol adds a symbol to the index
func (idx *Indexer) AddSymbol(sym *Symbol) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	// Generate ID if not set
	if sym.ID == "" {
		sym.ID = idx.generateID(sym)
	}

	// Add to primary indexes
	idx.symbolsByID[sym.ID] = sym
	idx.symbolsByName[sym.Name] = append(idx.symbolsByName[sym.Name], sym)
	idx.symbolsByFile[sym.FilePath] = append(idx.symbolsByFile[sym.FilePath], sym)

	// Add to type-specific indexes
	switch sym.Type {
	case SymbolFunction:
		idx.functionIndex[sym.Name] = append(idx.functionIndex[sym.Name], sym)
	case SymbolClass, SymbolInterface:
		idx.classIndex[sym.Name] = sym
	case SymbolMethod:
		if sym.ClassName != "" {
			if idx.methodIndex[sym.ClassName] == nil {
				idx.methodIndex[sym.ClassName] = make(map[string]*Symbol)
			}
			idx.methodIndex[sym.ClassName][sym.Name] = sym
		}
	}

	// Add to signature index
	if sym.Signature != "" {
		normSig := idx.normalizeSignature(sym.Signature)
		idx.signatureIndex[normSig] = append(idx.signatureIndex[normSig], sym)
	}

	idx.totalSymbols++
	idx.invalidateCache()
}

// AddReference adds a reference to a symbol
func (idx *Indexer) AddReference(symbolID string, ref Reference) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.referenceIndex[symbolID] = append(idx.referenceIndex[symbolID], ref)
	idx.totalReferences++
}

// GetSymbol retrieves a symbol by ID
func (idx *Indexer) GetSymbol(id string) *Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.symbolsByID[id]
}

// GetByName retrieves all symbols with a given name
func (idx *Indexer) GetByName(name string) []*Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.symbolsByName[name]
}

// GetFunction retrieves functions by name
func (idx *Indexer) GetFunction(name string) []*Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.functionIndex[name]
}

// GetClass retrieves a class by name
func (idx *Indexer) GetClass(name string) *Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.classIndex[name]
}

// GetMethod retrieves a method by class and method name
func (idx *Indexer) GetMethod(className, methodName string) *Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	if methods, ok := idx.methodIndex[className]; ok {
		return methods[methodName]
	}
	return nil
}

// GetSymbolsInFile retrieves all symbols in a file
func (idx *Indexer) GetSymbolsInFile(filePath string) []*Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.symbolsByFile[filePath]
}

// GetReferences retrieves all references to a symbol
func (idx *Indexer) GetReferences(symbolID string) []Reference {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.referenceIndex[symbolID]
}

// Search performs a search query against the index
func (idx *Indexer) Search(query *SearchQuery) []*SearchResult {
	// Check cache
	cacheKey := idx.queryCacheKey(query)
	idx.mu.RLock()
	if cached, ok := idx.searchCache[cacheKey]; ok {
		idx.mu.RUnlock()
		idx.touchCache(cacheKey)
		return cached
	}
	idx.mu.RUnlock()

	// Perform search
	results := idx.performSearch(query)

	// Cache results
	idx.mu.Lock()
	idx.cacheResults(cacheKey, results)
	idx.mu.Unlock()

	return results
}

// performSearch executes the actual search
func (idx *Indexer) performSearch(query *SearchQuery) []*SearchResult {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var results []*SearchResult
	var candidates []*Symbol

	// Start with candidates based on most restrictive filter
	if query.Name != "" {
		// Exact name match
		candidates = idx.symbolsByName[query.Name]
	} else if query.NamePattern != "" {
		// Pattern match
		re, err := regexp.Compile(query.NamePattern)
		if err == nil {
			for name, syms := range idx.symbolsByName {
				if re.MatchString(name) {
					candidates = append(candidates, syms...)
				}
			}
		}
	} else if query.ClassName != "" && query.Type == SymbolMethod {
		// Method lookup
		if methods, ok := idx.methodIndex[query.ClassName]; ok {
			for _, sym := range methods {
				candidates = append(candidates, sym)
			}
		}
	} else if query.FilePath != "" {
		// File lookup
		candidates = idx.symbolsByFile[query.FilePath]
	} else if query.FilePattern != "" {
		// Pattern file lookup
		re, err := regexp.Compile(idx.globToRegex(query.FilePattern))
		if err == nil {
			for path, syms := range idx.symbolsByFile {
				if re.MatchString(path) {
					candidates = append(candidates, syms...)
				}
			}
		}
	} else {
		// All symbols
		for _, sym := range idx.symbolsByID {
			candidates = append(candidates, sym)
		}
	}

	// Filter and score candidates
	for _, sym := range candidates {
		score, matched := idx.matchAndScore(sym, query)
		if score > 0 {
			results = append(results, &SearchResult{
				Symbol:    sym,
				Score:     score,
				MatchedBy: matched,
			})
		}
	}

	// Sort by score (highest first)
	idx.sortResults(results)

	// Apply limit
	if query.Limit > 0 && len(results) > query.Limit {
		results = results[:query.Limit]
	}

	return results
}

// matchAndScore checks if a symbol matches and returns a score
func (idx *Indexer) matchAndScore(sym *Symbol, query *SearchQuery) (float64, string) {
	score := 0.0
	matchedBy := ""

	// Type filter
	if query.Type != "" && sym.Type != query.Type {
		return 0, ""
	}

	// Language filter
	if query.Language != "" && sym.Language != query.Language {
		return 0, ""
	}

	// Class filter
	if query.ClassName != "" && sym.ClassName != query.ClassName {
		return 0, ""
	}

	// Return type filter
	if query.ReturnType != "" && sym.ReturnType != query.ReturnType {
		return 0, ""
	}

	// File path filter
	if query.FilePath != "" && sym.FilePath != query.FilePath {
		return 0, ""
	}

	// Name matching
	if query.Name != "" {
		if sym.Name == query.Name {
			score += 100
			matchedBy = "exact_name"
		} else if strings.Contains(strings.ToLower(sym.Name), strings.ToLower(query.Name)) {
			score += 50
			matchedBy = "partial_name"
		}
	}

	// Pattern matching
	if query.NamePattern != "" {
		re, err := regexp.Compile(query.NamePattern)
		if err == nil && re.MatchString(sym.Name) {
			score += 75
			matchedBy = "pattern"
		}
	}

	// Parameter matching
	if query.HasParameter != "" {
		for _, param := range sym.Parameters {
			if strings.Contains(param, query.HasParameter) {
				score += 30
				if matchedBy != "" {
					matchedBy += "+parameter"
				} else {
					matchedBy = "parameter"
				}
				break
			}
		}
	}

	// If no specific matching but passed filters
	if score == 0 && (query.Type != "" || query.ClassName != "" || query.FilePath != "") {
		score = 10
		matchedBy = "filter"
	}

	return score, matchedBy
}

// sortResults sorts results by score descending
func (idx *Indexer) sortResults(results []*SearchResult) {
	// Simple bubble sort for small result sets
	n := len(results)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if results[j].Score < results[j+1].Score {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}
}

// generateID generates a unique ID for a symbol
func (idx *Indexer) generateID(sym *Symbol) string {
	id := sym.FilePath + ":" + string(sym.Type) + ":" + sym.Name
	if sym.ClassName != "" {
		id = sym.FilePath + ":" + sym.ClassName + "::" + sym.Name
	}
	return id
}

// normalizeSignature normalizes a signature for matching
func (idx *Indexer) normalizeSignature(sig string) string {
	// Remove whitespace and normalize
	sig = strings.ReplaceAll(sig, " ", "")
	sig = strings.ReplaceAll(sig, "\t", "")
	sig = strings.ToLower(sig)
	return sig
}

// globToRegex converts a glob pattern to regex
func (idx *Indexer) globToRegex(glob string) string {
	glob = regexp.QuoteMeta(glob)
	glob = strings.ReplaceAll(glob, "\\*\\*", ".*")
	glob = strings.ReplaceAll(glob, "\\*", "[^/]*")
	glob = strings.ReplaceAll(glob, "\\?", ".")
	return "^" + glob + "$"
}

// queryCacheKey generates a cache key for a query
func (idx *Indexer) queryCacheKey(query *SearchQuery) string {
	return query.Name + "|" + query.NamePattern + "|" + string(query.Type) + "|" +
		query.ClassName + "|" + query.FilePath + "|" + query.Language
}

// cacheResults stores results in the LRU cache
func (idx *Indexer) cacheResults(key string, results []*SearchResult) {
	// Evict if necessary
	for idx.searchCacheLRU.Len() >= idx.maxCacheSize {
		oldest := idx.searchCacheLRU.Back()
		if oldest != nil {
			oldKey := oldest.Value.(string)
			idx.searchCacheLRU.Remove(oldest)
			delete(idx.searchCacheMap, oldKey)
			delete(idx.searchCache, oldKey)
		}
	}

	// Add to cache
	idx.searchCache[key] = results
	elem := idx.searchCacheLRU.PushFront(key)
	idx.searchCacheMap[key] = elem
}

// touchCache moves an entry to front of LRU
func (idx *Indexer) touchCache(key string) {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	if elem, ok := idx.searchCacheMap[key]; ok {
		idx.searchCacheLRU.MoveToFront(elem)
	}
}

// invalidateCache clears the search cache
func (idx *Indexer) invalidateCache() {
	idx.searchCache = make(map[string][]*SearchResult)
	idx.searchCacheLRU = list.New()
	idx.searchCacheMap = make(map[string]*list.Element)
}

// FindDefinition finds the definition of a symbol referenced at a location
func (idx *Indexer) FindDefinition(filePath string, line int, name string) *Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	// Try exact match first
	if syms, ok := idx.symbolsByName[name]; ok {
		// Prefer definition in same file
		for _, sym := range syms {
			if sym.FilePath == filePath {
				return sym
			}
		}
		// Return first match
		if len(syms) > 0 {
			return syms[0]
		}
	}

	// Try as method
	if idx.methodIndex != nil {
		for _, methods := range idx.methodIndex {
			if sym, ok := methods[name]; ok {
				return sym
			}
		}
	}

	return nil
}

// FindUsages finds all usages of a symbol
func (idx *Indexer) FindUsages(symbolID string) []Reference {
	idx.mu.RLock()
	defer idx.mu.RUnlock()
	return idx.referenceIndex[symbolID]
}

// GetAllFunctions returns all indexed functions
func (idx *Indexer) GetAllFunctions() []*Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var functions []*Symbol
	for _, syms := range idx.functionIndex {
		functions = append(functions, syms...)
	}
	return functions
}

// GetAllClasses returns all indexed classes
func (idx *Indexer) GetAllClasses() []*Symbol {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	var classes []*Symbol
	for _, sym := range idx.classIndex {
		classes = append(classes, sym)
	}
	return classes
}

// Stats returns indexer statistics
func (idx *Indexer) Stats() map[string]int {
	idx.mu.RLock()
	defer idx.mu.RUnlock()

	return map[string]int{
		"total_symbols":    idx.totalSymbols,
		"total_references": idx.totalReferences,
		"unique_names":     len(idx.symbolsByName),
		"indexed_files":    len(idx.symbolsByFile),
		"functions":        len(idx.functionIndex),
		"classes":          len(idx.classIndex),
		"cache_size":       idx.searchCacheLRU.Len(),
	}
}

// Clear clears the entire index
func (idx *Indexer) Clear() {
	idx.mu.Lock()
	defer idx.mu.Unlock()

	idx.symbolsByID = make(map[string]*Symbol)
	idx.symbolsByName = make(map[string][]*Symbol)
	idx.symbolsByFile = make(map[string][]*Symbol)
	idx.functionIndex = make(map[string][]*Symbol)
	idx.classIndex = make(map[string]*Symbol)
	idx.methodIndex = make(map[string]map[string]*Symbol)
	idx.signatureIndex = make(map[string][]*Symbol)
	idx.referenceIndex = make(map[string][]Reference)
	idx.invalidateCache()
	idx.totalSymbols = 0
	idx.totalReferences = 0
}
