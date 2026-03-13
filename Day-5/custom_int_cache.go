package main

import "sync"

type CustomIntCache struct {
	mapStorage *HashMap[int]
	mu         sync.RWMutex
	hits       int64
	misses     int64
}

func NewCustomIntCache() *CustomIntCache {
	c := &CustomIntCache{
		mapStorage: NewHashMap[int](),
	}

	return c
}

func (c *CustomIntCache) Get(key int) (int, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	val, ok := c.mapStorage.Get(key)
	if !ok {
		return 0, false
	}
	return val.(int), true
}

func (c *CustomIntCache) Put(key int, value int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.mapStorage.Put(key, value)
}
