package main

import (
	"fmt"
	"time"
)

type Protocol string

const (
	TCP  Protocol = "TCP"
	UDP  Protocol = "UDP"
	ICMP Protocol = "ICMP"
)

type Packet struct {
	ID        string
	SourceIP  string
	DestIP    string
	Protocol  Protocol
	Priority  int
	Payload   []byte
	Timestamp time.Time
	TTL       int
}

func (p Packet) String() string {
	return fmt.Sprintf("[%s] %s -> %s | Proto: %s | Prio: %d | TTL: %d | Time: %s",
		p.ID, p.SourceIP, p.DestIP, p.Protocol, p.Priority, p.TTL, p.Timestamp.Format(time.RFC3339))
}
