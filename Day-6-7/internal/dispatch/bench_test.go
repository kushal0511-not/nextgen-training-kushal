package dispatch

import (
	"fmt"
	"math/rand"
	"ride-sharing/internal/driver"
	"ride-sharing/internal/models"
	"ride-sharing/internal/queue"
	"ride-sharing/internal/rides"
	"testing"
)

func BenchmarkRequestRide(b *testing.B) {
	ds := driver.NewMemoryStore()
	rq := queue.NewMinHeap()
	rt := rides.NewLinkedTracker()
	dp := NewDispatcher(ds, rq, rt)
	rider := &models.Rider{ID: "U1"}
	pickup := models.Location{Lat: 0, Lng: 0}
	dropoff := models.Location{Lat: 1, Lng: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		dp.RequestRide(rider, pickup, dropoff)
	}
}

func BenchmarkProcessQueue(b *testing.B) {
	numDrivers := []int{100, 1000, 10000}
	for _, n := range numDrivers {
		b.Run(fmt.Sprintf("Drivers-%d", n), func(b *testing.B) {
			ds := driver.NewMemoryStore()
			rq := queue.NewMinHeap()
			rt := rides.NewLinkedTracker()
			dp := NewDispatcher(ds, rq, rt)

			// Pre-populate drivers
			for i := 0; i < n; i++ {
				ds.Register(&models.Driver{
					ID:       fmt.Sprintf("D-%d", i),
					Location: models.Location{Lat: rand.Float64() * 100, Lng: rand.Float64() * 100},
					Status:   models.DriverStatusAvailable,
				})
			}

			rider := &models.Rider{ID: "U1"}
			pickup := models.Location{Lat: 50, Lng: 50}
			dropoff := models.Location{Lat: 51, Lng: 51}

			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				dp.RequestRide(rider, pickup, dropoff)
				b.StartTimer()
				dp.ProcessQueue()
			}
		})
	}
}

func BenchmarkCompleteRide(b *testing.B) {
	ds := driver.NewMemoryStore()
	rq := queue.NewMinHeap()
	rt := rides.NewLinkedTracker()
	dp := NewDispatcher(ds, rq, rt)

	ds.Register(&models.Driver{
		ID:       "D1",
		Location: models.Location{Lat: 0, Lng: 0},
		Status:   models.DriverStatusAvailable,
	})

	rider := &models.Rider{ID: "U1"}
	pickup := models.Location{Lat: 0.1, Lng: 0.1}
	dropoff := models.Location{Lat: 1, Lng: 1}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		ride := dp.RequestRide(rider, pickup, dropoff)
		dp.ProcessQueue()
		b.StartTimer()
		dp.CompleteRide(ride.ID)
	}
}
