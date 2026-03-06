package main

import "strings"

type DefaultPacketClassifier struct{}

func (c *DefaultPacketClassifier) Classify(packet *Packet) int {
	// Rule 1: ICMP ping -> priority 1
	if packet.Protocol == ICMP {
		return 1
	}

	// Rule 2: TCP SYN -> priority 2
	// For simplicity, we'll check if the payload contains "SYN" or if it's a specific flag
	// In a real router, this would be a bit set in the TCP header.
	if packet.Protocol == TCP && strings.Contains(string(packet.Payload), "SYN") {
		return 2
	}

	// Default priorities based on protocol if not specifically handled
	switch packet.Protocol {
	case TCP:
		return 3
	case UDP:
		return 4
	default:
		return 5
	}
}
