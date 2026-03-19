package queue

import "errors"

type CircularQueue[T any] struct {
	items    []T
	capacity int
	front    int
	rear     int
	size     int
}

func NewCircularQueue[T any](capacity int) *CircularQueue[T] {
	return &CircularQueue[T]{
		items:    make([]T, capacity),
		capacity: capacity,
		front:    0,
		rear:     0,
		size:     0,
	}
}

func (q *CircularQueue[T]) Enqueue(item T) error {
	if q.IsFull() {
		return errors.New("queue is full")
	}
	q.items[q.rear] = item
	q.rear = (q.rear + 1) % q.capacity
	q.size++
	return nil
}

func (q *CircularQueue[T]) Dequeue() (T, error) {
	if q.IsEmpty() {
		var zero T
		return zero, errors.New("queue is empty")
	}
	item := q.items[q.front]
	q.front = (q.front + 1) % q.capacity
	q.size--
	return item, nil
}

func (q *CircularQueue[T]) Peek() (T, error) {
	if q.IsEmpty() {
		var zero T
		return zero, errors.New("queue is empty")
	}
	return q.items[q.front], nil
}

func (q *CircularQueue[T]) IsFull() bool {
	return q.size == q.capacity
}

func (q *CircularQueue[T]) IsEmpty() bool {
	return q.size == 0
}

func (q *CircularQueue[T]) Size() int {
	return q.size
}
