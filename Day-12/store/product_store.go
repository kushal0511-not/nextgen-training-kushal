package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/nextgen-training-kushal/Day-12/models"
	"github.com/rs/xid"
)

type ProductStore struct {
	mu       sync.RWMutex
	products map[string]*models.Product
	priceIdx *ProductPriceBTree
}

func NewProductStore(degree int) *ProductStore {
	return NewProductStoreWithCapacity(degree, 0)
}

func NewProductStoreWithCapacity(degree int, capacity int) *ProductStore {
	return &ProductStore{
		products: make(map[string]*models.Product, capacity),
		priceIdx: NewProductPriceBTree(degree),
	}
}

func (s *ProductStore) AddProduct(p *models.Product) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p.ID == "" {
		p.ID = xid.New().String()
	}
	p.CreatedAt = time.Now()
	s.products[p.ID] = p
	s.priceIdx.Insert(p.Price, p.ID)
}

func (s *ProductStore) GetProduct(id string) (*models.Product, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.products[id]
	return p, ok
}

func (s *ProductStore) DeleteProduct(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if p, ok := s.products[id]; ok {
		s.priceIdx.Delete(p.Price, id)
		delete(s.products, id)
	}
}

func (s *ProductStore) QueryByPriceRange(min, max float64) []*models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.priceIdx.RangeQuery(min, max)
	var results []*models.Product
	for _, id := range ids {
		if p, ok := s.products[id]; ok {
			results = append(results, p)
		}
	}
	return results
}

func (s *ProductStore) QueryByPriceRangeLinear(min, max float64) []*models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var results []*models.Product
	for _, p := range s.products {
		if p.Price >= min && p.Price <= max {
			results = append(results, p)
		}
	}
	return results
}

func (s *ProductStore) UpdateProduct(id string, updated *models.Product) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	old, ok := s.products[id]
	if !ok {
		return false
	}
	updated.ID = id
	if old.Price != updated.Price {
		s.priceIdx.Delete(old.Price, id)
		s.priceIdx.Insert(updated.Price, id)
	}
	s.products[id] = updated
	return true
}

func (s *ProductStore) GetAllPaginated(page, size int) []*models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if page < 1 {
		page = 1
	}
	if size < 1 {
		size = 20
	}
	offset := (page - 1) * size
	ids := s.priceIdx.InOrderTraversalPaginated(offset, size)

	var results []*models.Product
	for _, id := range ids {
		if p, ok := s.products[id]; ok {
			results = append(results, p)
		}
	}
	return results
}

func (s *ProductStore) GetByCategorySorted(category string) []*models.Product {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ids := s.priceIdx.InOrderTraversal()
	var results []*models.Product
	for _, id := range ids {
		if p, ok := s.products[id]; ok && p.Category == category {
			results = append(results, p)
		}
	}
	return results
}

type Stats struct {
	CategoryCounts    map[string]int
	PriceDistribution map[string]int // e.g., "0-50", "50-100", "100+"
	AverageRating     float64
}

func (s *ProductStore) GetStats() Stats {
	s.mu.RLock()
	defer s.mu.RUnlock()
	stats := Stats{
		CategoryCounts:    make(map[string]int),
		PriceDistribution: make(map[string]int),
		AverageRating:     0,
	}
	totalRating := 0.0
	for _, p := range s.products {
		stats.CategoryCounts[p.Category]++
		totalRating += p.Rating
		if p.Price <= 50 {
			stats.PriceDistribution["0-50"]++
		} else if p.Price <= 100 {
			stats.PriceDistribution["51-100"]++
		} else {
			stats.PriceDistribution["100+"]++
		}
	}
	if len(s.products) > 0 {
		stats.AverageRating = totalRating / float64(len(s.products))
	}
	return stats
}

func (s *ProductStore) PrintAllByPrice() {
	ids := s.priceIdx.InOrderTraversal()
	for _, id := range ids {
		p := s.products[id]
		fmt.Printf("ID: %s, Name: %s, Price: %.2f\n", p.ID, p.Name, p.Price)
	}
}
