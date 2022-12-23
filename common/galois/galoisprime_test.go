package galois

import (
	"log"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/holiman/uint256"
	"github.com/stretchr/testify/assert"
)

func TestGFPAddSub(t *testing.T) {
	gf := GFP()

	tenc := int64(0)
	tdec := int64(0)

	for i := 0; i < 100000; i++ {
		r := rand.Int63() % int64(gf.Prime.Uint64())
		a := uint256.NewInt(uint64(r))
		r = rand.Int63() % int64(gf.Prime.Uint64())
		b := uint256.NewInt(uint64(r))

		start := time.Now().UnixNano()
		exp1 := gf.Add(a, b)
		end := time.Now().UnixNano()
		tenc += (end - start)

		start = time.Now().UnixNano()
		exp2 := gf.Sub(exp1, b)
		end = time.Now().UnixNano()
		tdec += (end - start)

		start = time.Now().UnixNano()
		exp3 := gf.Sub(b, exp1)
		end = time.Now().UnixNano()
		tdec += (end - start)

		start = time.Now().UnixNano()
		exp4 := gf.Add(exp3, exp1)
		end = time.Now().UnixNano()
		tenc += (end - start)

		assert.Equal(t, a, exp2)
		assert.Equal(t, b, exp4)
		// log.Printf("%v, %v, %v", a, exp1, exp2)
	}
	log.Printf("enc : %v, dec : %v", tenc/1000000, tdec/1000000)
}

func TestGFPMulDiv(t *testing.T) {
	gf := GFP()

	tenc := int64(0)
	tdec := int64(0)

	for i := 0; i < 100000; i++ {
		r := rand.Int63() % int64(gf.Prime.Uint64())
		a := uint256.NewInt(uint64(r))
		r = rand.Int63() % int64(gf.Prime.Uint64())
		b := uint256.NewInt(uint64(r))

		start := time.Now().UnixNano()
		exp1 := gf.Mul(a, b)
		end := time.Now().UnixNano()
		tenc += (end - start)

		start = time.Now().UnixNano()
		exp2 := gf.Div(exp1, b)
		end = time.Now().UnixNano()
		tdec += (end - start)

		start = time.Now().UnixNano()
		exp3 := gf.Div(b, exp1)
		end = time.Now().UnixNano()
		tdec += (end - start)

		start = time.Now().UnixNano()
		exp4 := gf.Mul(exp3, exp1)
		end = time.Now().UnixNano()
		tenc += (end - start)

		assert.Equal(t, a, exp2)
		assert.Equal(t, b, exp4)
		// log.Printf("%v, %v, %v", a, exp1, exp2)
	}
	log.Printf("enc : %v, dec : %v", tenc/1000000, tdec/1000000)
}

func TestGFPExp(t *testing.T) {
	gf := GFP()

	tenc := int64(0)
	tdec := int64(0)

	for i := 0; i < 10000; i++ {
		r := rand.Int63() % int64(gf.Prime.Uint64())
		a := uint256.NewInt(uint64(r))
		// r = rand.Int63() % int64(gf.Prime.Uint64())
		b := uint256.NewInt(uint64(i))

		start := time.Now().UnixNano()
		exp1 := gf.Exp(a, b)
		end := time.Now().UnixNano()
		tenc += (end - start)

		start = time.Now().UnixNano()
		exp2 := uint256.NewInt(1)
		for j := 0; j < int(b.Uint64()); j++ {
			exp2 = gf.Mul(exp2, a)
		}

		end = time.Now().UnixNano()
		tdec += (end - start)

		assert.Equal(t, exp1, exp2)
		// log.Printf("%v, %v, %v", a, exp1, exp2)
	}
	log.Printf("enc : %v, dec : %v", tenc/1000000, tdec/1000000)
}

func TestGFPX3Inv(t *testing.T) {
	gf := GFP()
	// Inv x^3 = x^((2P-1)/3)
	P1 := gf.Prime.ToBig()
	P1.Mul(P1, big.NewInt(2))
	P1.Sub(P1, big.NewInt(1))
	P1.Div(P1, big.NewInt(3))
	P1.Mod(P1, gf.Prime.ToBig())
	P, _ := uint256.FromBig(P1)
	log.Printf("InvP  : %x", P)

	b := uint256.NewInt(3)
	tenc := int64(0)
	tdec := int64(0)

	for i := 0; i < 10000; i++ {
		r := rand.Int63() % int64(gf.Prime.Uint64())
		// r := 3
		a := uint256.NewInt(uint64(r))

		start := time.Now().UnixNano()
		exp1 := gf.Exp(a, b)
		end := time.Now().UnixNano()
		tenc += (end - start)

		start = time.Now().UnixNano()
		exp2 := gf.Exp(exp1, P)
		end = time.Now().UnixNano()
		tdec += (end - start)

		assert.Equal(t, a, exp2)
		// log.Printf("%v, %v, %v", a, exp1, exp2)
	}

	log.Printf("enc : %v, dec : %v", tenc/1000000, tdec/1000000)

}

func TestGFPFarmatLittle(t *testing.T) {
	gf := GFP()
	P := gf.Prime.Clone()
	P.Sub(P, uint256.NewInt(1))
	tenc := int64(0)

	for i := 0; i < 1000; i++ {
		r := rand.Uint64()
		a := uint256.NewInt(r)

		start := time.Now().UnixNano()
		exp1 := gf.Exp(a, P)
		end := time.Now().UnixNano()
		tenc += (end - start)

		assert.Equal(t, exp1, uint256.NewInt(1))
	}

	log.Printf("enc : %v", tenc/1000000)

}

func TestGFPInvInvF(t *testing.T) {
	gf := GFP()
	tm1 := int64(0)
	tm2 := int64(0)

	for i := 2; i < 100000; i++ {
		r := rand.Uint64() % gf.Prime.Uint64()
		a := uint256.NewInt(uint64(r))

		start := time.Now().UnixNano()
		inv := gf.Inv(a)
		end := time.Now().UnixNano()
		tm1 += (end - start)
		// log.Printf("x*inv_1 = %x, %x, %x", a, inv, gf.Mul(a, inv))

		start = time.Now().UnixNano()
		inv2 := gf.InvF(a)
		end = time.Now().UnixNano()
		tm2 += (end - start)
		// log.Printf("x*inv_2 = %x, %x, %x", a, inv2, gf.Mul(a, inv2))
		assert.Equal(t, inv, inv2)
	}

	log.Printf("Time1 : %v", tm1/1000000)
	log.Printf("Time2 : %v", tm2/1000000)

}

func TestGFPZPoly(t *testing.T) {
	gf := GFP()
	tm1 := int64(0)
	tm2 := int64(0)

	for i := 0; i < 1; i++ {
		var xs []*uint256.Int
		for j := 0; j < 3; j++ {
			r := rand.Uint64() % gf.Prime.Uint64()
			a := uint256.NewInt(r)
			a = uint256.NewInt(uint64(j + 1))
			xs = append(xs, a)
		}

		start := time.Now().UnixNano()
		zp := gf.ZPoly(xs)
		end := time.Now().UnixNano()
		tm1 += (end - start)

		log.Printf("zpoly : %v", zp)

		start = time.Now().UnixNano()
		for k := 0; k < len(xs); k++ {
			// xs[k].Add(xs[k], uint256.NewInt(1))
			ev := gf.EvalPolyAt(zp, xs[k])
			log.Printf("%v - ev : %v", xs[k], ev)
			assert.Equal(t, 0, ev.BitLen())
		}

		end = time.Now().UnixNano()
		tm2 += (end - start)

	}

	log.Printf("enc : %v, dec : %v", tm1/1000000, tm2/1000000)
}

func TestGFPDivPolys(t *testing.T) {
	gf := GFP()
	tm1 := int64(0)

	for i := 0; i < 100; i++ {
		var xs []*uint256.Int
		for j := 0; j < 100; j++ {
			r := rand.Uint64() % gf.Prime.Uint64()
			a := uint256.NewInt(r)
			// a = uint256.NewInt(uint64(j + 1))

			xs = append(xs, a)
		}

		start := time.Now().UnixNano()
		zp := gf.ZPoly(xs)
		// log.Printf("xs %v", xs)
		// log.Printf("zp %v", zp)
		for k := 0; k < len(xs); k++ {
			var sub []*uint256.Int
			for l := 0; l < len(xs); l++ {
				if l != k {
					sub = append(sub, xs[l])
				}
			}
			// log.Printf("sub : %v", sub)

			sp := gf.ZPoly(sub)
			ix := gf.Sub(gf.Prime, xs[k])
			dp := gf.DivPolys(zp, []*uint256.Int{ix, uint256.NewInt(1)})
			// log.Printf("sp %v", sp)
			// log.Printf("dp %v", dp)
			assert.Equal(t, sp, dp)
		}

		end := time.Now().UnixNano()
		tm1 += (end - start)

		// log.Printf("div polys : %v, %v", poly, zp)
	}

	log.Printf("enc : %v", tm1/1000000)
}

// func TestGFPLagrangeInterp(t *testing.T) {
// 	gf := GFP(256)
// 	tm1 := int64(0)

// 	for i := 0; i < 100; i++ {
// 		var xs []*uint256.Int
// 		var ys []*uint256.Int

// 		for j := 0; j < 50; j++ {
// 			r := rand.Uint64()
// 			a := uint256.NewInt(r)
// 			ys = append(ys, a)
// 			b := uint256.NewInt(uint64(j + 1))
// 			xs = append(xs, b)
// 		}

// 		start := time.Now().UnixNano()
// 		lp := gf.LagrangeInterp(xs, ys)
// 		log.Printf("lp : %v", lp)
// 		end := time.Now().UnixNano()
// 		tm1 += (end - start)

// 		for j := 0; j < len(xs); j++ {
// 			ev := gf.EvalPolyAt(lp, xs[j])
// 			assert.Equal(t, 0, ev.Cmp(ys[j]))
// 			log.Printf("evaulate poly %v : %v", ys[j], ev)
// 		}
// 	}

// 	log.Printf("lp : %v", tm1/1000000)
// }

// func TestGFPExtRootUnity(t *testing.T) {
// 	gf := GFP(16)
// 	tm1 := int64(0)

// 	log.Printf("P-1 : %v", gf.GMask)

// 	for i := uint64(2); i < gf.GMask.Uint64(); i++ {
// 		for j := i; j <= gf.GMask.Uint64(); j *= j {
// 			a := uint256.NewInt(uint64(j))
// 			start := time.Now().UnixNano()
// 			size, rus := gf.ExtRootUnity(a, false)
// 			end := time.Now().UnixNano()
// 			tm1 += (end - start)

// 			if len(rus) < 1300 {
// 				log.Printf("%v of rus[%v] - [%v]", a, size, rus)
// 			}
// 		}
// 	}

// 	log.Printf("lp : %v", tm1/1000000)
// }

// func TestGFPExtFindRootUnity(t *testing.T) {
// 	gf := GFP(16)

// 	g1 := gf.Exp256(uint256.NewInt(2), uint256.NewInt(257))
// 	ev := gf.Exp256(g1, uint256.NewInt(255))
// 	log.Printf("==>G(%v): %v", g1, ev)
// 	size, rus := gf.ExtRootUnity(g1, false)
// 	log.Printf("==>G(%v), %v: %v", g1, size, rus)

// 	g2 := gf.Exp256(uint256.NewInt(2), uint256.NewInt(1285))
// 	ev2 := gf.Exp256(g2, uint256.NewInt(51))
// 	log.Printf("==>G(%v): %v", g2, ev2)
// 	size2, rus2 := gf.ExtRootUnity(g2, false)
// 	log.Printf("==>G(%v), %v: %v", g2, size2, rus2)

// 	g3 := gf.Exp256(uint256.NewInt(2), uint256.NewInt(3855))
// 	ev3 := gf.Exp256(g3, uint256.NewInt(17))
// 	log.Printf("==>G(%v): %v", g3, ev3)
// 	size3, rus3 := gf.ExtRootUnity(g2, false)
// 	log.Printf("==>G(%v), %v: %v", g3, size3, rus3)
// }
// func TestGFPExtFindRootUnity2(t *testing.T) {
// 	gf := GFP(8)

// 	g1 := gf.Exp256(uint256.NewInt(2), uint256.NewInt(85))
// 	ev := gf.Exp256(g1, uint256.NewInt(48))
// 	log.Printf("==>G(%v): %v", g1, ev)
// 	size, rus := gf.ExtRootUnity(g1, false)
// 	log.Printf("==>G(%v), %v: %v", g1, size, rus)

// 	// g2 := gf.Exp256(uint256.NewInt(2), uint256.NewInt(85))
// 	// ev2 := gf.Exp256(g2, uint256.NewInt(12))
// 	// log.Printf("==>G(%v): %v", g2, ev2)
// 	// size2, rus2 := gf.ExtRootUnity(g2, false)
// 	// log.Printf("==>G(%v), %v: %v", g2, size2, rus2)
// }

// func TestGFPExtFindRootUnityAll(t *testing.T) {
// 	gf := GFP(6)

// 	for i := uint64(1); i <= gf.GMask.Uint64(); i++ {
// 		g1 := uint256.NewInt(i)
// 		// log.Printf("==>G: %v", g1)
// 		size, rus := gf.ExtRootUnity(g1, false)
// 		log.Printf("==>G(%v, %v) :(%v) %v", g1, gf.Inv256(g1), size, rus[:3])
// 	}

// 	// log.Println("=================================")
// 	// for i := int64(1); i <= gf.GMask.Int64(); i++ {
// 	// 	g1 := uint256.NewInt(i)
// 	// 	// log.Printf("==>G: %v", g1)
// 	// 	size, rus := gf.ExtRootUnity(g1, true)
// 	// 	log.Printf("==>G(%v, %v)-(%v) : %v", g1, gf.Inv256(g1), size, rus)

// 	// }
// }

// func TestGFPExtFindRootUnityAny(t *testing.T) {
// 	fd := 32
// 	gf := GFP(fd)

// 	for i := uint64(100); i <= 1000; i++ {
// 		any := ((1<<fd - 1) * i) / 3 % (1<<fd - 1)

// 		g1 := uint256.NewInt(any)
// 		// log.Printf("==>G: %v", g1)
// 		size, _ := gf.ExtRootUnity(g1, false)
// 		log.Printf("==>G(%v, %v) : %v", g1, gf.Inv256(g1), size)
// 	}

// 	// log.Println("=================================")
// 	// for i := int64(1); i <= gf.GMask.Int64(); i++ {
// 	// 	g1 := uint256.NewInt(i)
// 	// 	// log.Printf("==>G: %v", g1)
// 	// 	size, rus := gf.ExtRootUnity(g1, true)
// 	// 	log.Printf("==>G(%v, %v)-(%v) : %v", g1, gf.Inv256(g1), size, rus)

// 	// }
// }

// func TestGFPExtFindRootUnityAll2(t *testing.T) {
// 	gf := GFP(8)

// 	for i := uint64(1); i <= gf.GMask.Uint64(); i++ {
// 		// for i := uint64(1<<16 - 100); i <= uint64(1<<16+10); i++ {
// 		g1 := uint256.NewInt(i)
// 		// log.Printf("==>G: %v", g1)
// 		size, _ := gf.ExtRootUnity(g1, false)
// 		if 0 < size {
// 			log.Printf("==>G(%b, %v)-(%v)", g1, gf.Inv256(g1), size)
// 		}
// 	}
// }

// func TestGFPFFT(t *testing.T) {
// 	gf := GFP(4)

// 	g1 := uint256.NewInt(3)
// 	size, xs := gf.ExtRootUnity(g1, false)
// 	ys := make([]*uint256.Int, 0, size)

// 	for j := 0; j < size; j++ {
// 		r := rand.Uint64() % (1 << 4)
// 		a := uint256.NewInt(r)
// 		ys = append(ys, a)
// 	}

// 	log.Printf("==>xs:%v, ys:%v", xs, ys)
// 	start := time.Now().Nanosecond()
// 	os1 := gf.LagrangeInterp(xs, ys)
// 	end := time.Now().Nanosecond()
// 	log.Printf("LagrangeInterp(%v) : f(x)=%v", (end-start)/1000, os1)

// 	start = time.Now().Nanosecond()
// 	os2 := gf.IDFT(ys, g1)
// 	end = time.Now().Nanosecond()
// 	log.Printf("IDFT(%v) : f(x)=%v", (end-start)/1000, os2)

// 	os3 := gf.DFT(os1, g1)
// 	log.Printf("DFT : %v", os3)

// }

// func TestGFPFFTPerf(t *testing.T) {
// 	gf := GFP(8)

// 	g1 := uint256.NewInt(3)
// 	size, xs := gf.ExtRootUnity(g1, false)
// 	log.Printf("xs(%v) :  %v", size, xs[:10])

// 	if size == -1 {
// 		return
// 	}

// 	ys := make([]*uint256.Int, 0, size)
// 	for j := 0; j < size; j++ {
// 		r := rand.Uint64() % (1 << gf.Size)
// 		a := uint256.NewInt(r)
// 		ys = append(ys, a)
// 	}

// 	log.Printf("ys(%v) :  %v", size, ys[:10])

// 	start := time.Now().Nanosecond()
// 	os2 := gf.IDFT(ys, g1)
// 	end := time.Now().Nanosecond()
// 	log.Printf("IDFT(%v) : f(x)=%v", (end-start)/1000, os2[:10])

// 	os3 := gf.DFT(os2, g1)
// 	log.Printf("DFT :  %v", os3[:10])

// }
