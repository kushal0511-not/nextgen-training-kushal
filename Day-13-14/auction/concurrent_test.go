package auction

import (
	"fmt"
	"math/rand"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/nextgen-training-kushal/Day-13/category"
	"github.com/nextgen-training-kushal/Day-13/models"
	"github.com/nextgen-training-kushal/Day-13/sse"
	"github.com/nextgen-training-kushal/Day-13/user"
)

func TestConcurrentBidding(t *testing.T) {
	runtime.SetMutexProfileFraction(1)
	um := user.NewUserManager()
	ct := category.NewCategoryTree()
	br := sse.NewBroker(nil)
	am := NewAuctionManager(um, ct, br.Broadcast)

	// Register 5 items
	numItems := 5
	for i := 1; i <= numItems; i++ {
		item := &models.Item{
			ID:         i,
			Name:       fmt.Sprintf("ConcurrentItem%d", i),
			Category:   "Concurrent",
			StartPrice: 10,
			Status:     models.StatusActive,
			StartTime:  time.Now().Add(-1 * time.Hour),
			EndTime:    time.Now().Add(1 * time.Hour),
		}
		am.RegisterItem(item)
	}

	// Register 50 users
	numUsers := 50
	for i := 1; i <= numUsers; i++ {
		// Provide each user with a high balance to guarantee bids won't fail due to funds
		um.AddUser(&models.User{ID: i, Name: fmt.Sprintf("User%d", i), Balance: 1000000})
	}

	var wg sync.WaitGroup
	wg.Add(numUsers)

	// Keep track of the highest valid bid placed per item to verify at the end
	highestExpectedBids := make([]float64, numItems+1)
	var mu sync.Mutex

	// Each user makes 10 bids
	numBidsPerUser := 10

	for i := 1; i <= numUsers; i++ {
		go func(userID int) {
			defer wg.Done()
			r := rand.New(rand.NewSource(time.Now().UnixNano() + int64(userID)))

			for j := 0; j < numBidsPerUser; j++ {
				itemID := r.Intn(numItems) + 1

				// Calculate a bid amount based on user and base amount so it keeps incrementing
				// and multiple users will try to outbid each other.
				baseAmount := float64(100 + r.Intn(1000))

				bid, err := am.PlaceBid(itemID, userID, baseAmount)

				if err == nil {
					mu.Lock()
					if bid.Amount > highestExpectedBids[itemID] {
						highestExpectedBids[itemID] = bid.Amount
					}
					mu.Unlock()
				}

				// Simulating bidding over 30 seconds frame.
				// To keep test fast, we will sleep 10 milliseconds instead of actual 3 seconds per bid.
				time.Sleep(10 * time.Millisecond)
			}
		}(i)
	}

	wg.Wait()

	// Verify state and winners
	for i := 1; i <= numItems; i++ {
		winner, err := am.EndAuction(i)
		if err != nil {
			t.Fatalf("Failed to end auction for item %d: %v", i, err)
		}

		mu.Lock()
		expectedHighest := highestExpectedBids[i]
		mu.Unlock()

		if expectedHighest > 0 {
			if winner == nil {
				t.Errorf("Item %d should have a winner but got nil", i)
			} else if winner.Amount != expectedHighest {
				t.Errorf("Item %d winner amount %.2f does not match expected highest amount %.2f", i, winner.Amount, expectedHighest)
			}
		} else {
			// No successful bids were placed for this item
			if winner != nil {
				t.Errorf("Item %d should not have a winner but got %v", i, winner)
			}
		}
	}
}
