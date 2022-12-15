package galois

import (
	"log"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGF256Add(t *testing.T) {
	gf := GF256()
	a := big.NewInt(100)
	b := big.NewInt(2)

	sum := gf.Add256(a, b)
	assert.Equal(t, sum, a.Xor(a, b))
	log.Printf("%v XOR %v = %v", a, b, sum)
}

func TestGF256ExpInv(t *testing.T) {
	gf := GF256()
	P := big.NewInt(1)
	P = P.Lsh(P, 256)
	P = P.Mul(P, big.NewInt(3))
	P = P.Sub(P, big.NewInt(2))

	b := big.NewInt(2)
	tenc := int64(0)
	tdec := int64(0)

	for i := 0; i < 1000; i++ {
		r := rand.Int63()
		a := big.NewInt(r)

		start := time.Now().UnixNano()
		exp1 := gf.Exp256(a, b)
		end := time.Now().UnixNano()
		tenc += (end - start)

		start = time.Now().UnixNano()
		exp2 := gf.Exp256(exp1, P)
		end = time.Now().UnixNano()
		tdec += (end - start)

		assert.Equal(t, exp1, exp2)
	}

	log.Printf("enc : %v, dec : %v", tenc/1000000, tdec/1000000)

}

func TestGF256FarmatLittle(t *testing.T) {
	gf := GF256()
	P := big.NewInt(1)
	P = P.Lsh(P, 256)
	tenc := int64(0)

	for i := 0; i < 1000; i++ {
		r := rand.Int63()
		a := big.NewInt(r)

		start := time.Now().UnixNano()
		exp1 := gf.Exp256(a, P)
		end := time.Now().UnixNano()
		tenc += (end - start)

		assert.Equal(t, exp1, a)
	}

	log.Printf("enc : %v", tenc/1000000)

}

func TestGF256Inv(t *testing.T) {
	gf := GF256()
	tm1 := int64(0)
	tm2 := int64(0)

	for i := 0; i < 1000; i++ {
		r := rand.Int63()
		a := big.NewInt(r)

		start := time.Now().UnixNano()
		inv := gf.Inv256(a)
		end := time.Now().UnixNano()
		tm1 += (end - start)
		log.Printf("x*inv_1 = %x, %x, %x", a, inv, gf.Mul256(a, inv))

		start = time.Now().UnixNano()
		inv2 := gf.InvF256(a)
		end = time.Now().UnixNano()
		tm2 += (end - start)
		log.Printf("x*inv_2 = %x, %x, %x", a, inv2, gf.Mul256(a, inv2))
		// assert.Equal(t, exp1, a)
	}

	log.Printf("Time1 : %v", tm1/1000000)
	log.Printf("Time2 : %v", tm2/1000000)

}
