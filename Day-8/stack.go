package main

// Node represents a node in the linked list
type Node[T any] struct {
	Value T
	Next  *Node[T]
}

// Stack interface defines the behavior for both implementations
type Stack[T any] interface {
	Push(item T)
	Pop() (T, bool)
	Peek() (T, bool)
	IsEmpty() bool
	Size() int
}

// SliceStack is a slice-based implementation of Stack
type SliceStack[T any] struct {
	items []T
}

func NewSliceStack[T any]() *SliceStack[T] {
	return &SliceStack[T]{}
}

func (s *SliceStack[T]) Push(item T) {
	s.items = append(s.items, item)
}

func (s *SliceStack[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	item := s.items[len(s.items)-1]
	s.items = s.items[:len(s.items)-1]
	return item, true
}

func (s *SliceStack[T]) Peek() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	return s.items[len(s.items)-1], true
}

func (s *SliceStack[T]) IsEmpty() bool {
	return len(s.items) == 0
}

func (s *SliceStack[T]) Size() int {
	return len(s.items)
}

// LinkedListStack is a linked-list based implementation of Stack
type LinkedListStack[T any] struct {
	top  *Node[T]
	size int
}

func NewLinkedListStack[T any]() *LinkedListStack[T] {
	return &LinkedListStack[T]{}
}

func (s *LinkedListStack[T]) Push(item T) {
	s.top = &Node[T]{Value: item, Next: s.top}
	s.size++
}

func (s *LinkedListStack[T]) Pop() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	item := s.top.Value
	s.top = s.top.Next
	s.size--
	return item, true
}

func (s *LinkedListStack[T]) Peek() (T, bool) {
	if s.IsEmpty() {
		var zero T
		return zero, false
	}
	return s.top.Value, true
}

func (s *LinkedListStack[T]) IsEmpty() bool {
	return s.top == nil
}

func (s *LinkedListStack[T]) Size() int {
	return s.size
}
