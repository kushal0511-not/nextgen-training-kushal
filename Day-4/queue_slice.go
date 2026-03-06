package main

type SlicePacketQueue struct {
	packets []*Packet
}

func NewSlicePacketQueue() *SlicePacketQueue {
	return &SlicePacketQueue{
		packets: make([]*Packet, 0),
	}
}

func (q *SlicePacketQueue) Enqueue(packet *Packet) {
	q.packets = append(q.packets, packet)
}

func (q *SlicePacketQueue) Dequeue() *Packet {
	if len(q.packets) == 0 {
		return nil
	}
	packet := q.packets[0]
	q.packets = q.packets[1:]
	return packet
}

func (q *SlicePacketQueue) Peek() *Packet {
	if len(q.packets) == 0 {
		return nil
	}
	return q.packets[0]
}

func (q *SlicePacketQueue) Len() int {
	return len(q.packets)
}

func (q *SlicePacketQueue) Drop(id string) bool {
	for i, p := range q.packets {
		if p.ID == id {
			q.packets = append(q.packets[:i], q.packets[i+1:]...)
			return true
		}
	}
	return false
}
