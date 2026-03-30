package traffic

import (
	"testing"

	"github.com/nextgen-training-kushal/Day-15/models"
)

// buildTestCity creates a deterministic 5-node graph:
//
//	1 --NS--> 2 --EW--> 3
//	|                   |
//	NS                  EW
//	v                   v
//	4 ----------NS----> 5
func buildTestCity() *CityModel {
	city := NewCityModel(5, nil)
	city.AddRoad(1, 2, 1.0, 60, 1, "NS")
	city.AddRoad(2, 3, 1.0, 60, 1, "EW")
	city.AddRoad(3, 5, 2.0, 60, 1, "EW")
	city.AddRoad(1, 4, 3.0, 60, 1, "NS")
	city.AddRoad(4, 5, 1.0, 60, 1, "NS")
	return city
}

func TestDijkstra_ShortestPath(t *testing.T) {
	city := buildTestCity()
	path, cost := city.Dijkstra(models.IntersectionID(1), models.IntersectionID(5))

	if path == nil {
		t.Fatal("expected a path, got nil")
	}
	if path[0] != 1 || path[len(path)-1] != 5 {
		t.Errorf("path starts at %d and ends at %d, expected 1→5", path[0], path[len(path)-1])
	}
	if cost <= 0 {
		t.Errorf("expected positive travel cost, got %f", cost)
	}

	// Via 1→2→3→5 = 1+1+2 = 4km / 60*(1.1-0.1) = 54 km/h → 4/54 hours ≈ 0.074h
	// Via 1→4→5   = 3+1   = 4km / same speed → same cost
	// Both are equal so just verify path length makes sense
	if len(path) < 3 {
		t.Errorf("expected path length >= 3, got %d", len(path))
	}
}

func TestDijkstra_NoPath(t *testing.T) {
	city := NewCityModel(3, nil)
	city.AddRoad(1, 2, 1.0, 60, 1, "NS")
	// node 3 is isolated
	path, cost := city.Dijkstra(models.IntersectionID(1), models.IntersectionID(3))
	if path != nil {
		t.Errorf("expected nil path for disconnected nodes, got %v", path)
	}
	if cost != 0 {
		t.Errorf("expected 0 cost for no path, got %f", cost)
	}
}

func TestDijkstra_SameNode(t *testing.T) {
	city := buildTestCity()
	path, _ := city.Dijkstra(models.IntersectionID(2), models.IntersectionID(2))
	if len(path) != 1 || path[0] != 2 {
		t.Errorf("expected [2] for same-node route, got %v", path)
	}
}

func TestCalculateTravelTime_HighCongestion(t *testing.T) {
	r := models.Road{Distance: 10, SpeedLimit: 60, CongestionLevel: 10}
	t1 := CalculateTravelTime(r)

	r2 := models.Road{Distance: 10, SpeedLimit: 60, CongestionLevel: 1}
	t2 := CalculateTravelTime(r2)

	if t1 <= t2 {
		t.Errorf("high congestion should make travel slower: t1=%f t2=%f", t1, t2)
	}
}
