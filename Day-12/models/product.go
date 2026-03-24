package models

import "time"

// Product represents the product model
type Product struct {
	ID        string
	Name      string
	Category  string
	Price     float64
	Rating    float64
	Stock     int
	Tags      []string
	CreatedAt time.Time
}
