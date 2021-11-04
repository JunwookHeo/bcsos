package blockchain

type BlockHeader struct {
	Timestamp int64
	Hash      []byte
	Nonce     int
}

type Block struct {
	Header       BlockHeader
	Transactions []*Transaction
}

func CreateBlock(ts []*Transaction, prevhash []byte) *Block {
	return nil
}

func Genesis() *Block {
	return nil
}

func SerializeBlock(b *Block) []byte {
	return nil
}

func DeserializeBlock(d []byte) *Block {
	return nil
}
