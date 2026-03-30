package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/nextgen-training-kushal/Day-15/cmd/api"
	"github.com/nextgen-training-kushal/Day-15/cmd/cli"
	"github.com/nextgen-training-kushal/Day-15/models"
	"github.com/nextgen-training-kushal/Day-15/traffic"
	"go.uber.org/zap"
)

func main() {
	mode := flag.String("mode", "cli", "Run mode: cli | api")
	flag.Parse()

	// --- Logger ---
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

	// --- Build city ---
	city := traffic.NewCityModel(20, logger)
	city.GenerateRandomCity()

	// --- Graceful shutdown via OS signals ---
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		sig := <-quit
		logger.Info("shutdown signal received", zap.String("signal", sig.String()))
		cancel()
	}()

	// --- Register normal vehicles ---
	vehicles := []struct {
		plate    string
		from, to int
	}{
		{"V-101", 1, 15}, {"V-102", 5, 18},
		{"V-103", 2, 16}, {"V-104", 10, 7},
	}
	for _, v := range vehicles {
		if err := city.RegisterVehicle(v.plate,
			models.IntersectionID(v.from), models.IntersectionID(v.to)); err != nil {
			logger.Warn("vehicle registration failed", zap.String("plate", v.plate), zap.Error(err))
		}
	}

	// --- Start all goroutines ---
	city.StartTrafficSignals(ctx)
	go city.StartCongestionSimulation(ctx)
	go city.StartCongestionTracking(ctx)
	city.StartVehicleSimulation(ctx)

	// Dispatch emergency vehicle after 3 seconds
	go func() {
		time.Sleep(3 * time.Second)
		if err := city.RegisterEmergencyVehicle(ctx, "E-911",
			models.IntersectionID(3), models.IntersectionID(20)); err != nil {
			logger.Error("emergency dispatch failed", zap.Error(err))
		}
	}()

	switch *mode {
	case "api":
		runAPI(ctx, city, logger)
	default:
		runCLI(ctx, city, logger)
	}

	logger.Info("simulation terminated")
}

func runAPI(ctx context.Context, city *traffic.CityModel, logger *zap.Logger) {
	srv := api.New(city, logger, ":8080")
	fmt.Println("================================================================")
	fmt.Println("🚀 City Traffic Simulation — API Mode")
	fmt.Println("  POST /vehicles        Register a vehicle")
	fmt.Println("  GET  /route?from=&to= Calculate best route")
	fmt.Println("  POST /emergency       Dispatch emergency vehicle")
	fmt.Println("  GET  /congestion      Current congestion data")
	fmt.Println("  GET  /signals/{id}    Signal state at intersection")
	fmt.Println("  GET  /stats           System-wide statistics")
	fmt.Println("================================================================")
	if err := srv.Run(ctx); err != nil {
		logger.Error("API server error", zap.Error(err))
	}
}

func runCLI(ctx context.Context, city *traffic.CityModel, logger *zap.Logger) {
	fmt.Println("================================================================")
	fmt.Println("🚀 City Traffic Simulation — CLI Mode  (Ctrl+C to quit)")
	fmt.Println("================================================================")

	ticker := time.NewTicker(3 * time.Second)
	reportTicker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	defer reportTicker.Stop()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("\nSimulation Complete. Final Status:")
			cli.PrintVehicles(city)
			city.PrintCongestionReport()
			return
		case <-ticker.C:
			fmt.Printf("\n─── %s ───\n", time.Now().Format("15:04:05"))
			cli.PrintSignals(city)
			cli.PrintVehicles(city)
		case <-reportTicker.C:
			cli.PrintCityMap(city)
			city.PrintCongestionReport()
		}
	}
}
