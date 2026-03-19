package queue

import (
	"testing"
)

func TestCircularQueue(t *testing.T) {
	tests := []struct {
		name     string
		capacity int
		ops      func(*CircularQueue[int]) []interface{}
		expected []interface{}
	}{
		{
			name:     "Basic operations",
			capacity: 3,
			ops: func(q *CircularQueue[int]) []interface{} {
				var results []interface{}
				results = append(results, q.Enqueue(1))
				results = append(results, q.Enqueue(2))
				val, err := q.Dequeue()
				results = append(results, val, err)
				results = append(results, q.Enqueue(3))
				return results
			},
			expected: []interface{}{nil, nil, 1, nil, nil},
		},
		{
			name:     "Queue Full",
			capacity: 2,
			ops: func(q *CircularQueue[int]) []interface{} {
				q.Enqueue(1)
				q.Enqueue(2)
				return []interface{}{q.IsFull(), q.Enqueue(3)}
			},
			expected: []interface{}{true, "queue is full"},
		},
		{
			name:     "Queue Empty",
			capacity: 2,
			ops: func(q *CircularQueue[int]) []interface{} {
				val, err := q.Dequeue()
				return []interface{}{q.IsEmpty(), val, err}
			},
			expected: []interface{}{true, 0, "queue is empty"},
		},
		{
			name:     "Circular Wrap",
			capacity: 2,
			ops: func(q *CircularQueue[int]) []interface{} {
				q.Enqueue(1)
				q.Enqueue(2)
				q.Dequeue()
				q.Enqueue(3)
				v1, _ := q.Dequeue()
				v2, _ := q.Dequeue()
				return []interface{}{v1, v2}
			},
			expected: []interface{}{2, 3},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := NewCircularQueue[int](tt.capacity)
			results := tt.ops(q)
			for i, res := range results {
				var resVal interface{} = res
				if err, ok := res.(error); ok {
					resVal = err.Error()
				}
				if resVal != tt.expected[i] {
					t.Errorf("op %d: expected %v, got %v", i, tt.expected[i], resVal)
				}
			}
		})
	}
}
