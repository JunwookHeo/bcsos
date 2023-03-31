package wallet

import (
	"hash/fnv"
	"math/big"

	"github.com/junwookheo/bcsos/common/config"
)

func DistanceXor(h1 string, h2 string) uint64 {
	if config.RANDOM_DISTANCE_NODES {
		return DistanceXor2(h1, h2)
	} else {
		return DistanceXor1(h1, h2)
	}
}

// Calcuate distance between h1 and h2
func DistanceXor1(h1 string, h2 string) uint64 {
	n1, _ := new(big.Int).SetString(h1, 16)
	n2, _ := new(big.Int).SetString(h2, 16)

	x := new(big.Int).Xor(n1, n2)
	return x.Uint64()
}

// Calcuate random distance
func DistanceXor2(h1 string, h2 string) uint64 {
	n1, _ := new(big.Int).SetString(h1, 16)
	n2, _ := new(big.Int).SetString(h2, 16)
	d := new(big.Int).Xor(n1, n2)
	h := fnv.New64a()
	h.Write([]byte(d.Bytes()))
	return h.Sum64()
}
