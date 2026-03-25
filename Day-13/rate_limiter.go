package main

// import (
// 	"fmt"
// 	"time"
// )

// type RateLimiter struct {
// 	Tokens chan struct{}
// 	Rate   int //  tokens per second
// }

// func NewRateLimiter(capacity int, rate int) *RateLimiter {
// 	r := &RateLimiter{
// 		Tokens: make(chan struct{}, capacity),
// 		Rate:   rate,
// 	}
// 	for range capacity {
// 		r.Tokens <- struct{}{}
// 	}
// 	go r.Refill()

// 	return r
// }

// func (r *RateLimiter) Refill() {
// 	timer := time.NewTicker(time.Second / time.Duration(r.Rate))

// 	for range timer.C {
// 		select {
// 		case r.Tokens <- struct{}{}:
// 		default:
// 		}

// 	}
// }

// func (r *RateLimiter) Allow() {
// 	select {
// 	case <-r.Tokens:
// 		fmt.Println("Allowed")
// 	default:
// 		fmt.Println("Not Allowed")
// 	}
// }

// func main() {
// 	rl := NewRateLimiter(5, 1) // capacity=5, rate=1 token/sec

// 	for i := 0; i < 10; i++ {
// 		rl.Allow()
// 		time.Sleep(300 * time.Millisecond)
// 	}

// 	for i := 0; i < 5; i++ {
// 		rl.Allow()
// 	}
// }
