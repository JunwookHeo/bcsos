package galois

import (
	"log"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
)

var fields = []uint{8, 16, 32, 40, 48, 56, 64}

func TestGFCreate(t *testing.T) {
	for _, f := range fields {
		gf := GFN(f)
		assert.NotNilf(t, gf, "GF(%v) is nill", f)
		log.Printf("Align %v", gf.GetAlign())
	}
}

func TestGFAddSub(t *testing.T) {
	for _, f := range fields {
		gf := GFN(f)

		for i := 0; i < 100; i++ {
			n := rand.Uint64() & gf.size
			m := rand.Uint64() & gf.size
			c1 := gf.AddN(n, m)
			c2 := gf.SubN(c1, n)
			c3 := gf.SubN(c1, m)
			assert.Equal(t, c2, m)
			assert.Equal(t, c3, n)
			// log.Printf("%v + %v = %v", n, m, c1)
			// log.Printf("%v - %v = %v", c1, n, m)
		}
	}
}

func TestGFMulDiv(t *testing.T) {
	for _, f := range fields {
		gf := GFN(f)

		for i := 0; i < 100; i++ {
			n := rand.Uint64() & gf.size
			m := rand.Uint64() & gf.size
			c1 := gf.Mul(n, m)
			c2 := gf.Div(c1, n)
			c3 := gf.Div(c1, m)
			assert.Equal(t, c2, m)
			assert.Equal(t, c3, n)
			log.Printf("%v * %v = %v", n, m, c1)
			log.Printf("%v / %v = %v", c1, n, m)
		}
	}
}

func TestGFExp(t *testing.T) {
	for _, f := range fields {
		gf := GFN(f)

		for i := 0; i < 100; i++ {
			n := rand.Uint64() & gf.size
			m := (rand.Uint64() & gf.size) % 1000
			c1 := gf.Exp(n, m)
			c2 := uint64(1)
			for j := uint64(0); j < m; j++ {
				c2 = gf.Mul(c2, n)
			}
			c3 := c1
			for j := uint64(0); j < m; j++ {
				c3 = gf.Div(c3, n)
			}
			assert.Equal(t, c1, c2)
			assert.Equal(t, uint64(1), c3)
		}
	}
}

func TestGFSqrt(t *testing.T) {
	for _, f := range fields {
		gf := GFN(f)

		for i := 0; i < 1000; i++ {
			n := rand.Uint64() & gf.size
			c1 := gf.Exp(n, 2)
			c2 := gf.SqrR(c1)
			assert.Equal(t, n, c2)

			// log.Printf("%v * 2 = %v", n, c1)
			// log.Printf("Square Root(%v) = %v", c1, c2)
		}
	}
}
