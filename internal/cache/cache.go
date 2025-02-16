package cache

import (
	"sync"
	"time"
)

// CacheItem хранит значение и время срока годности.
type CacheItem struct {
	Value     interface{}
	ExpiresAt time.Time
}

// Cache - простенький in-memory кэш.
type Cache struct {
	data map[string]CacheItem
	mu   sync.RWMutex
}

// NewCache создает новый кэш.
func NewCache() *Cache {
	return &Cache{
		data: make(map[string]CacheItem),
	}
}

// cleanupExpiredItems периодически очищает элементы, срок действия которых истек.
func (c *Cache) СleanupExpiredItems() {
	ticker := time.NewTicker(5 * time.Minute) // Каждые 5 минут чекаем кэш
	for {
		<-ticker.C
		c.mu.Lock()
		for key, item := range c.data {
			if time.Now().After(item.ExpiresAt) {
				delete(c.data, key)
			}
		}
		c.mu.Unlock()
	}
}

// Set добавляет элемент в кэш с заданным TTL.
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.data[key] = CacheItem{
		Value:     value,
		ExpiresAt: time.Now().Add(ttl),
	}
}

// Get возвращает элемент из кэша, если он существует и не истек.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	item, exists := c.data[key]
	if !exists || time.Now().After(item.ExpiresAt) {
		return nil, false
	}
	return item.Value, true
}

// Delete удаляет элемент из кэша.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.data, key)
}

