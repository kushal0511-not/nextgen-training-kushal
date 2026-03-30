package models

import "time"

// RoadKey uniquely identifies a directed road segment.
type RoadKey struct {
	From IntersectionID
	To   IntersectionID
}

// CongestionRecord is a single time-stamped congestion snapshot for a road.
type CongestionRecord struct {
	Timestamp time.Time
	Level     int
}

// CongestedRoad is used for sorted output of the top-N congested roads.
type CongestedRoad struct {
	Key           RoadKey
	AvgCongestion float64
}
