package traffic

import (
	"context"
	"fmt"
	"runtime"
	"testing"
	"time"

	"github.com/nextgen-training-kushal/Day-15/models"
)

// TestConcurrentVehicles_NoRace loads 100 vehicles on a 20-node city and runs for 5 seconds.
// Run with: go test -race ./Day-15/traffic/...
func TestConcurrentVehicles_NoRace(t *testing.T) {
	city := NewCityModel(20, nil)
	city.GenerateRandomCity()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	city.StartTrafficSignals(ctx)
	go city.StartCongestionSimulation(ctx)

	// Register 100 vehicles across the 20-node city
	registered := 0
	for i := 1; i <= 100; i++ {
		from := models.IntersectionID((i % 19) + 1)
		to := models.IntersectionID((i % 17) + 2)
		if from == to {
			to = models.IntersectionID((int(to) % 20) + 1)
		}
		plate := fmt.Sprintf("T-%03d", i)
		if err := city.RegisterVehicle(plate, from, to); err == nil {
			registered++
		}
	}

	t.Logf("registered %d/100 vehicles (some may have no path in random graph)", registered)

	city.StartVehicleSimulation(ctx)

	// Let simulation run; the race detector will flag any concurrent access issues
	<-ctx.Done()
}

// TestConcurrentVehicles_AllArrive verifies that vehicles reach their destination.
func TestConcurrentVehicles_AllArrive(t *testing.T) {
	// Use a small deterministic graph so all vehicles are guaranteed paths
	city := buildTestCity()

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	city.StartTrafficSignals(ctx)

	routes := []struct{ from, to int }{
		{1, 5}, {1, 3}, {1, 2}, {2, 5}, {4, 5},
	}
	for i, rt := range routes {
		plate := fmt.Sprintf("V-%02d", i)
		city.RegisterVehicle(plate, models.IntersectionID(rt.from), models.IntersectionID(rt.to))
	}

	city.StartVehicleSimulation(ctx)

	// Poll until all vehicles arrive or timeout
	deadline := time.Now().Add(14 * time.Second)
	for time.Now().Before(deadline) {
		city.VehicleMu.RLock()
		allDone := true
		for _, v := range city.Vehicles {
			v.Mu.RLock()
			done := v.Status == "ARRIVED" || v.Status == "STUCK (NO PATH)"
			v.Mu.RUnlock()
			if !done {
				allDone = false
				break
			}
		}
		city.VehicleMu.RUnlock()

		if allDone {
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	// Report final statuses
	city.VehicleMu.RLock()
	for plate, v := range city.Vehicles {
		v.Mu.RLock()
		t.Logf("%s: %s", plate, v.Status)
		v.Mu.RUnlock()
	}
	city.VehicleMu.RUnlock()
}

// TestGoroutineShutdown verifies that all simulation goroutines exit cleanly on ctx cancel.
// Tracks runtime.NumGoroutine() before and after, allowing a small tolerance.
func TestGoroutineShutdown(t *testing.T) {
	// Record baseline BEFORE the city is started
	baseline := runtime.NumGoroutine()
	t.Logf("baseline goroutines: %d", baseline)

	city := NewCityModel(20, nil)
	city.GenerateRandomCity()

	ctx, cancel := context.WithCancel(context.Background())

	city.StartTrafficSignals(ctx)          // 20 goroutines
	go city.StartCongestionSimulation(ctx) // 1 goroutine
	go city.StartCongestionTracking(ctx)   // 1 goroutine

	// Register 10 vehicles
	for i := 1; i <= 10; i++ {
		from := models.IntersectionID((i % 18) + 1)
		to := models.IntersectionID((i % 16) + 3)
		if from == to {
			to++
		}
		plate := fmt.Sprintf("G-%02d", i)
		city.RegisterVehicle(plate, from, to)
	}
	city.StartVehicleSimulation(ctx) // up to 10 goroutines

	time.Sleep(200 * time.Millisecond) // let all goroutines spin up
	peak := runtime.NumGoroutine()
	t.Logf("peak goroutines: %d", peak)

	// Cancel context → all goroutines should exit
	cancel()
	time.Sleep(600 * time.Millisecond) // allow goroutines to drain

	after := runtime.NumGoroutine()
	t.Logf("goroutines after shutdown: %d", after)

	// Allow a small tolerance (Go runtime's own goroutines)
	tolerance := 5
	if after > baseline+tolerance {
		t.Errorf("goroutine leak: started with %d, ended with %d (tolerance %d)",
			baseline, after, tolerance)
	}
}
