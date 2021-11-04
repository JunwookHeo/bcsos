package blockchain

import "time"

type BlockHeader struct {
	Timestamp int64
	Hash      []byte
	Nonce     int
}

type Block struct {
	Header       BlockHeader
	Transactions []*Transaction
}

func CreateBlock(trs []*Transaction, prevhash []byte) *Block {
	h := BlockHeader{time.Now().Unix(), prevhash, 0}
	// TODO : create hash
	block := &Block{h, trs}

	return block
}

func Genesis() *Block {
	return nil
}

func (b *Block) Serialize() []byte {
	return Serialize(*b)
}

func (b *Block) Deserialize(d []byte) {
	Deserialize(d, b)
}
