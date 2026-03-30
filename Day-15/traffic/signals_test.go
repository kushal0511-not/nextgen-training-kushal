package traffic

import (
	"context"
	"testing"
	"time"

	"github.com/nextgen-training-kushal/Day-15/models"
)

func TestSignalPhaseOrder(t *testing.T) {
	city := NewCityModel(1, nil)
	city.AddRoad(1, 1, 1, 60, 1, "NS") // self-loop so intersection has inbound
	inter := city.Intersections[models.IntersectionID(1)]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	go city.RunSignalCycle(ctx, inter)

	// Collect at least 4 distinct signal readings in order
	expected := []models.SignalState{
		models.GreenNS, models.YellowNS, models.GreenEW, models.YellowEW,
	}

	seen := []models.SignalState{}
	prev := models.SignalState("")
	deadline := time.Now().Add(9 * time.Second)

	for len(seen) < 4 && time.Now().Before(deadline) {
		inter.Mu.RLock()
		sig := inter.CurrentSignal
		inter.Mu.RUnlock()

		if sig != prev {
			seen = append(seen, sig)
			prev = sig
		}
		time.Sleep(50 * time.Millisecond)
	}

	if len(seen) < 4 {
		t.Fatalf("only saw %d distinct phases in 9s: %v", len(seen), seen)
	}

	for i, exp := range expected {
		if seen[i] != exp {
			t.Errorf("phase[%d]: expected %s, got %s", i, exp, seen[i])
		}
	}
}

func TestSignalPreemption_SkipsCycle(t *testing.T) {
	city := NewCityModel(1, nil)
	inter := city.Intersections[models.IntersectionID(1)]

	// Preempt the intersection
	inter.Mu.Lock()
	inter.Preempted = true
	inter.CurrentSignal = models.GreenEW
	inter.Mu.Unlock()

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go city.RunSignalCycle(ctx, inter)
	time.Sleep(400 * time.Millisecond)

	// Signal must NOT have changed (cycle was skipped due to preemption)
	inter.Mu.RLock()
	sig := inter.CurrentSignal
	inter.Mu.RUnlock()

	if sig != models.GreenEW {
		t.Errorf("preempted signal changed to %s — expected it to stay GreenEW", sig)
	}
}
