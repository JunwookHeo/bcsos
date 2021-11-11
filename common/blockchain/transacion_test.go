package blockchain

import (
	"log"
	"testing"
	"time"
	"unsafe"

	"github.com/junwookheo/bcsos/common/serial"
	"github.com/stretchr/testify/assert"
)

func TestCreateTransaction(t *testing.T) {
	s := "test creating transaction"
	tr := CreateTransaction([]byte(s))
	assert.Equal(t, []byte(s), tr.Data)
	assert.NotEmpty(t, tr.Hash)
}

func TestSerializeTransaction(t *testing.T) {
	s := "test creating transaction"
	tr := CreateTransaction([]byte(s))
	b := serial.Serialize(tr)
	tr2 := Transaction{}
	serial.Deserialize(b, &tr2)
	assert.Equal(t, *tr, tr2)

}

func TestCheckSize(t *testing.T) {
	ts := time.Now()
	var s uint64 = 1
	var loc *Transaction
	log.Printf("time : %v - %v - %v", unsafe.Sizeof(ts), unsafe.Sizeof(s), unsafe.Sizeof(loc))
}
