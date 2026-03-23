package heap

import "errors"

type MinHeap[T any] struct {
	Data       []T
	comparator func(T, T) bool
}

func NewMinHeap[T any](comparator func(T, T) bool) *MinHeap[T] {
	return &MinHeap[T]{
		Data:       make([]T, 0),
		comparator: comparator,
	}
}

func (h *MinHeap[T]) Insert(value T) {
	h.Data = append(h.Data, value)
	h.heapifyUp(len(h.Data) - 1)
}

func (h *MinHeap[T]) heapifyUp(index int) {
	for index > 0 {
		parent := (index - 1) / 2
		if !h.comparator(h.Data[index], h.Data[parent]) {
			return
		}
		h.Data[index], h.Data[parent] = h.Data[parent], h.Data[index]
		index = parent
	}
}

func (h *MinHeap[T]) ExtractMin() (T, error) {
	if len(h.Data) == 0 {
		var zeroValue T
		return zeroValue, errors.New("heap is empty")
	}
	min := h.Data[0]
	h.Data[0] = h.Data[len(h.Data)-1]
	h.Data = h.Data[:len(h.Data)-1]
	h.HeapifyDown(0)
	return min, nil
}

func (h *MinHeap[T]) HeapifyDown(index int) {
	for {
		left := 2*index + 1
		right := 2*index + 2
		min := index
		if left < len(h.Data) && h.comparator(h.Data[left], h.Data[min]) {
			min = left
		}
		if right < len(h.Data) && h.comparator(h.Data[right], h.Data[min]) {
			min = right
		}
		if min == index {
			return
		}
		h.Data[index], h.Data[min] = h.Data[min], h.Data[index]
		index = min
	}
}

func (h *MinHeap[T]) Size() int {
	return len(h.Data)
}

func (h *MinHeap[T]) Peek() (T, error) {
	if len(h.Data) == 0 {
		var zeroValue T
		return zeroValue, errors.New("heap is empty")
	}
	return h.Data[0], nil
}

func (h *MinHeap[T]) Update(index int, value T) error {
	if index < 0 || index >= len(h.Data) {
		return errors.New("invalid index")
	}
	h.Data[index] = value
	h.heapifyUp(index)
	h.HeapifyDown(index)
	return nil
}
