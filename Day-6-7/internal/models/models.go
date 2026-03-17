package models

import "time"

type DriverStatus string

const (
	DriverStatusAvailable DriverStatus = "available"
	DriverStatusBusy      DriverStatus = "busy"
	DriverStatusOffline   DriverStatus = "offline"
)

type RideStatus string

const (
	RideStatusRequested RideStatus = "requested"
	RideStatusAccepted  RideStatus = "accepted"
	RideStatusStarted   RideStatus = "started"
	RideStatusCompleted RideStatus = "completed"
	RideStatusCancelled RideStatus = "cancelled"
)

type Location struct {
	Lat float64
	Lng float64
}

type Driver struct {
	ID          string
	Name        string
	Location    Location
	Status      DriverStatus
	Rating      float64
	RideHistory []string
}

type Rider struct {
	ID            string
	Name          string
	Location      Location
	PaymentMethod string
}

type Ride struct {
	ID          string
	RiderID     string
	DriverID    string
	Pickup      Location
	Dropoff     Location
	Status      RideStatus
	RequestTime time.Time
	StartTime   time.Time
	EndTime     time.Time
	Fare        float64
}
