package dataController

import (
	"github.com/Hucaru/Valhalla/mpacket"
	"sync"
)

// SliceQueue is an unbounded queue which uses a slice as underlying.
type SliceQueue struct {
	data []mpacket.Packet
	mu   sync.Mutex
}

// NewSliceQueue returns an empty queue.
// You can give a initial capacity.
func NewSliceQueue(n int) (q *SliceQueue) {
	return &SliceQueue{data: make([]mpacket.Packet, 0, n)}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *SliceQueue) Enqueue(v mpacket.Packet) {
	q.mu.Lock()
	q.data = append(q.data, v)
	q.mu.Unlock()
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *SliceQueue) Dequeue() mpacket.Packet {
	q.mu.Lock()
	if len(q.data) == 0 {
		q.mu.Unlock()
		return nil
	}
	v := q.data[0]
	q.data = q.data[1:]
	q.mu.Unlock()
	return v
}
