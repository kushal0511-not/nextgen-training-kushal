package main

import (
	"bufio"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"strconv"
	"strings"
	"time"
)

const (
	TOTAL_SPOTS     = 500
	SPOTS_PER_FLOOR = 100
	TOTAL_FLOORS    = 5
	HOURLY_RATE     = 20
)

// ParkingSpot represents a single parking spot
type ParkingSpot struct {
	ID           int
	Floor        int
	Occupied     bool
	VehiclePlate string
	EntryTime    time.Time
}

var inputScanner = bufio.NewScanner(os.Stdin)

// Global parking array
var parkingGarage [TOTAL_SPOTS]ParkingSpot

// tracks the first potential free spot index for each floor
var nearestAvailable [TOTAL_FLOORS]int

// Initialize parking garage
func initializeParkingGarage() {
	for i := 0; i < TOTAL_FLOORS; i++ {
		nearestAvailable[i] = 0 // Initial relative index for each floor
	}

	for i := 0; i < TOTAL_SPOTS; i++ {
		floor := (i / SPOTS_PER_FLOOR) + 1
		spotInFloor := (i % SPOTS_PER_FLOOR) + 1
		parkingGarage[i] = ParkingSpot{
			ID:           spotInFloor,
			Floor:        floor,
			Occupied:     false,
			VehiclePlate: "",
			EntryTime:    time.Time{},
		}
	}
}

// EntryVehicle: Park a vehicle on the requested floor
func EntryVehicle(requestedFloor int, vehiclePlate string) error {
	// Validate floor
	if requestedFloor < 1 || requestedFloor > TOTAL_FLOORS {
		return fmt.Errorf("invalid floor: %d. Floors range from 1 to %d", requestedFloor, TOTAL_FLOORS)
	}

	// Validate vehicle plate
	if strings.TrimSpace(vehiclePlate) == "" {
		return fmt.Errorf("vehicle plate cannot be empty")
	}

	// Check if vehicle is already parked (prevent duplicates)
	normalizedPlate := strings.ToUpper(vehiclePlate)
	for i := 0; i < TOTAL_SPOTS; i++ {
		if parkingGarage[i].Occupied && strings.ToUpper(parkingGarage[i].VehiclePlate) == normalizedPlate {
			return fmt.Errorf("vehicle %s is already parked at Floor %d, Spot %d", vehiclePlate, parkingGarage[i].Floor, parkingGarage[i].ID)
		}
	}

	// Calculate floor range
	floorBaseIdx := (requestedFloor - 1) * SPOTS_PER_FLOOR
	startIdx := floorBaseIdx + nearestAvailable[requestedFloor-1]
	endIdx := floorBaseIdx + SPOTS_PER_FLOOR

	// Find nearest available spot on the floor starting from the tracked index
	for i := startIdx; i < endIdx; i++ {
		if !parkingGarage[i].Occupied {
			parkingGarage[i].Occupied = true
			parkingGarage[i].VehiclePlate = vehiclePlate
			parkingGarage[i].EntryTime = time.Now()

			// Update the nearest free spot for this floor
			nearestAvailable[requestedFloor-1] = (i - floorBaseIdx) + 1

			fmt.Printf("✓ Vehicle %s parked at Floor %d, Spot %d\n",
				vehiclePlate, parkingGarage[i].Floor, parkingGarage[i].ID)
			return nil
		}
	}

	// If we reached here, the floor might be full, or startIdx was too high
	return fmt.Errorf("no available spots on Floor %d", requestedFloor)
}

// ExitVehicle: Remove vehicle and calculate fee
func ExitVehicle(vehiclePlate string) error {
	// Validate vehicle plate
	if strings.TrimSpace(vehiclePlate) == "" {
		return fmt.Errorf("vehicle plate cannot be empty")
	}

	// Search for vehicle
	normalizedPlate := strings.ToUpper(vehiclePlate)
	for i := 0; i < TOTAL_SPOTS; i++ {
		if parkingGarage[i].Occupied && strings.ToUpper(parkingGarage[i].VehiclePlate) == normalizedPlate {
			// Calculate duration and fee
			duration := time.Since(parkingGarage[i].EntryTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60

			// Fee calculation (round up to nearest hour with minimum 1 hour)
			fee := hours
			if minutes > 0 || fee == 0 {
				fee++
			}
			totalFee := fee * HOURLY_RATE

			// Free the spot
			floor := parkingGarage[i].Floor
			spotID := parkingGarage[i].ID
			parkingGarage[i].Occupied = false
			parkingGarage[i].VehiclePlate = ""
			parkingGarage[i].EntryTime = time.Time{}

			// Optimization: Update nearestAvailable if this vacated spot is lower than current nearest
			spotIdxInFloor := spotID - 1
			if spotIdxInFloor < nearestAvailable[floor-1] {
				nearestAvailable[floor-1] = spotIdxInFloor
			}

			fmt.Printf("✓ Vehicle %s exited. Duration: %dh %dm. Fee: ₹%d\n",
				vehiclePlate, hours, minutes, totalFee)
			fmt.Printf("  (Vacated from Floor %d, Spot %d)\n", floor, spotID)
			return nil
		}
	}

	return fmt.Errorf("vehicle %s not found in parking garage", vehiclePlate)
}

// DisplayAvailability: Show floor-wise availability
func DisplayAvailability() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("PARKING GARAGE STATUS")
	fmt.Println(strings.Repeat("=", 60))

	for floor := 1; floor <= TOTAL_FLOORS; floor++ {
		available := 0
		occupied := 0

		startIdx := (floor - 1) * SPOTS_PER_FLOOR
		endIdx := startIdx + SPOTS_PER_FLOOR

		for i := startIdx; i < endIdx; i++ {
			if parkingGarage[i].Occupied {
				occupied++
			} else {
				available++
			}
		}

		// Display with visual indicator
		percentage := (occupied * 100) / SPOTS_PER_FLOOR
		barLength := 20
		filledLength := (percentage * barLength) / 100
		emptyLength := barLength - filledLength

		bar := "[" + strings.Repeat("█", filledLength) + strings.Repeat("░", emptyLength) + "]"

		fmt.Printf("Floor %d: %3d/100 available | %s %d%%\n",
			floor, available, bar, percentage)
	}

	// Total summary
	totalOccupied := 0
	for i := 0; i < TOTAL_SPOTS; i++ {
		if parkingGarage[i].Occupied {
			totalOccupied++
		}
	}
	totalAvailable := TOTAL_SPOTS - totalOccupied

	fmt.Println(strings.Repeat("-", 60))
	fmt.Printf("TOTAL: %d/500 available | %d occupied\n", totalAvailable, totalOccupied)
	fmt.Println(strings.Repeat("=", 60) + "\n")
}

// SearchVehicle: Find a vehicle by plate number
func SearchVehicle(vehiclePlate string) error {
	if strings.TrimSpace(vehiclePlate) == "" {
		return fmt.Errorf("vehicle plate cannot be empty")
	}

	normalizedPlate := strings.ToUpper(vehiclePlate)
	for i := 0; i < TOTAL_SPOTS; i++ {
		if parkingGarage[i].Occupied && strings.ToUpper(parkingGarage[i].VehiclePlate) == normalizedPlate {
			duration := time.Since(parkingGarage[i].EntryTime)
			hours := int(duration.Hours())
			minutes := int(duration.Minutes()) % 60

			fmt.Printf("\n✓ Found Vehicle: %s\n", vehiclePlate)
			fmt.Printf("  Location: Floor %d, Spot %d\n", parkingGarage[i].Floor, parkingGarage[i].ID)
			fmt.Printf("  Entry Time: %s\n", parkingGarage[i].EntryTime.Format("2006-01-02 15:04:05"))
			fmt.Printf("  Duration: %dh %dm\n\n", hours, minutes)
			return nil
		}
	}

	return fmt.Errorf("vehicle %s is not currently parked in the garage", vehiclePlate)
}

// displayMenu shows the CLI menu
func displayMenu() {
	fmt.Println("\n" + strings.Repeat("=", 60))
	fmt.Println("PARKING GARAGE MANAGEMENT SYSTEM")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Println("1. Vehicle Entry")
	fmt.Println("2. Vehicle Exit")
	fmt.Println("3. Display Floor Availability")
	fmt.Println("4. Search Vehicle by Plate")
	fmt.Println("5. Exit Program")
	fmt.Println(strings.Repeat("=", 60))
	fmt.Print("Enter your choice (1-5): ")
}

// getUserInput reads input from the user using the global scanner
func getUserInput() string {
	if inputScanner.Scan() {
		return strings.TrimSpace(inputScanner.Text())
	}
	return ""
}

func main() {
	// Start pprof server
	go func() {
		fmt.Println("[INFO] Starting pprof server on :6060")
		if err := http.ListenAndServe("localhost:6060", nil); err != nil {
			fmt.Printf("[ERROR] pprof server failed: %v\n", err)
		}
	}()
	initializeParkingGarage()

	fmt.Println("\nWelcome to Parking Garage Management System!")
	fmt.Printf("Capacity: %d spots across %d floors\n", TOTAL_SPOTS, TOTAL_FLOORS)
	fmt.Printf("Rate: ₹%d per hour\n", HOURLY_RATE)

	for {
		displayMenu()
		choice := getUserInput()

		switch choice {
		case "1":
			// Vehicle Entry
			fmt.Print("Enter floor number (1-5): ")
			floorStr := getUserInput()
			floor, err := strconv.Atoi(floorStr)
			if err != nil {
				fmt.Println("❌ Invalid floor number")
				continue
			}

			fmt.Print("Enter vehicle plate number: ")
			plate := getUserInput()

			if err := EntryVehicle(floor, plate); err != nil {
				fmt.Printf("❌ %s\n", err.Error())
			}

		case "2":
			// Vehicle Exit
			fmt.Print("Enter vehicle plate number: ")
			plate := getUserInput()

			if err := ExitVehicle(plate); err != nil {
				fmt.Printf("❌ %s\n", err.Error())
			}

		case "3":
			// Display Availability
			DisplayAvailability()

		case "4":
			// Search Vehicle
			fmt.Print("Enter vehicle plate number: ")
			plate := getUserInput()

			if err := SearchVehicle(plate); err != nil {
				fmt.Printf("❌ %s\n", err.Error())
			}

		case "5":
			fmt.Println("\n✓ Thank you for using Parking Garage Management System!")
			fmt.Println("Exiting...")
			return

		default:
			fmt.Println("❌ Invalid choice. Please select 1-5.")
		}
	}
}
