package main

type Node struct {
	packet *Packet
	next   *Node
	prev   *Node
}

type LinkedListPacketQueue struct {
	head *Node
	tail *Node
	size int
}

func NewLinkedListPacketQueue() *LinkedListPacketQueue {
	return &LinkedListPacketQueue{}
}

func (q *LinkedListPacketQueue) Enqueue(packet *Packet) {
	newNode := &Node{packet: packet}
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

func (q *LinkedListPacketQueue) Dequeue() *Packet {
	if q.head == nil {
		return nil
	}
	packet := q.head.packet
	q.head = q.head.next
	if q.head == nil {
		q.tail = nil
	} else {
	q.head.prev = nil
	}
	q.size--
	return packet
}

func (q *LinkedListPacketQueue) Peek() *Packet {
	if q.head == nil {
		return nil
	}
	return q.head.packet
}

func (q *LinkedListPacketQueue) Len() int {
	return q.size
}

func (q *LinkedListPacketQueue) Drop(id string) bool {
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
			return true
		}
		current = current.next
	}
	return false
}
