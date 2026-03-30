package traffic

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/nextgen-training-kushal/Day-15/models"
)

const (
	maxHistorySize   = 10 // Sliding window: last 10 samples
	trackingInterval = 2 * time.Second
)

// StartCongestionTracking starts a goroutine that snapshots all road congestion levels periodically.
func (cm *CityModel) StartCongestionTracking(ctx context.Context) {
	ticker := time.NewTicker(trackingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case t := <-ticker.C:
			cm.AdjMu.RLock()
			cm.AnalysisMu.Lock()

			for _, roads := range cm.AdjList {
				for _, r := range roads {
					key := models.RoadKey{From: r.From, To: r.To}
					record := models.CongestionRecord{Timestamp: t, Level: r.CongestionLevel}

					history := cm.RoadHistory[key]
					if len(history) >= maxHistorySize {
						// Shift left: drop oldest entry (sliding window)
						history = history[1:]
					}
					cm.RoadHistory[key] = append(history, record)
				}
			}

			cm.AnalysisMu.Unlock()
			cm.AdjMu.RUnlock()
		}
	}
}

// SlidingWindowAverage returns the average congestion for a road over recent history.
func (cm *CityModel) SlidingWindowAverage(key models.RoadKey) float64 {
	cm.AnalysisMu.RLock()
	defer cm.AnalysisMu.RUnlock()

	history := cm.RoadHistory[key]
	if len(history) == 0 {
		return 0
	}
	total := 0
	for _, r := range history {
		total += r.Level
	}
	return float64(total) / float64(len(history))
}

// Top5CongestedRoads returns the five roads with the highest sliding window average congestion.
func (cm *CityModel) Top5CongestedRoads() []models.CongestedRoad {
	cm.AnalysisMu.RLock()
	keys := make([]models.RoadKey, 0, len(cm.RoadHistory))
	for k := range cm.RoadHistory {
		keys = append(keys, k)
	}
	cm.AnalysisMu.RUnlock()

	roads := make([]models.CongestedRoad, 0, len(keys))
	for _, k := range keys {
		avg := cm.SlidingWindowAverage(k)
		roads = append(roads, models.CongestedRoad{Key: k, AvgCongestion: avg})
	}

	sort.Slice(roads, func(i, j int) bool {
		return roads[i].AvgCongestion > roads[j].AvgCongestion
	})

	if len(roads) > 5 {
		return roads[:5]
	}
	return roads
}

// SuggestAlternative finds a route avoiding the given congested road segment.
func (cm *CityModel) SuggestAlternative(key models.RoadKey) []models.IntersectionID {
	blocked := map[models.RoadKey]bool{key: true}
	path := cm.dijkstraAvoiding(key.From, key.To, blocked)
	return path
}

// PrintCongestionReport prints the top-5 most congested roads and their alternatives.
func (cm *CityModel) PrintCongestionReport() {
	top5 := cm.Top5CongestedRoads()
	if len(top5) == 0 {
		fmt.Println("[Congestion Report] No data yet.")
		return
	}
	fmt.Println("\n[Congestion Report - Top Congested Roads]")
	for i, cr := range top5 {
		alt := cm.SuggestAlternative(cr.Key)
		altStr := "none"
		if len(alt) > 0 {
			altStr = fmt.Sprintf("%v", alt)
		}
		fmt.Printf("  #%d Road %d->%d | Avg Congestion: %.1f/10 | Alt: %s\n",
			i+1, cr.Key.From, cr.Key.To, cr.AvgCongestion, altStr)
	}
}
