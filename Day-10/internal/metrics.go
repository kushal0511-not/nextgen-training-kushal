package internal

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nextgen-training-kushal/Day-10/models"
)

type Metrics struct {
	startTimeUnixNano int64
	contextSwitches   int32
	starvationCount   int32

	// Channels for lock-free slice operations
	completedTasks []models.Task
	taskChan       chan models.Task
	reportChan     chan chan struct{}
	resetChan      chan struct{}
	tasksReqChan   chan chan []models.Task
}

var (
	GlobalMetrics *Metrics
	once          sync.Once
)

func GetMetrics() *Metrics {
	once.Do(func() {
		GlobalMetrics = &Metrics{
			startTimeUnixNano: time.Now().UnixNano(),
			completedTasks:    make([]models.Task, 0),
			taskChan:          make(chan models.Task, 3), // Buffered to prevent blocking
			reportChan:        make(chan chan struct{}),
			resetChan:         make(chan struct{}),
			tasksReqChan:      make(chan chan []models.Task),
		}
		go GlobalMetrics.processLoop()
	})
	return GlobalMetrics
}

func (m *Metrics) processLoop() {
	for {
		select {
		case t := <-m.taskChan:
			m.completedTasks = append(m.completedTasks, t)
		case <-m.resetChan:
			m.completedTasks = make([]models.Task, 0)
			atomic.StoreInt32(&m.contextSwitches, 0)
			atomic.StoreInt32(&m.starvationCount, 0)
			atomic.StoreInt64(&m.startTimeUnixNano, time.Now().UnixNano())
		case done := <-m.reportChan:
			m.printReport()
			close(done)
		case req := <-m.tasksReqChan:
			// Return a copy to avoid race conditions
			tasks := make([]models.Task, len(m.completedTasks))
			copy(tasks, m.completedTasks)
			req <- tasks
		}
	}
}

func (m *Metrics) GetCompletedTasks() []models.Task {
	resp := make(chan []models.Task)
	m.tasksReqChan <- resp
	return <-resp
}

func (m *Metrics) Reset() {
	m.resetChan <- struct{}{}
}

func (m *Metrics) RecordCompletion(t models.Task) {
	t.CompletionTime = time.Now()
	m.taskChan <- t
}

func (m *Metrics) RecordContextSwitch() {
	atomic.AddInt32(&m.contextSwitches, 1)
}

func (m *Metrics) RecordStarvation() {
	atomic.AddInt32(&m.starvationCount, 1)
}

func (m *Metrics) GetRelativeTime() float64 {
	start := time.Unix(0, atomic.LoadInt64(&m.startTimeUnixNano))
	return time.Since(start).Seconds()
}

func (m *Metrics) Report() {
	done := make(chan struct{})
	m.reportChan <- done
	<-done
}

func (m *Metrics) printReport() {
	start := time.Unix(0, atomic.LoadInt64(&m.startTimeUnixNano))
	totalTime := time.Since(start).Seconds()
	if totalTime == 0 {
		totalTime = 0.001
	}

	fmt.Printf("\n========== SCHEDULING METRICS REPORT ==========\n")
	fmt.Printf("Total Time Elapsed: %.2fs\n", totalTime)
	fmt.Printf("Total Tasks Completed: %d\n", len(m.completedTasks))
	fmt.Printf("Throughput: %.2f tasks/sec\n", float64(len(m.completedTasks))/totalTime)

	ctxSwitches := atomic.LoadInt32(&m.contextSwitches)
	starvCount := atomic.LoadInt32(&m.starvationCount)

	fmt.Printf("Total Context Switches: %d\n", ctxSwitches)
	fmt.Printf("Total Starvation Events (>5s): %d\n", starvCount)

	// Wait time per priority level
	waitTimes := make(map[int][]time.Duration)
	for _, t := range m.completedTasks {
		waitTimes[t.Priority] = append(waitTimes[t.Priority], t.FirstScheduledTime.Sub(t.ArrivalTime))
	}

	fmt.Printf("\nAverage Wait Time by Priority Level:\n")
	for p := 1; p <= 10; p++ {
		times, ok := waitTimes[p]
		if !ok || len(times) == 0 {
			continue
		}
		var totalWait time.Duration
		for _, d := range times {
			totalWait += d
		}
		avgWait := totalWait / time.Duration(len(times))
		fmt.Printf("  Priority %2d: %v (Count: %d)\n", p, avgWait, len(times))
	}
	fmt.Printf("===============================================\n")
}
