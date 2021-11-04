package blockchain

import "crypto/sha256"

type Transaction struct {
	Id   []byte
	Data []byte
}

func (t *Transaction) Hash() []byte {
	var hash [32]byte
	hash = sha256.Sum256(t.Serialize())
	return hash[:]
}

func (t *Transaction) Serialize() []byte {
	return Serialize(*t)
}

func (t *Transaction) Deserialize(d []byte) {
	Deserialize(d, t)
}

func CreateTransaction(d []byte) *Transaction {
	t := Transaction{nil, d[:]}
	t.Id = t.Hash()
	return &t
}
