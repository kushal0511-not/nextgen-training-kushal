package main

import (
	"fmt"
)

const (
	defaultCapacity = 16
	loadFactor      = 0.75
)

// Node represents a single entry in the linked list bucket.
type Node struct {
	Key   string
	Value interface{}
	Next  *Node
}

// HashMap is our custom hash map implementation.
type HashMap struct {
	buckets []*Node
	size    int
}

// NewHashMap creates a new hash map with a default capacity.
func NewHashMap() *HashMap {
	return &HashMap{
		buckets: make([]*Node, defaultCapacity),
		size:    0,
	}
}

// hashFNV1a implements the FNV-1a hash function for strings.
// See: https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function
func hashFNV1a(key string) uint64 {
	const (
		offset64 uint64 = 14695981039346656037
		prime64  uint64 = 1099511628211
	)
	var hash uint64 = offset64
	for i := 0; i < len(key); i++ {
		hash ^= uint64(key[i])
		hash *= prime64
	}
	return hash
}

// getIndex returns the bucket index for a given key and current capacity.
func (m *HashMap) getIndex(key string, capacity int) int {
	return int(hashFNV1a(key) % uint64(capacity))
}

// Put adds or updates a key-value pair.
func (m *HashMap) Put(key string, value interface{}) {
	// Check if resizing is necessary
	if float64(m.size+1) > float64(len(m.buckets))*loadFactor {
		m.resize()
	}

	index := m.getIndex(key, len(m.buckets))

	// Check if the key already exists in the bucket
	current := m.buckets[index]
	var prev *Node
	for current != nil {
		if current.Key == key {
			current.Value = value // Update value
			return
		}
		prev = current
		current = current.Next
	}

	// Create a new node and add it to the bucket (chaining)
	newNode := &Node{Key: key, Value: value}
	if prev == nil {
		m.buckets[index] = newNode
	} else {
		prev.Next = newNode
	}
	m.size++
}

// Get retrieves the value associated with a key.
func (m *HashMap) Get(key string) (interface{}, bool) {
	index := m.getIndex(key, len(m.buckets))
	current := m.buckets[index]
	for current != nil {
		if current.Key == key {
			return current.Value, true
		}
		current = current.Next
	}
	return nil, false
}

// Delete removes a key and its value from the map.
func (m *HashMap) Delete(key string) bool {
	index := m.getIndex(key, len(m.buckets))
	current := m.buckets[index]
	var prev *Node

	for current != nil {
		if current.Key == key {
			if prev == nil {
				m.buckets[index] = current.Next
			} else {
				prev.Next = current.Next
			}
			m.size--
			return true
		}
		prev = current
		current = current.Next
	}
	return false
}

// Size returns the number of elements in the map.
func (m *HashMap) Size() int {
	return m.size
}

// resize doubles the bucket array size and rehashes all elements.
func (m *HashMap) resize() {
	oldBuckets := m.buckets
	newCapacity := len(oldBuckets) * 2
	m.buckets = make([]*Node, newCapacity)

	// Temporarily reset size to zero to correctly re-populate
	m.size = 0

	for _, firstNode := range oldBuckets {
		current := firstNode
		for current != nil {
			// Save the next pointer before moving the current node
			next := current.Next

			// Re-insert this node into the new bucket array
			index := m.getIndex(current.Key, newCapacity)

			// Chain it (insert at head for simplicity during resize)
			current.Next = m.buckets[index]
			m.buckets[index] = current
			m.size++

			current = next
		}
	}
}

// Iterate executes a callback for every key-value pair in the map.
func (m *HashMap) Iterate(fn func(key string, value interface{})) {
	for _, node := range m.buckets {
		current := node
		for current != nil {
			fn(current.Key, current.Value)
			current = current.Next
		}
	}
}

// Display prints the internal structure of the hash map (for debugging).
func (m *HashMap) Display() {
	fmt.Printf("HashMap (size: %d, capacity: %d):\n", m.size, len(m.buckets))
	for i, node := range m.buckets {
		if node != nil {
			fmt.Printf("Bucket %d: ", i)
			current := node
			for current != nil {
				fmt.Printf("[%s: %v] -> ", current.Key, current.Value)
				current = current.Next
			}
			fmt.Printf("nil\n")
		}
	}
}
