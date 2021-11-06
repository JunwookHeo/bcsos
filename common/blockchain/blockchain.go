package blockchain

import (
	"encoding/hex"
)

type BlockChain struct {
	Hash  []byte
	Chain map[string]*Block
}

var BC BlockChain

func InitBC() {
	BC.Chain = make(map[string]*Block)
	tr := Transaction{nil, []byte("This is Genesis Block")}
	b := Genesis(&tr)
	AddBlock(b)
}

func AddBlock(b *Block) {
	key := hex.EncodeToString(b.Header.Hash)
	BC.Chain[key] = b
	BC.Hash = b.Header.Hash[:]
}

func GetLatestHash() []byte {
	return BC.Hash
}

func SetLatestHash(hash []byte) {
	BC.Hash = hash[:]
}
