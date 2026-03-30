package models

// Road defines a directed edge in the city's traffic graph.
type Road struct {
	From            IntersectionID
	To              IntersectionID
	Direction       string  // "NS" or "EW"
	Distance        float64 // in km
	CongestionLevel int     // 1 to 10
	SpeedLimit      int     // in km/h
}
