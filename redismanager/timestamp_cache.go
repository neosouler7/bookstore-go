package redismanager

import (
	"container/list"
	"sync"
	"time"
)

// TimestampCache manages timestamps with an LRU cache
type TimestampCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mutex    sync.RWMutex
}

// cacheEntry is an item stored in the cache
type cacheEntry struct {
	key       string
	timestamp string
	lastUsed  time.Time
}

// NewTimestampCache creates a new timestamp cache
func NewTimestampCache(capacity int) *TimestampCache {
	return &TimestampCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Store saves a key-timestamp pair to the cache
func (tc *TimestampCache) Store(key, timestamp string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	// remove existing entry if present
	if element, exists := tc.cache[key]; exists {
		tc.list.Remove(element)
		delete(tc.cache, key)
	}

	// create and store new entry
	entry := &cacheEntry{
		key:       key,
		timestamp: timestamp,
		lastUsed:  time.Now(),
	}
	element := tc.list.PushFront(entry)
	tc.cache[key] = element

	// evict LRU entry if over capacity
	if tc.list.Len() > tc.capacity {
		tc.evictLRU()
	}
}

// Load retrieves the timestamp for a given key
func (tc *TimestampCache) Load(key string) (string, bool) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if element, exists := tc.cache[key]; exists {
		entry := element.Value.(*cacheEntry)
		entry.lastUsed = time.Now()

		// move accessed entry to front
		tc.list.MoveToFront(element)

		return entry.timestamp, true
	}
	return "", false
}

// evictLRU removes the least recently used entry
func (tc *TimestampCache) evictLRU() {
	if tc.list.Len() == 0 {
		return
	}

	// remove the back entry (LRU)
	element := tc.list.Back()
	entry := element.Value.(*cacheEntry)

	tc.list.Remove(element)
	delete(tc.cache, entry.key)
}

// Len returns the current cache size
func (tc *TimestampCache) Len() int {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	return tc.list.Len()
}

// Cleanup removes stale entries (optional)
func (tc *TimestampCache) Cleanup(maxAge time.Duration) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	now := time.Now()
	for element := tc.list.Back(); element != nil; {
		entry := element.Value.(*cacheEntry)
		if now.Sub(entry.lastUsed) > maxAge {
			next := element.Prev()
			tc.list.Remove(element)
			delete(tc.cache, entry.key)
			element = next
		} else {
			break
		}
	}
}
