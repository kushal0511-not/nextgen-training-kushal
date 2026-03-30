package auction

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/nextgen-training-kushal/Day-13-14/category"
	"github.com/nextgen-training-kushal/Day-13-14/models"
	"github.com/nextgen-training-kushal/Day-13-14/sse"
	"github.com/nextgen-training-kushal/Day-13-14/user"
)

func TestAuctionManager(t *testing.T) {
	um := user.NewUserManager()
	ct := category.NewCategoryTree()
	br := sse.NewBroker(nil)
	am := NewAuctionManager(um, ct, br.Broadcast)

	// Setup User
	u := &models.User{ID: 1, Name: "Kushal", Balance: 1000}
	um.AddUser(u)

	// Setup Item
	item := &models.Item{
		ID:         1,
		Name:       "Phone",
		Category:   "Phones",
		StartPrice: 100,
		Status:     "Active",
		StartTime:  time.Now().Add(-1 * time.Hour),
		EndTime:    time.Now().Add(1 * time.Hour),
	}
	am.RegisterItem(item)

	// Test Place Bid
	_, err := am.PlaceBid(1, 1, 150)
	if err != nil {
		t.Fatalf("Failed to place bid: %v", err)
	}

	if item.CurrentBid.Amount != 150 {
		t.Errorf("Expected current bid to be 150, got %v", item.CurrentBid.Amount)
	}

	if u.Balance != 850 {
		t.Errorf("Expected balance to be 850, got %v", u.Balance)
	}

	// Test Retract Bid
	err = am.RetractBid(1, 1)
	if err != nil {
		t.Fatalf("Failed to retract bid: %v", err)
	}

	if item.CurrentBid != nil {
		t.Errorf("Expected current bid to be nil after retraction")
	}

	if u.Balance != 1000 {
		t.Errorf("Expected balance to be restored to 1000, got %v", u.Balance)
	}

	// Test End Auction
	_, _ = am.PlaceBid(1, 1, 200)
	winner, err := am.EndAuction(1)
	if err != nil {
		t.Fatalf("Failed to end auction: %v", err)
	}
	if winner == nil || winner.Amount != 200 {
		t.Errorf("Expected winner with 200, got %v", winner)
	}
}

func TestAuctionManagerThunderingHerd(t *testing.T) {
	um := user.NewUserManager()
	ct := category.NewCategoryTree()
	br := sse.NewBroker(nil)
	am := NewAuctionManager(um, ct, br.Broadcast)

	item := &models.Item{
		ID:         1,
		Name:       "Rare Artifact",
		Category:   "Antiques",
		StartPrice: 100,
		Status:     "Active",
		StartTime:  time.Now().Add(-1 * time.Hour),
		EndTime:    time.Now().Add(1 * time.Hour),
	}
	am.RegisterItem(item)

	const numUsers = 100
	var wg sync.WaitGroup
	wg.Add(numUsers)

	for i := 1; i <= numUsers; i++ {
		u := &models.User{ID: i, Name: fmt.Sprintf("User%d", i), Balance: 1000000}
		um.AddUser(u)

		go func(userID int) {
			defer wg.Done()
			// Each user tries to bid slightly more than the last known bid
			// Using random sleep to stagger the "herd"
			time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
			amount := float64(200 + userID*10)
			_, _ = am.PlaceBid(1, userID, amount)
		}(i)
	}

	wg.Wait()

	winner, _ := am.EndAuction(1)
	if winner == nil {
		t.Fatal("Expected a winner from the herd")
	}
}

func TestAuctionManagerRaceToFinish(t *testing.T) {
	um := user.NewUserManager()
	ct := category.NewCategoryTree()
	br := sse.NewBroker(nil)
	am := NewAuctionManager(um, ct, br.Broadcast)

	item := &models.Item{
		ID:         1,
		Name:       "Timed Deal",
		Category:   "Deals",
		StartPrice: 100,
		Status:     "Active",
		StartTime:  time.Now().Add(-1 * time.Hour),
		EndTime:    time.Now().Add(50 * time.Millisecond), // Ends very soon
	}
	am.RegisterItem(item)

	um.AddUser(&models.User{ID: 1, Name: "Bidder", Balance: 10000})

	var wg sync.WaitGroup
	wg.Add(2)

	// Goroutine 1: Continuous Bidding
	go func() {
		defer wg.Done()
		for i := 1; i < 100; i++ {
			_, err := am.PlaceBid(1, 1, float64(100+i))
			if err != nil {
				// Auction likely ended
				return
			}
		}
	}()

	// Goroutine 2: Try to end auction
	go func() {
		defer wg.Done()
		time.Sleep(60 * time.Millisecond) // Wait for auction to expire
		_, _ = am.EndAuction(1)
	}()

	wg.Wait()
}

func TestAuctionManagerChaos(t *testing.T) {
	um := user.NewUserManager()
	ct := category.NewCategoryTree()
	br := sse.NewBroker(nil)
	am := NewAuctionManager(um, ct, br.Broadcast)

	for i := 1; i <= 5; i++ {
		item := &models.Item{
			ID:         i,
			Name:       fmt.Sprintf("Item%d", i),
			Category:   "Chaos",
			StartPrice: 10,
			Status:     models.StatusActive,
			StartTime:  time.Now().Add(-1 * time.Hour),
			EndTime:    time.Now().Add(1 * time.Hour),
		}
		am.RegisterItem(item)
	}

	for i := 1; i <= 20; i++ {
		um.AddUser(&models.User{ID: i, Name: fmt.Sprintf("User%d", i), Balance: 10000})
	}

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano()))
			for j := 0; j < 50; j++ {
				op := r.Intn(4)
				itemID := r.Intn(5) + 1
				userID := r.Intn(20) + 1

				switch op {
				case 0:
					_, _ = am.PlaceBid(itemID, userID, float64(100+r.Intn(1000)))
				case 1:
					_ = am.RetractBid(itemID, userID)
				case 2:
					_, _ = am.BrowseCategory([]string{"Chaos"})
				case 3:
					_, _ = am.EndAuction(itemID)
				}
			}
		}()
	}

	wg.Wait()
}
