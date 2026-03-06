package main

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

type CustomDNSCache struct {
	mapStorage *HashMap
	mu         sync.RWMutex
	hits       int64
	misses     int64
	stopChan   chan struct{}
}

func NewCustomDNSCache() *CustomDNSCache {
	c := &CustomDNSCache{
		mapStorage: NewHashMap(),
		stopChan:   make(chan struct{}),
	}

	// Start background cleanup ticker
	go c.backgroundCleanup()

	return c
}

func (c *CustomDNSCache) Resolve(domain string) (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Priority Rule: Exact Match > Wildcard Match
	// 1. Try exact match
	if record, ok := c.getRecord(domain); ok {
		c.hits++
		return record.IP, true
	}

	// 2. Try wildcard matches (lazy match)
	// Example: sub.example.com -> *.example.com, *.com
	parts := strings.Split(domain, ".")
	for i := 1; i < len(parts); i++ {
		wildcard := "*." + strings.Join(parts[i:], ".")
		if record, ok := c.getRecord(wildcard); ok {
			c.hits++
			return record.IP, true
		}
	}

	// 3. Simulation of Upstream Lookup on Miss
	c.misses++
	// In reality, this would query another DNS or root server
	return "", false
}

// Internal helper for cleanups and lookups with lazy expiration
func (c *CustomDNSCache) getRecord(domain string) (*DNSRecord, bool) {
	val, ok := c.mapStorage.Get(domain)
	if !ok {
		return nil, false
	}

	record := val.(*DNSRecord)
	if record.IsExpired() {
		// Lazy expiration
		c.mapStorage.Delete(domain)
		return nil, false
	}

	record.HitCount++
	return record, true
}

func (c *CustomDNSCache) AddRecord(domain, ip string, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	record := &DNSRecord{
		Domain:    domain,
		IP:        ip,
		TTL:       ttl,
		CreatedAt: time.Now(),
	}

	c.mapStorage.Put(domain, record)
}

func (c *CustomDNSCache) GetStats() CacheStats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	total := c.hits + c.misses
	hitRatio := 0.0
	if total > 0 {
		hitRatio = float64(c.hits) / float64(total)
	}

	// Estimate Memory:
	// NumBuckets * PointerSize (8) + NumEntries * NodeSize (approx 64)
	numEntries := c.mapStorage.Size()
	numBuckets := len(c.mapStorage.buckets)
	estBytes := (numBuckets * 8) + (numEntries * 64)

	return CacheStats{
		Hits:         c.hits,
		Misses:       c.misses,
		TotalEntries: numEntries,
		HitRate:      hitRatio,
		MemoryEst:    fmt.Sprintf("%.2f KB", float64(estBytes)/1024),
	}
}

func (c *CustomDNSCache) backgroundCleanup() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.mu.Lock()
			count := 0
			// Iterate through all buckets to find expired records
			c.mapStorage.Iterate(func(key string, value interface{}) {
				record := value.(*DNSRecord)
				if record.IsExpired() {
					c.mapStorage.Delete(key)
					count++
				}
			})
			if count > 0 {
				fmt.Printf("[Cleanup] Removed %d expired DNS records\n", count)
			}
			c.mu.Unlock()
		case <-c.stopChan:
			return
		}
	}
}

func (c *CustomDNSCache) Close() {
	close(c.stopChan)
}
