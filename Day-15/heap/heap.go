package heap

import (
	"errors"

	"github.com/nextgen-training-kushal/Day-15/models"
)

// PrioritizedItem represents an item in the minheap with its weight/priority value.
type PrioritizedItem struct {
	ID    models.IntersectionID
	Value float64 // Represents travel time or distance
}

// MinHeap provides a priority queue implementation.
type MinHeap struct {
	Data []PrioritizedItem
}

// NewMinHeap initializes a new instance of MinHeap.
func NewMinHeap() *MinHeap {
	return &MinHeap{
		Data: make([]PrioritizedItem, 0),
	}
}

// Insert adds an item to the heap and maintains the heap property.
func (h *MinHeap) Insert(item PrioritizedItem) {
	h.Data = append(h.Data, item)
	h.heapifyUp(len(h.Data) - 1)
}

func (h *MinHeap) heapifyUp(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		if h.Data[index].Value >= h.Data[parent].Value {
			return
		}
		h.Data[index], h.Data[parent] = h.Data[parent], h.Data[index]
		index = parent
	}
}

// ExtractMin removes and returns the smallest item from the heap.
func (h *MinHeap) ExtractMin() (PrioritizedItem, error) {
	if len(h.Data) == 0 {
		return PrioritizedItem{}, errors.New("heap is empty")
	}
	min := h.Data[0]
	h.Data[0] = h.Data[len(h.Data)-1]
	h.Data = h.Data[:len(h.Data)-1]
	h.heapifyDown(0)
	return min, nil
}

func (h *MinHeap) heapifyDown(index int) {
	for {
		left := 2*index + 1
		right := 2*index + 2
		min := index
		if left < len(h.Data) && h.Data[left].Value < h.Data[min].Value {
			min = left
		}
		if right < len(h.Data) && h.Data[right].Value < h.Data[min].Value {
			min = right
		}
		if min == index {
			return
		}
		h.Data[index], h.Data[min] = h.Data[min], h.Data[index]
		index = min
	}
}

// Size returns the count of elements in the heap.
func (h *MinHeap) Size() int {
	return len(h.Data)
}
