package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/nextgen-training-kushal/Day-15/models"
	"go.uber.org/zap"
)

// — helpers —

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, map[string]string{"error": msg})
}

// POST /vehicles
// Body: {"plate":"V-99","from":1,"to":20}
func (s *Server) RegisterVehicle(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Plate string `json:"plate"`
		From  int    `json:"from"`
		To    int    `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Plate == "" || req.From == 0 || req.To == 0 {
		writeErr(w, http.StatusBadRequest, "plate, from, and to are required")
		return
	}

	start, end := models.IntersectionID(req.From), models.IntersectionID(req.To)
	if err := s.city.RegisterVehicle(req.Plate, start, end); err != nil {
		writeErr(w, http.StatusConflict, err.Error())
		return
	}

	s.city.VehicleMu.RLock()
	v := s.city.Vehicles[req.Plate]
	s.city.VehicleMu.RUnlock()

	v.Mu.RLock()
	path := v.Path
	v.Mu.RUnlock()

	s.logger.Info("vehicle registered via API", zap.String("plate", req.Plate))
	writeJSON(w, http.StatusCreated, map[string]any{
		"plate": req.Plate,
		"from":  req.From,
		"to":    req.To,
		"path":  path,
	})
}

// GET /route?from=1&to=20
func (s *Server) GetRoute(w http.ResponseWriter, r *http.Request) {
	fromStr := r.URL.Query().Get("from")
	toStr := r.URL.Query().Get("to")
	from, err1 := strconv.Atoi(fromStr)
	to, err2 := strconv.Atoi(toStr)
	if err1 != nil || err2 != nil || from == 0 || to == 0 {
		writeErr(w, http.StatusBadRequest, "from and to must be valid intersection IDs")
		return
	}

	path, timeHours := s.city.Dijkstra(models.IntersectionID(from), models.IntersectionID(to))
	if path == nil {
		writeErr(w, http.StatusNotFound, fmt.Sprintf("no path from %d to %d", from, to))
		return
	}

	s.logger.Info("route calculated", zap.Int("from", from), zap.Int("to", to))
	writeJSON(w, http.StatusOK, map[string]any{
		"from":     from,
		"to":       to,
		"path":     path,
		"time_min": timeHours * 60,
	})
}

// POST /emergency
// Body: {"plate":"E-911","from":1,"to":20}
func (s *Server) DispatchEmergency(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Plate string `json:"plate"`
		From  int    `json:"from"`
		To    int    `json:"to"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
		return
	}
	if req.Plate == "" || req.From == 0 || req.To == 0 {
		writeErr(w, http.StatusBadRequest, "plate, from, and to are required")
		return
	}

	if err := s.city.RegisterEmergencyVehicle(r.Context(), req.Plate,
		models.IntersectionID(req.From), models.IntersectionID(req.To)); err != nil {
		writeErr(w, http.StatusConflict, err.Error())
		return
	}

	s.city.VehicleMu.RLock()
	v := s.city.Vehicles[req.Plate]
	s.city.VehicleMu.RUnlock()

	v.Mu.RLock()
	path := v.Path
	v.Mu.RUnlock()

	// Report which intersections were preempted
	preempted := []int{}
	for _, id := range path {
		inter := s.city.Intersections[id]
		inter.Mu.RLock()
		if inter.Preempted {
			preempted = append(preempted, int(id))
		}
		inter.Mu.RUnlock()
	}

	s.logger.Warn("emergency dispatched via API", zap.String("plate", req.Plate))
	writeJSON(w, http.StatusCreated, map[string]any{
		"plate":             req.Plate,
		"path":              path,
		"preempted_signals": preempted,
	})
}

// GET /congestion
func (s *Server) GetCongestion(w http.ResponseWriter, r *http.Request) {
	type roadInfo struct {
		From    int     `json:"from"`
		To      int     `json:"to"`
		Level   int     `json:"level"`
		AvgLast float64 `json:"avg_last_10"`
	}

	s.city.AdjMu.RLock()
	result := []roadInfo{}
	for _, roads := range s.city.AdjList {
		for _, road := range roads {
			key := models.RoadKey{From: road.From, To: road.To}
			avg := s.city.SlidingWindowAverage(key)
			result = append(result, roadInfo{
				From:    int(road.From),
				To:      int(road.To),
				Level:   road.CongestionLevel,
				AvgLast: avg,
			})
		}
	}
	s.city.AdjMu.RUnlock()

	writeJSON(w, http.StatusOK, result)
}

// GET /signals/{id}
func (s *Server) GetSignal(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		writeErr(w, http.StatusBadRequest, "id must be a number")
		return
	}

	inter, ok := s.city.Intersections[models.IntersectionID(id)]
	if !ok {
		writeErr(w, http.StatusNotFound, fmt.Sprintf("intersection %d not found", id))
		return
	}

	inter.Mu.RLock()
	sig := inter.CurrentSignal
	pre := inter.Preempted
	inter.Mu.RUnlock()

	writeJSON(w, http.StatusOK, map[string]any{
		"id":        id,
		"signal":    sig,
		"preempted": pre,
	})
}

// GET /stats
func (s *Server) GetStats(w http.ResponseWriter, r *http.Request) {
	s.city.VehicleMu.RLock()
	total := len(s.city.Vehicles)
	arrived := 0
	emergency := 0
	for _, v := range s.city.Vehicles {
		v.Mu.RLock()
		if v.Status == "ARRIVED" || v.Status == "ARRIVED (EMERGENCY)" {
			arrived++
		}
		if v.IsEmergency {
			emergency++
		}
		v.Mu.RUnlock()
	}
	s.city.VehicleMu.RUnlock()

	// Compute average congestion across all roads
	s.city.AdjMu.RLock()
	totalCong, roadCount := 0, 0
	for _, roads := range s.city.AdjList {
		for _, r := range roads {
			totalCong += r.CongestionLevel
			roadCount++
		}
	}
	s.city.AdjMu.RUnlock()

	avgCong := 0.0
	if roadCount > 0 {
		avgCong = float64(totalCong) / float64(roadCount)
	}

	top5 := s.city.Top5CongestedRoads()

	writeJSON(w, http.StatusOK, map[string]any{
		"total_vehicles":   total,
		"arrived":          arrived,
		"emergency_active": emergency,
		"roads":            roadCount,
		"avg_congestion":   fmt.Sprintf("%.1f", avgCong),
		"top5_congested":   top5,
	})
}
