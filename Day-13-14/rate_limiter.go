package main

import (
	"sync"
	"time"
)

var (
	userRateLimiters = make(map[string]*RateLimiter)
	mu               sync.RWMutex
)

type RateLimiter struct {
	Tokens chan struct{}
	Rate   int //  tokens per second
}

func NewRateLimiter(capacity int, rate int, userID string) *RateLimiter {
	mu.Lock()
	defer mu.Unlock()
	r := &RateLimiter{
		Tokens: make(chan struct{}, capacity),
		Rate:   rate,
	}
	for range capacity {
		r.Tokens <- struct{}{}
	}
	userRateLimiters[userID] = r
	go r.Refill()

	return r
}

func (r *RateLimiter) Refill() {
	timer := time.NewTicker(time.Second / time.Duration(r.Rate))
	for range timer.C {
		select {
		case r.Tokens <- struct{}{}:
		default:
		}

	}
}

func (r *RateLimiter) Allow() bool {
	select {
	case <-r.Tokens:
		return true
	default:
		return false
	}
}
