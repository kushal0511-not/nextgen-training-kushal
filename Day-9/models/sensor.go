package models

import (
	"fmt" // Added sync/atomic import
	"time"
)

const (
	Info     Severity = "info"
	Warning  Severity = "warning"
	Critical Severity = "critical"
)

type SensorType string

type Severity string

type Alert struct {
	SensorID        float64
	Type            SensorType
	Message         string
	Severity        Severity
	Time            time.Time
	SevereTimeStamp time.Time
}

type Stats struct {
	Sum             float64
	Count           uint64
	SevereTimeStamp time.Time
}
type SensorReading struct {
	SensorID  float64
	Type      SensorType
	Value     float64
	Timestamp float64
	Location  string
}

const (
	Temperature SensorType = "temperature"
	Humidity    SensorType = "humidity"
	Motion      SensorType = "motion"
	Pressure    SensorType = "pressure"
	Light       SensorType = "light"
)

func (s SensorReading) Validate() bool {
	if s.Value < 0 && s.Type != Temperature { // Temperature can be negative
		return false
	}
	if s.Location == "" {
		return false
	}
	return true
}

func (s SensorReading) CheckAnomalies(alerts chan<- Alert) {
	switch s.Type {

	case Temperature:
		if s.Value > 50 {
			alerts <- Alert{
				SensorID: s.SensorID,
				Type:     s.Type,
				Message:  fmt.Sprintf("High Temperature detected! Value: %.2f°C at %s", s.Value, s.Location),
				Severity: Warning,
				Time:     time.Unix(int64(s.Timestamp), 0),
			}
		}
	case Humidity:
		if s.Value > 95 {
			alerts <- Alert{
				SensorID: s.SensorID,
				Type:     s.Type,
				Message:  fmt.Sprintf("Critical Humidity! Value: %.2f%% at %s", s.Value, s.Location),
				Severity: Critical,
				Time:     time.Unix(int64(s.Timestamp), 0),
			}
		}
	case Motion:
		// Restricted hours: 10 PM (22) to 6 AM (6)
		t := time.Unix(int64(s.Timestamp), 0)
		hour := t.Hour()
		if hour >= 22 || hour < 6 {
			alerts <- Alert{
				SensorID: s.SensorID,
				Type:     s.Type,
				Message:  fmt.Sprintf("Restricted Motion detected! Time: %02d:%02d at %s", hour, t.Minute(), s.Location),
				Severity: Warning,
				Time:     t,
			}
		}
	}
}
