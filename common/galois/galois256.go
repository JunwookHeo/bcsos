package galois

import (
	"log"
	"math/big"
)

type gf256 struct {
}

// Irreducible Polynomial
// F(x) = x^256 + x^241 + x^178 + x^121 + 1
var P256 = []uint{241, 178, 121}
var GFPRI = big.NewInt(1)
var GFMASK = big.NewInt(1)
var GFINV = big.NewInt(1)

const GF_SIZE = 256

func GF256() *gf256 {
	poly := gf256{}
	return &poly
}

func init() {
	GFMASK = GFMASK.Lsh(GFMASK, GF_SIZE)
	GFMASK = GFMASK.Sub(GFMASK, big.NewInt(1))

	GFINV = GFINV.Lsh(GFINV, GF_SIZE)
	GFINV = GFINV.Sub(GFINV, big.NewInt(2))

	for _, x := range P256 {
		p := big.NewInt(1)
		GFPRI = p.Add(p.Lsh(p, x), GFPRI)
	}
	log.Printf("Prime %v", GFPRI.BitLen())
}

func (table *gf256) Add256(x, y *big.Int) *big.Int {
	s := big.NewInt(0)
	return s.Xor(x, y)
}

func (table *gf256) Sub256(x, y *big.Int) *big.Int {
	s := big.NewInt(0)
	return s.Xor(x, y)
}

func (table *gf256) lsh256(x *big.Int) *big.Int {
	t := new(big.Int)
	t = t.Set(x)
	if x.BitLen() < GF_SIZE {
		t = t.Lsh(t, 1)
	} else {
		t = t.Lsh(t, 1)
		t = t.And(t, GFMASK)
		t = t.Xor(t, GFPRI)
	}

	return t
}

func (table *gf256) Mul256(x, y *big.Int) *big.Int {
	r := big.NewInt(0)
	t := new(big.Int)
	t = t.Set(x)
	max := y.BitLen()

	for i := 0; i < max; i++ {
		if y.Bit(i) == 1 {
			r = r.Xor(r, t)
		}

		t = table.lsh256(t)
	}

	return r
}

// x**p = x, so x**(p-2) = x**(-1)
func (table *gf256) Div256(x, y *big.Int) *big.Int {
	if y.BitLen() == 0 {
		log.Printf("Div by zero")
		return big.NewInt(0)
	}

	inv := table.Exp256(y, GFINV)
	return table.Mul256(x, inv)
}

func (table *gf256) Exp256(x, y *big.Int) *big.Int {
	p := new(big.Int)
	p = p.Set(x)
	b := big.NewInt(1)
	max := y.BitLen()

	for i := 0; i < max; i++ {
		if y.Bit(i) != 0 {
			b = table.Mul256(b, p)
		}
		p = table.Mul256(p, p)
	}

	return b
}
