package dataController

import (
	"sync/atomic"
	"unsafe"
)

// LKQueue is a lock-free unbounded queue.
type GenericLKQueue[V any] struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}
type generic_node[V any] struct {
	value V
	next  unsafe.Pointer
}

// NewLKQueue returns an empty queue.
func NewGenericLKQueue[V any]() *GenericLKQueue[V] {
	n := unsafe.Pointer(&generic_node[V]{})
	return &GenericLKQueue[V]{head: n, tail: n}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *GenericLKQueue[V]) Enqueue(v V) {
	n := &generic_node[V]{value: v}
	for {
		tail := q.load_generic(&q.tail)
		next := q.load_generic(&tail.next)
		if tail == q.load_generic(&q.tail) { // are tail and next consistent?
			if next == nil {
				if q.cas_generic(&tail.next, next, n) {
					q.cas_generic(&q.tail, tail, n) // Enqueue is done.  try to swing tail to the inserted node
					return
				}
			} else { // tail was not pointing to the last node
				// try to swing Tail to the next node
				q.cas_generic(&q.tail, tail, next)
			}
		}
	}
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *GenericLKQueue[V]) Dequeue() *V {
	for {
		head := q.load_generic(&q.head)
		tail := q.load_generic(&q.tail)
		next := q.load_generic(&head.next)
		if head == q.load_generic(&q.head) { // are head, tail, and next consistent?
			if head == tail { // is queue empty or tail falling behind?
				if next == nil { // is queue empty?
					return nil
				}
				// tail is falling behind.  try to advance it
				q.cas_generic(&q.tail, tail, next)
			} else {
				// read value before CAS otherwise another dequeue might free the next node

				if q.cas_generic(&q.head, head, next) {
					return &next.value // Dequeue is done.  return
				}
			}
		}
	}
}
func (q *GenericLKQueue[V]) load_generic(p *unsafe.Pointer) (n *generic_node[V]) {
	return (*generic_node[V])(atomic.LoadPointer(p))
}
func (q *GenericLKQueue[V]) cas_generic(p *unsafe.Pointer, old, new *generic_node[V]) (ok bool) {
	return atomic.CompareAndSwapPointer(
		p, unsafe.Pointer(old), unsafe.Pointer(new))
}
