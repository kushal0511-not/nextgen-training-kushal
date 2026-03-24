package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/nextgen-training-kushal/Day-12/middleware"
	"github.com/nextgen-training-kushal/Day-12/models"
	"github.com/nextgen-training-kushal/Day-12/store"
	"go.uber.org/zap"
)

type App struct {
	store  *store.ProductStore
	router *http.ServeMux
}

func NewApp() *App {
	app := &App{
		store:  store.NewProductStore(50), // order 50 B-Tree for production
		router: http.NewServeMux(),
	}
	app.routes()
	return app
}

func (a *App) routes() {
	runtime.SetMutexProfileFraction(1)
	runtime.SetBlockProfileRate(1)
	go func() {
		log.Println("Starting pprof on :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Fatalf("pprof server failed: %v", err)
		}
	}()

	a.router.HandleFunc("POST /products", a.createProduct)
	a.router.HandleFunc("GET /products/stats", a.getStats)
	a.router.HandleFunc("GET /products/{id}", a.getProduct)
	a.router.HandleFunc("GET /products", a.listProducts)
	a.router.HandleFunc("PUT /products/{id}", a.updateProduct)
	a.router.HandleFunc("DELETE /products/{id}", a.deleteProduct)
	a.router.HandleFunc("POST /seed", a.seedProducts)
}

func (a *App) createProduct(w http.ResponseWriter, r *http.Request) {
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.store.AddProduct(&p)
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func (a *App) getProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	p, ok := a.store.GetProduct(id)
	if !ok {
		http.Error(w, `{"error": "Product not found"}`, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(p)
}

func (a *App) updateProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var p models.Product
	if err := json.NewDecoder(r.Body).Decode(&p); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if ok := a.store.UpdateProduct(id, &p); !ok {
		http.Error(w, `{"error": "Product not found"}`, http.StatusNotFound)
		return
	}
	json.NewEncoder(w).Encode(p)
}

func (a *App) deleteProduct(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	a.store.DeleteProduct(id)
	w.WriteHeader(http.StatusNoContent)
}

func (a *App) listProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Parse pagination
	page := 1
	size := 100 // Default larger size for list
	if query.Has("page") {
		page, _ = strconv.Atoi(query.Get("page"))
	}
	if query.Has("size") {
		size, _ = strconv.Atoi(query.Get("size"))
	}

	// 1. Price Range Query: GET /products?min_price=X&max_price=Y
	if query.Has("min_price") && query.Has("max_price") {
		min, errMin := strconv.ParseFloat(query.Get("min_price"), 64)
		max, errMax := strconv.ParseFloat(query.Get("max_price"), 64)
		if errMin != nil || errMax != nil {
			http.Error(w, `{"error": "Invalid price parameters"}`, http.StatusBadRequest)
			return
		}

		var results []*models.Product
		if query.Get("scan") == "linear" {
			results = a.store.QueryByPriceRangeLinear(min, max)
		} else {
			results = a.store.QueryByPriceRange(min, max)
		}

		// Apply pagination to range results
		start := (page - 1) * size
		end := start + size
		if start >= len(results) {
			results = []*models.Product{}
		} else {
			if end > len(results) {
				end = len(results)
			}
			results = results[start:end]
		}

		json.NewEncoder(w).Encode(results)
		return
	}

	// 2. Category Filter & Sort: GET /products?category=X&sort=price
	if query.Has("category") {
		category := query.Get("category")
		results := a.store.GetByCategorySorted(category)

		// Apply pagination
		start := (page - 1) * size
		end := start + size
		if start >= len(results) {
			results = []*models.Product{}
		} else {
			if end > len(results) {
				end = len(results)
			}
			results = results[start:end]
		}

		json.NewEncoder(w).Encode(results)
		return
	}

	// 3. Paginated Listing: GET /products?page=1&size=20
	results := a.store.GetAllPaginated(page, size)
	if results == nil {
		results = []*models.Product{}
	}
	json.NewEncoder(w).Encode(results)
}

func (a *App) getStats(w http.ResponseWriter, r *http.Request) {
	stats := a.store.GetStats()
	json.NewEncoder(w).Encode(stats)
}

func (a *App) seedProducts(w http.ResponseWriter, r *http.Request) {
	countStr := r.URL.Query().Get("count")
	count, _ := strconv.Atoi(countStr)
	if count <= 0 {
		count = 1000
	}

	for i := 0; i < count; i++ {
		p := models.Product{
			Name:     fmt.Sprintf("Product-%d", i),
			Category: "Electronics",
			Price:    rand.Float64() * 1000,
			Rating:   rand.Float64() * 5,
			Stock:    rand.Intn(100),
		}
		a.store.AddProduct(&p)
	}
	fmt.Fprintf(w, "Seeded %d products", count)
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()
	undo := zap.ReplaceGlobals(logger)
	defer undo()

	app := NewApp()

	// Apply Middlewares in order: Error -> Logging -> Timing -> JSON
	var handler http.Handler = app.router
	handler = middleware.JSONContentTypeMiddleware(handler)
	handler = middleware.RequestTimingMiddleware(handler)
	handler = middleware.LoggingMiddleware(handler)
	handler = middleware.ErrorRecoveryMiddleware(handler)

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	// Start server on a separate goroutine
	go func() {
		zap.S().Infof("Starting Server on port 8080...")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zap.S().Fatalf("Server error: %v", err)
		}
	}()

	// Graceful Shutdown Setup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zap.S().Info("Shutting down server...")

	// 5 second grace period to finish existing requests
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting gracefully.")
}
