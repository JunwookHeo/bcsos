package galois

import (
	"log"
	"math/big"

	"github.com/holiman/uint256"
)

type gfp struct {
	Size  uint
	Prime *uint256.Int
}

// PRIME = 2^256 - 351*2^32 + 1

func GFP() *gfp {
	p := big.NewInt(1)
	p1 := big.NewInt(1)
	p1.Lsh(p1, 32)
	p1.Mul(p1, big.NewInt(351))
	p.Lsh(p, 256)
	p.Sub(p, p1)
	p.Add(p, big.NewInt(1))
	size := p.BitLen()
	// p.Set(big.NewInt(11))

	log.Printf("Prime : %x", p)
	prime, _ := uint256.FromBig(p)

	gf := gfp{uint(size), prime}
	return &gf
}

func (gf *gfp) Add(x, y *uint256.Int) *uint256.Int {
	s := new(uint256.Int)
	return s.AddMod(x, y, gf.Prime)
}

func (gf *gfp) Sub(x, y *uint256.Int) *uint256.Int {
	s := new(uint256.Int)
	if x.Lt(y) {
		s.Sub(y, x)
		return s.Sub(gf.Prime, s)
	} else {
		return s.Sub(x, y)
	}
}

func (gf *gfp) Mul(x, y *uint256.Int) *uint256.Int {
	s := new(uint256.Int)
	return s.MulMod(x, y, gf.Prime)
}

func (gf *gfp) Div(x, y *uint256.Int) *uint256.Int {
	return gf.Mul(x, gf.Inv(y))

}

func (gf *gfp) Exp(x, y *uint256.Int) *uint256.Int {
	a := x.ToBig()
	b := y.ToBig()
	m := gf.Prime.ToBig()
	a.Exp(a, b, m)
	s, _ := uint256.FromBig(a)
	return s
}

// Extended Euclidian Algorithm to calculate Inverse
func (gf *gfp) Inv(x *uint256.Int) *uint256.Int {
	if x.BitLen() == 0 {
		return uint256.NewInt(0)
	}
	lm := uint256.NewInt(1)
	hm := uint256.NewInt(0)
	low := x.Clone()
	high := gf.Prime.Clone()
	r := high.Clone()

	for low.Gt(uint256.NewInt(1)) {
		r.Div(high, low)
		nm := gf.Sub(hm, gf.Mul(lm, r))
		nw := gf.Sub(high, gf.Mul(low, r))
		lm, low, hm, high = nm, nw, lm, low
	}

	return lm.Mod(lm, gf.Prime)
}

func (gf *gfp) InvF(x *uint256.Int) *uint256.Int {
	if x.BitLen() == 0 {
		return uint256.NewInt(0)
	}
	inv := gf.Prime.Clone()
	inv.Sub(inv, uint256.NewInt(2))
	return gf.Exp(x, inv)
}

// Evaluate Polynomial at a point x
func (gf *gfp) EvalPolyAt(cs []*uint256.Int, x *uint256.Int) *uint256.Int {
	y := uint256.NewInt(0)
	pox := uint256.NewInt(1)

	for _, c := range cs {
		y = gf.Add(gf.Mul(c, pox), y)
		pox = gf.Mul(pox, x)
	}

	return y
}

func (gf *gfp) ZPoly(xs []*uint256.Int) []*uint256.Int {
	cs := make([]*uint256.Int, 1, len(xs)+1)
	// cs = append(cs, uint256.NewInt(1))
	cs[0] = uint256.NewInt(1)
	for j, x := range xs {
		cs = append([]*uint256.Int{uint256.NewInt(0)}, cs...)
		for i := 0; i < j+1; i++ {
			t := gf.Mul(cs[i+1], x)
			cs[i] = gf.Sub(cs[i], t)
		}
	}

	return cs
}

// def div_polys(self, a, b):
// 	a = [x for x in a]
// 	o = []
// 	apos = len(a) - 1
// 	bpos = len(b) - 1
// 	diff = apos - bpos
// 	while diff >= 0:
// 		quot = self.div(a[apos], b[bpos])
// 		o.insert(0, quot)
// 		for i in range(bpos, -1, -1):
// 			a[diff+i] -= b[i] * quot
// 		apos -= 1
// 		diff -= 1
// 	return [x % self.modulus for x in o]

// div polys
// D(x) = (x-1)(x-2)(x-3)....(x-n)/(x-k)
func (gf *gfp) DivPolys(a, b []*uint256.Int) []*uint256.Int {
	if len(a) < len(b) {
		return nil
	}

	var out []*uint256.Int
	cs := make([]*uint256.Int, len(a))
	for i := 0; i < len(a); i++ {
		cs[i] = a[i]
	}

	apos := len(cs) - 1
	bpos := len(b) - 1
	diff := apos - bpos

	for diff >= 0 {
		qout := gf.Div(cs[apos], b[bpos])
		out = append([]*uint256.Int{qout}, out...)
		for i := bpos; i >= 0; i-- {
			cs[diff+i] = gf.Sub(cs[diff+i], gf.Mul(b[i], qout))
		}
		apos -= 1
		diff -= 1
	}
	return out
}

// func (gf *gfp) LagrangeInterp(xs, ys []*uint256.Int) []*uint256.Int {
// 	zp := gf.ZPoly(xs)
// 	if len(zp) != len(ys)+1 {
// 		return nil
// 	}

// 	// var lp []*uint256.Int
// 	lp := make([]*uint256.Int, len(ys))
// 	for i := 0; i < len(ys); i++ {
// 		lp[i] = uint256.NewInt(0)
// 	}

// 	// var dps [][]*uint256.Int
// 	for i, x := range xs {
// 		ps := make([]*uint256.Int, 2)
// 		// var ps []*uint256.Int
// 		// ps = append(ps, x)
// 		// ps = append(ps, uint256.NewInt(1))
// 		ps[0], ps[1] = x, uint256.NewInt(1)

// 		// Get divid polynomial
// 		// dp = (x-x1)(x-x2).....(x-xn) / (x-xk)
// 		dp := gf.DivPolys(zp, ps)
// 		// dps = append([]*uint256.Int{dp}, dps...)
// 		// Evaluate each divided polynomial
// 		// denom = (xk-x1)(xk-x2)....(xk-xn)  without (xk-xk)
// 		denom := gf.EvalPolyAt(dp, x)
// 		// invdenom = 1/denom
// 		invdenom := gf.Inv256(denom)
// 		// yk = yk * 1/denom
// 		yk := gf.Mul256(ys[i], invdenom)
// 		// Add all coeficient of each x^n
// 		for j := range ys {
// 			lp[j] = gf.Add256(lp[j], gf.Mul256(dp[j], yk))
// 		}
// 	}

// 	return lp
// }

// func (gf *gfp) ExtRootUnity2(x *uint256.Int, inv bool) (int, []*uint256.Int) {
// 	maxc := 65536
// 	if gf.Size < 16 {
// 		maxc = int(gf.GMask.Uint64()) + 1
// 	}
// 	var cmpos int

// 	roots := make([]*uint256.Int, 2, maxc)
// 	if inv {
// 		roots[0] = x
// 		roots[1] = uint256.NewInt(1)
// 		cmpos = 0
// 	} else {
// 		roots[0] = uint256.NewInt(1)
// 		roots[1] = x
// 		cmpos = len(roots) - 1
// 	}

// 	one := uint256.NewInt(1)
// 	i := 2
// 	for ; one.Cmp(roots[cmpos]) != 0; i++ {
// 		if i < maxc {
// 			if inv {
// 				roots = append([]*uint256.Int{gf.Mul256(x, roots[cmpos])}, roots...)
// 			} else {
// 				roots = append(roots, gf.Mul256(x, roots[cmpos]))
// 			}
// 		} else {
// 			return -1, roots
// 		}
// 		if !inv {
// 			cmpos = len(roots) - 1
// 		}
// 	}
// 	// return i - 1, roots[:len(roots)-1]
// 	return i, roots
// }

// func (gf *gfp) ExtRootUnity(root *uint256.Int, inv bool) (int, []*uint256.Int) {
// 	maxc := 65536
// 	if gf.Size < 16 {
// 		maxc = int(gf.GMask.Uint64()) + 1
// 	}
// 	x := new(uint256.Int)
// 	if inv {
// 		x.Set(gf.Inv256(root))
// 	} else {
// 		x.Set(root)
// 	}

// 	roots := make([]*uint256.Int, maxc)
// 	roots[0] = uint256.NewInt(1)
// 	roots[1] = x

// 	one := uint256.NewInt(1)
// 	i := 2
// 	// for ; one.Cmp(roots[len(roots)-1]) != 0; i++ {
// 	for ; one.Cmp(roots[i-1]) != 0; i++ {
// 		if i < maxc {
// 			// roots = append(roots, gf.Mul256(x, roots[len(roots)-1]))
// 			roots[i] = gf.Mul256(x, roots[i-1])
// 		} else {
// 			return -1, roots
// 		}
// 	}
// 	return i - 1, roots[:i-1]
// 	// return i, roots[:i]
// }

// // FFT algorithm with root of unity
// // xs should be root of unit : x^n = 1
// func (gf *gfp) fft(xs, ys []*uint256.Int) []*uint256.Int {
// 	l := len(xs)
// 	os := make([]*uint256.Int, l) // outputs

// 	for i := 0; i < l; i++ {
// 		sum := uint256.NewInt(0)
// 		for j := 0; j < len(ys); j++ {
// 			m := gf.Mul256(ys[j], xs[(i*j)%l])
// 			sum = gf.Add256(sum, m)
// 		}
// 		os[i] = sum
// 	}
// 	return os
// }

// // DFT evaluates a polynomial at xs(root of unity)
// // cs is coefficients of a polynominal : [c0, c1, c2, c3 ... cn-1]
// // xs is root of unity, so x^n = 1 : [x^0, x^1, x^2, .... x^n-1]
// func (gf *gfp) DFT(cs []*uint256.Int, root *uint256.Int) []*uint256.Int {
// 	size, xs := gf.ExtRootUnity(root, false)
// 	if size == -1 {
// 		log.Printf("Wrong root of unity !!!")
// 		return nil
// 	}
// 	return gf.fft(xs, cs)
// }

// // IDFT generates a polynomial with points [(x0, y0), (x1, y1)....(xn-1, yn-1)]
// // xs is root of unity, so x^n = 1 : [x^0, x^1, x^2, .... x^n-1]
// // ys : [y0, y1, y2, y3 ... yn-1]
// // Output is coefficients of a polynominal : [c0, c1, c2, c3 ... cn-1]
// func (gf *gfp``) IDFT(ys []*uint256.Int, root *uint256.Int) []*uint256.Int {
// 	size, xs := gf.ExtRootUnity(root, true)
// 	if size != len(ys) {
// 		log.Println("The length of xs and ys should be the same")
// 		return nil
// 	}

// 	return gf.fft(xs, ys)
// }
