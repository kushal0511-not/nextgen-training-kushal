package user

import (
	"errors"
	"fmt"
	"sync"

	"github.com/nextgen-training-kushal/Day-13-14/models"
)

// UserManager manages users using a HashMap.
type UserManager struct {
	users map[int]*models.User
	mu    sync.RWMutex
}

// NewUserManager creates a new UserManager.
func NewUserManager() *UserManager {
	return &UserManager{
		users: make(map[int]*models.User),
	}
}

// AddUser adds a user to the HashMap.
func (m *UserManager) AddUser(user *models.User) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, exists := m.users[user.ID]; exists {
		return fmt.Errorf("user with ID %d already exists", user.ID)
	}
	m.users[user.ID] = user
	return nil
}

// GetUser retrieves a user from the HashMap.
func (m *UserManager) GetUser(id int) (*models.User, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	user, exists := m.users[id]
	if !exists {
		return nil, fmt.Errorf("user with ID %d not found", id)
	}
	return user, nil
}

// PushBidToUndoStack adds a bid to the user's undo stack.
func (m *UserManager) PushBidToUndoStack(userID int, bid models.Bid) error {
	user, err := m.GetUser(userID)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	user.ActiveBids = append(user.ActiveBids, bid.ID) // Using ActiveBids as a stack for undo
	return nil
}

// PopBidFromUndoStack removes and returns the last bid from the user's undo stack.
func (m *UserManager) PopBidFromUndoStack(userID int) (int, error) {
	user, err := m.GetUser(userID)
	if err != nil {
		return 0, err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	n := len(user.ActiveBids)
	if n == 0 {
		return 0, errors.New("no bids to undo")
	}
	bidID := user.ActiveBids[n-1]
	user.ActiveBids = user.ActiveBids[:n-1]
	return bidID, nil
}

// DeductBalance deducts the amount from the user's balance.
func (m *UserManager) DeductBalance(userID int, amount float64) error {
	user, err := m.GetUser(userID)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if user.Balance < amount {
		return fmt.Errorf("insufficient balance: have %.2f, need %.2f", user.Balance, amount)
	}
	user.Balance -= amount
	return nil
}

// RestoreBalance adds the amount back to the user's balance.
func (m *UserManager) RestoreBalance(userID int, amount float64) error {
	user, err := m.GetUser(userID)
	if err != nil {
		return err
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	user.Balance += amount
	return nil
}
