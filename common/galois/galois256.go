package galois

import (
	"log"
	"math/big"
)

type gf256 struct {
	Size uint
}

// http://poincare.matf.bg.ac.rs/~ezivkovm/publications/primpol1.pdf
// Irreducible Polynomial
// F(x) = x^256 + x^241 + x^178 + x^121 + 1
// var P256 = []uint{256, 241, 178, 121}
// F(x) = x^128 + x^77 + x^35 + x^11 + 1
var P256 = []uint{128, 77, 35, 11}

var GFPRI = big.NewInt(1)
var GFMASK = big.NewInt(1)
var GFINV = big.NewInt(1)

var GF_SIZE = uint(256)

func GF256() *gf256 {
	gf := gf256{GF_SIZE}
	return &gf
}

func init() {
	for _, x := range P256 {
		p := big.NewInt(1)
		GFPRI.Add(p.Lsh(p, x), GFPRI)
	}
	GF_SIZE = uint(GFPRI.BitLen() - 1)
	log.Printf("Prime %v", GF_SIZE)

	GFMASK.Lsh(GFMASK, GF_SIZE)
	GFMASK.Sub(GFMASK, big.NewInt(1))

	GFINV.Lsh(GFINV, GF_SIZE)
	GFINV.Sub(GFINV, big.NewInt(2))

}

func (gf *gf256) Add256(x, y *big.Int) *big.Int {
	s := big.NewInt(0)
	return s.Xor(x, y)
}

func (gf *gf256) Sub256(x, y *big.Int) *big.Int {
	s := big.NewInt(0)
	return s.Xor(x, y)
}

func (gf *gf256) lsh256(x *big.Int) *big.Int {
	t := new(big.Int)
	t = t.Set(x)
	if x.BitLen() < int(GF_SIZE) {
		t.Lsh(t, 1)
	} else {
		t.Lsh(t, 1)
		t.Xor(t, GFPRI)
	}

	return t
}

func (gf *gf256) Mul256(x, y *big.Int) *big.Int {
	r := big.NewInt(0)
	t := new(big.Int)
	t = t.Set(x)
	max := y.BitLen()

	for i := 0; i < max; i++ {
		if y.Bit(i) == 1 {
			r.Xor(r, t)
		}

		t = gf.lsh256(t)
	}

	return r
}

// x**p = x, so x**(p-2) = x**(-1)
func (gf *gf256) Div256(x, y *big.Int) *big.Int {
	if y.BitLen() == 0 {
		log.Printf("Div by zero")
		return big.NewInt(0)
	}

	inv := gf.Exp256(y, GFINV)
	return gf.Mul256(x, inv)
}

func (gf *gf256) Exp256(x, y *big.Int) *big.Int {
	p := new(big.Int)
	p = p.Set(x)
	q := new(big.Int)
	q = q.Mod(y, GFMASK)
	// q.Set(y)

	b := big.NewInt(1)
	max := q.BitLen()

	for i := 0; i < max; i++ {
		if q.Bit(i) != 0 {
			b = gf.Mul256(b, p)
		}
		p = gf.Mul256(p, p)
	}

	return b
}

// y/x = q*x + r
func (gf *gf256) divid(x, y *big.Int) (*big.Int, *big.Int) {
	d := new(big.Int)
	d = d.Set(x)
	q := new(big.Int)
	q = q.Set(y)
	r := new(big.Int)
	// log.Printf("x : %x", x)
	nq := big.NewInt(0)

	for {
		lq := big.NewInt(1)
		dif := q.BitLen() - d.BitLen()
		if dif < 0 {
			break
		}
		r.Lsh(d, uint(dif))
		r.Xor(r, q)
		q.Set(r)
		nq.Add(lq.Lsh(lq, uint(dif)), nq)
	}

	return nq, r
}

// Farmat's Little Theorem to calculate Inverse
func (gf *gf256) InvF256(x *big.Int) *big.Int {
	return gf.Exp256(x, GFINV)
}

// Extended Euclidian Algorithm to calculate Inverse
func (gf *gf256) Inv256(x *big.Int) *big.Int {
	if x.BitLen() == 0 {
		return big.NewInt(0)
	}

	b := new(big.Int)
	b = b.Set(x)
	a := new(big.Int)
	a = a.Set(GFPRI)
	t := big.NewInt(1)
	t1 := big.NewInt(0)
	t2 := big.NewInt(1)

	for b.BitLen() > 1 {
		q, r := gf.divid(b, a)
		t = gf.Sub256(t1, gf.Mul256(q, t2))
		a.Set(b)
		b.Set(r)
		t1.Set(t2)
		t2.Set(t)
	}

	// log.Printf("b : %x, x : %x, Inv : %x", b, x, t)
	return t
}
