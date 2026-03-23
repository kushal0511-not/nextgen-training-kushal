package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"runtime/pprof"
	"sync"
	"syscall"
	"time"

	"github.com/nextgen-training-kushal/Day-10/internal"
	"github.com/nextgen-training-kushal/Day-10/models"
)

func taskProducer(ctx context.Context, schedulers []internal.Scheduler, wg *sync.WaitGroup) {
	defer wg.Done()
	pid := 1
	timer := time.NewTimer((time.Duration(200+rand.Intn(600)) * time.Millisecond))
	defer timer.Stop()
	for {
		timer.Reset(time.Duration(200+rand.Intn(600)) * time.Millisecond)
		select {
		case <-ctx.Done():
			fmt.Println("\n[Producer] Stopping task generation...")
			return
		case <-timer.C:
			task := models.Task{
				PID:         pid,
				Name:        fmt.Sprintf("Task-%d", pid),
				Priority:    1 + rand.Intn(10), // 1 to 10
				CPUBurst:    time.Duration(50+rand.Intn(450)) * time.Millisecond,
				ArrivalTime: time.Now(),
			}
			pid++

			// Distribute tasks to all schedulers for simultaneous testing
			for _, s := range schedulers {
				s.AddTask(task)
			}
			fmt.Printf("[Producer] Added: %s (Priority: %d, Burst: %v) at T=%.2fs\n",
				task.Name, task.Priority, task.CPUBurst, internal.GetMetrics().GetRelativeTime())
		}
	}
}

func main() {
	// 1. CPU Profile
	cpuFile, err := os.Create("cpu.prof")
	if err == nil {
		pprof.StartCPUProfile(cpuFile)
		defer pprof.StopCPUProfile()
	}

	// 2. Heap Profile (Snapshot at end)
	defer func() {
		heapFile, err := os.Create("heap.prof")
		if err == nil {
			pprof.WriteHeapProfile(heapFile)
			heapFile.Close()
		}
	}()

	// Initialize Schedulers
	ps := internal.NewPriorityScheduler()
	as := internal.NewAgingScheduler()
	rr := internal.NewRoundRobinScheduler()

	schedulers := []internal.Scheduler{ps, as, rr}

	// Initialize Metrics
	metrics := internal.GetMetrics()
	metrics.Reset()

	// Context for producer shutdown
	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	fmt.Println("===============================================================")
	fmt.Println("--- OS Scheduler Simulation: Metrics & Timeline Visualization ---")
	fmt.Println("===============================================================")

	// Start Schedulers in background
	for _, s := range schedulers {
		go s.Schedule()
	}

	// Start Producer
	wg.Add(1)
	go taskProducer(ctx, schedulers, &wg)

	// Graceful Shutdown Setup
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	fmt.Println("\nSimulation Running (Press Ctrl+C to initiate graceful shutdown and see report)")

	// Wait for interrupt
	<-sigChan
	fmt.Printf("\nInterrupt received. Closing task production and draining queues...\n")

	// 1. Stop Producer
	cancel()
	wg.Wait()

	// 2. Notify Schedulers to Shutdown (finish remaining tasks)
	for _, s := range schedulers {
		s.Shutdown()
	}

	// Wait a bit for schedulers to finish and print their exit messages
	time.Sleep(2 * time.Second)

	// 3. Final Report
	metrics.Report()

	fmt.Println("\nSimulation finished.")
}
