package blockchain

import (
	"crypto/sha256"

	"github.com/junwookheo/bcsos/common/serial"
)

type Transaction struct {
	Id   []byte
	Data []byte
}

func (t *Transaction) Hash() []byte {
	hash := sha256.Sum256(serial.Serialize(t))
	return hash[:]
}

func CreateTransaction(d []byte) *Transaction {
	t := Transaction{nil, d[:]}
	t.Id = t.Hash()
	return &t
}
