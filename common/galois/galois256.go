package galois

import (
	"log"
	"math/big"

	"github.com/holiman/uint256"
)

type gf256 struct {
	Size    uint
	PrimeW  *big.Int     // with the highest bit
	PrimeWo *uint256.Int // without the highest bit
	GMask   *uint256.Int
	GInv    *uint256.Int
}

// http://poincare.matf.bg.ac.rs/~ezivkovm/publications/primpol1.pdf
// Irreducible Polynomial
// F(x) = x^256 + x^241 + x^178 + x^121 + 1
// var P256 = []uint{256, 241, 178, 121}
// F(x) = x^128 + x^77 + x^35 + x^11 + 1
// var P256 = []uint{128, 77, 35, 11}

var P256 = map[int][]uint{
	4:   {4, 1},
	8:   {8, 7, 2, 1},
	16:  {16, 15, 12, 10},
	32:  {32, 16, 7, 2},
	64:  {64, 61, 34, 9},
	128: {128, 77, 35, 11},
	256: {256, 241, 178, 121},
}

func GF256(id int) *gf256 {
	field := P256[id]
	if field == nil {
		log.Panicf("Not support the prime field : %v", id)
		return nil
	}

	pw := big.NewInt(1)
	for _, x := range field {
		p := big.NewInt(1)
		pw.Add(p.Lsh(p, x), pw)
	}
	size := uint(pw.BitLen() - 1)
	log.Printf("Prime %v", size)

	mask := uint256.NewInt(1)
	mask.Lsh(mask, size)
	mask.Sub(mask, uint256.NewInt(1))

	inv := uint256.NewInt(1)
	inv.Lsh(inv, size)
	inv.Sub(inv, uint256.NewInt(2))

	pwo, err := uint256.FromBig(pw)
	pwo.And(pwo, mask)
	log.Printf("%b, %b, %v", pw, pwo, err)

	gf := gf256{size, pw, pwo, mask, inv}
	return &gf
}

func (gf *gf256) Add256(x, y *uint256.Int) *uint256.Int {
	s := new(uint256.Int)
	return s.Xor(x, y)
}

func (gf *gf256) Sub256(x, y *uint256.Int) *uint256.Int {
	s := new(uint256.Int)
	return s.Xor(x, y)
}

func (gf *gf256) lsh256(x *uint256.Int) *uint256.Int {
	t := x.Clone()
	if x.BitLen() < int(gf.Size) {
		t.Lsh(t, 1)
	} else {
		t.Lsh(t, 1)
		t.And(t, gf.GMask)
		t.Xor(t, gf.PrimeWo)
	}

	return t
}

func (gf *gf256) Mul256(x, y *uint256.Int) *uint256.Int {
	r := uint256.NewInt(0)
	t := x.Clone()
	m := y.Clone()
	max := y.BitLen()

	for i := 0; i < max; i++ {
		um := m.Uint64()
		if um&0x1 == 1 {
			r.Xor(r, t)
		}

		t = gf.lsh256(t)
		m.Rsh(m, uint(1))
	}

	return r
}

// x**p = x, so x**(p-2) = x**(-1)
func (gf *gf256) Div256(x, y *uint256.Int) *uint256.Int {
	if y.BitLen() == 0 {
		log.Printf("Div by zero")
		return uint256.NewInt(0)
	}

	inv := gf.Exp256(y, gf.GInv)
	return gf.Mul256(x, inv)
}

func (gf *gf256) Exp256(x, y *uint256.Int) *uint256.Int {
	p := x.Clone()
	q := new(uint256.Int)
	q = q.Mod(y, gf.GMask)
	b := uint256.NewInt(1)
	max := q.BitLen()

	for i := 0; i < max; i++ {
		um := q.Uint64()
		if um&0x01 != 0 {
			b = gf.Mul256(b, p)
		}
		p = gf.Mul256(p, p)
		q.Rsh(q, 1)
	}

	return b
}

// y/x = q*x + r
// carry is true for 256bit irreducable polynomials
func (gf *gf256) divid(x, y *uint256.Int, carry bool) (*uint256.Int, *uint256.Int) {
	d := new(uint256.Int)
	d = d.Set(x)
	q := new(uint256.Int)
	q = q.Set(y)
	r := new(uint256.Int)
	// log.Printf("x : %x", x)
	nq := uint256.NewInt(0)

	for {
		lq := uint256.NewInt(1)
		dif := q.BitLen() - d.BitLen()
		if carry {
			dif = 257 - d.BitLen()
			carry = false
		}

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
func (gf *gf256) InvF256(x *uint256.Int) *uint256.Int {
	return gf.Exp256(x, gf.GInv)
}

// Extended Euclidian Algorithm to calculate Inverse
func (gf *gf256) Inv256(x *uint256.Int) *uint256.Int {
	if x.BitLen() == 0 {
		return uint256.NewInt(0)
	}

	b := new(uint256.Int)
	b = b.Set(x)
	a := new(uint256.Int)
	// a = a.Set(gf.PrimeWo) // TODO : deal with prime
	a.SetFromBig(gf.PrimeW)
	ca := false
	if gf.PrimeW.BitLen() == 257 {
		ca = true
	}

	t := uint256.NewInt(1)
	t1 := uint256.NewInt(0)
	t2 := uint256.NewInt(1)

	for b.BitLen() > 1 {
		q, r := gf.divid(b, a, ca) // TODO :: fist thereis no the highest field.
		t = gf.Sub256(t1, gf.Mul256(q, t2))
		a.Set(b)
		b.Set(r)
		t1.Set(t2)
		t2.Set(t)
		ca = false
	}

	// log.Printf("b : %x, x : %x, Inv : %x", b, x, t)
	return t
}

// Evaluate Polynomial at a point x
func (gf *gf256) EvalPolyAt(poly []*uint256.Int, x *uint256.Int) *uint256.Int {
	y := uint256.NewInt(0)
	pox := uint256.NewInt(1)

	for _, c := range poly {
		y = gf.Add256(gf.Mul256(c, pox), y)
		pox = gf.Mul256(pox, x)
	}

	return y
}

// Z(x) = (x -a1)(x-a2)(x-a3)....(x-an)
func (gf *gf256) ZPoly(xs []*uint256.Int) []*uint256.Int {
	var poly []*uint256.Int
	poly = append(poly, uint256.NewInt(1))

	for _, x := range xs {
		poly = append([]*uint256.Int{uint256.NewInt(0)}, poly...)
		for i := 0; i < len(poly)-1; i++ {
			t := gf.Mul256(poly[i+1], x)
			poly[i] = gf.Sub256(poly[i], t)
		}
	}

	return poly
}

// div polys
// D(x) = (x-1)(x-2)(x-3)....(x-n)/(x-k)
func (gf *gf256) DivPolys(a, b []*uint256.Int) []*uint256.Int {
	if len(a) < len(b) {
		return nil
	}

	var out []*uint256.Int
	var ad []*uint256.Int
	ad = append(ad, a...)

	apos := len(ad) - 1
	bpos := len(b) - 1
	diff := apos - bpos

	for diff >= 0 {
		qout := gf.Div256(ad[apos], b[bpos])
		out = append([]*uint256.Int{qout}, out...)
		for i := 1; i >= 0; i-- {
			ad[diff+i] = gf.Sub256(ad[diff+i], gf.Mul256(b[i], qout))
		}
		apos -= 1
		diff -= 1
	}
	return out
}

func (gf *gf256) LagrangeInterp(xs, ys []*uint256.Int) []*uint256.Int {
	zp := gf.ZPoly(xs)
	if len(zp) != len(ys)+1 {
		return nil
	}

	// var lp []*uint256.Int
	lp := make([]*uint256.Int, 0, len(ys))
	for i := 0; i < len(ys); i++ {
		lp = append(lp, uint256.NewInt(0))
	}

	// var dps [][]*uint256.Int
	for i, x := range xs {
		var ps []*uint256.Int
		ps = append(ps, x)
		ps = append(ps, uint256.NewInt(1))

		// Get divid polynomial
		// dp = (x-x1)(x-x2).....(x-xn) / (x-xk)
		dp := gf.DivPolys(zp, ps)
		// dps = append([]*uint256.Int{dp}, dps...)
		// Evaluate each divided polynomial
		// denom = (xk-x1)(xk-x2)....(xk-xn)  without (xk-xk)
		denom := gf.EvalPolyAt(dp, x)
		// invdenom = 1/denom
		invdenom := gf.Inv256(denom)
		// yk = yk * 1/denom
		yk := gf.Mul256(ys[i], invdenom)
		// Add all coeficient of each x^n
		for j := range ys {
			lp[j] = gf.Add256(lp[j], gf.Mul256(dp[j], yk))
		}
	}

	return lp
}

func (gf *gf256) ExtRootUnity2(x *uint256.Int, inv bool) (int, []*uint256.Int) {
	maxc := 65536
	if gf.Size < 16 {
		maxc = int(gf.GMask.Uint64()) + 1
	}
	var cmpos int

	roots := make([]*uint256.Int, 2, maxc)
	if inv {
		roots[0] = x
		roots[1] = uint256.NewInt(1)
		cmpos = 0
	} else {
		roots[0] = uint256.NewInt(1)
		roots[1] = x
		cmpos = len(roots) - 1
	}

	one := uint256.NewInt(1)
	i := 2
	for ; one.Cmp(roots[cmpos]) != 0; i++ {
		if i < maxc {
			if inv {
				roots = append([]*uint256.Int{gf.Mul256(x, roots[cmpos])}, roots...)
			} else {
				roots = append(roots, gf.Mul256(x, roots[cmpos]))
			}
		} else {
			return -1, roots
		}
		if !inv {
			cmpos = len(roots) - 1
		}
	}
	// return i - 1, roots[:len(roots)-1]
	return i, roots
}

func (gf *gf256) ExtRootUnity(root *uint256.Int, inv bool) (int, []*uint256.Int) {
	maxc := 65536
	if gf.Size < 16 {
		maxc = int(gf.GMask.Uint64()) + 1
	}
	x := new(uint256.Int)
	if inv {
		x.Set(gf.Inv256(root))
	} else {
		x.Set(root)
	}

	roots := make([]*uint256.Int, 2, maxc)
	roots[0] = uint256.NewInt(1)
	roots[1] = x

	one := uint256.NewInt(1)
	i := 2
	for ; one.Cmp(roots[len(roots)-1]) != 0; i++ {
		if i < maxc {
			roots = append(roots, gf.Mul256(x, roots[len(roots)-1]))
		} else {
			return -1, roots
		}
	}
	return i - 1, roots[:len(roots)-1]
	// return i-1, roots
}

// FFT algorithm with root of unity
// xs should be root of unit : x^n = 1
func (gf *gf256) fft(xs, ys []*uint256.Int, inv bool) []*uint256.Int {
	l := len(xs)
	os := make([]*uint256.Int, 0, l) // outputs

	for i := 0; i < l; i++ {
		sum := uint256.NewInt(0)
		for j := 0; j < len(ys); j++ {
			m := gf.Mul256(ys[j], xs[(i*j)%l])
			sum = gf.Add256(sum, m)
		}
		os = append(os, sum)
	}
	return os
}

// DFT evaluates a polynomial at xs(root of unity)
// cs is coefficients of a polynominal : [c0, c1, c2, c3 ... cn-1]
// xs is root of unity, so x^n = 1 : [x^0, x^1, x^2, .... x^n-1]
func (gf *gf256) DFT(cs []*uint256.Int, root *uint256.Int) []*uint256.Int {
	size, xs := gf.ExtRootUnity(root, false)
	if size == -1 {
		log.Printf("Wrong root of unity !!!")
		return nil
	}
	return gf.fft(xs, cs, false)
}

// IDFT generates a polynomial with points [(x0, y0), (x1, y1)....(xn-1, yn-1)]
// xs is root of unity, so x^n = 1 : [x^0, x^1, x^2, .... x^n-1]
// ys : [y0, y1, y2, y3 ... yn-1]
// Output is coefficients of a polynominal : [c0, c1, c2, c3 ... cn-1]
func (gf *gf256) IDFT(ys []*uint256.Int, root *uint256.Int) []*uint256.Int {
	size, xs := gf.ExtRootUnity(root, true)
	if size != len(ys) {
		log.Println("The length of xs and ys should be the same")
		return nil
	}

	return gf.fft(xs, ys, true)
}
