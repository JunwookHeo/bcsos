package datalib

import (
	"errors"
	"log"
	"sync"

	"github.com/google/go-cmp/cmp"
)

type BcQueue struct {
	mutex    sync.Mutex
	capacity int
	items    []interface{}
}

func (q *BcQueue) Push(item interface{}) {
	if len(q.items) == int(q.capacity) {
		q.Pop()
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, item)
}

func (q *BcQueue) Pop() (interface{}, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.items) > 0 {
		item := q.items[0]
		q.items = q.items[1:]
		return item, nil
	}

	return nil, errors.New("queue is empty")
}

func (q *BcQueue) Find(item interface{}) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for _, d := range q.items {
		if cmp.Equal(item, d) {
			return true
		}
	}

	return false
}

func (q *BcQueue) ShowAll() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for _, d := range q.items {
		log.Printf("BcQueleu : %v", d)
	}

}

func NewBcQueue(capacity int) *BcQueue {
	return &BcQueue{
		capacity: capacity,
		items:    make([]interface{}, capacity),
	}
}
