package sse

import (
	"log"
	"sync"

	"github.com/nextgen-training-kushal/Day-13-14/models"
)

type Broker struct {
	mu       sync.RWMutex
	watchers map[int]map[chan models.Bid]struct{} // itemID -> set of channels
	logger   *log.Logger
}

func NewBroker(logger *log.Logger) *Broker {
	return &Broker{
		watchers: make(map[int]map[chan models.Bid]struct{}),
		logger:   logger,
	}
}

func (b *Broker) AddWatcher(itemID int) chan models.Bid {
	b.mu.Lock()
	defer b.mu.Unlock()
	ch := make(chan models.Bid, 10)
	if _, ok := b.watchers[itemID]; !ok {
		b.watchers[itemID] = make(map[chan models.Bid]struct{})
	}
	b.watchers[itemID][ch] = struct{}{}
	return ch
}

func (b *Broker) RemoveWatcher(itemID int, ch chan models.Bid) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if itemWatchers, ok := b.watchers[itemID]; ok {
		delete(itemWatchers, ch)
		close(ch)
		if len(itemWatchers) == 0 {
			delete(b.watchers, itemID)
		}
	}
}

func (b *Broker) Broadcast(itemID int, bid models.Bid) {
	b.mu.RLock()
	defer b.mu.RUnlock()
	if itemWatchers, ok := b.watchers[itemID]; ok {
		for ch := range itemWatchers {
			select {
			case ch <- bid:
			default:
				b.logger.Printf("Channel is full, skipping bid %v", bid)
			}
		}
	}
}
