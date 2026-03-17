package driver

import (
	"fmt"
	"math"
	"ride-sharing/internal/models"
	"sync"
)

type DriverStore interface {
	Register(driver *models.Driver) error
	UpdateLocation(id string, loc models.Location) error
	UpdateStatus(id string, status models.DriverStatus) error
	GetByID(id string) (*models.Driver, error)
	FindNearest(loc models.Location, radius float64) ([]*models.Driver, error)
	GetBusiestZones() map[string]int
}

const ZoneSize = 5

type MemoryStore struct {
	drivers    map[string]*models.Driver
	mu         sync.RWMutex
	zoneScores map[string]int
	zones      map[string][]*models.Driver
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		drivers:    make(map[string]*models.Driver),
		zoneScores: make(map[string]int),
		zones:      make(map[string][]*models.Driver),
	}
}

func (s *MemoryStore) getZone(loc models.Location) string {
	// Simple grid: 0.1 degree resolution (~11km approx, or use 0.01 for ~1km)
	// Let's use 0.05 for a decent balance
	lat := math.Floor(loc.Lat/ZoneSize) * ZoneSize
	lng := math.Floor(loc.Lng/ZoneSize) * ZoneSize
	return fmt.Sprintf("%.0f:%.0f", lat, lng)
}

func (s *MemoryStore) Register(d *models.Driver) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.drivers[d.ID]; exists {
		return fmt.Errorf("driver %s already exists", d.ID)
	}

	s.drivers[d.ID] = d
	zone := s.getZone(d.Location)
	s.zones[zone] = append(s.zones[zone], d)
	return nil
}

func (s *MemoryStore) UpdateLocation(id string, loc models.Location) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, exists := s.drivers[id]
	if !exists {
		return fmt.Errorf("driver %s not found", id)
	}

	oldZone := s.getZone(d.Location)
	newZone := s.getZone(loc)

	if oldZone != newZone {
		for i, driver := range s.zones[oldZone] {
			if driver.ID == id {
				s.zones[oldZone] = append(s.zones[oldZone][:i], s.zones[oldZone][i+1:]...)
				break
			}
		}
		s.zones[newZone] = append(s.zones[newZone], d)
	}

	d.Location = loc
	return nil
}

func (s *MemoryStore) UpdateStatus(id string, status models.DriverStatus) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	d, exists := s.drivers[id]
	if !exists {
		return fmt.Errorf("driver %s not found", id)
	}

	d.Status = status
	return nil
}

func (s *MemoryStore) GetByID(id string) (*models.Driver, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	d, exists := s.drivers[id]
	if !exists {
		return nil, fmt.Errorf("driver %s not found", id)
	}
	return d, nil
}

func (s *MemoryStore) FindNearest(loc models.Location, radius float64) ([]*models.Driver, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var nearest []*models.Driver
	zones := s.getZones(loc)
	for _, zone := range zones {
		for _, d := range s.zones[zone] {
			if d.Status != models.DriverStatusAvailable {
				continue
			}

			dist := s.euclideanDistance(loc, d.Location)
			if dist <= radius {
				nearest = append(nearest, d)
			}
		}
	}

	return nearest, nil
}

func (s *MemoryStore) euclideanDistance(l1, l2 models.Location) float64 {
	// Simple Euclidean for flat grid (as requested)
	// In reality, 1 degree lat is ~111km, lng varies.
	// We'll treat units as km for the sake of the exercise.
	return math.Pow(l1.Lat-l2.Lat, 2) + math.Pow(l1.Lng-l2.Lng, 2)
}

func (s *MemoryStore) GetBusiestZones() map[string]int {
	// This might be better tracked in dispatch, but we'll provide a getter.
	return s.zoneScores
}

func (s *MemoryStore) IncrementZoneScore(loc models.Location) {
	s.mu.Lock()
	defer s.mu.Unlock()
	zone := s.getZone(loc)
	s.zoneScores[zone]++
}

func (s *MemoryStore) DecrementZoneScore(loc models.Location) {
	s.mu.Lock()
	defer s.mu.Unlock()
	zone := s.getZone(loc)
	if s.zoneScores[zone] > 0 {
		s.zoneScores[zone]--
	}
	if s.zoneScores[zone] == 0 {
		delete(s.zoneScores, zone)
	}
}

func (s *MemoryStore) getZones(loc models.Location) []string {
	var zones []string
	lat := math.Floor(loc.Lat/ZoneSize) * ZoneSize
	lng := math.Floor(loc.Lng/ZoneSize) * ZoneSize
	for i := -1; i <= 1; i++ {
		for j := -1; j <= 1; j++ {
			zone := fmt.Sprintf("%.0f:%.0f", lat+float64(i)*ZoneSize, lng+float64(j)*ZoneSize)
			if _, exists := s.zones[zone]; exists {
				zones = append(zones, zone)
			}
		}
	}
	return zones
}
