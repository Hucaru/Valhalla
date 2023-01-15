package dataController

import (
	"github.com/Hucaru/Valhalla/mpacket"
	"golang.org/x/exp/rand"
	"sync"
)

type mapElement struct {
	nextNum int64
	ele     mpacket.Packet
}

type MapQueue struct {
	data    map[int64]mapElement
	lastNum int64
	mapLock sync.Mutex
}

// NewSliceQueue returns an empty queue.
// You can give a initial capacity.
func NewMapQueue() (q *MapQueue) {
	return &MapQueue{data: map[int64]mapElement{}, lastNum: 0, mapLock: sync.Mutex{}}
}

// Enqueue puts the given value v at the tail of the queue.
func (q *MapQueue) Enqueue(v mpacket.Packet) {
	q.mapLock.Lock()
	curNum := q.lastNum
	for {
		q.lastNum = rand.Int63()
		_, ok := q.data[q.lastNum]
		if ok {
			continue
		}

		break;
	}

	q.data[q.lastNum] = mapElement{ele: v, nextNum: curNum}
	q.mapLock.Unlock()
}

// Dequeue removes and returns the value at the head of the queue.
// It returns nil if the queue is empty.
func (q *MapQueue) Dequeue() mpacket.Packet {
	q.mapLock.Lock()
	defer q.mapLock.Unlock()
	v, ok := q.data[q.lastNum]
	if ok {
		q.lastNum = v.nextNum
		delete(q.data, q.lastNum)
		return v.ele
	}
	return nil
}
