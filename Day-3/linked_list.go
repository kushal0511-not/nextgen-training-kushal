package main

import "reflect"

type SimpleNode[T any] struct {
	data T
	next *SimpleNode[T]
	prev *SimpleNode[T]
}
type LinkedList[T any] struct {
	Head *SimpleNode[T]
}

func (l *LinkedList[T]) Insert(data T) {
	newNode := &SimpleNode[T]{data: data}
	if l.Head == nil {
		l.Head = newNode
		return
	}
	newNode.next = l.Head
	l.Head.prev = newNode
	l.Head = newNode
}

func (l *LinkedList[T]) Delete(data T) {
	if l.Head == nil {
		return
	}
	nodeToDelete := l.Find(data)
	if nodeToDelete == nil {
		return
	}
	if nodeToDelete.prev != nil {
		nodeToDelete.prev.next = nodeToDelete.next
	}
	if nodeToDelete.next != nil {
		nodeToDelete.next.prev = nodeToDelete.prev
	}
	if nodeToDelete == l.Head {
		l.Head = nodeToDelete.next
	}
	
}

func (l *LinkedList[T]) Size() int {
	count := 0
	current := l.Head
	for current != nil {
		count++
		current = current.next
	}
	return count
}

func (l *LinkedList[T]) Find(data T) *SimpleNode[T] {
	current := l.Head
	for current != nil {
		if reflect.DeepEqual(current.data, data) {
			return current
		}
		current = current.next
	}
	return nil
}
