package sensor

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"

	metrics "github.com/nextgen-training-kushal/Day-9/internal/metrics"
	"github.com/nextgen-training-kushal/Day-9/models"
)

var (
	mu             sync.Mutex
	data           map[models.SensorType][]models.SensorReading
	stats          map[models.SensorType]*models.Stats
	lastAlertTimes map[float64]time.Time // Added for per-sensor deduplication
)

var Readings chan models.SensorReading

func init() {
	Readings = make(chan models.SensorReading, 10)
	data = make(map[models.SensorType][]models.SensorReading)
	stats = make(map[models.SensorType]*models.Stats)
	lastAlertTimes = make(map[float64]time.Time) // Initialized lastAlertTimes
	// Initialize stats for each type
	types := []models.SensorType{models.Temperature, models.Humidity, models.Motion, models.Pressure, models.Light}
	for _, t := range types {
		stats[t] = &models.Stats{}
	}
}

func GenerateReading(ctx context.Context, readings chan models.SensorReading) {
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	// Randomly generate sensor readings for random type
	types := []models.SensorType{models.Temperature, models.Humidity, models.Motion, models.Pressure, models.Light}

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Generator stopping...")
			return
		case <-ticker.C:
			if float64(len(readings)) >= float64(cap(readings))*0.85 {
				fmt.Println("Queue 85% full — applying backpressure")
			}

			sensorID := rand.Intn(5) + 1
			sensortype := rand.Intn(len(types))
			value := float64(rand.Intn(100))
			location := strconv.Itoa(rand.Intn(10))

			reading := models.SensorReading{
				SensorID:  float64(sensorID),
				Type:      types[sensortype],
				Value:     value,
				Timestamp: float64(time.Now().Unix()),
				Location:  "Room " + location,
			}

			select {
			case readings <- reading:
				mu.Lock()
				data[types[sensortype]] = append(data[types[sensortype]], reading)
				mu.Unlock()
			case <-ctx.Done():
				fmt.Println("Generator stopping while sending...")
				return
			default:
				fmt.Println("Channel is full, skipping reading")
			}
		}
	}
}

func ProcessReading(readings chan models.SensorReading, alerts chan<- models.Alert) {
	for reading := range readings {
		start := time.Now()
		if !reading.Validate() {
			fmt.Printf("❌ Invalid reading received: %+v\n", reading)
			continue
		}

		// Detect Anomalies
		reading.CheckAnomalies(alerts)

		// Update Rolling Average
		mu.Lock()
		s := stats[reading.Type]
		s.Sum += reading.Value
		s.Count++
		avg := s.Sum / float64(s.Count)
		mu.Unlock()
		metrics.DefaultMetrics.RecordReading(time.Since(start))
		fmt.Printf("✅ Processed %s: Value=%.2f, Rolling Average=%.2f\n", reading.Type, reading.Value, avg)
	}
}

func StartAlertHandler(ctx context.Context, alerts <-chan models.Alert, wg *sync.WaitGroup) {
	defer wg.Done()
	// The global 'mu' is used for 'lastAlertTimes' to ensure thread safety.
	// The local 'mu' was removed to avoid confusion and ensure global state protection.

	for {
		select {
		case alert, ok := <-alerts:
			if !ok {
				fmt.Println("Alert channel closed, handler stopping...")
				return
			}

			mu.Lock()
			// Deduplicate based on SensorID, not SensorType
			if lastTime, found := lastAlertTimes[alert.SensorID]; found && alert.Time.Sub(lastTime) < 60*time.Second {
				mu.Unlock()
				continue // Deduplicate
			}
			lastAlertTimes[alert.SensorID] = time.Now() // Update last alert time for this sensor
			mu.Unlock()

			metrics.DefaultMetrics.RecordAlert(alert.Type) // Record the alert
			fmt.Printf("🚨 ALERT [%s] Sensor ID %.0f: %s at %v\n",
				alert.Severity, alert.SensorID, alert.Message, alert.Time.Format(time.RFC822))

		case <-ctx.Done():
			fmt.Println("Alert handler stopping due to context cancellation...")
			return
		}
	}
}
