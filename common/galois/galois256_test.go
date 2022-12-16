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
	log.Printf("%v XOR %v = %v", a, b, gf.Size)
}

func TestGF256ExpInv(t *testing.T) {
	gf := GF256()
	P := big.NewInt(1)
	P = P.Lsh(P, gf.Size)
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
	P = P.Lsh(P, gf.Size)
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

func TestGF256Inv2(t *testing.T) {
	gf := GF256()
	tm1 := int64(0)
	tm2 := int64(0)

	for i := 0; i < 10; i++ {
		// r := rand.Int63()
		a := big.NewInt(int64(i))

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

func TestGF256InvF2Inv(t *testing.T) {
	gf := GF256()
	tm1 := int64(0)
	tm2 := int64(0)

	// P1 = (3*P - 2)/2  <--> Inverse of x^2
	P1 := big.NewInt(1)
	P1.Lsh(P1, gf.Size)
	P1.Mul(P1, big.NewInt(3))
	P1.Sub(P1, big.NewInt(2))
	P1.Div(P1, big.NewInt(2))

	// X^2
	P2 := big.NewInt(2)

	log.Printf("P1 : %x, P2 : %x", P1, P2)

	for i := 0; i < 10000; i++ {
		r := rand.Int63()
		a := big.NewInt(r)

		start := time.Now().UnixNano()
		exp1 := gf.Exp256(a, P1)
		end := time.Now().UnixNano()
		tm1 += (end - start)

		start = time.Now().UnixNano()
		exp2 := gf.Exp256(exp1, P2)
		end = time.Now().UnixNano()
		tm2 += (end - start)

		assert.Equal(t, a, exp2)
	}

	log.Printf("enc : %v, dec : %v", tm1/1000000, tm2/1000000)

}
