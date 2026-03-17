package rides

import (
	"ride-sharing/internal/models"
	"testing"
)

func TestLinkedTracker(t *testing.T) {
	tr := NewLinkedTracker()
	r1 := &models.Ride{ID: "R1"}
	r2 := &models.Ride{ID: "R2"}

	tr.Add(r1)
	tr.Add(r2)

	if len(tr.List()) != 2 {
		t.Errorf("Expected 2 rides, got %d", len(tr.List()))
	}

	res, _ := tr.Get("R1")
	if res.ID != "R1" {
		t.Errorf("Expected R1, got %s", res.ID)
	}

	removed, _ := tr.Remove("R1")
	if removed.ID != "R1" {
		t.Errorf("Expected R1 removed, got %s", removed.ID)
	}

	if len(tr.List()) != 1 {
		t.Errorf("Expected 1 ride, got %d", len(tr.List()))
	}
}
