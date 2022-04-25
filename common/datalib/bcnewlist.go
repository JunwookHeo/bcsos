package datalib

import (
	"errors"
	"log"
	"sync"

	"github.com/google/go-cmp/cmp"
)

type BcNewList struct {
	mutex    sync.Mutex
	capacity int
	pos      int
	items    []*BlockData
}

func (q *BcNewList) Push(item *BlockData) {
	if len(q.items) == int(q.capacity) {
		q.Pop()
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()
	q.items = append(q.items, item)
}

func (q *BcNewList) Pop() (*BlockData, error) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.items) > 0 {
		item := q.items[0]
		q.items = q.items[1:]
		return item, nil
	}

	return nil, errors.New("queue is empty")
}

func (sl *BcNewList) Find(item *BlockData) int {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	for i, d := range sl.items {
		if cmp.Equal(item, d) {
			return i
		}
	}

	return -1
}

func (sl *BcNewList) Next() (int, *BlockData) {
	sl.pos++
	if sl.pos >= len(sl.items) {
		sl.pos = -1
		return sl.pos, nil
	}
	return sl.pos, sl.items[sl.pos]
}

func (sl *BcNewList) ShowAll() {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	log.Printf("BcNewList Start [[[[[[ ")
	for i, d := range sl.items {
		log.Printf("BcNewList %v: %v", i, d)
	}
	log.Printf("]]]]]]]] BcNewList End")

}

func NewBcNewList() *BcNewList {
	return &BcNewList{
		capacity: 10,
		pos:      -1,
		items:    make([]*BlockData, 0),
	}
}
