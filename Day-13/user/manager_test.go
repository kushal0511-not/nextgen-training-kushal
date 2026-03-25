package user

import (
	"sync"
	"testing"

	"github.com/nextgen-training-kushal/Day-13/models"
)

func TestUserManager(t *testing.T) {
	m := NewUserManager()
	u := &models.User{ID: 1, Name: "Test User", Balance: 1000}

	err := m.AddUser(u)
	if err != nil {
		t.Fatalf("Failed to add user: %v", err)
	}

	retrieved, err := m.GetUser(1)
	if err != nil || retrieved.Name != "Test User" {
		t.Errorf("Failed to retrieve user correctly")
	}

	// Test Undo Stack
	bid1 := models.Bid{ID: 101, Amount: 100}
	bid2 := models.Bid{ID: 102, Amount: 110}

	m.PushBidToUndoStack(1, bid1)
	m.PushBidToUndoStack(1, bid2)

	undoneID, err := m.PopBidFromUndoStack(1)
	if err != nil || undoneID != 102 {
		t.Errorf("Expected to undo bid 102, got %v", undoneID)
	}

	undoneID, err = m.PopBidFromUndoStack(1)
	if err != nil || undoneID != 101 {
		t.Errorf("Expected to undo bid 101, got %v", undoneID)
	}

	_, err = m.PopBidFromUndoStack(1)
	if err == nil {
		t.Errorf("Expected error when popping from empty stack")
	}
}

func TestUserManagerConcurrency(t *testing.T) {
	m := NewUserManager()
	u := &models.User{ID: 1, Name: "Wealthy User", Balance: 1000000}
	m.AddUser(u)

	const numGoroutines = 100
	const opsPerGoroutine = 100
	var wg sync.WaitGroup

	wg.Add(numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func() {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				m.DeductBalance(1, 10.0)
				m.RestoreBalance(1, 10.0)
				m.PushBidToUndoStack(1, models.Bid{ID: j})
				m.PopBidFromUndoStack(1)
			}
		}()
	}

	wg.Wait()

	if u.Balance != 1000000 {
		t.Errorf("Expected balance to be 1000000, got %v", u.Balance)
	}
}
