package datalib

import (
	"log"
	"testing"
	"time"
)

func TestNewBcSortedList(t *testing.T) {
	tc := NewBcSortedList()
	log.Printf("Start TestNewBcSortedList")
	b1 := BlockData{-1, 1, "a", ""}
	tc.Add(&b1)
	//tc.PrintAll()
	b2 := BlockData{-1, 0, "b", "a"}
	tc.Add(&b2)
	//tc.PrintAll()
	b3 := BlockData{-1, 3, "c", "b"}
	tc.Add(&b3)
	//tc.PrintAll()
	b4 := BlockData{-1, 4, "c2", "b"}
	tc.Add(&b4)
	//tc.PrintAll()
	b5 := BlockData{-1, 2, "c3", "b"}
	tc.Add(&b5)

	b6 := BlockData{-1, 5, "d", "c2"}
	tc.Add(&b6)
	b7 := BlockData{-1, 7, "e", "d"}
	tc.Add(&b7)
	b8 := BlockData{-1, 6, "f", "e"}
	tc.Add(&b8)
	tc.ShowAll()

	pos := tc.Find(&b8)
	if pos >= 0 {
		log.Printf("Find : %v - %v", b8, pos)
	}

	tc.Remove(&b8)
	tc.ShowAll()

	for {
		i, d := tc.Next()
		if i < 0 {
			break
		}
		log.Printf("index %v, %v", i, d)
	}

	tc.Update(0)
	tc.ShowAll()
}

func TestUpdateBcSortedList(t *testing.T) {
	tc := NewBcSortedList()
	log.Printf("Start TestUpdateBcSortedList")
	b1 := BlockData{-1, time.Now().UnixNano(), "a", ""}
	tc.Add(&b1)
	//tc.PrintAll()
	b2 := BlockData{-1, time.Now().UnixNano(), "b", "a"}
	tc.Add(&b2)
	//tc.PrintAll()
	b3 := BlockData{-1, time.Now().UnixNano(), "c", "b"}
	tc.Add(&b3)
	//tc.PrintAll()
	b4 := BlockData{-1, time.Now().UnixNano(), "c2", "b"}
	tc.Add(&b4)
	//tc.PrintAll()
	b5 := BlockData{-1, time.Now().UnixNano(), "c3", "b"}
	tc.Add(&b5)

	b6 := BlockData{-1, time.Now().UnixNano(), "d", "c2"}
	tc.Add(&b6)
	b7 := BlockData{-1, time.Now().UnixNano(), "e", "d"}
	tc.Add(&b7)
	b8 := BlockData{-1, time.Now().UnixNano(), "f", "e"}
	tc.Add(&b8)
	tc.ShowAll()

	for {
		i, d := tc.Next()
		if i < 0 {
			break
		}

		if i == 1 || i == 2 || i == 5 || i == 4 {
			d.Height = i
		}
		log.Printf("index %v, %v", i, d)
	}

	tc.ShowAll()
	tc.Update(time.Now().UnixNano())
	tc.ShowAll()
}

func TestUpdate2BcSortedList(t *testing.T) {
	tc := NewBcSortedList()
	log.Printf("Start TestUpdate2BcSortedList")
	b1 := BlockData{-1, time.Now().UnixNano(), "a", ""}
	tc.Add(&b1)
	//tc.PrintAll()
	b2 := BlockData{-1, time.Now().UnixNano() - 50000000000, "b", "a"}
	tc.Add(&b2)
	//tc.PrintAll()
	b3 := BlockData{-1, time.Now().UnixNano() - 54000000000, "c", "b"}
	tc.Add(&b3)
	//tc.PrintAll()
	b4 := BlockData{-1, time.Now().UnixNano(), "c2", "b"}
	tc.Add(&b4)
	//tc.PrintAll()
	b5 := BlockData{-1, time.Now().UnixNano(), "c3", "b"}
	tc.Add(&b5)

	b6 := BlockData{-1, time.Now().UnixNano() - 61000000000, "d", "c2"}
	tc.Add(&b6)
	b7 := BlockData{-1, time.Now().UnixNano() - 60000000000, "e", "d"}
	tc.Add(&b7)
	b8 := BlockData{-1, time.Now().UnixNano(), "f", "e"}
	tc.Add(&b8)
	tc.ShowAll()

	tc.Update(time.Now().UnixNano())
	tc.ShowAll()
}
