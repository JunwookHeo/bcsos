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

	// i := 0
	// for k, v := range BC.Chain {
	// 	log.Printf("%04d : %s", i, k)
	// 	log.Printf("prev : %s", hex.EncodeToString(v.Header.PrvHash))
	// 	i++
	// }
}

func GetLatestHash() []byte {
	return BC.Hash
}

func SetLatestHash(hash []byte) {
	BC.Hash = hash[:]
}
