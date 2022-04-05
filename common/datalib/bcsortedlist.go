package datalib

import (
	"log"
	"sync"

	"github.com/google/go-cmp/cmp"
	"github.com/junwookheo/bcsos/common/config"
)

type BcSortedList struct {
	mutex    sync.Mutex
	capacity int
	pos      int
	items    []*BlockData
}

func (sl *BcSortedList) swap(i, j int) {
	sl.items[i], sl.items[j] = sl.items[j], sl.items[i]
}

func (sl *BcSortedList) Add(item *BlockData) {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()
	for i, d := range sl.items {
		tarr := make([]interface{}, 1)
		tarr[0] = item
		if d.Timestamp > item.Timestamp {
			sl.items = append(sl.items[:i], append([]*BlockData{item}, sl.items[i:]...)...)
			return
		}
	}

	sl.items = append(sl.items, item)
}

func (sl *BcSortedList) Remove(item *BlockData) bool {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	for i, d := range sl.items {
		if cmp.Equal(item, d) {
			sl.items = append(sl.items[:i], sl.items[i+1:]...)
			return true
		}
	}

	return false
}

func (sl *BcSortedList) Find(item *BlockData) int {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	for i, d := range sl.items {
		if cmp.Equal(item, d) {
			return i
		}
	}

	return -1
}

func (sl *BcSortedList) Next() (int, *BlockData) {
	sl.pos++
	if sl.pos >= len(sl.items) {
		sl.pos = -1
		return sl.pos, nil
	}
	return sl.pos, sl.items[sl.pos]
}

func (sl *BcSortedList) Update(ts int64) {
	// gap := int64(((MAX_TREE_DEPTH + 2) * config.BLOCK_CREATE_PERIOD) * 1000000000) // Nano second
	gap := int64(config.BLOCK_CREATE_PERIOD * 1000000000)

	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	for i := 0; i < len(sl.items); i++ {
		// ts := time.Now().UnixNano()
		if (sl.items[i].Height >= 0) || (ts-sl.items[i].Timestamp > gap) {
			sl.items = append(sl.items[:i], sl.items[i+1:]...)
			i-- // the item is deleted, so the next item's position is current
		}
	}
}

func (sl *BcSortedList) ShowAll() {
	sl.mutex.Lock()
	defer sl.mutex.Unlock()

	log.Printf("BcSortedList Start [[[[[[ ")
	for i, d := range sl.items {
		log.Printf("BcSortedList %v: %v", i, d)
	}
	log.Printf("]]]]]]]] BcSortedList End")

}

func NewBcSortedList() *BcSortedList {
	return &BcSortedList{
		capacity: 0,
		pos:      -1,
		items:    make([]*BlockData, 0),
	}
}
