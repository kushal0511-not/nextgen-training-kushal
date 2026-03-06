package main

import (
	"fmt"
	"testing"
)

func runBenchmark(b *testing.B, queueType string, numPackets int) {
	classifier := &DefaultPacketClassifier{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router := NewRouter(classifier, queueType)
		b.StartTimer()
		// Enqueue
		for j := 0; j < numPackets; j++ {
			p := &Packet{
				ID:       fmt.Sprintf("pkt-%d", j),
				Protocol: TCP,
				TTL:      10,
			}
			router.Enqueue(p)
		}

		// Dequeue
		for j := 0; j < numPackets; j++ {
			router.Dequeue()
		}
		b.StopTimer()
	}
}

func BenchmarkLinkedList_100(b *testing.B)    { runBenchmark(b, "linkedlist", 100) }
func BenchmarkLinkedList_1000(b *testing.B)   { runBenchmark(b, "linkedlist", 1000) }
func BenchmarkLinkedList_10000(b *testing.B)  { runBenchmark(b, "linkedlist", 10000) }
func BenchmarkLinkedList_100000(b *testing.B) { runBenchmark(b, "linkedlist", 100000) }

func BenchmarkSlice_100(b *testing.B)    { runBenchmark(b, "slice", 100) }
func BenchmarkSlice_1000(b *testing.B)   { runBenchmark(b, "slice", 1000) }
func BenchmarkSlice_10000(b *testing.B)  { runBenchmark(b, "slice", 10000) }
func BenchmarkSlice_100000(b *testing.B) { runBenchmark(b, "slice", 100000) }

func BenchmarkPooled_100(b *testing.B)    { runBenchmark(b, "linkedlist_pooled", 100) }
func BenchmarkPooled_1000(b *testing.B)   { runBenchmark(b, "linkedlist_pooled", 1000) }
func BenchmarkPooled_10000(b *testing.B)  { runBenchmark(b, "linkedlist_pooled", 10000) }
func BenchmarkPooled_100000(b *testing.B) { runBenchmark(b, "linkedlist_pooled", 100000) }
