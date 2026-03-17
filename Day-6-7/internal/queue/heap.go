package queue

import (
	"fmt"
	"ride-sharing/internal/models"
)

type RideQueue interface {
	Enqueue(ride *models.Ride)
	Dequeue() (*models.Ride, error)
	Peek() (*models.Ride, error)
	Size() int
	IsEmpty() bool
}

type MinHeap struct {
	items []*models.Ride
}

func NewMinHeap() *MinHeap {
	return &MinHeap{
		items: make([]*models.Ride, 0),
	}
}

func (h *MinHeap) Size() int {
	return len(h.items)
}

func (h *MinHeap) IsEmpty() bool {
	return len(h.items) == 0
}

func (h *MinHeap) Enqueue(ride *models.Ride) {
	h.items = append(h.items, ride)
	h.heapifyUp(len(h.items) - 1)
}

func (h *MinHeap) Dequeue() (*models.Ride, error) {
	if h.IsEmpty() {
		return nil, fmt.Errorf("queue is empty")
	}

	root := h.items[0]
	last := h.items[len(h.items)-1]
	h.items = h.items[:len(h.items)-1]

	if !h.IsEmpty() {
		h.items[0] = last
		h.heapifyDown(0)
	}

	return root, nil
}

func (h *MinHeap) Peek() (*models.Ride, error) {
	if h.IsEmpty() {
		return nil, fmt.Errorf("queue is empty")
	}
	return h.items[0], nil
}

func (h *MinHeap) heapifyUp(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		// Priority based on RequestTime (Min-Heap: oldest first)
		if h.items[index].RequestTime.Before(h.items[parent].RequestTime) {
			h.items[index], h.items[parent] = h.items[parent], h.items[index]
			index = parent
		} else {
			break
		}
	}
}

func (h *MinHeap) heapifyDown(index int) {
	lastIndex := len(h.items) - 1
	for {
		left := 2*index + 1
		right := 2*index + 2
		smallest := index

		if left <= lastIndex && h.items[left].RequestTime.Before(h.items[smallest].RequestTime) {
			smallest = left
		}
		if right <= lastIndex && h.items[right].RequestTime.Before(h.items[smallest].RequestTime) {
			smallest = right
		}

		if smallest != index {
			h.items[index], h.items[smallest] = h.items[smallest], h.items[index]
			index = smallest
		} else {
			break
		}
	}
}
