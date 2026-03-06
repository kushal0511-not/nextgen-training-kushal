package main

import (
	"fmt"
	"testing"
	"time"
)

// --- AddRecord (Put) Benchmarks ---

func BenchmarkCustomCache_AddRecord(b *testing.B) {
	cache := NewCustomDNSCache()
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("domain-%d.com", i%1000)
		cache.AddRecord(key, "1.2.3.4", 10*time.Minute)
	}
}

func BenchmarkBuiltInCache_AddRecord(b *testing.B) {
	cache := NewBuiltInDNSCache()
	defer cache.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("domain-%d.com", i%1000)
		cache.AddRecord(key, "1.2.3.4", 10*time.Minute)
	}
}

// --- Resolve (Get) Benchmarks ---

func BenchmarkCustomCache_Resolve(b *testing.B) {
	cache := NewCustomDNSCache()
	defer cache.Close()

	// Pre-fill
	for i := 0; i < 1000; i++ {
		cache.AddRecord(fmt.Sprintf("domain-%d.com", i), "1.2.3.4", 10*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("domain-%d.com", i%1000)
		_, _ = cache.Resolve(key)
	}
}

func BenchmarkBuiltInCache_Resolve(b *testing.B) {
	cache := NewBuiltInDNSCache()
	defer cache.Close()

	// Pre-fill
	for i := 0; i < 1000; i++ {
		cache.AddRecord(fmt.Sprintf("domain-%d.com", i), "1.2.3.4", 10*time.Minute)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		key := fmt.Sprintf("domain-%d.com", i%1000)
		_, _ = cache.Resolve(key)
	}
}
