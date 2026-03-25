package models

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

func TestItemMarshalJSON(t *testing.T) {
	bids := []Bid{
		{
			ID:        12,
			ItemID:    122,
			UserID:    2,
			Amount:    132,
			Timestamp: time.Now(),
		},
		{
			ID:        11,
			ItemID:    121,
			UserID:    2,
			Amount:    144,
			Timestamp: time.Now(),
		},
	}
	i := Item{
		ID:          12,
		Name:        "Kushal",
		Category:    "Kushal",
		Description: "Kushal",
		SellerID:    2,
		StartPrice:  23,
		CurrentBid:  &bids[0],
		BidHistory:  &BidNode{Bid: bids[0], Next: &BidNode{Bid: bids[1], Next: nil}},
		StartTime:   time.Now(),
		EndTime:     time.Now(),
		Status:      "Pending",
	}
	b, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))

}

func TestItemUnmarshalJSON(t *testing.T) {
	s := `{"bid_history":[{"id":12,"item_id":122,"user_id":2,"amount":132,"timestamp":"2026-03-25T12:44:41.135089497+05:30"},{"id":11,"item_id":121,"user_id":2,"amount":144,"timestamp":"2026-03-25T12:44:41.135089531+05:30"}],"id":12,"name":"Kushal","category":"Kushal","description":"Kushal","seller_id":2,"start_price":23,"current_bid":{"id":12,"item_id":122,"user_id":2,"amount":132,"timestamp":"2026-03-25T12:44:41.135089497+05:30"},"start_time":"2026-03-25T12:44:41.135089613+05:30","end_time":"2026-03-25T12:44:41.135089649+05:30","status":"Pending"}`

	var i Item
	err := json.Unmarshal([]byte(s), &i)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(i)
}

func TestItemUnmarshalMarshalJSON(t *testing.T) {

	s := `{"bid_history":[{"id":12,"item_id":122,"user_id":2,"amount":132,"timestamp":"2026-03-25T12:44:41.135089497+05:30"},{"id":11,"item_id":121,"user_id":2,"amount":144,"timestamp":"2026-03-25T12:44:41.135089531+05:30"}],"id":12,"name":"Kushal","category":"Kushal","description":"Kushal","seller_id":2,"start_price":23,"current_bid":{"id":12,"item_id":122,"user_id":2,"amount":132,"timestamp":"2026-03-25T12:44:41.135089497+05:30"},"start_time":"2026-03-25T12:44:41.135089613+05:30","end_time":"2026-03-25T12:44:41.135089649+05:30","status":"Pending"}`

	var i Item
	err := json.Unmarshal([]byte(s), &i)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(i)
	b, err := json.Marshal(i)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(b))

}
