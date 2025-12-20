// Package symbolic provides symbolic execution for deep semantic tracing
package symbolic

import (
	"container/list"
	"context"
	"os"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/php"
)

// LRUFileCache provides memory-efficient file and AST caching with O(1) operations
// It uses lazy loading with LRU eviction to prevent unbounded memory growth
type LRUFileCache struct {
	maxEntries int
	maxMemory  int64 // Maximum memory in bytes
	currentMem int64 // Current memory usage

	entries   map[string]*list.Element
	evictList *list.List
	mu        sync.RWMutex
	parser    *sitter.Parser

	// Stats
	hits   int64
	misses int64
}

type fileCacheEntry struct {
	key      string
	root     *sitter.Node
	tree     *sitter.Tree // Keep reference to close properly
	content  []byte
	memory   int64
	refCount int32 // Track active references
}

// NewLRUFileCache creates a new file cache with specified limits
// MEMORY FIX: Reduced defaults for multi-threaded usage
func NewLRUFileCache(maxEntries int) *LRUFileCache {
	if maxEntries <= 0 {
		maxEntries = 30 // Default: keep max 30 files in memory (reduced from 100 for multi-thread)
	}
	parser := sitter.NewParser()
	parser.SetLanguage(php.GetLanguage())

	return &LRUFileCache{
		maxEntries: maxEntries,
		maxMemory:  64 * 1024 * 1024, // 64MB default max memory (reduced from 512MB for multi-thread)
		entries:    make(map[string]*list.Element, maxEntries),
		evictList:  list.New(),
		parser:     parser,
	}
}

// NewLRUFileCacheWithMemoryLimit creates a cache with custom memory limit
func NewLRUFileCacheWithMemoryLimit(maxEntries int, maxMemory int64) *LRUFileCache {
	cache := NewLRUFileCache(maxEntries)
	if maxMemory > 0 {
		cache.maxMemory = maxMemory
	}
	return cache
}

// Get retrieves or lazily loads a file's AST and content - O(1) for cached files
func (c *LRUFileCache) Get(filePath string) (*sitter.Node, []byte, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check cache first
	if elem, ok := c.entries[filePath]; ok {
		// Move to front (most recently used) - O(1)
		c.evictList.MoveToFront(elem)
		c.hits++
		entry := elem.Value.(*fileCacheEntry)
		return entry.root, entry.content, nil
	}

	c.misses++

	// Lazy load from disk
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, nil, err
	}

	tree, err := c.parser.ParseCtx(context.Background(), nil, content)
	if err != nil {
		return nil, nil, err
	}

	memUsage := estimateFileMemory(content)

	// Evict oldest entries if at capacity (by entries or memory)
	for len(c.entries) >= c.maxEntries || (c.currentMem+memUsage > c.maxMemory && c.evictList.Len() > 0) {
		c.evictOldest()
	}

	// Add new entry
	entry := &fileCacheEntry{
		key:     filePath,
		root:    tree.RootNode(),
		tree:    tree,
		content: content,
		memory:  memUsage,
	}
	elem := c.evictList.PushFront(entry)
	c.entries[filePath] = elem
	c.currentMem += memUsage

	return entry.root, entry.content, nil
}

// GetContent retrieves file content with lazy loading
func (c *LRUFileCache) GetContent(filePath string) ([]byte, error) {
	_, content, err := c.Get(filePath)
	return content, err
}

// GetParsedFile retrieves parsed AST with lazy loading
func (c *LRUFileCache) GetParsedFile(filePath string) (*sitter.Node, error) {
	root, _, err := c.Get(filePath)
	return root, err
}

// Has checks if a file is in the cache - O(1)
func (c *LRUFileCache) Has(filePath string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.entries[filePath]
	return exists
}

// evictOldest removes the least recently used entry - O(1)
func (c *LRUFileCache) evictOldest() {
	elem := c.evictList.Back()
	if elem == nil {
		return
	}
	entry := elem.Value.(*fileCacheEntry)

	// Close the tree to free AST memory
	if entry.tree != nil {
		entry.tree.Close()
	}

	c.evictList.Remove(elem)
	delete(c.entries, entry.key)
	c.currentMem -= entry.memory
}

// Remove removes a specific file from the cache - O(1)
func (c *LRUFileCache) Remove(filePath string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.entries[filePath]; exists {
		entry := elem.Value.(*fileCacheEntry)
		if entry.tree != nil {
			entry.tree.Close()
		}
		c.evictList.Remove(elem)
		delete(c.entries, filePath)
		c.currentMem -= entry.memory
	}
}

// Clear removes all entries from the cache
func (c *LRUFileCache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Close all trees
	for elem := c.evictList.Front(); elem != nil; elem = elem.Next() {
		entry := elem.Value.(*fileCacheEntry)
		if entry.tree != nil {
			entry.tree.Close()
		}
	}

	c.entries = make(map[string]*list.Element, c.maxEntries)
	c.evictList = list.New()
	c.currentMem = 0
}

// Size returns the current number of cached files
func (c *LRUFileCache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.entries)
}

// MemoryUsage returns current estimated memory usage in bytes
func (c *LRUFileCache) MemoryUsage() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMem
}

// Stats returns cache hit/miss statistics
func (c *LRUFileCache) Stats() (hits, misses int64, memUsage int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses, c.currentMem
}

// estimateFileMemory estimates memory usage for a cached file
func estimateFileMemory(content []byte) int64 {
	// Rough estimate: content + AST overhead (typically 5-8x content size for PHP)
	return int64(len(content)) * 7
}
