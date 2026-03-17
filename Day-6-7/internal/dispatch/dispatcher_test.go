package dispatch

import (
	"ride-sharing/internal/driver"
	"ride-sharing/internal/models"
	"ride-sharing/internal/queue"
	"ride-sharing/internal/rides"
	"testing"
	"time"
)

func TestDispatcherMatch(t *testing.T) {
	ds := driver.NewMemoryStore()
	rq := queue.NewMinHeap()
	rt := rides.NewLinkedTracker()
	dp := NewDispatcher(ds, rq, rt)

	// Register driver
	d := &models.Driver{
		ID:       "D1",
		Location: models.Location{Lat: 0, Lng: 0},
		Status:   models.DriverStatusAvailable,
	}
	ds.Register(d)

	// Request ride near driver
	rider := &models.Rider{ID: "U1"}
	ride := dp.RequestRide(rider, models.Location{Lat: 0.01, Lng: 0.01}, models.Location{Lat: 1, Lng: 1})

	// Process queue
	dp.ProcessQueue()

	// Check if assigned
	if ride.Status != models.RideStatusAccepted {
		t.Errorf("Expected ride status accepted, got %s", ride.Status)
	}
	if ride.DriverID != "D1" {
		t.Errorf("Expected driver D1, got %s", ride.DriverID)
	}

	// Complete ride
	dp.CompleteRide(ride.ID)
	if ride.Status != models.RideStatusCompleted {
		t.Errorf("Expected ride status completed, got %s", ride.Status)
	}
	if d.Status != models.DriverStatusAvailable {
		t.Errorf("Expected driver status available, got %s", d.Status)
	}
	if dp.GetDriverEarnings("D1") <= 0 {
		t.Error("Expected earnings > 0")
	}
}

func TestDispatcherTimeout(t *testing.T) {
	ds := driver.NewMemoryStore()
	rq := queue.NewMinHeap()
	rt := rides.NewLinkedTracker()
	dp := NewDispatcher(ds, rq, rt)

	rider := &models.Rider{ID: "U1"}
	ride := dp.RequestRide(rider, models.Location{Lat: 0, Lng: 0}, models.Location{Lat: 1, Lng: 1})

	// Manually set request time to > 10 mins ago
	ride.RequestTime = time.Now().Add(-11 * time.Minute)

	dp.ProcessQueue()

	if ride.Status != models.RideStatusCancelled {
		t.Errorf("Expected ride status cancelled, got %s", ride.Status)
	}
}

func TestDispatcherClosestDriver(t *testing.T) {
	ds := driver.NewMemoryStore()
	rq := queue.NewMinHeap()
	rt := rides.NewLinkedTracker()
	dp := NewDispatcher(ds, rq, rt)

	// D1 at 1km (Euclidean: 0.01^2 + 0.01^2 = 0.0002. Wait, 1km is approx 0.01 degree)
	// Let's use simple coords: (0.1, 0.1) vs (0.2, 0.2)
	ds.Register(&models.Driver{ID: "D1", Location: models.Location{Lat: 0.1, Lng: 0.1}, Status: models.DriverStatusAvailable})
	ds.Register(&models.Driver{ID: "D2", Location: models.Location{Lat: 0.2, Lng: 0.2}, Status: models.DriverStatusAvailable})
	ds.Register(&models.Driver{ID: "D3", Location: models.Location{Lat: 1, Lng: 1}, Status: models.DriverStatusAvailable})

	ride := dp.RequestRide(&models.Rider{ID: "U1"}, models.Location{Lat: 0, Lng: 0}, models.Location{Lat: 0.1, Lng: 0.1})
	dp.ProcessQueue()
	if ride.DriverID != "D1" {
		t.Errorf("Expected closest driver D1, got %s", ride.DriverID)
	}
}

func TestDispatcherNoDriverInRange(t *testing.T) {
	ds := driver.NewMemoryStore()
	rq := queue.NewMinHeap()
	rt := rides.NewLinkedTracker()
	dp := NewDispatcher(ds, rq, rt)

	ds.Register(&models.Driver{ID: "D1", Location: models.Location{Lat: 10, Lng: 10}, Status: models.DriverStatusAvailable})

	ride := dp.RequestRide(&models.Rider{ID: "U1"}, models.Location{Lat: 0, Lng: 0}, models.Location{Lat: 1, Lng: 1})
	dp.ProcessQueue()

	if ride.Status != models.RideStatusRequested {
		t.Errorf("Expected ride status requested (no driver in range), got %s", ride.Status)
	}
}

func TestDispatcherExcludedBusyDriver(t *testing.T) {
	ds := driver.NewMemoryStore()
	rq := queue.NewMinHeap()
	rt := rides.NewLinkedTracker()
	dp := NewDispatcher(ds, rq, rt)

	ds.Register(&models.Driver{ID: "D1", Location: models.Location{Lat: 0.1, Lng: 0.1}, Status: models.DriverStatusBusy})

	ride := dp.RequestRide(&models.Rider{ID: "U1"}, models.Location{Lat: 0, Lng: 0}, models.Location{Lat: 1, Lng: 1})
	dp.ProcessQueue()

	if ride.Status != models.RideStatusRequested {
		t.Errorf("Expected ride status requested (available driver is busy), got %s", ride.Status)
	}
}

func TestDispatcherAnalytics(t *testing.T) {
	ds := driver.NewMemoryStore()
	rq := queue.NewMinHeap()
	rt := rides.NewLinkedTracker()
	dp := NewDispatcher(ds, rq, rt)

	ds.Register(&models.Driver{ID: "D1", Location: models.Location{Lat: 0, Lng: 0}, Status: models.DriverStatusAvailable})

	// Ride 1
	r1 := dp.RequestRide(&models.Rider{ID: "U1"}, models.Location{Lat: 0.01, Lng: 0.01}, models.Location{Lat: 1, Lng: 1})
	dp.ProcessQueue()
	dp.CompleteRide(r1.ID)

	// Ride 2
	r2 := dp.RequestRide(&models.Rider{ID: "U2"}, models.Location{Lat: 0.01, Lng: 0.01}, models.Location{Lat: 2, Lng: 2})
	dp.ProcessQueue()
	dp.CompleteRide(r2.ID)

	avgWait := dp.GetAverageWaitTime()
	if avgWait < 0 {
		t.Errorf("Average wait time should be positive, got %v", avgWait)
	}

	earnings := dp.GetDriverEarnings("D1")
	if earnings <= 0 {
		t.Errorf("Driver D1 should have positive earnings, got %v", earnings)
	}

	zones := dp.GetBusiestZones()
	if len(zones) != 0 {
		t.Error("Zones should be empty")
	}
}

func TestDispatcherErrors(t *testing.T) {
	ds := driver.NewMemoryStore()
	rq := queue.NewMinHeap()
	rt := rides.NewLinkedTracker()
	dp := NewDispatcher(ds, rq, rt)

	err := dp.CompleteRide("NON-EXISTENT-ID")
	if err == nil {
		t.Error("Expected error when completing non-existent ride")
	}
}

func TestDispatcherCrossZoneMatch(t *testing.T) {
	ds := driver.NewMemoryStore()
	rq := queue.NewMinHeap()
	rt := rides.NewLinkedTracker()
	dp := NewDispatcher(ds, rq, rt)

	// Pickup is at (4.9, 4.9), which is in zone "0:0"
	// Driver is at (5.1, 5.1), which is in zone "5:5"
	// Distance is sqrt(0.2^2 + 0.2^2) = 0.28km, which is < 5km
	ds.Register(&models.Driver{
		ID:       "D_NEXT_ZONE",
		Location: models.Location{Lat: 5.1, Lng: 5.1},
		Status:   models.DriverStatusAvailable,
	})

	ride := dp.RequestRide(&models.Rider{ID: "U1"}, models.Location{Lat: 4.9, Lng: 4.9}, models.Location{Lat: 10, Lng: 10})
	dp.ProcessQueue()

	if ride.Status != models.RideStatusAccepted {
		t.Errorf("Expected ride to be accepted, even if driver is in next zone, but status is %s", ride.Status)
	}
	if ride.DriverID != "D_NEXT_ZONE" {
		t.Errorf("Expected driver D_NEXT_ZONE, got %s", ride.DriverID)
	}
}

var ssink string

func BenchmarkMapAccessAlloc(b *testing.B) {
	b.ReportAllocs()

	m := map[int]string{
		1: "north",
		2: "south",
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ssink = m[1] + m[2]
	}
}
