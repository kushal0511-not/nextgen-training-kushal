package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestRateLimitMiddleware(t *testing.T) {
	// Reset the globals for a clean test state definition
	mu.Lock()
	userRateLimiters = make(map[string]*RateLimiter)
	mu.Unlock()

	// A dummy handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Wrap handler with middleware
	middleware := RateLimitMiddleware(handler)

	userID := "test-user"

	// 1. Send 5 allowed requests
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		req.Header.Set("X-User-ID", userID)
		rr := httptest.NewRecorder()

		middleware.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("Request %d: expected status OK, got %v", i+1, status)
		}
	}

	// 2. 6th request should fail with TooManyRequests immediately
	reqFail := httptest.NewRequest(http.MethodGet, "/", nil)
	reqFail.Header.Set("X-User-ID", userID)
	rrFail := httptest.NewRecorder()
	middleware.ServeHTTP(rrFail, reqFail)

	if status := rrFail.Code; status != http.StatusTooManyRequests {
		t.Errorf("Request 6 (rate limited): expected status Too Many Requests (429), got %v", status)
	}

	// 3. Wait for 1 second to allow tokens to replenish
	time.Sleep(1 * time.Second)

	// 4. Request should succeed again
	reqSuccess := httptest.NewRequest(http.MethodGet, "/", nil)
	reqSuccess.Header.Set("X-User-ID", userID)
	rrSuccess := httptest.NewRecorder()
	middleware.ServeHTTP(rrSuccess, reqSuccess)

	if status := rrSuccess.Code; status != http.StatusOK {
		t.Errorf("Request after waiting: expected status OK, got %v", status)
	}
}
