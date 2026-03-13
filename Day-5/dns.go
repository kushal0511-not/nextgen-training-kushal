package main

import (
	"time"
)

// DNSRecord represents a single DNS entry in our cache.
type DNSRecord struct {
	Domain    string        `json:"domain"`
	IP        string        `json:"ip"`
	TTL       time.Duration `json:"ttl"`
	CreatedAt time.Time     `json:"created_at"`
	HitCount  int           `json:"hit_count"`
}

// IsExpired checks if the record has lived past its TTL.
func (r *DNSRecord) IsExpired() bool {
	return time.Since(r.CreatedAt) >= r.TTL
}

// CacheStats provides visibility into the cache performance.
type CacheStats struct {
	Hits         int64   `json:"hits"`
	Misses       int64   `json:"misses"`
	TotalEntries int     `json:"total_entries"`
	HitRate      float64 `json:"hit_rate"`
	MemoryEst    string  `json:"memory_estimate"`
}

// DNSCache defines the required behavior for our DNS systems.
type DNSCache interface {
	Resolve(domain string) (string, bool)
	AddRecord(domain, ip string, ttl time.Duration)
	GetStats() CacheStats
	Close() // To stop the background cleaner
}
