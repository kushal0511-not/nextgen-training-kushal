package traffic

import (
	"fmt"
	"runtime"
	"sync"
	"testing"

	"github.com/nextgen-training-kushal/Day-15/models"
)

// buildNNodeCity creates a random fully-connected city of N intersections.
func buildNNodeCity(n int) *CityModel {
	city := NewCityModel(n, nil)
	city.GenerateRandomCity()
	return city
}

// --- Dijkstra scaling benchmarks ---

func BenchmarkDijkstra_20nodes(b *testing.B) {
	benchDijkstra(b, 20)
}

func BenchmarkDijkstra_100nodes(b *testing.B) {
	benchDijkstra(b, 100)
}

func BenchmarkDijkstra_500nodes(b *testing.B) {
	benchDijkstra(b, 500)
}

func benchDijkstra(b *testing.B, n int) {
	city := buildNNodeCity(n)
	start := models.IntersectionID(1)
	end := models.IntersectionID(n)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		city.Dijkstra(start, end)
	}
}

// --- Mutex vs RWMutex comparison under 8 readers + 1 writer ---

// BenchmarkRWMutex_GraphEdgeUpdates benchmarks the current RWMutex approach.
// 8 concurrent readers call Dijkstra while 1 writer updates edge weights.
func BenchmarkRWMutex_GraphEdgeUpdates(b *testing.B) {
	city := buildNNodeCity(20)

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Mix of reads (Dijkstra) and writes (congestion update)
			if runtime.NumGoroutine()%8 == 0 {
				// 1 in 8 goroutines writes (simulates congestion updater)
				city.AdjMu.Lock()
				for from, roads := range city.AdjList {
					for j := range roads {
						roads[j].CongestionLevel = (roads[j].CongestionLevel % 10) + 1
					}
					city.AdjList[from] = roads
				}
				city.AdjMu.Unlock()
			} else {
				// 7 in 8 goroutines read (simulates Dijkstra calls)
				city.Dijkstra(models.IntersectionID(1), models.IntersectionID(20))
			}
		}
	})
}

// BenchmarkMutex_GraphEdgeUpdates uses a plain sync.Mutex for comparison.
func BenchmarkMutex_GraphEdgeUpdates(b *testing.B) {
	city := buildNNodeCity(20)

	var mu sync.Mutex // plain Mutex instead of RWMutex

	b.ReportAllocs()
	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			if runtime.NumGoroutine()%8 == 0 {
				mu.Lock()
				for from, roads := range city.AdjList {
					for j := range roads {
						roads[j].CongestionLevel = (roads[j].CongestionLevel % 10) + 1
					}
					city.AdjList[from] = roads
				}
				mu.Unlock()
			} else {
				// Read without AdjMu (uses local mu instead)
				mu.Lock()
				_ = fmt.Sprintf("%d", len(city.AdjList)) // simulate minimal read work
				mu.Unlock()
			}
		}
	})
}

// --- Heap vs Graph traversal: Dijkstra internals breakdown ---

// BenchmarkDijkstra_HeapPressure measures allocations per Dijkstra call,
// which are dominated by the priority queue heap insertions.
func BenchmarkDijkstra_HeapPressure(b *testing.B) {
	city := buildNNodeCity(100)
	start := models.IntersectionID(1)
	end := models.IntersectionID(100)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		city.Dijkstra(start, end)
	}
}
