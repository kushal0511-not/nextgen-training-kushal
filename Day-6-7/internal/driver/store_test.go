package driver

import (
	"fmt"
	"math/rand"
	"ride-sharing/internal/models"
	"testing"
)

func TestMemoryStore(t *testing.T) {
	s := NewMemoryStore()
	d1 := &models.Driver{ID: "D1", Name: "Alice", Location: models.Location{Lat: 10, Lng: 10}, Status: models.DriverStatusAvailable}
	d2 := &models.Driver{ID: "D2", Name: "Bob", Location: models.Location{Lat: 10.01, Lng: 10.01}, Status: models.DriverStatusAvailable}

	s.Register(d1)
	s.Register(d2)

	// Test GetByID
	res, _ := s.GetByID("D1")
	if res.Name != "Alice" {
		t.Errorf("Expected Alice, got %s", res.Name)
	}

	// Test FindNearest
	nearest, _ := s.FindNearest(models.Location{Lat: 10, Lng: 10}, 0.02)
	if len(nearest) != 2 {
		t.Errorf("Expected 2 nearest, got %d", len(nearest))
	}

	// Test UpdateLocation and zone change
	s.UpdateLocation("D1", models.Location{Lat: 20, Lng: 20})
	res, _ = s.GetByID("D1")
	if res.Location.Lat != 20 {
		t.Errorf("Expected Lat 20, got %.2f", res.Location.Lat)
	}

	nearest, _ = s.FindNearest(models.Location{Lat: 10, Lng: 10}, 0.05)
	if len(nearest) != 1 {
		t.Errorf("Expected 1 nearest (Bob), got %d", len(nearest))
	}
}

func BenchmarkRegister(b *testing.B) {
	store := NewMemoryStore()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		d := &models.Driver{
			ID:       fmt.Sprintf("D-%d", i),
			Location: models.Location{Lat: rand.Float64() * 100, Lng: rand.Float64() * 100},
			Status:   models.DriverStatusAvailable,
		}
		store.Register(d)
	}
}

func BenchmarkFindNearest(b *testing.B) {
	numDrivers := []int{100, 1000, 10000}
	for _, n := range numDrivers {
		b.Run(fmt.Sprintf("Drivers-%d", n), func(b *testing.B) {
			store := NewMemoryStore()
			for i := 0; i < n; i++ {
				d := &models.Driver{
					ID:       fmt.Sprintf("D-%d", i),
					Location: models.Location{Lat: rand.Float64() * 100, Lng: rand.Float64() * 100},
					Status:   models.DriverStatusAvailable,
				}
				store.Register(d)
			}

			loc := models.Location{Lat: 50, Lng: 50}
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				_, _ = store.FindNearest(loc, 5.0)
			}
		})
	}
}

func BenchmarkUpdateLocation(b *testing.B) {
	store := NewMemoryStore()
	d := &models.Driver{
		ID:       "D-1",
		Location: models.Location{Lat: 10, Lng: 10},
		Status:   models.DriverStatusAvailable,
	}
	store.Register(d)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		newLoc := models.Location{Lat: rand.Float64() * 100, Lng: rand.Float64() * 100}
		store.UpdateLocation("D-1", newLoc)
	}
}
