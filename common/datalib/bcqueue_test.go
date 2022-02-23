package datalib

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBlQueue(t *testing.T) {
	q := NewBcQueue(10)

	q.Push("aaaa")
	q.Push("bbbb")
	q.Push("cccc")
	assert.Equal(t, true, q.Find("bbbb"))
}

type CS struct {
	Item1 string
	Item2 []byte
	Item3 string
	Item4 *int
}

func TestNewBlQueueStruct(t *testing.T) {
	q := NewBcQueue(10)
	i := 100

	q.Push(CS{"a", []byte("aaa"), "a", &i})
	q.Push(CS{"b", []byte("aaa"), "b", nil})
	q.Push(CS{"c", []byte("aaa"), "c", &i})

	i = 1000
	assert.Equal(t, true, q.Find(CS{"c", []byte("aaa"), "c", &i}))
	
	log.Println(q)
	assert.Equal(t, false, q.Find(CS{"c", []byte("bbb"), "c", &i}))
}
