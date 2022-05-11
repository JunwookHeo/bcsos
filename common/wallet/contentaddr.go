package wallet

import (
	"hash/fnv"
	"math/big"
)

const START_DIST = 52

func DistanceXor(h1 string, h2 string) *big.Int {
	n1, _ := new(big.Int).SetString(h1[len(h1)-START_DIST:], 16)
	n2, _ := new(big.Int).SetString(h2[len(h2)-START_DIST:], 16)
	return new(big.Int).Xor(n1, n2)
}

func DistanceXor2(h1 string, h2 string) uint64 {
	n1, _ := new(big.Int).SetString(h1, 16)
	n2, _ := new(big.Int).SetString(h2, 16)
	d := new(big.Int).Xor(n1, n2)
	h := fnv.New64a()
	h.Write([]byte(d.Bytes()))
	return h.Sum64()
}
