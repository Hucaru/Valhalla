package dataController

import (
	"sync/atomic"
	"unsafe"
)

type NewGridInfo struct {
	OldRegionId int64
	NewRegionId int64
	OldGridX    int
	OldGridY    int
	NewGridX    int
	NewGridY    int
	AccountID   int64
}

// GridLKQueue is a lock-free unbounded queue.
type GridLKQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}
type grid_node struct {
	value NewGridInfo
	next  unsafe.Pointer
}

// NewGridLKQueue returns an empty queue.
func NewGridLKQueue() *GridLKQueue {
	n := unsafe.Pointer(&grid_node{})
	return &GridLKQueue{head: n, tail: n}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *GridLKQueue) Enqueue(v NewGridInfo) {
	n := &grid_node{value: v}
	for {
		tail := load_grid_node(&q.tail)
		next := load_grid_node(&tail.next)
		if tail == load_grid_node(&q.tail) { // are tail and next consistent?
			if next == nil {
				if cas_grid_node(&tail.next, next, n) {
					cas_grid_node(&q.tail, tail, n) // Enqueue is done.  try to swing tail to the inserted node
					return
				}
			} else { // tail was not pointing to the last node
				// try to swing Tail to the next node
				cas_grid_node(&q.tail, tail, next)
			}
		}
	}
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *GridLKQueue) Dequeue() *NewGridInfo {
	for {
		head := load_grid_node(&q.head)
		tail := load_grid_node(&q.tail)
		next := load_grid_node(&head.next)
		if head == load_grid_node(&q.head) { // are head, tail, and next consistent?
			if head == tail { // is queue empty or tail falling behind?
				if next == nil { // is queue empty?
					return nil
				}
				// tail is falling behind.  try to advance it
				cas_grid_node(&q.tail, tail, next)
			} else {
				// read value before CAS otherwise another dequeue might free the next node
				v := next.value
				if cas_grid_node(&q.head, head, next) {
					return &v // Dequeue is done.  return
				}
			}
		}
	}
}
func load_grid_node(p *unsafe.Pointer) (n *grid_node) {
	return (*grid_node)(atomic.LoadPointer(p))
}
func cas_grid_node(p *unsafe.Pointer, old, new *grid_node) (ok bool) {
	return atomic.CompareAndSwapPointer(
		p, unsafe.Pointer(old), unsafe.Pointer(new))
}
