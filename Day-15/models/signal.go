package models

// SignalState represents the possible states for traffic signals.
type SignalState string

const (
	GreenNS  SignalState = "GREEN-NS" // North-South Green
	YellowNS SignalState = "YELLOW-NS"
	GreenEW  SignalState = "GREEN-EW" // East-West Green
	YellowEW SignalState = "YELLOW-EW"
)
