package main

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

var packetPool = sync.Pool{
	New: func() interface{} {
		return &Packet{}
	},
}

func main() {
	rand.Seed(time.Now().UnixNano())

	classifier := &DefaultPacketClassifier{}
	router := NewRouter(classifier, "linkedlist_pooled") // Defaulting to linkedlist_pooled

	// Simulate 10K packets
	numPackets := 10000
	fmt.Printf("Simulating %d packets with Packet and Node pooling...\n", numPackets)

	start := time.Now()
	for i := 0; i < numPackets; i++ {
		p := packetPool.Get().(*Packet)

		p.ID = fmt.Sprintf("pkt-%d", i)
		p.SourceIP = "192.168.1.1"
		p.DestIP = "10.0.0.1"
		p.Protocol = randomProtocol()
		p.Payload = []byte("sample payload")
		p.Timestamp = time.Now()
		p.TTL = 10

		// Inject some TCP SYN packets
		if p.Protocol == TCP && rand.Float32() < 0.2 {
			p.Payload = []byte("SYN packet")
		}

		router.Enqueue(p)
	}

	enqueueDuration := time.Since(start)
	fmt.Printf("Enqueued %d packets in %v\n", numPackets, enqueueDuration)

	router.DisplayStatus()

	// Process (Dequeue) all packets
	fmt.Println("Processing packets and returning to pool...")
	start = time.Now()
	processedCount := 0
	for {
		p := router.Dequeue()
		if p == nil {
			break
		}

		// Return to pool after processing
		packetPool.Put(p)

		processedCount++
		// Occasionally log processing
		if processedCount%2000 == 0 {
			fmt.Printf("Processed %d/%d packets...\n", processedCount, numPackets)
		}
	}
	dequeueDuration := time.Since(start)

	fmt.Printf("Processed %d packets in %v\n", processedCount, dequeueDuration)
	fmt.Printf("Total time: %v\n", enqueueDuration+dequeueDuration)
}

func randomProtocol() Protocol {
	protos := []Protocol{TCP, UDP, ICMP}
	return protos[rand.Intn(len(protos))]
}
