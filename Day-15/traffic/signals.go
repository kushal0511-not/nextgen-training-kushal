package traffic

import (
	"context"
	"time"

	"github.com/nextgen-training-kushal/Day-15/models"
)

// StartTrafficSignals initializes and starts the goroutines for each intersection.
func (cm *CityModel) StartTrafficSignals(ctx context.Context) {
	for _, inter := range cm.Intersections {
		go cm.RunSignalCycle(ctx, inter)
	}
}

// RunSignalCycle managed a single intersection's signal timing loop.
func (cm *CityModel) RunSignalCycle(ctx context.Context, i *models.Intersection) {
	for {
		// If an emergency vehicle has preempted this intersection, pause normal cycling
		i.Mu.RLock()
		preempted := i.Preempted
		i.Mu.RUnlock()
		if preempted {
			if !SleepWithContext(ctx, 200*time.Millisecond) {
				return
			}
			continue
		}

		// 1. Green NS (Adaptive Timing)
		i.Mu.Lock()
		i.CurrentSignal = models.GreenNS
		i.Mu.Unlock()
		nsCong := cm.GetCongestion(i.ID, i.InboundNS)
		if !SleepWithContext(ctx, CalculateAdaptiveDuration(nsCong)) {
			return
		}

		// 2. Yellow NS
		i.Mu.Lock()
		i.CurrentSignal = models.YellowNS
		i.Mu.Unlock()
		if !SleepWithContext(ctx, 500*time.Millisecond) {
			return
		}

		// 3. Green EW (Adaptive Timing)
		i.Mu.Lock()
		i.CurrentSignal = models.GreenEW
		i.Mu.Unlock()
		ewCong := cm.GetCongestion(i.ID, i.InboundEW)
		if !SleepWithContext(ctx, CalculateAdaptiveDuration(ewCong)) {
			return
		}

		// 4. Yellow EW
		i.Mu.Lock()
		i.CurrentSignal = models.YellowEW
		i.Mu.Unlock()
		if !SleepWithContext(ctx, 500*time.Millisecond) {
			return
		}
	}
}

// GetCongestion calculates total congestion for a list of inbound nodes.
func (cm *CityModel) GetCongestion(target models.IntersectionID, inbound []models.IntersectionID) int {
	cm.AdjMu.RLock()
	defer cm.AdjMu.RUnlock()

	total := 0
	for _, src := range inbound {
		for _, road := range cm.AdjList[src] {
			if road.To == target {
				total += road.CongestionLevel
			}
		}
	}
	return total
}

// CalculateAdaptiveDuration computes the duration of the green light based on traffic density.
func CalculateAdaptiveDuration(congestion int) time.Duration {
	base := 1000 // 1s
	extra := float64(congestion) / 5.0
	return time.Duration(base+int(extra*500)) * time.Millisecond
}
