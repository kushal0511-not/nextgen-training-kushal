package sensor

import (
	"testing"
	"time"

	"github.com/nextgen-training-kushal/Day-9/models"
)

func TestCheckAnomalies(t *testing.T) {
	tests := []struct {
		name     string
		reading  models.SensorReading
		expects  bool
		severity models.Severity
	}{
		{
			name: "Normal Temperature",
			reading: models.SensorReading{
				Type:     models.Temperature,
				Value:    25.0,
				Location: "Living Room",
			},
			expects: false,
		},
		{
			name: "High Temperature",
			reading: models.SensorReading{
				Type:     models.Temperature,
				Value:    55.0,
				Location: "Kitchen",
			},
			expects:  true,
			severity: models.Warning,
		},
		{
			name: "Normal Humidity",
			reading: models.SensorReading{
				Type:     models.Humidity,
				Value:    45.0,
				Location: "Basement",
			},
			expects: false,
		},
		{
			name: "Critical Humidity",
			reading: models.SensorReading{
				Type:     models.Humidity,
				Value:    98.0,
				Location: "Basement",
			},
			expects:  true,
			severity: models.Critical,
		},
		{
			name: "Motion Normal Hours",
			reading: models.SensorReading{
				Type:      models.Motion,
				Value:     1.0,
				Location:  "Patio",
				Timestamp: float64(time.Date(2023, 10, 27, 12, 0, 0, 0, time.UTC).Unix()),
			},
			expects: false,
		},
		{
			name: "Motion Restricted Hours 11 PM",
			reading: models.SensorReading{
				Type:      models.Motion,
				Value:     1.0,
				Location:  "Patio",
				Timestamp: float64(time.Date(2023, 10, 27, 23, 0, 0, 0, time.UTC).Unix()),
			},
			expects:  true,
			severity: models.Warning,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			alerts := make(chan models.Alert, 1)
			tt.reading.CheckAnomalies(alerts)

			select {
			case alert := <-alerts:
				if !tt.expects {
					t.Errorf("expected no alert, but got: %v", alert.Message)
				}
				if alert.Severity != tt.severity {
					t.Errorf("expected severity %v, but got: %v", tt.severity, alert.Severity)
				}
			default:
				if tt.expects {
					t.Error("expected alert, but none received")
				}
			}
		})
	}
}
