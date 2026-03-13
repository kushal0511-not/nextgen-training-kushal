package main

import (
	"testing"
	"time"
)

func TestCustomDNSCache_WildcardResolve(t *testing.T) {
	cache := NewCustomDNSCache()
	defer cache.Close()

	// Add some specific subdomains to the cache
	cache.AddRecord("sub.example.com", "192.168.1.10", 5*time.Minute)
	cache.AddRecord("api.example.com", "192.168.1.11", 5*time.Minute)
	cache.AddRecord("test.com", "10.0.0.1", 5*time.Minute)

	tests := []struct {
		name        string
		query       string
		expectFound bool
		expectedIPs []string // Using a list of possible IPs because hashmap iteration order is not guaranteed
	}{
		{
			name:        "Exact match",
			query:       "sub.example.com",
			expectFound: true,
			expectedIPs: []string{"192.168.1.10"},
		},
		{
			name:        "Wildcard match *.example.com",
			query:       "*.example.com",
			expectFound: true,
			expectedIPs: []string{"192.168.1.10", "192.168.1.11"},
		},
		{
			name:        "Wildcard match *.com",
			query:       "*.com",
			expectFound: true,
			expectedIPs: []string{"192.168.1.10", "192.168.1.11", "10.0.0.1"},
		},
		{
			name:        "Wildcard match no result",
			query:       "*.net",
			expectFound: false,
			expectedIPs: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ip, found := cache.Resolve(tt.query)

			if found != tt.expectFound {
				t.Errorf("expected found:%v but got %v", tt.expectFound, found)
			}

			if found {
				matchFound := false
				for _, expectedIP := range tt.expectedIPs {
					if ip == expectedIP {
						matchFound = true
						break
					}
				}
				if !matchFound {
					t.Errorf("got IP %s, expected one of %v", ip, tt.expectedIPs)
				}
			}
		})
	}
}
