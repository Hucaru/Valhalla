package dataController

import (
	"github.com/Hucaru/Valhalla/mpacket"
	"sync"
)

// CQueue is a concurrent unbounded queue which uses two-Lock concurrent queue qlgorithm.
type CQueue struct {
	head  *cnode
	tail  *cnode
	hlock sync.Mutex
	tlock sync.Mutex
}
type cnode struct {
	value mpacket.Packet
	next  *cnode
}

// NewCQueue returns an empty CQueue.
func NewCQueue() *CQueue {
	n := cnode{}
	return &CQueue{head: &n, tail: &n}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *CQueue) Enqueue(v mpacket.Packet) {
	n := cnode{value: v}
	q.tlock.Lock()
	q.tail.next = &n // Link node at the end of the linked list
	q.tail = &n      // Swing Tail to node
	q.tlock.Unlock()
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *CQueue) Dequeue() mpacket.Packet {
	q.hlock.Lock()
	n := q.head
	newHead := n.next
	if newHead == nil {
		q.hlock.Unlock()
		return nil
	}
	v := newHead.value
	newHead.value = nil
	q.head = newHead
	q.hlock.Unlock()
	return v
}
