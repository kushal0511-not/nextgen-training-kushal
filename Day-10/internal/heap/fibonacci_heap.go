package heap

import (
	"errors"
	"math"
)

type Node[T any] struct {
	Value  T
	Degree int
	Parent *Node[T]
	Child  *Node[T]
	Left   *Node[T]
	Right  *Node[T]
	Marked bool
}

type FibonacciHeap[T any] struct {
	Min        *Node[T]
	NumNodes   int
	comparator func(T, T) bool
}

func NewFibonacciHeap[T any](comparator func(T, T) bool) *FibonacciHeap[T] {
	return &FibonacciHeap[T]{
		comparator: comparator,
	}
}

func (h *FibonacciHeap[T]) Insert(value T) {
	node := &Node[T]{
		Value: value,
	}
	node.Left = node
	node.Right = node

	if h.Min == nil {
		h.Min = node
	} else {
		h.Min.Left.Right = node
		node.Left = h.Min.Left
		node.Right = h.Min
		h.Min.Left = node

		if h.comparator(value, h.Min.Value) {
			h.Min = node
		}
	}
	h.NumNodes++
}

func (h *FibonacciHeap[T]) ExtractMin() (T, error) {
	z := h.Min
	if z == nil {
		var zeroValue T
		return zeroValue, errors.New("heap is empty")
	}

	if z.Child != nil {
		child := z.Child
		firstChild := child
		for {
			nextChild := child.Right
			// Add child to root list
			child.Left = h.Min.Left
			child.Right = h.Min
			h.Min.Left.Right = child
			h.Min.Left = child
			child.Parent = nil
			if nextChild == firstChild {
				break
			}
			child = nextChild
		}
	}

	// Remove z from root list
	z.Left.Right = z.Right
	z.Right.Left = z.Left

	if z == z.Right {
		h.Min = nil
	} else {
		h.Min = z.Right
		h.consolidate()
	}

	h.NumNodes--
	return z.Value, nil
}

func (h *FibonacciHeap[T]) consolidate() {
	// D(n) <= log_phi(n)
	maxDegree := int(math.Log2(float64(h.NumNodes))) + 2
	a := make([]*Node[T], maxDegree)

	rootList := make([]*Node[T], 0)
	curr := h.Min
	if curr != nil {
		first := curr
		for {
			rootList = append(rootList, curr)
			curr = curr.Right
			if curr == first {
				break
			}
		}
	}

	for _, w := range rootList {
		x := w
		d := x.Degree
		for d < len(a) && a[d] != nil {
			y := a[d]
			if h.comparator(y.Value, x.Value) {
				x, y = y, x
			}
			h.link(y, x)
			a[d] = nil
			d++
		}
		if d < len(a) {
			a[d] = x
		}
	}

	h.Min = nil
	for _, node := range a {
		if node != nil {
			if h.Min == nil {
				h.Min = node
				node.Left = node
				node.Right = node
			} else {
				node.Left = h.Min.Left
				node.Right = h.Min
				h.Min.Left.Right = node
				h.Min.Left = node
				if h.comparator(node.Value, h.Min.Value) {
					h.Min = node
				}
			}
		}
	}
}

func (h *FibonacciHeap[T]) link(y, x *Node[T]) {
	// Remove y from root list
	y.Left.Right = y.Right
	y.Right.Left = y.Left

	// Make y a child of x
	y.Parent = x
	if x.Child == nil {
		x.Child = y
		y.Left = y
		y.Right = y
	} else {
		y.Left = x.Child.Left
		y.Right = x.Child
		x.Child.Left.Right = y
		x.Child.Left = y
	}
	x.Degree++
	y.Marked = false
}

func (h *FibonacciHeap[T]) Size() int {
	return h.NumNodes
}

func (h *FibonacciHeap[T]) IsEmpty() bool {
	return h.NumNodes == 0
}
