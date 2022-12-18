package galois

import (
	"log"
	"math/big"
)

type gf256 struct {
	Size  uint
	Prime *big.Int
	GMask *big.Int
	GInv  *big.Int
}

// http://poincare.matf.bg.ac.rs/~ezivkovm/publications/primpol1.pdf
// Irreducible Polynomial
// F(x) = x^256 + x^241 + x^178 + x^121 + 1
// var P256 = []uint{256, 241, 178, 121}
// F(x) = x^128 + x^77 + x^35 + x^11 + 1
// var P256 = []uint{128, 77, 35, 11}

var P256 = map[int][]uint{
	4:    {4, 1},
	8:    {8, 7, 2, 1},
	16:   {16, 15, 12, 10},
	32:   {32, 16, 7, 2},
	64:   {64, 61, 34, 9},
	128:  {128, 77, 35, 11},
	256:  {256, 241, 178, 121},
	512:  {512, 419, 321, 125},
	1024: {1024, 333, 135, 73},
}

func GF256(id int) *gf256 {
	field := P256[id]
	if field == nil {
		log.Panicf("Not support the prime field : %v", id)
		return nil
	}

	prime := big.NewInt(1)
	for _, x := range field {
		p := big.NewInt(1)
		prime.Add(p.Lsh(p, x), prime)
	}
	size := uint(prime.BitLen() - 1)
	log.Printf("Prime %v", size)

	mask := big.NewInt(1)
	mask.Lsh(mask, size)
	mask.Sub(mask, big.NewInt(1))

	inv := big.NewInt(1)
	inv.Lsh(inv, size)
	inv.Sub(inv, big.NewInt(2))

	gf := gf256{size, prime, mask, inv}
	return &gf
}

func (gf *gf256) Add256(x, y *big.Int) *big.Int {
	// s := big.NewInt(0)
	s := new(big.Int)
	return s.Xor(x, y)
}

func (gf *gf256) Sub256(x, y *big.Int) *big.Int {
	// s := big.NewInt(0)
	s := new(big.Int)
	return s.Xor(x, y)
}

func (gf *gf256) lsh256(x *big.Int) *big.Int {
	t := new(big.Int)
	t = t.Set(x)
	if x.BitLen() < int(gf.Size) {
		t.Lsh(t, 1)
	} else {
		t.Lsh(t, 1)
		t.Xor(t, gf.Prime)
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

	inv := gf.Exp256(y, gf.GInv)
	return gf.Mul256(x, inv)
}

func (gf *gf256) Exp256(x, y *big.Int) *big.Int {
	p := new(big.Int)
	p = p.Set(x)
	q := new(big.Int)
	q = q.Mod(y, gf.GMask)

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
	return gf.Exp256(x, gf.GInv)
}

// Extended Euclidian Algorithm to calculate Inverse
func (gf *gf256) Inv256(x *big.Int) *big.Int {
	if x.BitLen() == 0 {
		return big.NewInt(0)
	}

	b := new(big.Int)
	b = b.Set(x)
	a := new(big.Int)
	a = a.Set(gf.Prime)
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

// Evaluate Polynomial at a point x
func (gf *gf256) EvalPolyAt(poly []*big.Int, x *big.Int) *big.Int {
	y := big.NewInt(0)
	pox := big.NewInt(1)

	for _, c := range poly {
		y = gf.Add256(gf.Mul256(c, pox), y)
		pox = gf.Mul256(pox, x)
	}

	return y
}

// Z(x) = (x -a1)(x-a2)(x-a3)....(x-an)
func (gf *gf256) ZPoly(xs []*big.Int) []*big.Int {
	var poly []*big.Int
	poly = append(poly, big.NewInt(1))

	for _, x := range xs {
		poly = append([]*big.Int{big.NewInt(0)}, poly...)
		for i := 0; i < len(poly)-1; i++ {
			t := gf.Mul256(poly[i+1], x)
			poly[i] = gf.Sub256(poly[i], t)
		}
	}

	return poly
}

// div polys
// D(x) = (x-1)(x-2)(x-3)....(x-n)/(x-k)
func (gf *gf256) DivPolys(a, b []*big.Int) []*big.Int {
	if len(a) < len(b) {
		return nil
	}

	var out []*big.Int
	var ad []*big.Int
	ad = append(ad, a...)

	apos := len(ad) - 1
	bpos := len(b) - 1
	diff := apos - bpos

	for diff >= 0 {
		qout := gf.Div256(ad[apos], b[bpos])
		out = append([]*big.Int{qout}, out...)
		for i := 1; i >= 0; i-- {
			ad[diff+i] = gf.Sub256(ad[diff+i], gf.Mul256(b[i], qout))
		}
		apos -= 1
		diff -= 1
	}
	return out
}

func (gf *gf256) LagrangeInterp(xs, ys []*big.Int) []*big.Int {
	zp := gf.ZPoly(xs)
	if len(zp) != len(ys)+1 {
		return nil
	}

	var lp []*big.Int
	for i := 0; i < len(ys); i++ {
		lp = append(lp, big.NewInt(0))
	}

	// var dps [][]*big.Int
	for i, x := range xs {
		var ps []*big.Int
		ps = append(ps, x)
		ps = append(ps, big.NewInt(1))

		// Get divid polynomial
		// dp = (x-x1)(x-x2).....(x-xn) / (x-xk)
		dp := gf.DivPolys(zp, ps)
		// dps = append([]*big.Int{dp}, dps...)
		// Evaluate each divided polynomial
		// denom = (xk-x1)(xk-x2)....(xk-xn)  without (xk-xk)
		denom := gf.EvalPolyAt(dp, x)
		// invdenom = 1/denom
		invdenom := gf.Inv256(denom)
		// yk = yk * 1/denom
		yk := gf.Mul256(ys[i], invdenom)
		// Add all coeficient of each x^n
		for j, _ := range ys {
			lp[j] = gf.Add256(lp[j], gf.Mul256(dp[j], yk))
		}
	}

	return lp
}

func (gf *gf256) ExtRootUnity(x *big.Int) (int, []*big.Int) {
	maxc := 65536
	if gf.Size < 16 {
		maxc = int(gf.GMask.Uint64()) + 1
	}

	roots := make([]*big.Int, 2, maxc)
	roots[0] = big.NewInt(1)
	roots[1] = x
	one := big.NewInt(1)
	i := 2
	for ; one.Cmp(roots[len(roots)-1]) != 0; i++ {
		if i < maxc {
			roots = append(roots, gf.Mul256(x, roots[len(roots)-1]))
		} else {
			return -1, roots
		}
	}
	return i - 1, roots[:len(roots)-1]
	// return i, roots
}
