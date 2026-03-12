package main

import (
	"sync"
)

var nodePool = sync.Pool{
	New: func() any {
		return &Node{}
	},
}

type LinkedListPooledPacketQueue struct {
	head *Node
	tail *Node
	size int
}

func NewLinkedListPooledPacketQueue() *LinkedListPooledPacketQueue {
	return &LinkedListPooledPacketQueue{}
}

func (q *LinkedListPooledPacketQueue) Enqueue(packet *Packet) {
	// Reusing Node from pool instead of 'new(Node)'
	newNode := nodePool.Get().(*Node)
	newNode.packet = packet
	newNode.next = nil
	newNode.prev = nil

	if q.tail == nil {
		q.head = newNode
		q.tail = newNode
	} else {
		newNode.prev = q.tail
		q.tail.next = newNode
		q.tail = newNode
	}
	q.size++
}

func (q *LinkedListPooledPacketQueue) Dequeue() *Packet {
	if q.head == nil {
		return nil
	}

	node := q.head
	packet := node.packet

	q.head = q.head.next
	if q.head == nil {
		q.tail = nil
	} else {
		q.head.prev = nil
	}
	q.size--

	// Clear node and return it to pool to reduce GC pressure
	node.packet = nil
	node.next = nil
	node.prev = nil
	nodePool.Put(node)

	return packet
}

func (q *LinkedListPooledPacketQueue) Peek() *Packet {
	if q.head == nil {
		return nil
	}
	return q.head.packet
}

func (q *LinkedListPooledPacketQueue) Len() int {
	return q.size
}

func (q *LinkedListPooledPacketQueue) Drop(id string) bool {
	current := q.head
	for current != nil {
		if current.packet.ID == id {
			if current.prev != nil {
				current.prev.next = current.next
			} else {
				q.head = current.next
			}

			if current.next != nil {
				current.next.prev = current.prev
			} else {
				q.tail = current.prev
			}

			q.size--

			// Return dropped node back to pool
			node := current
			node.packet = nil
			node.next = nil
			node.prev = nil
			nodePool.Put(node)

			return true
		}
		current = current.next
	}
	return false
}
