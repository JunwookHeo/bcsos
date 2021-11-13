package blockchain

import (
	"bytes"
	"crypto/sha256"
	"time"
)

type Transaction struct {
	Hash      []byte
	Timestamp int64
	Data      []byte
}

func (t *Transaction) GetHash() []byte {
	data := bytes.Join(
		[][]byte{
			toHex(t.Timestamp),
			t.Data,
		},
		[]byte{},
	)

	hash := sha256.Sum256(data)
	return hash[:]
}

func CreateTransaction(d []byte) *Transaction {
	t := Transaction{nil, time.Now().UnixNano(), d[:]}
	t.Hash = t.GetHash()
	return &t
}
