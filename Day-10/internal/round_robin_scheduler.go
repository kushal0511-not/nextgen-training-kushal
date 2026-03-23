package internal

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nextgen-training-kushal/Day-10/models"
)

var completedTasksRoundRobin int32

// RoundRobinScheduler implements a multi-level feedback queue-like scheduler.
type RoundRobinScheduler struct {
	mu          sync.RWMutex
	cond        *sync.Cond
	highQueue   []models.Task
	mediumQueue []models.Task
	lowQueue    []models.Task
	runningTask *models.Task
	interrupt   chan bool
	shutdown    chan bool
}

func NewRoundRobinScheduler() *RoundRobinScheduler {
	ps := &RoundRobinScheduler{
		highQueue:   make([]models.Task, 0),
		mediumQueue: make([]models.Task, 0),
		lowQueue:    make([]models.Task, 0),
		interrupt:   make(chan bool, 1),
		shutdown:    make(chan bool, 1),
	}
	ps.cond = sync.NewCond(&ps.mu)
	return ps
}

func (ps *RoundRobinScheduler) AddTask(task models.Task) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if task.Priority <= 3 {
		ps.highQueue = append(ps.highQueue, task)
		if ps.runningTask != nil && ps.runningTask.Priority > 3 {
			select {
			case ps.interrupt <- true:
			default:
			}
		}
	} else if task.Priority <= 7 {
		ps.mediumQueue = append(ps.mediumQueue, task)
	} else {
		ps.lowQueue = append(ps.lowQueue, task)
	}
	ps.cond.Signal()
}

func (ps *RoundRobinScheduler) Schedule() {
	metrics := GetMetrics()
	for {
		ps.mu.Lock()
		for len(ps.highQueue) == 0 && len(ps.mediumQueue) == 0 && len(ps.lowQueue) == 0 {
			select {
			case <-ps.shutdown:
				ps.mu.Unlock()
				fmt.Printf("[T=%.2fs] Round Robin Scheduler: Shutdown signal received. Exiting.\n", metrics.GetRelativeTime())
				return
			default:
				ps.cond.Wait()
			}
		}
		ps.mu.Unlock()

		task, quantum := ps.getNextTask()
		if task != nil {
			ps.executeTask(task, quantum)
		}
	}
}

func (ps *RoundRobinScheduler) getNextTask() (*models.Task, time.Duration) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if len(ps.highQueue) > 0 {
		t := ps.highQueue[0]
		ps.highQueue = ps.highQueue[1:]
		return &t, 50 * time.Millisecond
	}
	if len(ps.mediumQueue) > 0 {
		t := ps.mediumQueue[0]
		ps.mediumQueue = ps.mediumQueue[1:]
		return &t, 100 * time.Millisecond
	}
	if len(ps.lowQueue) > 0 {
		t := ps.lowQueue[0]
		ps.lowQueue = ps.lowQueue[1:]
		return &t, 200 * time.Millisecond
	}
	return nil, 0
}

func (ps *RoundRobinScheduler) executeTask(t *models.Task, quantum time.Duration) {
	metrics := GetMetrics()
	ps.mu.Lock()
	ps.runningTask = t
	ps.mu.Unlock()

	// Metrics tracking
	currentTime := time.Now()
	if t.FirstScheduledTime.IsZero() {
		t.FirstScheduledTime = currentTime
		if currentTime.Sub(t.ArrivalTime) > 5*time.Second {
			metrics.RecordStarvation()
		}
	}
	metrics.RecordContextSwitch()

	fmt.Printf("[T=%.2fs] PID=%d (P%d) START | Round Robin (Q=%v)\n", metrics.GetRelativeTime(), t.PID, t.Priority, quantum)

	runTime := t.CPUBurst
	if runTime > quantum {
		runTime = quantum
	}

	timer := time.NewTimer(runTime)
	select {
	case <-timer.C:
		t.CPUBurst -= runTime
		fmt.Printf("[T=%.2fs] PID=%d QUANTUM DONE | Round Robin\n", metrics.GetRelativeTime(), t.PID)
	case <-ps.interrupt:
		timer.Stop()
		fmt.Printf("[T=%.2fs] PID=%d PRE-EMPTED | Round Robin\n", metrics.GetRelativeTime(), t.PID)
		t.CPUBurst -= (runTime / 2)
	}

	ps.mu.Lock()
	ps.runningTask = nil
	if t.CPUBurst > 0 {
		if t.Priority <= 3 {
			ps.highQueue = append(ps.highQueue, *t)
		} else if t.Priority <= 7 {
			ps.mediumQueue = append(ps.mediumQueue, *t)
		} else {
			ps.lowQueue = append(ps.lowQueue, *t)
		}
		ps.cond.Signal()
	} else {
		fmt.Printf("[T=%.2fs] PID=%d DONE | Round Robin\n", metrics.GetRelativeTime(), t.PID)
		atomic.AddInt32(&completedTasksRoundRobin, 1)
		metrics.RecordCompletion(*t)
	}
	ps.mu.Unlock()
}

func (ps *RoundRobinScheduler) Shutdown() {
	ps.shutdown <- true
	ps.mu.Lock()
	ps.cond.Broadcast()
	ps.mu.Unlock()
}

func GetCompletedTasksRoundRobin() int32 {
	return atomic.LoadInt32(&completedTasksRoundRobin)
}
