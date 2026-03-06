package main

type PacketQueue interface {
	Enqueue(packet *Packet)
	Dequeue() *Packet
	Peek() *Packet
	Len() int
	Drop(id string) bool
}

type PacketClassifier interface {
	Classify(packet *Packet) int
}
