package galois

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGFNAdd(t *testing.T) {
	gf := GFN(32)
	a := uint64(100)
	b := uint64(2)

	sum := gf.AddN(a, b)
	assert.Equal(t, sum, (a ^ b))
	log.Printf("%v XOR %v = %v", a, b, sum)
}

func TestGFNExpInv(t *testing.T) {
	gf := GFN(32)
	P := uint64(1 << 32)
	P = 3*P - 2

	b := 2
	tenc := int64(0)
	tdec := int64(0)

	for i := 0; i < 1000; i++ {
		a := uint64(rand.Int31())

		start := time.Now().UnixNano()
		exp1 := gf.Exp(a, uint64(b))
		end := time.Now().UnixNano()
		tenc += (end - start)

		start = time.Now().UnixNano()
		exp2 := gf.Exp(exp1, P)
		end = time.Now().UnixNano()
		tdec += (end - start)

		assert.Equal(t, exp1, exp2)
	}

	log.Printf("enc : %v, dec : %v", tenc/1000000, tdec/1000000)

}

func TestGFNFarmatLittle(t *testing.T) {
	gf := GFN(32)
	P := uint64(1 << 32)
	tenc := int64(0)

	for i := 0; i < 1000; i++ {
		a := uint64(rand.Int31())

		start := time.Now().UnixNano()
		exp1 := gf.Exp(a, P)
		end := time.Now().UnixNano()
		tenc += (end - start)

		assert.Equal(t, exp1, a)
	}

	log.Printf("enc : %v", tenc/1000000)

}
