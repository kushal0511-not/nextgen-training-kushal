package main

import (
	"fmt"
)

const (
	defaultCapacity = 16
	loadFactor      = 0.75
)

//EXTRA : Key can be any comparable type

// Node represents a single entry in the linked list bucket.
type Slot[T comparable] struct {
	Key      T
	Value    interface{}
	Occupied bool
}

type Bucket[T comparable] struct {
	Values     [8]Slot[T]
	isOverflow bool
	Next       *Bucket[T]
}

// HashMap is our custom hash map implementation.
type HashMap[T comparable] struct {
	buckets []Bucket[T]
	size    int
}

// NewHashMap creates a new hash map with a default capacity.
func NewHashMap[T comparable]() *HashMap[T] {
	return &HashMap[T]{
		buckets: make([]Bucket[T], defaultCapacity),
		size:    0,
	}
}

// hashFNV1a implements the FNV-1a hash function for strings.
// See: https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function
func hashFNV1a[T comparable](key T) uint64 {
	const (
		offset64 uint64 = 14695981039346656037
		prime64  uint64 = 1099511628211
	)
	var hash uint64 = offset64
	keyStr := fmt.Sprintf("%v", key)
	for i := 0; i < len(keyStr); i++ {
		hash ^= uint64(keyStr[i])
		hash *= prime64
	}
	return hash
}

// getIndex returns the bucket index for a given key and current capacity.
func (m *HashMap[T]) getIndex(key T, capacity int) int {
	return int(hashFNV1a(key) % uint64(capacity))
}

// Put adds or updates a key-value pair.
func (m *HashMap[T]) Put(key T, value interface{}) {
	// Check if resizing is necessary
	// A bucket has 8 slots, so we factor that into the threshold
	if float64(m.size+1) > float64(len(m.buckets)*8)*loadFactor {
		m.resize()
	}

	index := m.getIndex(key, len(m.buckets))
	currentBucket := &m.buckets[index]

	// First pass: try to find an existing key to update
	for cb := currentBucket; cb != nil; cb = cb.Next {
		for i := 0; i < 8; i++ {
			if cb.Values[i].Occupied && cb.Values[i].Key == key {
				cb.Values[i].Value = value // Update value
				return
			}
		}
	}

	// Second pass: find the first available slot to insert
	for cb := currentBucket; ; cb = cb.Next {
		for i := 0; i < 8; i++ {
			if !cb.Values[i].Occupied {
				cb.Values[i].Key = key
				cb.Values[i].Value = value
				cb.Values[i].Occupied = true
				m.size++
				return
			}
		}

		if cb.Next == nil {
			// Bucket and all existing overflow are full, so append a new overflow bucket
			newBucket := &Bucket[T]{
				isOverflow: true,
			}
			newBucket.Values[0] = Slot[T]{Key: key, Value: value, Occupied: true}
			cb.Next = newBucket
			m.size++
			return
		}
	}
}

// Get retrieves the value associated with a key.
func (m *HashMap[T]) Get(key T) (interface{}, bool) {
	index := m.getIndex(key, len(m.buckets))
	currentBucket := &m.buckets[index]

	for currentBucket != nil {
		for i := 0; i < 8; i++ {
			if currentBucket.Values[i].Occupied && currentBucket.Values[i].Key == key {
				return currentBucket.Values[i].Value, true
			}
		}
		currentBucket = currentBucket.Next
	}
	return nil, false
}

// Delete removes a key and its value from the map.
func (m *HashMap[T]) Delete(key T) bool {
	index := m.getIndex(key, len(m.buckets))
	currentBucket := &m.buckets[index]

	for currentBucket != nil {
		for i := 0; i < 8; i++ {
			if currentBucket.Values[i].Occupied && currentBucket.Values[i].Key == key {
				currentBucket.Values[i].Occupied = false
				var zeroKey T
				currentBucket.Values[i].Key = zeroKey
				currentBucket.Values[i].Value = nil
				m.size--
				return true
			}
		}
		currentBucket = currentBucket.Next
	}
	return false
}

// Size returns the number of elements in the map.
func (m *HashMap[T]) Size() int {
	return m.size
}

// resize doubles the bucket array size and rehashes all elements.
func (m *HashMap[T]) resize() {
	oldBuckets := m.buckets
	newCapacity := len(oldBuckets) * 2
	m.buckets = make([]Bucket[T], newCapacity)

	// Temporarily reset size to zero to correctly re-populate
	m.size = 0

	for i := range oldBuckets {
		currentBucket := &oldBuckets[i]
		for currentBucket != nil {
			for j := 0; j < 8; j++ {
				slot := &currentBucket.Values[j]
				if slot.Occupied {
					m.Put(slot.Key, slot.Value)
				}
			}
			currentBucket = currentBucket.Next
		}
	}
}

// Iterate executes a callback for every key-value pair in the map.
func (m *HashMap[T]) Iterate(fn func(key T, value interface{})) {
	for i := range m.buckets {
		currentBucket := &m.buckets[i]
		for currentBucket != nil {
			for j := 0; j < 8; j++ {
				slot := &currentBucket.Values[j]
				if slot.Occupied {
					fn(slot.Key, slot.Value)
				}
			}
			currentBucket = currentBucket.Next
		}
	}
}

// Display prints the internal structure of the hash map (for debugging).
func (m *HashMap[T]) Display() {
	fmt.Printf("HashMap (size: %d, capacity: %d buckets):\n", m.size, len(m.buckets))
	for i := range m.buckets {
		currentBucket := &m.buckets[i]
		fmt.Printf("Bucket %d: ", i)
		for currentBucket != nil {
			if currentBucket.isOverflow {
				fmt.Printf("(Overflow) ")
			}
			for j := 0; j < 8; j++ {
				slot := &currentBucket.Values[j]
				if slot.Occupied {
					fmt.Printf("[%v: %v] ", slot.Key, slot.Value)
				}
			}
			fmt.Printf(" -> ")
			currentBucket = currentBucket.Next
		}
		fmt.Printf("nil\n")
	}
}
