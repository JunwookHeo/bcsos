package blockchain

import (
	"time"
)

type BlockHeader struct {
	Hash       []byte
	PrvHash    []byte
	MerkleRoot []byte
	Timestamp  int64
	Difficulty int
	Nonce      int
}

type Block struct {
	Header       BlockHeader
	Transactions []*Transaction
}

func CreateBlock(trs []*Transaction, prevhash []byte) *Block {
	h := BlockHeader{nil, prevhash, nil, time.Now().Unix(), 0, 0}
	block := &Block{h, trs}
	block.Header.MerkleRoot = block.MerkleRoot()

	nonce, hash := ProofWork(block)

	block.Header.Hash = hash[:]
	block.Header.Nonce = nonce

	return block
}

func Genesis(t *Transaction) *Block {
	return CreateBlock([]*Transaction{t}, []byte{})
}

func (b *Block) MerkleRoot() []byte {
	var hashes [][]byte

	for _, tr := range b.Transactions {
		hashes = append(hashes, tr.Serialize())
	}
	root := CalMerkleRootHash(hashes)

	return root
}

func (b *Block) Serialize() []byte {
	return Serialize(*b)
}

func (b *Block) Deserialize(d []byte) {
	Deserialize(d, b)
}

func (b *Block) PoW() {

}
