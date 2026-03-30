package cli

import (
	"fmt"
	"strings"

	"github.com/nextgen-training-kushal/Day-15/models"
	"github.com/nextgen-training-kushal/Day-15/traffic"
)

// congestionColor returns an ASCII color code based on level (1-10)
func congestionColor(level int) string {
	switch {
	case level <= 3:
		return "\033[32m" // green
	case level <= 6:
		return "\033[33m" // yellow
	default:
		return "\033[31m" // red
	}
}

const reset = "\033[0m"

// PrintCityMap prints a text-based overview of intersections and their outbound congestion.
func PrintCityMap(city *traffic.CityModel) {
	city.AdjMu.RLock()
	defer city.AdjMu.RUnlock()

	fmt.Println("\n┌──────────────── CITY MAP (Congestion) ────────────────┐")
	numIntersections := len(city.AdjList)
	for i := 1; i <= numIntersections; i++ {
		id := models.IntersectionID(i)
		roads := city.AdjList[id]
		if len(roads) == 0 {
			continue
		}
		parts := make([]string, 0, len(roads))
		for _, r := range roads {
			col := congestionColor(r.CongestionLevel)
			parts = append(parts, fmt.Sprintf("%s→%d(%d)%s", col, r.To, r.CongestionLevel, reset))
		}
		fmt.Printf("│ Int %2d: %s\n", i, strings.Join(parts, "  "))
	}
	fmt.Println("└────────────────────────────────────────────────────────┘")
}

// PrintVehicles prints a formatted table of all vehicle positions and statuses.
func PrintVehicles(city *traffic.CityModel) {
	city.VehicleMu.RLock()
	defer city.VehicleMu.RUnlock()

	if len(city.Vehicles) == 0 {
		return
	}
	fmt.Println("\n┌──────────────── VEHICLE TRACKING ──────────────────────┐")
	fmt.Printf("│ %-8s  %-4s %-4s %-22s %-4s │\n", "Plate", "Cur", "Dst", "Status", "Emer")
	fmt.Println("│ " + strings.Repeat("─", 52) + " │")
	for _, v := range city.Vehicles {
		v.Mu.RLock()
		em := " "
		if v.IsEmergency {
			em = "🚨"
		}
		fmt.Printf("│ %-8s  %-4d %-4d %-22s %-4s │\n",
			v.Plate, v.Current, v.Destination, v.Status, em)
		v.Mu.RUnlock()
	}
	fmt.Println("└────────────────────────────────────────────────────────┘")
}

// PrintSignals prints signal state for the first 6 intersections.
func PrintSignals(city *traffic.CityModel) {
	fmt.Println("\n┌──────────────── SIGNAL STATUS ─────────────────────────┐")
	for id := models.IntersectionID(1); id <= 6; id++ {
		inter, ok := city.Intersections[id]
		if !ok {
			continue
		}
		inter.Mu.RLock()
		sig := inter.CurrentSignal
		pre := inter.Preempted
		inter.Mu.RUnlock()
		preStr := ""
		if pre {
			preStr = " ⚡PREEMPTED"
		}
		fmt.Printf("│  Int %2d: %-12s%s\n", id, sig, preStr)
	}
	fmt.Println("└────────────────────────────────────────────────────────┘")
}
