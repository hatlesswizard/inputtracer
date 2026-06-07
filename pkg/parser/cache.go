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

// estimateMemory estimates memory usage of a cached entry.
// The multiplier of 6 is a conservative mid-point of the observed 5-10× ratio
// between raw source bytes and the memory occupied by the corresponding
// Tree-Sitter AST (each node is ~48 bytes; a typical file has ~1 node per
// 8-10 source bytes, yielding roughly 5-6× overhead). Using 6× keeps the
// cache memory budget accurate without requiring a full AST walk.
func (cp *CachedParse) estimateMemory() int64 {
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
// Uses an optimistic read → upgrade pattern: check existence under RLock, then
// re-acquire a write lock only when a hit is confirmed (MoveToFront mutates the list).
func (c *Cache) Get(key string) *CachedParse {
	// Fast path: shared read lock for the existence check.
	c.mu.RLock()
	_, exists := c.items[key]
	c.mu.RUnlock()

	if !exists {
		atomic.AddInt64(&c.misses, 1)
		return nil
	}

	// Slow path: we have a hit, but MoveToFront mutates the list so we need an
	// exclusive lock. Re-check that the key still exists after the lock upgrade
	// to guard against a concurrent eviction between RUnlock and Lock (TOCTOU).
	c.mu.Lock()
	elem, exists := c.items[key]
	if !exists {
		// Evicted between the two critical sections – treat as a miss.
		c.mu.Unlock()
		atomic.AddInt64(&c.misses, 1)
		return nil
	}
	c.evictList.MoveToFront(elem)
	data := elem.Value.(*cacheEntry).data
	c.mu.Unlock()

	atomic.AddInt64(&c.hits, 1)
	return data
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

// evictOldest removes the least recently used entry - O(1).
//
// It drops the cache's reference to the entry but does NOT close the tree:
// a previously returned ParseResult may still hold Root nodes that point into
// it, and closing here would free that C memory out from under an in-flight
// traversal (use-after-free -> SIGSEGV). The go-tree-sitter finalizer frees the
// tree's C memory once it becomes unreachable, bounding memory safely.
func (c *Cache) evictOldest() {
	elem := c.evictList.Back()
	if elem == nil {
		return
	}
	entry := elem.Value.(*cacheEntry)
	c.evictList.Remove(elem)
	delete(c.items, entry.key)
	c.currentMem -= entry.memory
}

// Remove removes an entry from the cache - O(1).
//
// As with evictOldest, the tree is not closed here: outstanding ParseResults
// may still reference its nodes. The finalizer reclaims it once unreachable.
func (c *Cache) Remove(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if elem, exists := c.items[key]; exists {
		entry := elem.Value.(*cacheEntry)
		c.evictList.Remove(elem)
		delete(c.items, key)
		c.currentMem -= entry.memory
	}
}

// Clear clears all entries from the cache - O(n) but infrequent.
//
// Trees are not closed here: outstanding ParseResults may still reference their
// nodes. Dropping the cache's references lets the finalizer reclaim each tree
// once it is no longer used anywhere.
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()

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
