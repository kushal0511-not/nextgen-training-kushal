package bid

import (
	"sync"
	"testing"

	"github.com/nextgen-training-kushal/Day-13/models"
)

func TestBidHeap(t *testing.T) {
	h := &BidHeap{}
	bids := []models.Bid{
		{ID: 1, Amount: 100},
		{ID: 2, Amount: 250},
		{ID: 3, Amount: 150},
		{ID: 4, Amount: 200},
	}

	for _, b := range bids {
		h.PushBid(b)
	}

	if h.Peek().Amount != 250 {
		t.Errorf("Expected root to be 250, got %v", h.Peek().Amount)
	}

	highest := h.PopBid()
	if highest.Amount != 250 {
		t.Errorf("Expected popped bid to be 250, got %v", highest.Amount)
	}

	if h.Peek().Amount != 200 {
		t.Errorf("Expected new root to be 200, got %v", h.Peek().Amount)
	}
}

func TestBidHeapConcurrency(t *testing.T) {
	h := &BidHeap{}
	const workers = 100
	done := make(chan bool)

	for i := 0; i < workers; i++ {
		go func(id int) {
			h.PushBid(models.Bid{ID: id, Amount: float64(id)})
			h.Peek()
			done <- true
		}(i)
	}

	for i := 0; i < workers; i++ {
		<-done
	}
}

func TestBidHeapStress(t *testing.T) {
	h := &BidHeap{}
	const numGoroutines = 50
	const opsPerGoroutine = 100
	var wg sync.WaitGroup

	wg.Add(numGoroutines * 2)

	// Concurrent Pushers
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				h.PushBid(models.Bid{ID: id*1000 + j, Amount: float64(j)})
			}
		}(i)
	}

	// Concurrent Poppers/Peekers
	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				h.Peek()
				if h.SafeLen() > 0 {
					h.PopBid()
				}
			}
		}()
	}

	wg.Wait()
}
