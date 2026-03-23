package internal

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nextgen-training-kushal/Day-10/internal/heap"
	"github.com/nextgen-training-kushal/Day-10/models"
)

var completedTasksPriority int32

type PriorityScheduler struct {
	mu           sync.Mutex
	cond         *sync.Cond
	queue        *heap.MinHeap[models.Task]
	shutdownChan chan struct{}
	stopOnce     sync.Once
}

func NewPriorityScheduler() *PriorityScheduler {
	ps := &PriorityScheduler{
		queue: heap.NewMinHeap(func(t1, t2 models.Task) bool {
			return t1.Priority < t2.Priority
		}),
		shutdownChan: make(chan struct{}),
	}
	ps.cond = sync.NewCond(&ps.mu)
	return ps
}

func (ps *PriorityScheduler) AddTask(task models.Task) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.queue.Insert(task)
	ps.cond.Signal()
}

func (ps *PriorityScheduler) Schedule() {
	metrics := GetMetrics()
	for {
		ps.mu.Lock()
		for ps.queue.Size() == 0 {
			select {
			case <-ps.shutdownChan:
				ps.mu.Unlock()
				fmt.Printf("[T=%.2fs] Priority Scheduler: Shutdown signal received. Exiting.\n", metrics.GetRelativeTime())
				return
			default:
				ps.cond.Wait()
			}
		}

		t, err := ps.queue.ExtractMin()
		ps.mu.Unlock()

		if err != nil {
			continue
		}

		// Metrics tracking
		currentTime := time.Now()
		if t.FirstScheduledTime.IsZero() {
			t.FirstScheduledTime = currentTime
			if currentTime.Sub(t.ArrivalTime) > 5*time.Second {
				metrics.RecordStarvation()
			}
		}
		metrics.RecordContextSwitch()

		fmt.Printf("[T=%.2fs] PID=%d (P%d) START | Priority Scheduler\n", metrics.GetRelativeTime(), t.PID, t.Priority)
		time.Sleep(t.CPUBurst / 10) // Simulate processing
		fmt.Printf("[T=%.2fs] PID=%d DONE | Priority Scheduler\n", metrics.GetRelativeTime(), t.PID)

		atomic.AddInt32(&completedTasksPriority, 1)
		metrics.RecordCompletion(t)
	}
}

func (ps *PriorityScheduler) Shutdown() {
	ps.stopOnce.Do(func() {
		close(ps.shutdownChan)
		ps.mu.Lock()
		ps.cond.Broadcast()
		ps.mu.Unlock()
	})
}

func GetCompletedTasksPriority() int32 {
	return atomic.LoadInt32(&completedTasksPriority)
}
