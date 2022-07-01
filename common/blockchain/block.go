package blockchain

import (
	"bytes"
	"crypto/sha256"
	"time"

	"github.com/junwookheo/bcsos/common/wallet"
)

type BlockHeader struct {
	Hash       []byte
	PrvHash    []byte
	MerkleRoot []byte
	Timestamp  int64
	Difficulty int
	Nonce      int
	Height     int
}

type Block struct {
	Header       BlockHeader
	Transactions []*Transaction
}

func CreateBlock(trs []*Transaction, prevhash []byte, height int) *Block {
	h := BlockHeader{nil, prevhash, nil, time.Now().UnixNano(), 0, 0, height}
	block := &Block{h, trs}
	block.Header.MerkleRoot = block.MerkleRoot()

	difficulty, nonce, hash := ProofWork(block)

	block.Header.Hash = hash[:]
	block.Header.Difficulty = difficulty
	block.Header.Nonce = nonce

	return block
}

func genesis(t *Transaction) *Block {
	return CreateBlock([]*Transaction{t}, []byte{}, 0)
}

func CreateGenesis(w *wallet.Wallet) *Block {
	tr := CreateTransaction(w, []byte("This is Genesis Block"))
	return genesis(tr)
}

func (bh *BlockHeader) GetHash() []byte {
	data := bytes.Join(
		[][]byte{
			bh.Hash,
			bh.PrvHash,
			bh.MerkleRoot,
			toHex(bh.Timestamp),
			toHex(int64(bh.Difficulty)),
			toHex(int64(bh.Nonce)),
			toHex(int64(bh.Height)),
		},
		[]byte{},
	)

	hash := sha256.Sum256(data)
	return hash[:]
}

func (b *Block) MerkleRoot() []byte {
	var hashes [][]byte

	for _, tr := range b.Transactions {
		// hashes = append(hashes, serial.Serialize(tr))
		hashes = append(hashes, tr.Hash)
	}
	root := CalMerkleRootHash(hashes)

	return root
}

func (b *Block) PoW() {

}
