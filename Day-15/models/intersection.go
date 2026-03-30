package models

import (
	"sync"
)

// IntersectionID represents a unique identifier for an intersection.
type IntersectionID int

// Intersection represents a traffic node containing signal states.
type Intersection struct {
	ID            IntersectionID
	CurrentSignal SignalState
	InboundNS     []IntersectionID
	InboundEW     []IntersectionID
	Preempted     bool // true when an emergency vehicle has claimed this intersection
	Mu            sync.RWMutex
}
