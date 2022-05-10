package wallet

import "math/big"

const START_DIST = 52

func DistanceXor(h1 string, h2 string) *big.Int {
	n1, _ := new(big.Int).SetString(h1[len(h1)-START_DIST:], 16)
	n2, _ := new(big.Int).SetString(h2[len(h2)-START_DIST:], 16)
	return new(big.Int).Xor(n1, n2)
}
