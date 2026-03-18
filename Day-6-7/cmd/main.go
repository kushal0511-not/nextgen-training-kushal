package main

import (
	"bufio"
	"fmt"
	"os"
	"ride-sharing/internal/dispatch"
	"ride-sharing/internal/driver"
	"ride-sharing/internal/models"
	"ride-sharing/internal/queue"
	"ride-sharing/internal/rides"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var (
	driverStore *driver.MemoryStore
	rideQueue   *queue.MinHeap
	rideTracker *rides.LinkedTracker
	dispatcher  *dispatch.Dispatcher
)

func init() {
	driverStore = driver.NewMemoryStore()
	rideQueue = queue.NewMinHeap()
	rideTracker = rides.NewLinkedTracker()
	dispatcher = dispatch.NewDispatcher(driverStore, rideQueue, rideTracker)
}

func newRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "rideshare",
		Short: "Ride-Sharing Dispatch System",
		Run: func(cmd *cobra.Command, args []string) {
			startInteractiveShell()
		},
	}

	// Driver Commands
	rootCmd.AddCommand(registerDriverCmd())
	rootCmd.AddCommand(onlineCmd())
	rootCmd.AddCommand(offlineCmd())
	rootCmd.AddCommand(moveCmd())

	// Ride Commands
	rootCmd.AddCommand(requestCmd())
	rootCmd.AddCommand(dispatchCmd())
	rootCmd.AddCommand(completeCmd())

	// Query Commands
	rootCmd.AddCommand(queryNearestCmd())
	rootCmd.AddCommand(queryEarningsCmd())
	rootCmd.AddCommand(queryWaitsCmd())
	rootCmd.AddCommand(queryZonesCmd())

	return rootCmd
}

func main() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func startInteractiveShell() {
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Ride-Sharing Dispatch System (Cobra Interactive Shell)")
	fmt.Println("Type 'exit' to quit or 'help' for available commands.")

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			break
		}

		input := scanner.Text()
		trimmedInput := strings.TrimSpace(input)
		if trimmedInput == "" {
			continue
		}
		if trimmedInput == "exit" {
			break
		}

		args := strings.Fields(trimmedInput)
		// If the user types the program name "rideshare", strip it
		if len(args) > 0 && args[0] == "rideshare" {
			args = args[1:]
		}

		if len(args) == 0 {
			continue
		}

		// Create a fresh command tree for each execution to avoid flag persistence
		cmd := newRootCmd()
		cmd.SilenceUsage = true
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			// Errors are already printed by Cobra because SilenceErrors is false by default
		}
	}
}

// Command Definitions

func registerDriverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "register-driver <id> <name> <lat> <lng>",
		Short: "Register a new driver",
		Args:  cobra.ExactArgs(4),
		RunE: func(cmd *cobra.Command, args []string) error {
			lat, err := strconv.ParseFloat(args[2], 64)
			if err != nil {
				return fmt.Errorf("invalid latitude: %v", err)
			}
			lng, err := strconv.ParseFloat(args[3], 64)
			if err != nil {
				return fmt.Errorf("invalid longitude: %v", err)
			}
			d := &models.Driver{
				ID:       args[0],
				Name:     args[1],
				Location: models.Location{Lat: lat, Lng: lng},
				Status:   models.DriverStatusAvailable,
			}
			if err := driverStore.Register(d); err != nil {
				return err
			}
			fmt.Printf("Driver %s registered successfully\n", d.Name)
			return nil
		},
	}
}

func onlineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "online <driverID>",
		Short: "Bring a driver online",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := driverStore.UpdateStatus(args[0], models.DriverStatusAvailable); err != nil {
				return err
			}
			fmt.Println("Driver is now online")
			return nil
		},
	}
}

func offlineCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "offline <driverID>",
		Short: "Bring a driver offline",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := driverStore.UpdateStatus(args[0], models.DriverStatusOffline); err != nil {
				return err
			}
			fmt.Println("Driver is now offline")
			return nil
		},
	}
}

func moveCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "move <driverID> <lat> <lng>",
		Short: "Update driver location",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			lat, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return fmt.Errorf("invalid latitude: %v", err)
			}
			lng, err := strconv.ParseFloat(args[2], 64)
			if err != nil {
				return fmt.Errorf("invalid longitude: %v", err)
			}
			if err := driverStore.UpdateLocation(args[0], models.Location{Lat: lat, Lng: lng}); err != nil {
				return err
			}
			fmt.Println("Location updated successfully")
			return nil
		},
	}
}

func requestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "request <riderID> <plat> <plng> <dlat> <dlng>",
		Short: "Request a ride",
		Args:  cobra.ExactArgs(5),
		RunE: func(cmd *cobra.Command, args []string) error {
			plat, _ := strconv.ParseFloat(args[1], 64)
			plng, _ := strconv.ParseFloat(args[2], 64)
			dlat, _ := strconv.ParseFloat(args[3], 64)
			dlng, _ := strconv.ParseFloat(args[4], 64)

			rider := &models.Rider{ID: args[0]}
			ride := dispatcher.RequestRide(rider, models.Location{Lat: plat, Lng: plng}, models.Location{Lat: dlat, Lng: dlng})
			fmt.Printf("Ride requested successfully: %s\n", ride.ID)
			return nil
		},
	}
}

func dispatchCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "dispatch",
		Short: "Process the ride queue and match drivers",
		Run: func(cmd *cobra.Command, args []string) {
			dispatcher.ProcessQueue()
		},
	}
}

func completeCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "complete <rideID>",
		Short: "Complete a ride",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return dispatcher.CompleteRide(args[0])
		},
	}
}

func queryNearestCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query-nearest <lat> <lng> <n>",
		Short: "Find nearest available drivers",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			lat, _ := strconv.ParseFloat(args[0], 64)
			lng, _ := strconv.ParseFloat(args[1], 64)
			n, _ := strconv.Atoi(args[2])
			drivers := dispatcher.GetNearestDrivers(models.Location{Lat: lat, Lng: lng}, n)
			if len(drivers) == 0 {
				fmt.Println("No available drivers found in the vicinity.")
			} else {
				fmt.Printf("Nearest %d drivers:\n", len(drivers))
				for _, d := range drivers {
					fmt.Printf("- %s (%s) at (%.2f, %.2f)\n", d.Name, d.ID, d.Location.Lat, d.Location.Lng)
				}
			}
			return nil
		},
	}
}

func queryEarningsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query-earnings <driverID>",
		Short: "Check driver earnings",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := driverStore.GetByID(args[0]); err != nil {
				return err
			}
			fmt.Printf("Total Earnings for Driver %s: ₹%.2f\n", args[0], dispatcher.GetDriverEarnings(args[0]))
			return nil
		},
	}
}

func queryWaitsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query-waits",
		Short: "Check average rider wait time",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("Average Wait Time: %v\n", dispatcher.GetAverageWaitTime())
		},
	}
}

func queryZonesCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "query-zones",
		Short: "Identify busiest zones",
		Run: func(cmd *cobra.Command, args []string) {
			zones := dispatcher.GetBusiestZones()
			fmt.Println("Busiest Zones:")
			for zone, count := range zones {
				fmt.Printf("- %s: %d requests\n", zone, count)
			}
		},
	}
}
