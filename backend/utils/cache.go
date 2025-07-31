package utils

import (
	"container/list"
	"sync"
	"time"
)

type LRUCache struct {
	capacity int
	ttl      time.Duration
	cache    map[string]*list.Element
	list     *list.List
	mutex    sync.RWMutex
	stopChan chan bool
}

type cacheEntry struct {
	key       string
	value     interface{}
	timestamp time.Time
}

func NewLRUCache(capacity int, ttl time.Duration) *LRUCache {
	return &LRUCache{
		capacity: capacity,
		ttl:      ttl,
		cache:    make(map[string]*list.Element, capacity),
		list:     list.New(),
		stopChan: make(chan bool),
	}
}

func (c *LRUCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	if element, exists := c.cache[key]; exists {
		entry := element.Value.(*cacheEntry)
		
		// Check if entry has expired
		if time.Since(entry.timestamp) > c.ttl {
			// Remove expired entry
			c.mutex.RUnlock()
			c.mutex.Lock()
			c.removeElement(element)
			c.mutex.Unlock()
			c.mutex.RLock()
			return nil, false
		}

		// Move to front (most recently used)
		c.list.MoveToFront(element)
		return entry.value, true
	}

	return nil, false
}

func (c *LRUCache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// Check if key already exists
	if element, exists := c.cache[key]; exists {
		// Update existing entry
		entry := element.Value.(*cacheEntry)
		entry.value = value
		entry.timestamp = time.Now()
		c.list.MoveToFront(element)
		return
	}

	// Create new entry
	entry := &cacheEntry{
		key:       key,
		value:     value,
		timestamp: time.Now(),
	}

	element := c.list.PushFront(entry)
	c.cache[key] = element

	// Remove least recently used if capacity exceeded
	if c.list.Len() > c.capacity {
		c.removeLRU()
	}
}

func (c *LRUCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if element, exists := c.cache[key]; exists {
		c.removeElement(element)
	}
}

func (c *LRUCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.cache = make(map[string]*list.Element, c.capacity)
	c.list.Init()
}

func (c *LRUCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.list.Len()
}

func (c *LRUCache) removeElement(element *list.Element) {
	c.list.Remove(element)
	entry := element.Value.(*cacheEntry)
	delete(c.cache, entry.key)
}

func (c *LRUCache) removeLRU() {
	element := c.list.Back()
	if element != nil {
		c.removeElement(element)
	}
}

func (c *LRUCache) StartCleanup(interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				c.cleanup()
			case <-c.stopChan:
				return
			}
		}
	}()
}

func (c *LRUCache) StopCleanup() {
	close(c.stopChan)
}

func (c *LRUCache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	element := c.list.Back()

	// Remove expired entries from the back of the list
	for element != nil {
		entry := element.Value.(*cacheEntry)
		if now.Sub(entry.timestamp) > c.ttl {
			prev := element.Prev()
			c.removeElement(element)
			element = prev
		} else {
			break // All remaining entries are not expired
		}
	}
}

// GetStats returns cache statistics for monitoring
func (c *LRUCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return map[string]interface{}{
		"size":     c.list.Len(),
		"capacity": c.capacity,
		"ttl":      c.ttl.String(),
	}
}