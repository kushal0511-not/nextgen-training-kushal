package models

import (
	"encoding/json"
	"time"
)

const (
	StatusActive    = "Active"
	StatusEnded     = "Ended"
	StatusCancelled = "Cancelled"
)

type User struct {
	ID         int     `json:"id"`
	Name       string  `json:"name"`
	Balance    float64 `json:"balance"`
	ActiveBids []int   `json:"active_bids"`
}

type Bid struct {
	ID        int       `json:"id"`
	ItemID    int       `json:"item_id"`
	UserID    int       `json:"user_id"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
}

type AuctionStats struct {
	TotalItems   int     `json:"total_items"`
	ActiveItems  int     `json:"active_items"`
	EndedItems   int     `json:"ended_items"`
	TotalBids    int64   `json:"total_bids"`
	TotalRevenue float64 `json:"total_revenue"`
}

// BidNode represents a node in the linked list for bid history.
type BidNode struct {
	Bid  Bid
	Next *BidNode
}

type Item struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	SellerID    int       `json:"seller_id"`
	StartPrice  float64   `json:"start_price"`
	CurrentBid  *Bid      `json:"current_bid"`
	BidHistory  *BidNode  `json:"-"` // Hidden from JSON by default
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
}

// MarshalJSON customizes the JSON representation of Item.
// It converts the internal BidHistory linked list to a slice for external consumers.
func (i Item) MarshalJSON() ([]byte, error) {
	type Alias Item
	return json.Marshal(&struct {
		BidHistory []Bid `json:"bid_history"`
		Alias
	}{
		BidHistory: i.GetBidHistorySlice(),
		Alias:      (Alias)(i),
	})
}

// GetBidHistorySlice converts the internal linked list to a slice.
func (i *Item) GetBidHistorySlice() []Bid {
	var bids []Bid
	current := i.BidHistory
	for current != nil {
		bids = append(bids, current.Bid)
		current = current.Next
	}
	return bids
}

// AddBid adds a new bid to the linked list history.
func (i *Item) AddBid(bid Bid) {
	newNode := &BidNode{Bid: bid, Next: i.BidHistory}
	i.BidHistory = newNode
	i.CurrentBid = &bid
}

// UnmarshalJSON customizes the JSON unmarshaling of Item.
// It reconstructs the internal BidHistory linked list from the slice.
func (i *Item) UnmarshalJSON(data []byte) error {
	type Alias Item
	aux := &struct {
		BidHistory []Bid `json:"bid_history"`
		*Alias
	}{
		Alias: (*Alias)(i),
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	i.BidHistory = nil
	var tail *BidNode
	for _, bid := range aux.BidHistory {
		newNode := &BidNode{Bid: bid}
		if i.BidHistory == nil {
			i.BidHistory = newNode
		} else {
			tail.Next = newNode
		}
		tail = newNode
	}

	return nil
}
