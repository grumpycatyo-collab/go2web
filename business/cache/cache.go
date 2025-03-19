package cache

import (
	"sync"
	"time"
)

type CacheItem struct {
	Response   interface{}
	Expiration time.Time
}

type Cache struct {
	items map[string]CacheItem
	mutex sync.RWMutex
}

func NewCache() *Cache {
	cache := &Cache{
		items: make(map[string]CacheItem),
	}

	go func() {
		for {
			time.Sleep(5 * time.Minute)
			cache.cleanup()
		}
	}()

	return cache
}

func (c *Cache) Get(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, found := c.items[key]
	if !found {
		return nil
	}

	if time.Now().After(item.Expiration) {
		delete(c.items, key)
		return nil
	}

	return item.Response
}

func (c *Cache) Set(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	expiration := time.Now().Add(5 * time.Minute)
	c.items[key] = CacheItem{
		Response:   value,
		Expiration: expiration,
	}
}

func (c *Cache) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, item := range c.items {
		if now.After(item.Expiration) {
			delete(c.items, key)
		}
	}
}
