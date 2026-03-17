package dispatch

import (
	"fmt"
	"math"
	"ride-sharing/internal/driver"
	"ride-sharing/internal/models"
	"ride-sharing/internal/queue"
	"ride-sharing/internal/rides"
	"time"
)

type Dispatcher struct {
	drivers  driver.DriverStore
	queue    queue.RideQueue
	tracker  rides.RideTracker
	earnings map[string]float64
	waits    []time.Duration
}

func NewDispatcher(d driver.DriverStore, q queue.RideQueue, t rides.RideTracker) *Dispatcher {
	return &Dispatcher{
		drivers:  d,
		queue:    q,
		tracker:  t,
		earnings: make(map[string]float64),
	}
}

func (dp *Dispatcher) RequestRide(r *models.Rider, pickup, dropoff models.Location) *models.Ride {
	ride := &models.Ride{
		ID:          fmt.Sprintf("RIDE-%d", time.Now().UnixNano()),
		RiderID:     r.ID,
		Pickup:      pickup,
		Dropoff:     dropoff,
		Status:      models.RideStatusRequested,
		RequestTime: time.Now(),
	}
	dp.queue.Enqueue(ride)

	// Increment zone score for analytics
	if ms, ok := dp.drivers.(*driver.MemoryStore); ok {
		ms.IncrementZoneScore(pickup)
	}

	return ride
}

func (dp *Dispatcher) ProcessQueue() {
	if dp.queue.IsEmpty() {
		return
	}
	ride, _ := dp.queue.Peek()

	// Check timeout (10 minutes)
	if time.Since(ride.RequestTime) > 10*time.Minute {
		dp.queue.Dequeue()
		ride.Status = models.RideStatusCancelled
		if ms, ok := dp.drivers.(*driver.MemoryStore); ok {
			ms.DecrementZoneScore(ride.Pickup)
		}
		fmt.Printf("Ride %s cancelled due to timeout\n", ride.ID)
		return
	}

	// Find nearest driver within 5km
	drivers, err := dp.drivers.FindNearest(ride.Pickup, 5.0)
	if err != nil {
		fmt.Printf("Error searching for drivers for ride %s: %v\n", ride.ID, err)
		return
	}

	if len(drivers) == 0 {
		return // Keep in queue
	}

	// Simple match: First available in the distance (already filtered by FindNearest)
	// We can refine this to pick the absolute closest if multiple found
	var bestDriver *models.Driver
	minDist := 999.0
	for _, d := range drivers {
		dist := dp.euclideanDistance(ride.Pickup, d.Location)
		if dist < minDist {
			minDist = dist
			bestDriver = d
		}
	}
	fmt.Println(bestDriver)
	if bestDriver != nil {
		dp.queue.Dequeue()
		ride.DriverID = bestDriver.ID
		ride.Status = models.RideStatusAccepted
		ride.StartTime = time.Now()
		dp.drivers.UpdateStatus(bestDriver.ID, models.DriverStatusBusy)
		dp.tracker.Add(ride)
		dp.waits = append(dp.waits, time.Since(ride.RequestTime))
		fmt.Printf("Ride %s assigned to Driver %s\n", ride.ID, bestDriver.ID)
	}
}

func (dp *Dispatcher) CompleteRide(rideID string) error {
	ride, err := dp.tracker.Remove(rideID)
	if err != nil {
		return err
	}

	ride.EndTime = time.Now()
	ride.Status = models.RideStatusCompleted
	ride.Fare = dp.calculateFare(ride.Pickup, ride.Dropoff)

	// Update Driver
	d, err := dp.drivers.GetByID(ride.DriverID)
	if err != nil {
		fmt.Printf("Internal error: driver %s assigned to ride %s not found in store\n", ride.DriverID, ride.ID)
		return nil // Still return nil as the ride is conceptually complete
	}
	d.Status = models.DriverStatusAvailable
	d.RideHistory = append(d.RideHistory, ride.ID)

	if err := dp.drivers.UpdateLocation(d.ID, ride.Dropoff); err != nil {
		fmt.Printf("Warning: failed to update driver %s location: %v\n", d.ID, err)
	}
	dp.earnings[d.ID] += ride.Fare
	// Decrement zone score as the request is fulfilled
	if ms, ok := dp.drivers.(*driver.MemoryStore); ok {
		ms.DecrementZoneScore(ride.Pickup)
	}
	fmt.Printf("Ride %s completed. Fare: ₹%.2f\n", ride.ID, ride.Fare)
	return nil
}

func (dp *Dispatcher) calculateFare(p, d models.Location) float64 {
	dist := dp.euclideanDistance(p, d)
	return 50.0 + (12.0 * dist)
}

func (dp *Dispatcher) euclideanDistance(l1, l2 models.Location) float64 {
	return math.Sqrt(math.Pow(l1.Lat-l2.Lat, 2) + math.Pow(l1.Lng-l2.Lng, 2))
}

// Query Functions

func (dp *Dispatcher) GetNearestDrivers(loc models.Location, n int) []*models.Driver {
	drivers, _ := dp.drivers.FindNearest(loc, 5) // Large radius for "nearest N"
	// Sort by distance (could be optimized)
	// For simplicity, we'll just return first N from the slice
	if len(drivers) > n {
		return drivers[:n]
	}
	return drivers
}

func (dp *Dispatcher) GetDriverEarnings(driverID string) float64 {
	return dp.earnings[driverID]
}

func (dp *Dispatcher) GetAverageWaitTime() time.Duration {
	if len(dp.waits) == 0 {
		return 0
	}
	var total time.Duration
	for _, w := range dp.waits {
		total += w
	}
	return total / time.Duration(len(dp.waits))
}

func (dp *Dispatcher) GetBusiestZones() map[string]int {
	if ms, ok := dp.drivers.(*driver.MemoryStore); ok {
		return ms.GetBusiestZones()
	}
	return nil
}
