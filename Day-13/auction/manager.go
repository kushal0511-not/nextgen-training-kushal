package auction

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/nextgen-training-kushal/Day-13/bid"
	"github.com/nextgen-training-kushal/Day-13/category"
	"github.com/nextgen-training-kushal/Day-13/models"
	"github.com/nextgen-training-kushal/Day-13/user"
)

// ItemContext manages an item's auction state with fine-grained locking.
type ItemContext struct {
	Item *models.Item
	Heap *bid.BidHeap
	mu   sync.Mutex // Fine-grained lock per item
}

// AuctionManager coordinates auction operations across items and users.
type AuctionManager struct {
	items        map[int]*ItemContext
	userManager  *user.UserManager
	categoryTree *category.CategoryTree
	mu           sync.RWMutex // Protects the items map itself
}

// NewAuctionManager creates a new AuctionManager.
func NewAuctionManager(userManager *user.UserManager, categoryTree *category.CategoryTree) *AuctionManager {
	return &AuctionManager{
		items:        make(map[int]*ItemContext),
		userManager:  userManager,
		categoryTree: categoryTree,
	}
}

// RegisterItem adds an item to the auction manager.
func (am *AuctionManager) RegisterItem(item *models.Item) {
	am.mu.Lock()
	defer am.mu.Unlock()
	am.items[item.ID] = &ItemContext{
		Item: item,
		Heap: &bid.BidHeap{},
	}
}

// PlaceBid validates and places a bid on an item.
func (am *AuctionManager) PlaceBid(itemID, userID int, amount float64) error {
	am.mu.RLock()
	ctx, exists := am.items[itemID]
	am.mu.RUnlock()

	if !exists {
		return fmt.Errorf("item %d not found", itemID)
	}

	// Fine-grained lock for this item
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// 1. Validate auction active
	now := time.Now()
	if now.Before(ctx.Item.StartTime) || now.After(ctx.Item.EndTime) || ctx.Item.Status != models.StatusActive {
		return errors.New("auction is not currently active")
	}

	// 2. Validate amount > current bid
	if ctx.Item.CurrentBid != nil && amount <= ctx.Item.CurrentBid.Amount {
		return fmt.Errorf("bid amount %.2f must be greater than current bid %.2f", amount, ctx.Item.CurrentBid.Amount)
	}
	if amount <= ctx.Item.StartPrice {
		return fmt.Errorf("bid amount %.2f must be greater than start price %.2f", amount, ctx.Item.StartPrice)
	}

	// 3. Validate and deduct user balance
	if err := am.userManager.DeductBalance(userID, amount); err != nil {
		return err
	}

	// 4. Create and record bid
	newBid := models.Bid{
		ID:        int(time.Now().UnixNano()), // Simple unique ID
		ItemID:    itemID,
		UserID:    userID,
		Amount:    amount,
		Timestamp: time.Now(),
	}

	// Add to item history (Linked List)
	ctx.Item.AddBid(newBid)

	// Add to item heap (Max-Heap)
	ctx.Heap.PushBid(newBid)

	// Add to user undo stack
	if err := am.userManager.PushBidToUndoStack(userID, newBid); err != nil {
		// Rollback balance if pushing to undo stack fails (unexpected)
		_ = am.userManager.RestoreBalance(userID, amount)
		return err
	}

	return nil
}

// RetractBid retracts the last bid for a user on specific item.
func (am *AuctionManager) RetractBid(itemID, userID int) error {
	am.mu.RLock()
	ctx, exists := am.items[itemID]
	am.mu.RUnlock()

	if !exists {
		return fmt.Errorf("item %d not found", itemID)
	}

	// Fine-grained lock
	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	// 1. Pop from user undo stack
	bidID, err := am.userManager.PopBidFromUndoStack(userID)
	if err != nil {
		return err
	}

	// 2. Validate it's for this item
	if ctx.Item.CurrentBid == nil || ctx.Item.CurrentBid.ID != bidID {
		return errors.New("can only retract the current leading bid if it was your last bid")
	}

	amountToRestore := ctx.Item.CurrentBid.Amount

	// 3. Restore previous bid from history
	if ctx.Item.BidHistory != nil {
		ctx.Item.BidHistory = ctx.Item.BidHistory.Next
		if ctx.Item.BidHistory != nil {
			ctx.Item.CurrentBid = &ctx.Item.BidHistory.Bid
		} else {
			ctx.Item.CurrentBid = nil
		}
	}

	// 4. Handle Heap
	if ctx.Heap.SafeLen() > 0 {
		_ = ctx.Heap.PopBid()
	}

	// 5. Restore balance
	return am.userManager.RestoreBalance(userID, amountToRestore)
}

// EndAuction extracts the winner from the heap.
func (am *AuctionManager) EndAuction(itemID int) (*models.Bid, error) {
	am.mu.RLock()
	ctx, exists := am.items[itemID]
	am.mu.RUnlock()

	if !exists {
		return nil, fmt.Errorf("item %d not found", itemID)
	}

	ctx.mu.Lock()
	defer ctx.mu.Unlock()

	ctx.Item.Status = models.StatusEnded
	if ctx.Heap.Len() == 0 {
		return nil, nil // No winner
	}

	winner := ctx.Heap.PopBid()
	return &winner, nil
}

// BrowseCategory returns items in a specific category.
func (am *AuctionManager) BrowseCategory(path []string) ([]*models.Item, error) {
	_, err := am.categoryTree.FindCategory(path)
	if err != nil {
		return nil, err
	}

	am.mu.RLock()
	defer am.mu.RUnlock()

	var results []*models.Item
	catStr := ""
	if len(path) > 0 {
		catStr = path[len(path)-1]
	}

	for _, ctx := range am.items {
		if ctx.Item.Category == catStr {
			results = append(results, ctx.Item)
		}
	}
	return results, nil
}
