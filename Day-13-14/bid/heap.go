package bid

import (
	"container/heap"
	"sync"

	"github.com/nextgen-training-kushal/Day-13-14/models"
)

// BidHeap is a max-heap of bids.
type BidHeap struct {
	mu   sync.RWMutex
	bids []models.Bid
}

func (h *BidHeap) Len() int           { return len(h.bids) }
func (h *BidHeap) Less(i, j int) bool { return h.bids[i].Amount > h.bids[j].Amount }
func (h *BidHeap) Swap(i, j int)      { h.bids[i], h.bids[j] = h.bids[j], h.bids[i] }

// SafeLen returns the length of the heap in a thread-safe way.
func (h *BidHeap) SafeLen() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.bids)
}

func (h *BidHeap) Push(x interface{}) {
	h.bids = append(h.bids, x.(models.Bid))
}

func (h *BidHeap) Pop() interface{} {
	old := h.bids
	n := len(old)
	x := old[n-1]
	h.bids = old[0 : n-1]
	return x
}

// PushBid adds a bid to the heap and maintains heap properties.
func (h *BidHeap) PushBid(bid models.Bid) {
	h.mu.Lock()
	defer h.mu.Unlock()
	heap.Push(h, bid)
}

// PopBid removes and returns the highest bid from the heap.
func (h *BidHeap) PopBid() models.Bid {
	h.mu.Lock()
	defer h.mu.Unlock()
	return heap.Pop(h).(models.Bid)
}

// Peek returns the highest bid without removing it.
func (h *BidHeap) Peek() models.Bid {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if len(h.bids) == 0 {
		return models.Bid{}
	}
	return h.bids[0]
}
