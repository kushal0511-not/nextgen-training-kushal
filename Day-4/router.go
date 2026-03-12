package main

import (
	"fmt"
	"log"
)

type Router struct {
	queues     map[int]PacketQueue
	classifier PacketClassifier
}

func NewRouter(classifier PacketClassifier, queueType string) *Router {
	queues := make(map[int]PacketQueue)
	for i := 1; i <= 5; i++ {
		switch queueType {
		case "linkedlist_pooled":
			queues[i] = NewLinkedListPooledPacketQueue()
		case "linkedlist":
			queues[i] = NewLinkedListPacketQueue()
		default:
			queues[i] = NewSlicePacketQueue()
		}
	}
	return &Router{
		queues:     queues,
		classifier: classifier,
	}
}

func (r *Router) Enqueue(packet *Packet) {
	if packet.TTL <= 0 {
		log.Printf("[DROP] Packet %s dropped. reason: TTL=0", packet.ID)
		return
	}

	priority := r.classifier.Classify(packet)
	packet.Priority = priority

	if q, ok := r.queues[priority]; ok {
		q.Enqueue(packet)
	} else {
		log.Printf("[ERROR] Invalid priority %d for packet %s", priority, packet.ID)
	}
}

func (r *Router) Dequeue() *Packet {
	// Process highest priority first (1 is highest)
	for i := 1; i <= 5; i++ {

		if r.queues[i].Len() > 0 {
			return r.queues[i].Dequeue()
		}
	}
	return nil
}

func (r *Router) Reorder(packetID string, fromPriority, toPriority int) bool {
	fromQ, fromOk := r.queues[fromPriority]
	toQ, toOk := r.queues[toPriority]
	if !fromOk || !toOk {
		return false
	}

	var found *Packet
	var temp []*Packet

	// Dequeue all to find the specific packet
	for fromQ.Len() > 0 {
		p := fromQ.Dequeue()
		if p.ID == packetID {
			found = p
		} else {
			temp = append(temp, p)
		}
	}

	// Re-enqueue everything else back
	for _, p := range temp {
		fromQ.Enqueue(p)
	}

	// Enqueue the found packet to the new priority queue
	if found != nil {
		found.Priority = toPriority
		toQ.Enqueue(found)
		return true
	}

	return false
}

func (r *Router) DisplayStatus() {
	fmt.Println("\n--- Router Queue Status ---")

	for i := 1; i <= 5; i++ {
		fmt.Printf("Priority %d: %d packets\n", i, r.queues[i].Len())
	}
	fmt.Println("---------------------------")
}
