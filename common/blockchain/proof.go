package blockchain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"log"
	"math"
	"math/big"

	"github.com/junwookheo/bcsos/common/serial"
)

const DIFFICULTY = 12

func getTarget() *big.Int {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-DIFFICULTY))

	return target
}

func initData(b *Block, nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			b.Header.PrvHash,
			b.Header.MerkleRoot,
			toHex(b.Header.Timestamp),
			toHex(int64(DIFFICULTY)),
			toHex(int64(nonce)),
			toHex(int64(b.Header.Height)),
		},
		[]byte{},
	)
	var bs []byte
	for _, t := range b.Transactions {
		bs = append(bs, t.Hash...)
		bs = append(bs, serial.Serialize(t.Timestamp)...)
		bs = append(bs, t.Data...)
	}

	data = append(data, bs...)

	return data
}

func ProofWork(b *Block) (int, int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0
	target := getTarget()

	for nonce < math.MaxInt64 {
		data := initData(b, nonce)
		hash = sha256.Sum256(data)

		intHash.SetBytes(hash[:])
		if intHash.Cmp(target) == -1 {
			break
		} else {
			nonce++
		}
	}

	return DIFFICULTY, nonce, hash[:]
}

func Validate(b *Block) bool {
	var intHash big.Int
	target := getTarget()

	data := initData(b, b.Header.Nonce)
	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(target) == -1
}

func toHex(n int64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, n)
	if err != nil {
		log.Panic(err)
	}

	return buf.Bytes()
}
