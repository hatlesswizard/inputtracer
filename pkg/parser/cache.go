package parser

import (
	"container/list"
	"sync"
	"sync/atomic"

	sitter "github.com/smacker/go-tree-sitter"
)

// CachedParse represents a cached parse result
// MEMORY FIX: Now stores the Tree reference to properly close it on eviction
type CachedParse struct {
	Root   *sitter.Node
	Tree   *sitter.Tree // Keep tree reference to close on eviction
	Source []byte
}

// estimateMemory estimates memory usage of a cached entry
func (cp *CachedParse) estimateMemory() int64 {
	// Rough estimate: source bytes + AST overhead (typically 5-10x source size)
	return int64(len(cp.Source)) * 6
}

// Cache is an LRU cache with O(1) operations and memory limits
type Cache struct {
	maxEntries int
	maxMemory  int64 // Maximum memory in bytes
	currentMem int64 // Current memory usage

	items     map[string]*list.Element
	evictList *list.List
	mu        sync.RWMutex

	hits   int64
	misses int64
}

type cacheEntry struct {
	key    string
	data   *CachedParse
	memory int64
}

// NewCache creates a new cache with entry and memory limits
// MEMORY FIX: Reduced default memory limit from 256MB to 32MB for multi-threaded usage
func NewCache(maxEntries int) *Cache {
	return NewCacheWithMemoryLimit(maxEntries, 32*1024*1024) // 32MB default - reduced for 100-thread usage
}

// NewCacheWithMemoryLimit creates a cache with custom memory limit
// MEMORY FIX: Reduced default max entries from 1000 to 100 for multi-threaded usage
func NewCacheWithMemoryLimit(maxEntries int, maxMemory int64) *Cache {
	if maxEntries <= 0 {
		maxEntries = 100 // Reduced from 1000 for multi-threaded usage
	}
	if maxMemory <= 0 {
		maxMemory = 32 * 1024 * 1024 // 32MB default - reduced for multi-threaded usage
	}
	return &Cache{
		maxEntries: maxEntries,
		maxMemory:  maxMemory,
		items:      make(map[string]*list.Element, maxEntries),
		evictList:  list.New(),
	}
}

// Get retrieves a cached parse result - O(1)
func (c *Cache) Get(key string) *CachedParse {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		// Move to front - O(1)
		c.evictList.MoveToFront(elem)
		atomic.AddInt64(&c.hits, 1)
		return elem.Value.(*cacheEntry).data
	}

	atomic.AddInt64(&c.misses, 1)
	return nil
}

// Put adds or updates a cached parse result - O(1)
func (c *Cache) Put(key string, data *CachedParse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	memUsage := data.estimateMemory()

	// Update existing entry
	if elem, exists := c.items[key]; exists {
		oldEntry := elem.Value.(*cacheEntry)
		c.currentMem -= oldEntry.memory
		c.currentMem += memUsage
		oldEntry.data = data
		oldEntry.memory = memUsage
		c.evictList.MoveToFront(elem)
		return
	}

	// Evict until we have space (by entries or memory)
	for len(c.items) >= c.maxEntries || c.currentMem+memUsage > c.maxMemory {
		if c.evictList.Len() == 0 {
			break
		}
		c.evictOldest()
	}

	// Add new entry
	entry := &cacheEntry{
		key:    key,
		data:   data,
		memory: memUsage,
	}
	elem := c.evictList.PushFront(entry)
	c.items[key] = elem
	c.currentMem += memUsage
}

// evictOldest removes the least recently used entry - O(1)
// MEMORY FIX: Now properly closes the tree to release AST memory
func (c *Cache) evictOldest() {
	elem := c.evictList.Back()
	if elem == nil {
		return
	}
	entry := elem.Value.(*cacheEntry)
	// CRITICAL: Close the tree to release AST memory
	if entry.data != nil && entry.data.Tree != nil {
		entry.data.Tree.Close()
	}
	c.evictList.Remove(elem)
	delete(c.items, entry.key)
	c.currentMem -= entry.memory
}

// Remove removes an entry from the cache - O(1)
// MEMORY FIX: Now properly closes the tree to release AST memory
func (c *Cache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		entry := elem.Value.(*cacheEntry)
		// CRITICAL: Close the tree to release AST memory
		if entry.data != nil && entry.data.Tree != nil {
			entry.data.Tree.Close()
		}
		c.evictList.Remove(elem)
		delete(c.items, key)
		c.currentMem -= entry.memory
	}
}

// Clear clears all entries from the cache - O(n) but infrequent
// MEMORY FIX: Now properly closes all trees to release AST memory
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

	// CRITICAL: Close all trees before clearing
	for elem := c.evictList.Front(); elem != nil; elem = elem.Next() {
		entry := elem.Value.(*cacheEntry)
		if entry.data != nil && entry.data.Tree != nil {
			entry.data.Tree.Close()
		}
	}

	c.items = make(map[string]*list.Element, c.maxEntries)
	c.evictList = list.New()
	c.currentMem = 0
}

// Size returns the current number of cached items
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// MemoryUsage returns current memory usage estimate
func (c *Cache) MemoryUsage() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.currentMem
}

// Stats returns cache statistics
func (c *Cache) Stats() (hits, misses int64) {
	return atomic.LoadInt64(&c.hits), atomic.LoadInt64(&c.misses)
}

// StatsWithMemory returns cache statistics including memory usage
func (c *Cache) StatsWithMemory() (hits, misses, memUsage int64) {
	c.mu.RLock()
	mem := c.currentMem
	c.mu.RUnlock()
	return atomic.LoadInt64(&c.hits), atomic.LoadInt64(&c.misses), mem
}
