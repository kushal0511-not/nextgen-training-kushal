package metrics

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nextgen-training-kushal/Day-9/models"
)

type Metrics struct {
	readingsProcessed uint64
	totalLatency      uint64 // in microseconds
	queueSize         uint64
	queueCap          uint64
	alertsPerSensor   map[models.SensorType]uint64
	mu                sync.Mutex
	startTime         time.Time
}

var DefaultMetrics = &Metrics{
	alertsPerSensor: make(map[models.SensorType]uint64),
	startTime:       time.Now(),
}

func (m *Metrics) RecordReading(latency time.Duration) {
	atomic.AddUint64(&m.readingsProcessed, 1)
	atomic.AddUint64(&m.totalLatency, uint64(latency.Microseconds()))
}

func (m *Metrics) UpdateQueueStatus(size, capacity int) {
	atomic.StoreUint64(&m.queueSize, uint64(size))
	atomic.StoreUint64(&m.queueCap, uint64(capacity))
}

func (m *Metrics) RecordAlert(Type models.SensorType) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.alertsPerSensor[Type]++
}

func (m *Metrics) Report() {
	processed := atomic.LoadUint64(&m.readingsProcessed)
	latency := atomic.LoadUint64(&m.totalLatency)
	qSize := atomic.LoadUint64(&m.queueSize)
	qCap := atomic.LoadUint64(&m.queueCap)

	duration := time.Since(m.startTime).Seconds()
	rps := float64(processed) / duration

	avgLatency := 0.0
	if processed > 0 {
		avgLatency = float64(latency) / float64(processed)
	}

	qUtilization := 0.0
	if qCap > 0 {
		qUtilization = (float64(qSize) / float64(qCap)) * 100
	}

	fmt.Println("\n--- 📊 Statistics Report ---")
	fmt.Printf("⏱️  Readings Processed: %d (%.2f ops/sec)\n", processed, rps)
	fmt.Printf("🧪 Average Latency: %.2f µs\n", avgLatency)
	fmt.Printf("📦 Queue Utilization: %.2f%% (%d/%d)\n", qUtilization, qSize, qCap)

	m.mu.Lock()
	fmt.Println("🚨 Alerts per Sensor Type:")
	for id, count := range m.alertsPerSensor {
		fmt.Printf("   - %s: %d alerts\n", id, count)
	}
	m.mu.Unlock()
	fmt.Println("---------------------------")
}
