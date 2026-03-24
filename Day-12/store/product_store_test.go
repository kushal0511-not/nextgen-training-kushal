package store

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/nextgen-training-kushal/Day-12/models"
)

func generateProducts(n int) []models.Product {
	var products []models.Product
	for i := 0; i < n; i++ {
		price := float64(rand.Intn(10000)) + rand.Float64()
		p := models.Product{
			ID:    fmt.Sprintf("prod-%d", i),
			Name:  fmt.Sprintf("Product %d", i),
			Price: price,
		}
		products = append(products, p)
	}
	return products
}

func BenchmarkRangeQueryBTree(b *testing.B) {
	b.StopTimer()
	store := NewProductStore(50) // Adjust B-Tree degree if needed
	products := generateProducts(100000)
	for _, p := range products {
		store.AddProduct(&p)
	}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		// Random range
		min := rand.Float64() * 9500
		max := min + 500
		_ = store.priceIdx.RangeQuery(min, max)
	}
}

func BenchmarkRangeQueryLinearScan(b *testing.B) {
	b.StopTimer()
	products := generateProducts(100000)
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		var results []string
		min := rand.Float64() * 9500
		max := min + 500
		for _, p := range products {
			if p.Price >= min && p.Price <= max {
				results = append(results, p.ID)
			}
		}
		_ = results
	}
}
