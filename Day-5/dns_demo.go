package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz0123456789"

func randomString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

func main() {

	cache := NewCustomDNSCache()
	defer cache.Close()

	fmt.Println("=== High-Load DNS Cache Demo ===")
	fmt.Println("Worker Pool Size: 1000")
	fmt.Println("Simulating high-noise traffic (valid domains + random junk)...")

	const (
		numWorkers = 1000
		duration   = 10 * time.Second
	)

	var wg sync.WaitGroup
	taskChan := make(chan string, 10000)

	// 1. Start periodic status display
	stopDisplay := make(chan struct{})
	go func() {
		ticker := time.NewTicker(2000 * time.Microsecond)
		defer ticker.Stop()
		start := time.Now()
		for {
			select {
			case <-ticker.C:
				stats := cache.GetStats()
				fmt.Printf("\n[STATUS] Runtime: %v | Entries: %d | Hits: %d | Misses: %d | HitRate: %.2f%% | Mem: %s\n",
					time.Since(start).Truncate(time.Second),
					stats.TotalEntries, stats.Hits, stats.Misses, stats.HitRate*100, stats.MemoryEst)
			case <-stopDisplay:
				return
			}
		}
	}()

	// 2. Start Worker Pool
	wg.Add(numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(id int) {
			defer wg.Done()
			for domain := range taskChan {
				// Randomly decide to Add or Resolve
				if rand.Float32() < 0.2 { // 20% Writes
					ip := fmt.Sprintf("192.168.%d.%d", rand.Intn(255), rand.Intn(255))
					cache.AddRecord(domain, ip, 5*time.Second)
				} else {
					// 80% Resolves
					cache.Resolve(domain)
				}
			}
		}(i)
	}

	// 3. Feed tasks for a set duration
	start := time.Now()
	domains := []string{"google.com", "github.com", "example.org", "openai.com", "golang.org", "*.domain.com"}

	for time.Since(start) < duration {
		var d string
		randVal := rand.Float32()

		if randVal < 0.4 {
			// 40% known domains
			d = domains[rand.Intn(len(domains))]
		} else if randVal < 0.7 {
			// 30% wildcard subdomains
			d = fmt.Sprintf("sub-%d.domain.com", rand.Intn(100))
		} else {
			// 30% completely random junk strings (guaranteed misses)
			d = randomString(8) + ".io"
		}

		taskChan <- d
	}

	close(taskChan)
	wg.Wait()
	close(stopDisplay)

	fmt.Println("\n--- Demo Complete ---")
	finalStats := cache.GetStats()
	fmt.Printf("Final Records: %d\n", finalStats.TotalEntries)
}
