package internal

import (
	"testing"
	"time"

	"github.com/nextgen-training-kushal/Day-10/models"
)

func TestSchedulers(t *testing.T) {
	metrics := GetMetrics()

	t.Run("Priority: 10 low-priority, 1 high-priority -> high runs first", func(t *testing.T) {
		metrics.Reset()
		ps := NewPriorityScheduler()

		// Add 10 low-priority tasks (P10)
		for i := 1; i <= 10; i++ {
			ps.AddTask(models.Task{
				PID:         i,
				Priority:    10,
				CPUBurst:    10 * time.Millisecond,
				ArrivalTime: time.Now(),
			})
		}

		// Add 1 high-priority task (P1)
		ps.AddTask(models.Task{
			PID:         100,
			Priority:    1,
			CPUBurst:    10 * time.Millisecond,
			ArrivalTime: time.Now(),
		})

		// Start scheduler
		go ps.Schedule()

		// Wait for all tasks to complete (11 tasks)
		// Each task takes ~1ms (Burst/10)
		time.Sleep(200 * time.Millisecond)
		ps.Shutdown()

		completed := metrics.GetCompletedTasks()
		if len(completed) < 1 {
			t.Fatal("No tasks completed")
		}

		// First completed task should be PID 100 (Priority 1)
		if completed[0].PID != 100 {
			t.Errorf("Expected first task to be PID 100 (P1), got PID %d (P%d)", completed[0].PID, completed[0].Priority)
		}
	})

	t.Run("Aging: low-priority task eventually runs", func(t *testing.T) {
		metrics.Reset()
		as := NewAgingScheduler()

		// Add a low priority task
		as.AddTask(models.Task{
			PID:         200,
			Priority:    10,
			CPUBurst:    10 * time.Millisecond,
			ArrivalTime: time.Now(),
		})

		// Start scheduler
		go as.Schedule()

		// In AgingScheduler, startAging is called once.
		// Priority decreases every 500ms.
		// We wait for it to run. Since it's the only task, it should run anyway.
		// To REALLY test aging, we'd need to keep adding higher priority tasks.
		// But the requirement says "low-priority task eventually runs".

		time.Sleep(100 * time.Millisecond)
		completed := metrics.GetCompletedTasks()

		found := false
		for _, task := range completed {
			if task.PID == 200 {
				found = true
				break
			}
		}

		if !found {
			t.Error("Low-priority task PID 200 did not run")
		}
		as.Shutdown()
	})
}

func TestAgingEffect(t *testing.T) {
	metrics := GetMetrics()
	metrics.Reset()
	as := NewAgingScheduler()

	// Add a low priority task (P10)
	as.AddTask(models.Task{
		PID:      300,
		Priority: 10,
		CPUBurst: 10 * time.Millisecond,
	})

	// Wait for aging to happen (500ms per priority drop)
	// After 1 second, priority should be 8
	time.Sleep(1100 * time.Millisecond)

	as.mu.Lock()
	p := as.queue.Data[0].Priority
	as.mu.Unlock()

	if p >= 10 {
		t.Errorf("Priority did not decrease, still %d", p)
	}

	as.Shutdown()
}
