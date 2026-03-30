package traffic

import (
	"context"
	"math"
	"math/rand"
	"sync"
	"time"

	"github.com/nextgen-training-kushal/Day-15/heap"
	"github.com/nextgen-training-kushal/Day-15/models"
	"go.uber.org/zap"
)

// CityModel manages the graph representation of intersections and roads.
type CityModel struct {
	AdjList       map[models.IntersectionID][]models.Road
	Intersections map[models.IntersectionID]*models.Intersection
	Vehicles      map[string]*models.Vehicle
	VehicleMu     sync.RWMutex
	AdjMu         sync.RWMutex
	RoadHistory   map[models.RoadKey][]models.CongestionRecord
	AnalysisMu    sync.RWMutex
	Logger        *zap.Logger
}

// NewCityModel initializes a new CityModel with N intersections.
func NewCityModel(numIntersections int, logger *zap.Logger) *CityModel {
	if logger == nil {
		logger, _ = zap.NewProduction()
	}
	city := &CityModel{
		AdjList:       make(map[models.IntersectionID][]models.Road),
		Intersections: make(map[models.IntersectionID]*models.Intersection),
		Vehicles:      make(map[string]*models.Vehicle),
		RoadHistory:   make(map[models.RoadKey][]models.CongestionRecord),
		Logger:        logger,
	}
	for i := 1; i <= numIntersections; i++ {
		id := models.IntersectionID(i)
		city.AdjList[id] = []models.Road{}
		city.Intersections[id] = &models.Intersection{
			ID:            id,
			CurrentSignal: models.GreenNS,
			InboundNS:     []models.IntersectionID{},
			InboundEW:     []models.IntersectionID{},
		}
	}
	return city
}

// AddRoad connects two intersections and assigns them to signal orientation.
func (cm *CityModel) AddRoad(from, to models.IntersectionID, dist float64, speedLimit int, congestion int, direction string) {
	cm.AdjMu.Lock()
	defer cm.AdjMu.Unlock()

	road := models.Road{
		From:            from,
		To:              to,
		Direction:       direction,
		Distance:        dist,
		CongestionLevel: congestion,
		SpeedLimit:      speedLimit,
	}
	cm.AdjList[from] = append(cm.AdjList[from], road)

	inter := cm.Intersections[to]
	if direction == "NS" {
		inter.InboundNS = append(inter.InboundNS, from)
	} else {
		inter.InboundEW = append(inter.InboundEW, from)
	}
}

// StartCongestionSimulation periodically changes congestion across the city.
func (cm *CityModel) StartCongestionSimulation(ctx context.Context) {
	ticker := time.NewTicker(4 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			cm.AdjMu.Lock()
			for from, roads := range cm.AdjList {
				for j := range roads {
					change := rand.Intn(3) - 1
					lvl := roads[j].CongestionLevel + change
					if lvl < 1 {
						lvl = 1
					}
					if lvl > 10 {
						lvl = 10
					}
					roads[j].CongestionLevel = lvl
				}
				cm.AdjList[from] = roads
			}
			cm.AdjMu.Unlock()
		}
	}
}

// Dijkstra calculates the shortest path between start and end using dynamic travel times.
func (cm *CityModel) Dijkstra(start, end models.IntersectionID) ([]models.IntersectionID, float64) {
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
			alt := dist[u] + CalculateTravelTime(road)
			if alt < dist[road.To] {
				dist[road.To] = alt
				prev[road.To] = u
				pq.Insert(heap.PrioritizedItem{ID: road.To, Value: alt})
			}
		}
	}

	path := []models.IntersectionID{}
	curr := end
	if prev[curr] == 0 && curr != start {
		return nil, 0
	}

	for curr != 0 {
		path = append(path, curr)
		curr = prev[curr]
	}

	// Reverse path
	for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
		path[i], path[j] = path[j], path[i]
	}

	return path, dist[end]
}

// CalculateTravelTime computes travel time in hours factoring in congestion.
func CalculateTravelTime(r models.Road) float64 {
	effSpeed := float64(r.SpeedLimit) * (1.1 - float64(r.CongestionLevel)/10.0)
	if effSpeed < 5.0 {
		effSpeed = 5.0
	}
	return r.Distance / effSpeed
}

// GenerateRandomCity populates the city with mock roads and data.
func (cm *CityModel) GenerateRandomCity() {
	numIntersections := len(cm.AdjList)
	for i := 1; i <= numIntersections; i++ {
		connections := rand.Intn(3) + 2
		for c := 0; c < connections; c++ {
			to := rand.Intn(numIntersections) + 1
			if models.IntersectionID(to) == models.IntersectionID(i) {
				continue
			}
			dist := rand.Float64()*4.5 + 0.5
			speed := []int{30, 40, 50, 60}[rand.Intn(4)]
			congestion := rand.Intn(10) + 1

			dir := "NS"
			if rand.Intn(2) == 0 {
				dir = "EW"
			}

			cm.AddRoad(models.IntersectionID(i), models.IntersectionID(to), dist, speed, congestion, dir)
		}
	}
}

// SleepWithContext provides a context-aware way to delay execution in simulation goroutines.
func SleepWithContext(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}
