package queue

import (
	"ride-sharing/internal/models"
	"testing"
	"time"
)

func TestMinHeap(t *testing.T) {
	h := NewMinHeap()
	now := time.Now()

	rides := []*models.Ride{
		{ID: "1", RequestTime: now.Add(10 * time.Minute)},
		{ID: "2", RequestTime: now},
		{ID: "3", RequestTime: now.Add(5 * time.Minute)},
	}

	for _, r := range rides {
		h.Enqueue(r)
	}

	if h.Size() != 3 {
		t.Errorf("Expected size 3, got %d", h.Size())
	}

	// Dequeue should return rides in order of RequestTime (oldest first)
	r, _ := h.Dequeue()
	if r.ID != "2" {
		t.Errorf("Expected ID 2, got %s", r.ID)
	}

	r, _ = h.Dequeue()
	if r.ID != "3" {
		t.Errorf("Expected ID 3, got %s", r.ID)
	}

	r, _ = h.Dequeue()
	if r.ID != "1" {
		t.Errorf("Expected ID 1, got %s", r.ID)
	}

	if !h.IsEmpty() {
		t.Error("Heap should be empty")
	}
}

func BenchmarkMinHeap(b *testing.B) {
	h := NewMinHeap()
	r := &models.Ride{RequestTime: time.Now()}
	for i := 0; i < b.N; i++ {
		h.Enqueue(r)
	}
	for i := 0; i < b.N; i++ {
		h.Dequeue()
	}
}
