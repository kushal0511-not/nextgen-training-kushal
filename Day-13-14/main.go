package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/nextgen-training-kushal/Day-13/auction"
	"github.com/nextgen-training-kushal/Day-13/category"
	"github.com/nextgen-training-kushal/Day-13/sse"
	"github.com/nextgen-training-kushal/Day-13/user"
)

type App struct {
	router         *http.ServeMux
	userManager    *user.UserManager
	categoryTree   *category.CategoryTree
	auctionManager *auction.AuctionManager
	broker         *sse.Broker
	logger         *log.Logger
	rateLimiter    *RateLimiter
}

func main() {
	logger := log.New(os.Stdout, "Auction: ", log.LstdFlags)
	userManager := user.NewUserManager()
	categoryTree := category.NewCategoryTree()
	broker := sse.NewBroker(logger)
	auctionManager := auction.NewAuctionManager(userManager, categoryTree, broker.Broadcast)
	app := &App{
		userManager:    userManager,
		categoryTree:   categoryTree,
		auctionManager: auctionManager,
		broker:         broker,
		logger:         logger,
	}

	go func() {
		for {
			time.Sleep(time.Second * 10)
			Seed(userManager, categoryTree, auctionManager)
		}
	}()
	app.router = NewRouter(app)
	fmt.Println("Server started on :8080")
	http.ListenAndServe(":8080", app.router)
}
