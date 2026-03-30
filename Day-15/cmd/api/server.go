package api

import (
	"context"
	"net/http"
	_ "net/http/pprof" // side-effect: registers debug/pprof handlers
	"time"

	"github.com/nextgen-training-kushal/Day-15/traffic"
	"go.uber.org/zap"
)

// Server wraps the HTTP server and city model.
type Server struct {
	city   *traffic.CityModel
	logger *zap.Logger
	http   *http.Server
}

// New creates a configured API Server with all routes registered.
func New(city *traffic.CityModel, logger *zap.Logger, addr string) *Server {
	s := &Server{city: city, logger: logger}

	mux := http.NewServeMux()

	// Business routes
	mux.HandleFunc("POST /vehicles", s.RegisterVehicle)
	mux.HandleFunc("GET /route", s.GetRoute)
	mux.HandleFunc("POST /emergency", s.DispatchEmergency)
	mux.HandleFunc("GET /congestion", s.GetCongestion)
	mux.HandleFunc("GET /signals/{id}", s.GetSignal)
	mux.HandleFunc("GET /stats", s.GetStats)

	// Profiling endpoints: CPU, memory, goroutine, block, mutex
	mux.Handle("/debug/pprof/", http.DefaultServeMux)

	s.http = &http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	return s
}

// Run starts listening and shuts down when ctx is cancelled.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)
	go func() {
		s.logger.Info("API server started", zap.String("addr", s.http.Addr))
		if err := s.http.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		shutCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.logger.Info("Shutting down API server...")
		return s.http.Shutdown(shutCtx)
	case err := <-errCh:
		return err
	}
}
