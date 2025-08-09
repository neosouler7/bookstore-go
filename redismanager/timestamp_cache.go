package redismanager

import (
	"container/list"
	"sync"
	"time"
)

// TimestampCache LRU 캐시로 타임스탬프를 관리
type TimestampCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mutex    sync.RWMutex
}

// cacheEntry 캐시에 저장될 항목
type cacheEntry struct {
	key       string
	timestamp string
	lastUsed  time.Time
}

// NewTimestampCache 새로운 타임스탬프 캐시 생성
func NewTimestampCache(capacity int) *TimestampCache {
	return &TimestampCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Store 키와 타임스탬프를 캐시에 저장
func (tc *TimestampCache) Store(key, timestamp string) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	// 기존 키가 있으면 제거
	if element, exists := tc.cache[key]; exists {
		tc.list.Remove(element)
		delete(tc.cache, key)
	}

	// 새 항목 생성 및 저장
	entry := &cacheEntry{
		key:       key,
		timestamp: timestamp,
		lastUsed:  time.Now(),
	}
	element := tc.list.PushFront(entry)
	tc.cache[key] = element

	// 용량 초과 시 LRU 정리
	if tc.list.Len() > tc.capacity {
		tc.evictLRU()
	}
}

// Load 키에 해당하는 타임스탬프 조회
func (tc *TimestampCache) Load(key string) (string, bool) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if element, exists := tc.cache[key]; exists {
		entry := element.Value.(*cacheEntry)
		entry.lastUsed = time.Now()

		// 사용된 항목을 맨 앞으로 이동
		tc.list.MoveToFront(element)

		return entry.timestamp, true
	}
	return "", false
}

// evictLRU 가장 오래 사용되지 않은 항목 제거
func (tc *TimestampCache) evictLRU() {
	if tc.list.Len() == 0 {
		return
	}

	// 맨 뒤의 항목 제거 (LRU)
	element := tc.list.Back()
	entry := element.Value.(*cacheEntry)

	tc.list.Remove(element)
	delete(tc.cache, entry.key)
}

// Len 현재 캐시 크기 반환
func (tc *TimestampCache) Len() int {
	tc.mutex.RLock()
	defer tc.mutex.RUnlock()
	return tc.list.Len()
}

// Cleanup 오래된 항목들 정리 (선택적)
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
