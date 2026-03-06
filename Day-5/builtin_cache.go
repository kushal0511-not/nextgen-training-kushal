package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type BuiltInDNSCache struct {
	storage  map[string]*DNSRecord
	mu       sync.RWMutex
	hits     int64
	misses   int64
	stopChan chan struct{}
}

func NewBuiltInDNSCache() *BuiltInDNSCache {
	c := &BuiltInDNSCache{
		storage:  make(map[string]*DNSRecord),
		stopChan: make(chan struct{}),
	}

	// Start background cleanup
	go c.backgroundCleanup()

	return c
}

func (c *BuiltInDNSCache) Resolve(domain string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 1. Exact match
	if record, ok := c.getRecord(domain); ok {
		c.hits++
		return record.IP, true
	}

	// 2. Wildcard
	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		wildcard := "*." + strings.Join(parts[i:], ".")
		if record, ok := c.getRecord(wildcard); ok {
			c.hits++
			return record.IP, true
		}
	}

	c.misses++
	return "", false
}

func (c *BuiltInDNSCache) getRecord(domain string) (*DNSRecord, bool) {
	record, ok := c.storage[domain]
	if !ok {
		return nil, false
	}

	if record.IsExpired() {
		delete(c.storage, domain)
		return nil, false
	}

	record.HitCount++
	return record, true
}

func (c *BuiltInDNSCache) AddRecord(domain, ip string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.storage[domain] = &DNSRecord{
		Domain:    domain,
		IP:        ip,
		TTL:       ttl,
		CreatedAt: time.Now(),
	}
}

func (c *BuiltInDNSCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRatio := 0.0
	if total > 0 {
		hitRatio = float64(c.hits) / float64(total)
	}

	return CacheStats{
		Hits:         c.hits,
		Misses:       c.misses,
		TotalEntries: len(c.storage),
		HitRate:      hitRatio,
		MemoryEst:    fmt.Sprintf("%d (exact count, built-in map memory is opaque)", len(c.storage)),
	}
}

func (c *BuiltInDNSCache) backgroundCleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			count := 0
			for key, record := range c.storage {
				if record.IsExpired() {
					delete(c.storage, key)
					count++
				}
			}
			if count > 0 {
				fmt.Printf("[Cleanup-BuiltIn] Removed %d expired records\n", count)
			}
			c.mu.Unlock()
		case <-c.stopChan:
			return
		}
	}
}

func (c *BuiltInDNSCache) Close() {
	close(c.stopChan)
}
