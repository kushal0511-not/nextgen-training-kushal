package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/pprof"

	"github.com/nextgen-training-kushal/Day-13/models"
)

func (app *App) CreateUser(w http.ResponseWriter, r *http.Request) {
	var user models.User
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		app.logger.Println("Error decoding user:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := app.userManager.AddUser(&user); err != nil {
		app.logger.Println("Error adding user:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (app *App) CreateItem(w http.ResponseWriter, r *http.Request) {
	var item models.Item
	if err := json.NewDecoder(r.Body).Decode(&item); err != nil {
		app.logger.Println("Error decoding item:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	app.auctionManager.RegisterItem(&item)
	w.WriteHeader(http.StatusCreated)
}

func (app *App) PlaceBid(w http.ResponseWriter, r *http.Request) {
	var bid models.Bid
	if err := json.NewDecoder(r.Body).Decode(&bid); err != nil {
		app.logger.Println("Error decoding bid:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	placedBid, err := app.auctionManager.PlaceBid(bid.ItemID, bid.UserID, bid.Amount)
	_ = placedBid // used by manager broadcast internally now
	if err != nil {
		app.logger.Println("Error placing bid:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (app *App) RetractBid(w http.ResponseWriter, r *http.Request) {
	var bid models.Bid
	if err := json.NewDecoder(r.Body).Decode(&bid); err != nil {
		app.logger.Println("Error decoding bid:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := app.auctionManager.RetractBid(bid.ItemID, bid.UserID); err != nil {
		app.logger.Println("Error retracting bid:", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GET /items/:id → Item details + bid history
func (app *App) GetItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		app.logger.Println("Invalid item ID:", idStr)
		http.Error(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	item, err := app.auctionManager.GetItem(id)
	if err != nil {
		app.logger.Println("Error fetching item:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(item); err != nil {
		app.logger.Println("Error encoding item:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) GetItemsByCategory(w http.ResponseWriter, r *http.Request) {
	category := r.URL.Query().Get("category")
	if category == "" {
		app.logger.Println("Invalid category")
		http.Error(w, "invalid category", http.StatusBadRequest)
		return
	}

	items, err := app.auctionManager.BrowseCategory([]string{category})
	if err != nil {
		app.logger.Println("Error fetching items by category:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		app.logger.Println("Error encoding items by category:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

}

func (app *App) GetBidsByItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		app.logger.Println("Invalid item ID:", idStr)
		http.Error(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	bids, err := app.auctionManager.GetBidsByItem(id)
	if err != nil {
		app.logger.Println("Error fetching bids:", err)
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(bids); err != nil {
		app.logger.Println("Error encoding bids:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) EndAuction(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		app.logger.Println("Invalid item ID:", idStr)
		http.Error(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	winner, err := app.auctionManager.EndAuction(id)
	if err != nil {
		app.logger.Println("Error ending auction:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(winner); err != nil {
		app.logger.Println("Error encoding winner:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) GetStats(w http.ResponseWriter, r *http.Request) {
	stats := app.auctionManager.GetStats()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(stats); err != nil {
		app.logger.Println("Error encoding stats:", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (app *App) GetLiveBids(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	var id int
	if _, err := fmt.Sscanf(idStr, "%d", &id); err != nil {
		app.logger.Println("Invalid item ID:", idStr)
		http.Error(w, "invalid item ID", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	ch := app.broker.AddWatcher(id)
	defer app.broker.RemoveWatcher(id, ch)

	ctx := r.Context()
	for {
		select {
		case bid := <-ch:
			data, err := json.Marshal(bid)
			if err != nil {
				app.logger.Println("Error marshaling bid for SSE:", err)
				continue
			}
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		case <-ctx.Done():
			return
		}
	}
}

func NewRouter(app *App) *http.ServeMux {
	mux := http.NewServeMux()
	// pprof endpoints
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	mux.HandleFunc("/debug/pprof/trace", pprof.Trace)

	mux.Handle("POST /users", AuthMiddleware(LoggingMiddleware(PanicRecoveryMiddleware(http.HandlerFunc(app.CreateUser)))))
	mux.Handle("GET /items", AuthMiddleware(LoggingMiddleware(PanicRecoveryMiddleware(http.HandlerFunc(app.GetItemsByCategory)))))
	mux.Handle("POST /items", AuthMiddleware(LoggingMiddleware(PanicRecoveryMiddleware(http.HandlerFunc(app.CreateItem)))))
	mux.Handle("POST /items/{id}/bid", AuthMiddleware(LoggingMiddleware(RateLimitMiddleware(PanicRecoveryMiddleware(http.HandlerFunc(app.PlaceBid))))))
	mux.Handle("DELETE /items/{id}/bid/last", AuthMiddleware(LoggingMiddleware(PanicRecoveryMiddleware(http.HandlerFunc(app.RetractBid)))))
	mux.Handle("GET /items/{id}", AuthMiddleware(LoggingMiddleware(PanicRecoveryMiddleware(http.HandlerFunc(app.GetItem)))))
	mux.Handle("GET /items/{id}/bids", AuthMiddleware(LoggingMiddleware(PanicRecoveryMiddleware(http.HandlerFunc(app.GetBidsByItem)))))
	mux.Handle("POST /items/{id}/end", AuthMiddleware(LoggingMiddleware(PanicRecoveryMiddleware(http.HandlerFunc(app.EndAuction)))))
	mux.Handle("GET /stats", AuthMiddleware(LoggingMiddleware(PanicRecoveryMiddleware(http.HandlerFunc(app.GetStats)))))
	mux.Handle("GET /items/{id}/live", AuthMiddleware(LoggingMiddleware(PanicRecoveryMiddleware(http.HandlerFunc(app.GetLiveBids)))))
	return mux
}
