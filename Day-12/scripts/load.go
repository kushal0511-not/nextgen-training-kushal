package main

import (
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL     = "http://localhost:8080"
	numProducts = 100000
	numQueries  = 1000
	concurrency = 10
)

func main() {
	// 1. Seed Products
	fmt.Printf("Seeding %d products...\n", numProducts)
	seedProducts()

	// 2. Perform B-Tree Range Queries
	fmt.Printf("Performing %d B-Tree range queries...\n", numQueries)
	start := time.Now()
	runQueries(false)
	fmt.Printf("B-Tree Queries took: %v\n", time.Since(start))

	// 3. Perform Linear Scan Range Queries
	fmt.Printf("Performing %d Linear Scan range queries...\n", numQueries)
	start = time.Now()
	runQueries(true)
	fmt.Printf("Linear Scan Queries took: %v\n", time.Since(start))
}

func seedProducts() {
	resp, err := http.Post(fmt.Sprintf("%s/seed?count=%d", baseURL, numProducts), "application/json", nil)
	if err != nil {
		fmt.Printf("Error seeding products: %v\n", err)
		return
	}
	resp.Body.Close()
}

func runQueries(linear bool) {
	var wg sync.WaitGroup
	ch := make(chan int, numQueries)

	for i := 0; i < concurrency; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range ch {
				min := rand.Float64() * 500
				max := min + 50
				url := fmt.Sprintf("%s/products?min_price=%.2f&max_price=%.2f", baseURL, min, max)
				if linear {
					url += "&scan=linear"
				}
				resp, err := http.Get(url)
				if err != nil {
					return
				}
				resp.Body.Close()
			}
		}()
	}

	for i := 0; i < numQueries; i++ {
		ch <- i
	}
	close(ch)
	wg.Wait()
}
