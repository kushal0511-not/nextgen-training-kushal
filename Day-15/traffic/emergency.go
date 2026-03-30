package traffic

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/nextgen-training-kushal/Day-15/heap"
	"github.com/nextgen-training-kushal/Day-15/models"
)

// RegisterEmergencyVehicle registers an emergency vehicle with the highest priority.
// It calculates the fastest route (ignoring congestion), preempts signals, and
// forces all other vehicles to reroute around the emergency path.
func (cm *CityModel) RegisterEmergencyVehicle(ctx context.Context, plate string, start, end models.IntersectionID) error {
	// Calculate fastest path with congestion overridden to 0
	path, _ := cm.dijkstraEmergency(start, end)
	if path == nil {
		return fmt.Errorf("emergency vehicle %s: no path found from %d to %d", plate, start, end)
	}

	cm.VehicleMu.Lock()
	cm.Vehicles[plate] = &models.Vehicle{
		Plate:       plate,
		Current:     start,
		Destination: end,
		Path:        path,
		Status:      "EMERGENCY_DISPATCHED",
		IsEmergency: true,
	}
	cm.VehicleMu.Unlock()

	// Preempt all signals on the route
	cm.PreemptSignals(path)
	fmt.Printf("[EMERGENCY] %s dispatched on path %v\n", plate, path)

	// Reroute all non-emergency vehicles away from the emergency path
	cm.RerouteNormalVehicles(path)

	// Start the emergency vehicle's movement goroutine, which releases signals on arrival
	go cm.moveEmergencyVehicle(ctx, cm.Vehicles[plate], path)

	return nil
}

// dijkstraEmergency runs Dijkstra with all congestion set to 0 (pure time/distance).
func (cm *CityModel) dijkstraEmergency(start, end models.IntersectionID) ([]models.IntersectionID, float64) {
	cm.AdjMu.RLock()
	defer cm.AdjMu.RUnlock()

	numNodes := len(cm.Intersections) + 1
	dist := make([]float64, numNodes)
	prev := make([]models.IntersectionID, numNodes)
	pq := heap.NewMinHeap()

	for i := range dist {
		dist[i] = math.Inf(1)
	}

	dist[start] = 0
	pq.Insert(heap.PrioritizedItem{ID: start, Value: 0})

	for pq.Size() > 0 {
		u_item, _ := pq.ExtractMin()
		u := u_item.ID

		if u_item.Value > dist[u] {
			continue
		}
		if u == end {
			break
		}

		for _, road := range cm.AdjList[u] {
			// Override congestion to 1 → pure speed limit determines travel time
			cleared := road
			cleared.CongestionLevel = 1
			alt := dist[u] + CalculateTravelTime(cleared)
			if alt < dist[road.To] {
				dist[road.To] = alt
				prev[road.To] = u
				pq.Insert(heap.PrioritizedItem{ID: road.To, Value: alt})
			}
		}
	}

	curr := end
	if prev[curr] == 0 && curr != start {
		return nil, 0
	}
	path := []models.IntersectionID{}
	for curr != 0 {
		path = append(path, curr)
		curr = prev[curr]
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path, dist[end]
}

// PreemptSignals forces green signals on all intersections along the emergency path.
func (cm *CityModel) PreemptSignals(path []models.IntersectionID) {
	cm.AdjMu.RLock()
	defer cm.AdjMu.RUnlock()

	for i := 0; i < len(path)-1; i++ {
		from, to := path[i], path[i+1]
		inter := cm.Intersections[from]

		// Find the road to determine which direction to green
		dir := "NS"
		for _, r := range cm.AdjList[from] {
			if r.To == to {
				dir = r.Direction
				break
			}
		}

		inter.Mu.Lock()
		inter.Preempted = true
		if dir == "NS" {
			inter.CurrentSignal = models.GreenNS
		} else {
			inter.CurrentSignal = models.GreenEW
		}
		inter.Mu.Unlock()

		fmt.Printf("[PREEMPT] Intersection %d → GREEN-%s (emergency)\n", from, dir)
	}
}

// ReleaseSignals restores normal signal cycling after the emergency vehicle passes.
func (cm *CityModel) ReleaseSignals(path []models.IntersectionID) {
	for _, id := range path {
		inter := cm.Intersections[id]
		inter.Mu.Lock()
		inter.Preempted = false
		inter.Mu.Unlock()
	}
	fmt.Println("[PREEMPT RELEASED] Normal signal cycling restored.")
}

// RerouteNormalVehicles forces non-emergency vehicles to avoid the emergency path.
func (cm *CityModel) RerouteNormalVehicles(emergencyPath []models.IntersectionID) {
	// Build a set of road keys that make up the emergency path
	blocked := make(map[models.RoadKey]bool)
	for i := 0; i < len(emergencyPath)-1; i++ {
		blocked[models.RoadKey{From: emergencyPath[i], To: emergencyPath[i+1]}] = true
	}

	cm.VehicleMu.RLock()
	defer cm.VehicleMu.RUnlock()

	for _, v := range cm.Vehicles {
		v.Mu.RLock()
		isEmergency := v.IsEmergency
		current := v.Current
		dest := v.Destination
		v.Mu.RUnlock()

		if isEmergency {
			continue
		}

		// Recalculate path avoiding blocked roads
		newPath := cm.dijkstraAvoiding(current, dest, blocked)
		if len(newPath) >= 2 {
			v.Mu.Lock()
			v.Path = newPath
			v.Status = "RE-ROUTING (emergency)"
			v.Mu.Unlock()
		}
	}
}

// dijkstraAvoiding runs Dijkstra treating specified road keys as infinite cost.
func (cm *CityModel) dijkstraAvoiding(start, end models.IntersectionID, blocked map[models.RoadKey]bool) []models.IntersectionID {
	cm.AdjMu.RLock()
	defer cm.AdjMu.RUnlock()

	numNodes := len(cm.Intersections) + 1
	dist := make([]float64, numNodes)
	prev := make([]models.IntersectionID, numNodes)
	pq := heap.NewMinHeap()

	for i := range dist {
		dist[i] = math.Inf(1)
	}

	dist[start] = 0
	pq.Insert(heap.PrioritizedItem{ID: start, Value: 0})

	for pq.Size() > 0 {
		u_item, _ := pq.ExtractMin()
		u := u_item.ID
		if u_item.Value > dist[u] {
			continue
		}
		if u == end {
			break
		}

		for _, road := range cm.AdjList[u] {
			key := models.RoadKey{From: u, To: road.To}
			if blocked[key] {
				continue
			} // Skip emergency path
			alt := dist[u] + CalculateTravelTime(road)
			if alt < dist[road.To] {
				dist[road.To] = alt
				prev[road.To] = u
				pq.Insert(heap.PrioritizedItem{ID: road.To, Value: alt})
			}
		}
	}

	curr := end
	if prev[curr] == 0 && curr != start {
		return nil
	}
	path := []models.IntersectionID{}
	for curr != 0 {
		path = append(path, curr)
		curr = prev[curr]
	}
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}
	return path
}

// moveEmergencyVehicle moves an emergency vehicle along its precomputed path without signal waits.
func (cm *CityModel) moveEmergencyVehicle(ctx context.Context, v *models.Vehicle, path []models.IntersectionID) {
	for i := 0; i < len(path)-1; i++ {
		from, to := path[i], path[i+1]

		// Find road
		cm.AdjMu.RLock()
		var road *models.Road
		for j := range cm.AdjList[from] {
			if cm.AdjList[from][j].To == to {
				r := cm.AdjList[from][j]
				road = &r
				break
			}
		}
		cm.AdjMu.RUnlock()

		if road == nil {
			continue
		}

		v.Mu.Lock()
		v.Status = "EMERGENCY_MOVING"
		v.Mu.Unlock()

		// Emergency vehicles travel at full speed regardless of congestion
		cleared := *road
		cleared.CongestionLevel = 1
		simDuration := time.Duration(CalculateTravelTime(cleared)*3600*1000/SpeedScale) * time.Millisecond

		if !SleepWithContext(ctx, simDuration) {
			return
		}

		v.Mu.Lock()
		v.Current = to
		v.Mu.Unlock()
	}

	v.Mu.Lock()
	v.Status = "ARRIVED (EMERGENCY)"
	v.Mu.Unlock()

	// Release signal preemption
	cm.ReleaseSignals(path)
	fmt.Printf("[EMERGENCY] %s has arrived at destination.\n", v.Plate)
}
