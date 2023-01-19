package dataController

import (
	"sync/atomic"
	"unsafe"
)

type ActionSync struct {
	Fn   func()
	Time int64
}

// GridLKQueue is a lock-free unbounded queue.
type MoveSyncLKQueue struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}
type action_sync_node struct {
	value ActionSync
	next  unsafe.Pointer
}

// NewGridLKQueue returns an empty queue.
func NewMoveSyncLKQueue() *MoveSyncLKQueue {
	n := unsafe.Pointer(&grid_node{})
	return &MoveSyncLKQueue{head: n, tail: n}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *MoveSyncLKQueue) Enqueue(v ActionSync) {
	n := &action_sync_node{value: v}
	for {
		tail := load_action_sync_node(&q.tail)
		next := load_action_sync_node(&tail.next)
		if tail == load_action_sync_node(&q.tail) { // are tail and next consistent?
			if next == nil {
				if cas_action_sync_node(&tail.next, next, n) {
					cas_action_sync_node(&q.tail, tail, n) // Enqueue is done.  try to swing tail to the inserted node
					return
				}
			} else { // tail was not pointing to the last node
				// try to swing Tail to the next node
				cas_action_sync_node(&q.tail, tail, next)
			}
		}
	}
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *MoveSyncLKQueue) Dequeue() *ActionSync {
	for {
		head := load_action_sync_node(&q.head)
		tail := load_action_sync_node(&q.tail)
		next := load_action_sync_node(&head.next)
		if head == load_action_sync_node(&q.head) { // are head, tail, and next consistent?
			if head == tail { // is queue empty or tail falling behind?
				if next == nil { // is queue empty?
					return nil
				}
				// tail is falling behind.  try to advance it
				cas_action_sync_node(&q.tail, tail, next)
			} else {
				// read value before CAS otherwise another dequeue might free the next node
				v := next.value
				if cas_action_sync_node(&q.head, head, next) {
					return &v // Dequeue is done.  return
				}
			}
		}
	}
}
func load_action_sync_node(p *unsafe.Pointer) (n *action_sync_node) {
	return (*action_sync_node)(atomic.LoadPointer(p))
}
func cas_action_sync_node(p *unsafe.Pointer, old, new *action_sync_node) (ok bool) {
	return atomic.CompareAndSwapPointer(
		p, unsafe.Pointer(old), unsafe.Pointer(new))
}
