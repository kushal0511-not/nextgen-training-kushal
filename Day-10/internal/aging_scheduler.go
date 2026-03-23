package internal

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nextgen-training-kushal/Day-10/internal/heap"
	"github.com/nextgen-training-kushal/Day-10/models"
)

var completedTasksAging int32

type AgingScheduler struct {
	mu           sync.Mutex
	cond         *sync.Cond
	queue        *heap.MinHeap[models.Task]
	shutdownChan chan struct{}
	once         sync.Once
	stopOnce     sync.Once
}

func NewAgingScheduler() *AgingScheduler {
	as := &AgingScheduler{
		queue: heap.NewMinHeap(func(t1, t2 models.Task) bool {
			return t1.Priority < t2.Priority
		}),
		shutdownChan: make(chan struct{}),
	}
	as.cond = sync.NewCond(&as.mu)
	return as
}

func (as *AgingScheduler) AddTask(task models.Task) {
	as.mu.Lock()
	defer as.mu.Unlock()
	as.queue.Insert(task)
	as.cond.Signal()
	as.once.Do(as.startAging)
}

func (as *AgingScheduler) startAging() {
	// every 500ms increase priority (decrease the value)
	ticker := time.NewTicker(500 * time.Millisecond)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				as.mu.Lock()
				if as.queue.Size() == 0 {
					as.mu.Unlock()
					continue
				}
				for i := 0; i < len(as.queue.Data); i++ {
					if as.queue.Data[i].Priority > 1 {
						as.queue.Data[i].Priority -= 1
					}
				}
				as.rebuildHeap()
				as.mu.Unlock()
			case <-as.shutdownChan:
				return
			}
		}
	}()
}

// rebuildHeap is a helper to restore heap property after bulk modification
func (as *AgingScheduler) rebuildHeap() {
	for i := (len(as.queue.Data) / 2) - 1; i >= 0; i-- {
		as.queue.HeapifyDown(i)
	}
}

func (as *AgingScheduler) Schedule() {
	metrics := GetMetrics()
	for {
		as.mu.Lock()
		for as.queue.Size() == 0 {
			select {
			case <-as.shutdownChan:
				as.mu.Unlock()
				fmt.Printf("[T=%.2fs] Aging Scheduler: Shutdown signal received. Exiting.\n", metrics.GetRelativeTime())
				return
			default:
				as.cond.Wait()
			}
		}
		t, err := as.queue.ExtractMin()
		as.mu.Unlock()

		if err == nil {
			// Metrics tracking
			currentTime := time.Now()
			if t.FirstScheduledTime.IsZero() {
				t.FirstScheduledTime = currentTime
				if currentTime.Sub(t.ArrivalTime) > 5*time.Second {
					metrics.RecordStarvation()
				}
			}
			metrics.RecordContextSwitch()

			fmt.Printf("[T=%.2fs] PID=%d (P%d) START | Aging Scheduler\n", metrics.GetRelativeTime(), t.PID, t.Priority)
			time.Sleep(t.CPUBurst / 10)
			fmt.Printf("[T=%.2fs] PID=%d DONE | Aging Scheduler\n", metrics.GetRelativeTime(), t.PID)

			atomic.AddInt32(&completedTasksAging, 1)
			metrics.RecordCompletion(t)
		}
	}
}

func (as *AgingScheduler) Shutdown() {
	as.stopOnce.Do(func() {
		close(as.shutdownChan)
		as.mu.Lock()
		as.cond.Broadcast()
		as.mu.Unlock()
	})
}

func GetCompletedTasksAging() int32 {
	return atomic.LoadInt32(&completedTasksAging)
}
