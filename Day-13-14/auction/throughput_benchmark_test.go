package auction

import (
	"testing"
	"time"

	"github.com/nextgen-training-kushal/Day-13-14/category"
	"github.com/nextgen-training-kushal/Day-13-14/models"
	"github.com/nextgen-training-kushal/Day-13-14/sse"
	"github.com/nextgen-training-kushal/Day-13-14/user"
)

func BenchmarkConcurrentBidding(b *testing.B) {
	um := user.NewUserManager()
	ct := category.NewCategoryTree()
	br := sse.NewBroker(nil) // No logging to keep it fast
	am := NewAuctionManager(um, ct, br.Broadcast)

	// Setup 10 items
	for i := 1; i <= 10; i++ {
		am.RegisterItem(&models.Item{
			ID:         i,
			StartPrice: 10,
			Status:     models.StatusActive,
			StartTime:  time.Now().Add(-1 * time.Hour),
			EndTime:    time.Now().Add(1 * time.Hour),
		})
	}

	// Setup 100 users
	for i := 1; i <= 100; i++ {
		um.AddUser(&models.User{ID: i, Balance: 1000000000})
	}

	b.ResetTimer()

	b.RunParallel(func(pb *testing.PB) {
		i := 0
		for pb.Next() {
			itemID := (i % 10) + 1
			userID := (i % 100) + 1
			amount := float64(100 + i)
			_, _ = am.PlaceBid(itemID, userID, amount)
			i++
		}
	})
}
