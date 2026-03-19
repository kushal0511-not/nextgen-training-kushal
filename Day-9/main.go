package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/nextgen-training-kushal/Day-9/internal/metrics"
	"github.com/nextgen-training-kushal/Day-9/internal/sensor"
	"github.com/nextgen-training-kushal/Day-9/models"
)

func main() {
	// Create a context that is cancelled on ctrl+c
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	var wgData sync.WaitGroup
	var wgAlerts sync.WaitGroup

	readings := make(chan models.SensorReading, 10)
	alerts := make(chan models.Alert, 100)

	// Start generators
	for i := 0; i < 5; i++ {
		wgData.Add(1)
		go func() {
			defer wgData.Done()
			sensor.GenerateReading(ctx, readings)
		}()
	}

	// Start processors
	for i := 0; i < 2; i++ {
		wgData.Add(1)
		go func() {
			defer wgData.Done()
			sensor.ProcessReading(readings, alerts)
		}()
	}

	// Start Alert Handler (Deduplicates and logs alerts)
	wgAlerts.Add(1)
	go sensor.StartAlertHandler(ctx, alerts, &wgAlerts)

	// Start Metrics Reporter
	wgAlerts.Add(1)
	go func() {
		defer wgAlerts.Done()
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				metrics.DefaultMetrics.UpdateQueueStatus(len(readings), cap(readings))
				metrics.DefaultMetrics.Report()
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for context cancellation (e.g., Ctrl+C)
	<-ctx.Done()
	fmt.Println("\nShutdown signal received. Cleaning up...")

	// Close the readings channel first so generators and processors finish
	close(readings)
	wgData.Wait()
	fmt.Println("Processors finished. Closing alerts channel...")

	// Close the alerts channel so the handler can finish processing remaining alerts
	close(alerts)
	wgAlerts.Wait()
	fmt.Println("All goroutines finished. Exit.")
}
