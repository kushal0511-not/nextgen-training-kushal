package models

import (
	"sync"
)

// Vehicle represents a vehicle moving through the city.
type Vehicle struct {
	Plate       string
	Current     IntersectionID
	Destination IntersectionID
	Path        []IntersectionID
	Status      string // e.g., "MOVING", "ARRIVED", "WAITING_AT_SIGNAL"
	IsEmergency bool
	Mu          sync.RWMutex
}
