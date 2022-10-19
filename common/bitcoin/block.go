package bitcoin

import (
	"crypto/sha256"
	"encoding/hex"
)

type BlockHeader struct {
	Version    uint32
	PreHash    []byte
	MerkelRoot []byte
	Timestamp  uint32
	Difficulty uint32
	Nonce      uint32
}

type Block struct {
	Hash   []byte
	Header BlockHeader
}

func NewBlock() *Block {
	return &Block{}
}

func (b *Block) SetHash(buf []byte) {
	hash := sha256.Sum256(buf)
	hash = sha256.Sum256(hash[:])
	b.Hash = b.ReverseBuf(hash[:])
}

func (b *Block) GetHashString() string {
	return hex.EncodeToString(b.Hash)
}

func (b *Block) ReverseBuf(buf []byte) []byte {
	n := len(buf)
	des := make([]byte, n)
	copy(des, buf)
	for i := 0; i < n/2; i++ {
		des[i], des[n-1-i] = des[n-1-i], des[i]
	}
	return des
}
