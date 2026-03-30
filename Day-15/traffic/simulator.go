package traffic

import (
	"context"
	"fmt"
	"time"

	"github.com/nextgen-training-kushal/Day-15/models"
)

// Simulation settings
const SpeedScale = 3600.0 // 1 hour in real-time is 1 second in simulation

// RegisterVehicle adds a vehicle to the city and calculates its initial route.
func (cm *CityModel) RegisterVehicle(plate string, start, end models.IntersectionID) error {
	path, _ := cm.Dijkstra(start, end)
	if path == nil {
		return fmt.Errorf("no path found between %d and %d. Vehicle parked", start, end)
	}

	cm.VehicleMu.Lock()
	defer cm.VehicleMu.Unlock()

	cm.Vehicles[plate] = &models.Vehicle{
		Plate:       plate,
		Current:     start,
		Destination: end,
		Path:        path,
		Status:      "READY",
	}
	return nil
}

// StartVehicleSimulation initiates independent goroutines for each registered vehicle.
func (cm *CityModel) StartVehicleSimulation(ctx context.Context) {
	cm.VehicleMu.RLock()
	defer cm.VehicleMu.RUnlock()

	for _, v := range cm.Vehicles {
		go cm.moveVehicle(ctx, v)
	}
}

// moveVehicle handles the lifecycle of a vehicle's journey through the graph.
func (cm *CityModel) moveVehicle(ctx context.Context, v *models.Vehicle) {
	for {
		v.Mu.RLock()
		current := v.Current
		dest := v.Destination
		v.Mu.RUnlock()

		if current == dest {
			v.Mu.Lock()
			v.Status = "ARRIVED"
			v.Mu.Unlock()
			return
		}

		// Re-route at each intersection for real-time optimal path
		path, _ := cm.Dijkstra(current, dest)
		if len(path) < 2 {
			v.Mu.Lock()
			v.Status = "STUCK (NO PATH)"
			v.Mu.Unlock()
			if !SleepWithContext(ctx, 2*time.Second) {
				return
			}
			continue
		}

		// Decide next step
		next := path[1]

		// Find the road detail for the current segment
		cm.AdjMu.RLock()
		roads := cm.AdjList[current]
		var currentRoad *models.Road
		for j := range roads {
			if roads[j].To == next {
				currentRoad = &roads[j]
				break
			}
		}
		cm.AdjMu.RUnlock()

		if currentRoad == nil {
			v.Mu.Lock()
			v.Status = "ERROR (ROAD GONE)"
			v.Mu.Unlock()
			return
		}

		// Wait at signal based on ROAD DIRECTION
		if !cm.waitForSignal(ctx, v, current, currentRoad.Direction) {
			return
		}

		// Move along the road
		v.Mu.Lock()
		v.Status = "MOVING"
		v.Mu.Unlock()

		// Read congestion safely
		cm.AdjMu.RLock()
		travelTime := CalculateTravelTime(*currentRoad)
		cm.AdjMu.RUnlock()

		simDuration := time.Duration(travelTime*3600*1000/SpeedScale) * time.Millisecond

		select {
		case <-ctx.Done():
			return
		case <-time.After(simDuration):
			v.Mu.Lock()
			v.Current = next
			v.Mu.Unlock()
		}
	}
}

// waitForSignal pauses the vehicle goroutine if the signal is not green for its direction.
func (cm *CityModel) waitForSignal(ctx context.Context, v *models.Vehicle, intersectionID models.IntersectionID, roadDir string) bool {
	inter := cm.Intersections[intersectionID]

	for {
		inter.Mu.RLock()
		sig := inter.CurrentSignal
		inter.Mu.RUnlock()

		if (roadDir == "NS" && sig == models.GreenNS) || (roadDir == "EW" && sig == models.GreenEW) {
			return true // Proceed
		}

		v.Mu.Lock()
		v.Status = "WAITING_AT_SIGNAL"
		v.Mu.Unlock()

		if !SleepWithContext(ctx, 200*time.Millisecond) {
			return false
		}
	}
}

// GetVehicleStatus returns a summary of all vehicle positions.
func (cm *CityModel) GetVehicleStatus() map[string]string {
	cm.VehicleMu.RLock()
	defer cm.VehicleMu.RUnlock()

	status := make(map[string]string)
	for plate, v := range cm.Vehicles {
		v.Mu.RLock()
		status[plate] = fmt.Sprintf("At: %2d -> Dest: %2d | Status: %-18s", v.Current, v.Destination, v.Status)
		v.Mu.RUnlock()
	}
	return status
}
