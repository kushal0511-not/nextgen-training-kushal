package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/nextgen-training-kushal/Day-13-14/auction"
	"github.com/nextgen-training-kushal/Day-13-14/category"
	"github.com/nextgen-training-kushal/Day-13-14/models"
	"github.com/nextgen-training-kushal/Day-13-14/user"
)

func Seed(userManager *user.UserManager, categoryTree *category.CategoryTree, auctionManager *auction.AuctionManager) {
	fmt.Println("=== Bidding System: End-to-End Simulation ===")

	// 2. Setup Infrastructure
	categoryTree.AddCategory([]string{"Electronics", "Phones"})

	const numUsers = 100
	const itemID = 101

	// Item Registration
	item := &models.Item{
		ID:         itemID,
		Name:       "Super Smartphone",
		Category:   "Phones",
		StartPrice: 100,
		Status:     models.StatusActive,
		StartTime:  time.Now().Add(-1 * time.Hour),
		EndTime:    time.Now().Add(1 * time.Hour),
	}
	auctionManager.RegisterItem(item)

	// User Registration
	for i := 1; i <= numUsers; i++ {
		userManager.AddUser(&models.User{
			ID:      i,
			Name:    fmt.Sprintf("User%d", i),
			Balance: 1000000, // Rich users
		})
	}

	// 3. Test Concurrent Bidding: 100 Goroutines
	fmt.Printf("\n--- [Phase 1] Simulating 100 Concurrent Bidders ---\n")
	var wg sync.WaitGroup
	wg.Add(numUsers)

	startTime := time.Now()
	for i := 1; i <= numUsers; i++ {
		go func(uid int) {
			defer wg.Done()
			// Each user places 1-5 random bids
			numBids := rand.Intn(5) + 1
			for b := 0; b < numBids; b++ {
				amount := float64(500 + uid*10 + b*5)
				_, _ = auctionManager.PlaceBid(itemID, uid, amount)
				time.Sleep(time.Millisecond * time.Duration(rand.Intn(10)))
			}
		}(i)
	}
	wg.Wait()
	fmt.Printf("Phase 1 Complete in %v\n", time.Since(startTime))

	// 4. Test Bid Retraction & Undo
	fmt.Printf("\n--- [Phase 2] Testing Bid Retraction ---\n")

	// A new high bid from a specific user
	specialUserID := 999
	userManager.AddUser(&models.User{ID: specialUserID, Name: "Charlie", Balance: 50000})

	highAmount := 10000.0
	fmt.Printf("Charlie placing high bid of $%.2f...\n", highAmount)
	_, err := auctionManager.PlaceBid(itemID, specialUserID, highAmount)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	fmt.Printf("Current Leading Bid: $%.2f (ID: %d)\n", item.CurrentBid.Amount, item.CurrentBid.UserID)

	fmt.Println("Charlie retracting his last bid...")
	err = auctionManager.RetractBid(itemID, specialUserID)
	if err != nil {
		fmt.Printf("Retraction Failed: %v\n", err)
	} else {
		fmt.Printf("Retraction Success! New Leading Bid: $%.2f (ID: %d)\n",
			item.CurrentBid.Amount, item.CurrentBid.UserID)
	}

	// 5. Test Auction End: Winner Determination
	fmt.Printf("\n--- [Phase 3] Ending Auction & Winner Determination ---\n")
	winner, err := auctionManager.EndAuction(itemID)
	if err != nil {
		fmt.Printf("Error ending auction: %v\n", err)
	} else if winner != nil {
		wUser, _ := userManager.GetUser(winner.UserID)
		fmt.Printf("WINNER FOUND!\n")
		fmt.Printf("  Name:   %s\n", wUser.Name)
		fmt.Printf("  Bid:    $%.2f\n", winner.Amount)
		fmt.Printf("  Time:   %v\n", winner.Timestamp.Format(time.Kitchen))
	} else {
		fmt.Println("No qualified winner found for this item.")
	}

	fmt.Println("\n=== Simulation Finished Successfully ===")
}
