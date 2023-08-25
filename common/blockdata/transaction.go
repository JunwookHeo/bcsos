package blockdata

import (
	"crypto/sha256"
	"encoding/hex"
)

type TransactionHeader struct {
	Hash    string
	Witness uint32
}

type Transaction struct {
	Header TransactionHeader
	TxBuf  []byte
}

func NewTransaction() *Transaction {
	return &Transaction{}
}

func (t *Transaction) SetHash() {
	hash := sha256.Sum256(t.TxBuf)
	hash = sha256.Sum256(hash[:])
	t.Header.Hash = hex.EncodeToString(t.ReverseBuf(hash[:]))
}

func (t *Transaction) AppendBuf(buf []byte) {
	t.TxBuf = append(t.TxBuf, buf...)
}

func (t *Transaction) ReverseBuf(buf []byte) []byte {
	n := len(buf)
	des := make([]byte, n)
	copy(des, buf)
	for i := 0; i < n/2; i++ {
		des[i], des[n-1-i] = des[n-1-i], des[i]
	}
	return des
}
