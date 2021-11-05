package blockchain

import (
	"encoding/hex"
)

type BlockChain struct {
}

var BC map[string]*Block

func InitBC() {
	BC = make(map[string]*Block)
	tr := Transaction{nil, []byte("This is Genesis Block")}
	b := Genesis(&tr)
	key := hex.EncodeToString(b.Header.Hash)
	BC[key] = b
}
