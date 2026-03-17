package rides

import (
	"fmt"
	"ride-sharing/internal/models"
	"sync"
)

type RideTracker interface {
	Add(ride *models.Ride)
	Remove(id string) (*models.Ride, error)
	Get(id string) (*models.Ride, error)
	List() []*models.Ride
}

type Node struct {
	Ride *models.Ride
	Prev *Node
	Next *Node
}

type LinkedTracker struct {
	head  *Node
	tail  *Node
	size  int
	index map[string]*Node
	mu    sync.RWMutex
}

func NewLinkedTracker() *LinkedTracker {
	return &LinkedTracker{
		index: make(map[string]*Node),
	}
}

func (t *LinkedTracker) Add(ride *models.Ride) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node := &Node{Ride: ride}
	t.index[ride.ID] = node

	if t.head == nil {
		t.head = node
		t.tail = node
	} else {
		node.Prev = t.tail
		t.tail.Next = node
		t.tail = node
	}
	t.size++
}

func (t *LinkedTracker) Remove(id string) (*models.Ride, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	node, exists := t.index[id]
	if !exists {
		return nil, fmt.Errorf("ride %s not found in active tracker", id)
	}

	delete(t.index, id)

	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		t.head = node.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		t.tail = node.Prev
	}

	t.size--
	return node.Ride, nil
}

func (t *LinkedTracker) Get(id string) (*models.Ride, error) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	node, exists := t.index[id]
	if !exists {
		return nil, fmt.Errorf("ride %s not found", id)
	}
	return node.Ride, nil
}

func (t *LinkedTracker) List() []*models.Ride {
	t.mu.RLock()
	defer t.mu.RUnlock()

	rides := make([]*models.Ride, 0, t.size)
	curr := t.head
	for curr != nil {
		rides = append(rides, curr.Ride)
		curr = curr.Next
	}
	return rides
}
