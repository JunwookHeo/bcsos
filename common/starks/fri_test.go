package starks

import (
	"log"
	"math/rand"
	"testing"
	"time"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestFriMerklize(t *testing.T) {
	f := NewFri()

	xs := make([]*uint256.Int, 65536)
	for i := 0; i < len(xs); i++ {
		r := rand.Int63()
		xs[i] = uint256.NewInt(uint64(r))
	}

	ys := f.Merklize(xs)
	assert.Equal(t, len(xs)*2, len(ys))
}

func TestFriDivid(t *testing.T) {
	for i := 0; i < 100; i++ {
		log.Printf("%v, %v", i, i>>2)
	}
}

func TestFriGetPseudorandomIndices(t *testing.T) {
	f := NewFri()
	r := make([]byte, 32)
	for i := 0; i < len(r); i++ {
		r[i] = byte(rand.Int() % 256)
	}

	count := 40
	dices := f.GetPseudorandomIndices(r, 32, count)
	assert.Equal(t, len(dices), count)

	log.Printf("%v", dices)
}

func TestFriProveLowDegree(t *testing.T) {
	f := NewFri()
	length := 16384
	ys := make([]*uint256.Int, length)
	for i := 0; i < len(ys); i++ {
		r := rand.Int63()
		ys[i] = uint256.NewInt(uint64(r))
	}

	g := f.GFP.Prime.Clone()
	g.Sub(g, uint256.NewInt(1))
	g.Div(g, uint256.NewInt(uint64(length)))
	g1 := f.GFP.Exp(uint256.NewInt(7), g)

	tm1 := int64(0)
	start := time.Now().UnixNano()
	proof := f.ProveLowDegree(ys, g1)
	end := time.Now().UnixNano()
	tm1 = end - start
	log.Printf("size of Proof : %v, %v", len(proof), tm1/1000000)

	m1 := f.Merklize(ys)
	eval := f.VerifyLowDegreeProof(m1[1], proof, g1)
	assert.Equal(t, eval, true)
}
