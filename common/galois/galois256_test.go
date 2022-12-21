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

func TestGF256Add(t *testing.T) {
	gf := GF256(128)
	a := uint256.NewInt(4)
	b := uint256.NewInt(2)

	sum := gf.Add256(a, b)
	assert.Equal(t, sum, a.Xor(a, b))
	log.Printf("%v XOR %v = %v", a, b, sum)
}

func TestGF256ExpInv(t *testing.T) {
	gf := GF256(256)
	P := uint256.NewInt(1)
	P = P.Lsh(P, gf.Size-1) // p = (3P-2)/2

	b := uint256.NewInt(2)
	tenc := int64(0)
	tdec := int64(0)

	for i := 0; i < 1000; i++ {
		r := rand.Int63()
		a := uint256.NewInt(uint64(r))
		a.And(a, gf.GMask)

		start := time.Now().UnixNano()
		exp1 := gf.Exp256(a, b)
		end := time.Now().UnixNano()
		tenc += (end - start)

		start = time.Now().UnixNano()
		exp2 := gf.Exp256(exp1, P)
		end = time.Now().UnixNano()
		tdec += (end - start)

		assert.Equal(t, a, exp2)
		log.Printf("%v, %v, %v", a, exp1, exp2)
	}

	log.Printf("enc : %v, dec : %v", tenc/1000000, tdec/1000000)

}

func TestGF256FarmatLittle(t *testing.T) {
	gf := GF256(256)
	bp := big.NewInt(1)
	bp = bp.Lsh(bp, gf.Size)
	P, _ := uint256.FromBig(bp.Sub(bp, big.NewInt(1)))
	tenc := int64(0)

	for i := 0; i < 1000; i++ {
		r := rand.Uint64()
		a := uint256.NewInt(r)

		start := time.Now().UnixNano()
		exp1 := gf.Exp256(a, P)
		end := time.Now().UnixNano()
		tenc += (end - start)

		assert.Equal(t, exp1, uint256.NewInt(1))
	}

	log.Printf("enc : %v", tenc/1000000)

}

func TestGF256Inv(t *testing.T) {
	gf := GF256(256)
	tm1 := int64(0)
	tm2 := int64(0)

	for i := 0; i < 100; i++ {
		r := rand.Uint64()
		a := uint256.NewInt(r)

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
	gf := GF256(256)
	tm1 := int64(0)
	tm2 := int64(0)

	for i := 0; i < 100; i++ {
		r := rand.Uint64()
		a := uint256.NewInt(r)
		// a := uint256.NewInt(uint64(i))

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
		assert.Equal(t, inv, inv2)
	}

	log.Printf("Time1 : %v", tm1/1000000)
	log.Printf("Time2 : %v", tm2/1000000)

}

func TestGF256InvF2Inv(t *testing.T) {
	gf := GF256(256)
	tm1 := int64(0)
	tm2 := int64(0)

	// P1 = (3*P - 2)/2  <--> Inverse of x^2
	P1 := uint256.NewInt(1)
	P1.Lsh(P1, gf.Size-1)

	// X^2
	P2 := uint256.NewInt(2)

	log.Printf("P1 : %x, P2 : %x", P1, P2)

	for i := 0; i < 100; i++ {
		r := rand.Uint64()
		a := uint256.NewInt(r)

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

func TestGF256ZPoly(t *testing.T) {
	gf := GF256(256)
	tm1 := int64(0)
	tm2 := int64(0)

	for i := 0; i < 1; i++ {
		var xs []*uint256.Int
		for j := 0; j < 100; j++ {
			r := rand.Uint64() & (1<<gf.Size - 1)
			a := uint256.NewInt(r)
			// a := uint256.NewInt(uint64(2*j + 1))
			xs = append(xs, a)
		}

		start := time.Now().UnixNano()
		zp := gf.ZPoly(xs)
		end := time.Now().UnixNano()
		tm1 += (end - start)

		log.Printf("zpoly : %v", zp)

		start = time.Now().UnixNano()
		for k := 0; k < len(xs); k++ {
			ev := gf.EvalPolyAt(zp, xs[k])
			log.Printf("%v - ev : %v", k, ev)
			assert.Equal(t, 0, ev.BitLen())
		}

		end = time.Now().UnixNano()
		tm2 += (end - start)

	}

	log.Printf("enc : %v, dec : %v", tm1/1000000, tm2/1000000)
}

func TestGF256DivPolys(t *testing.T) {
	gf := GF256(64)
	tm1 := int64(0)

	for i := 0; i < 100; i++ {
		var poly []*uint256.Int
		for j := 0; j < 10; j++ {
			r := rand.Uint64()
			a := uint256.NewInt(r)
			// a := uint256.NewInt(int64(j + 1))

			poly = append(poly, a)
		}

		start := time.Now().UnixNano()
		zp := gf.ZPoly(poly)
		for k := 0; k < len(poly); k++ {
			var sub []*uint256.Int
			for l := 0; l < len(poly); l++ {
				if l != k {
					sub = append(sub, poly[l])
				}
			}

			sp := gf.ZPoly(sub)
			dp := gf.DivPolys(zp, []*uint256.Int{poly[k], uint256.NewInt(1)})
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

func TestGF256LagrangeInterp(t *testing.T) {
	gf := GF256(256)
	tm1 := int64(0)

	for i := 0; i < 100; i++ {
		var xs []*uint256.Int
		var ys []*uint256.Int

		for j := 0; j < 50; j++ {
			r := rand.Uint64()
			a := uint256.NewInt(r)
			ys = append(ys, a)
			b := uint256.NewInt(uint64(j + 1))
			xs = append(xs, b)
		}

		start := time.Now().UnixNano()
		lp := gf.LagrangeInterp(xs, ys)
		log.Printf("lp : %v", lp)
		end := time.Now().UnixNano()
		tm1 += (end - start)

		for j := 0; j < len(xs); j++ {
			ev := gf.EvalPolyAt(lp, xs[j])
			assert.Equal(t, 0, ev.Cmp(ys[j]))
			log.Printf("evaulate poly %v : %v", ys[j], ev)
		}
	}

	log.Printf("lp : %v", tm1/1000000)
}

func TestGF256ExtRootUnity(t *testing.T) {
	gf := GF256(16)
	tm1 := int64(0)

	log.Printf("P-1 : %v", gf.GMask)

	for i := uint64(2); i < gf.GMask.Uint64(); i++ {
		for j := i; j <= gf.GMask.Uint64(); j *= j {
			a := uint256.NewInt(uint64(j))
			start := time.Now().UnixNano()
			size, rus := gf.ExtRootUnity(a, false)
			end := time.Now().UnixNano()
			tm1 += (end - start)

			if len(rus) < 1300 {
				log.Printf("%v of rus[%v] - [%v]", a, size, rus)
			}
		}
	}

	log.Printf("lp : %v", tm1/1000000)
}

func TestGF256ExtFindRootUnity(t *testing.T) {
	gf := GF256(16)

	g1 := gf.Exp256(uint256.NewInt(2), uint256.NewInt(257))
	ev := gf.Exp256(g1, uint256.NewInt(255))
	log.Printf("==>G(%v): %v", g1, ev)
	size, rus := gf.ExtRootUnity(g1, false)
	log.Printf("==>G(%v), %v: %v", g1, size, rus)

	g2 := gf.Exp256(uint256.NewInt(2), uint256.NewInt(1285))
	ev2 := gf.Exp256(g2, uint256.NewInt(51))
	log.Printf("==>G(%v): %v", g2, ev2)
	size2, rus2 := gf.ExtRootUnity(g2, false)
	log.Printf("==>G(%v), %v: %v", g2, size2, rus2)

	g3 := gf.Exp256(uint256.NewInt(2), uint256.NewInt(3855))
	ev3 := gf.Exp256(g3, uint256.NewInt(17))
	log.Printf("==>G(%v): %v", g3, ev3)
	size3, rus3 := gf.ExtRootUnity(g2, false)
	log.Printf("==>G(%v), %v: %v", g3, size3, rus3)
}
func TestGF256ExtFindRootUnity2(t *testing.T) {
	gf := GF256(8)

	g1 := gf.Exp256(uint256.NewInt(2), uint256.NewInt(85))
	ev := gf.Exp256(g1, uint256.NewInt(48))
	log.Printf("==>G(%v): %v", g1, ev)
	size, rus := gf.ExtRootUnity(g1, false)
	log.Printf("==>G(%v), %v: %v", g1, size, rus)

	// g2 := gf.Exp256(uint256.NewInt(2), uint256.NewInt(85))
	// ev2 := gf.Exp256(g2, uint256.NewInt(12))
	// log.Printf("==>G(%v): %v", g2, ev2)
	// size2, rus2 := gf.ExtRootUnity(g2, false)
	// log.Printf("==>G(%v), %v: %v", g2, size2, rus2)
}

func TestGF256ExtFindRootUnityAll(t *testing.T) {
	gf := GF256(6)

	for i := uint64(1); i <= gf.GMask.Uint64(); i++ {
		g1 := uint256.NewInt(i)
		// log.Printf("==>G: %v", g1)
		size, rus := gf.ExtRootUnity(g1, false)
		log.Printf("==>G(%v, %v) :(%v) %v", g1, gf.Inv256(g1), size, rus[:3])
	}

	// log.Println("=================================")
	// for i := int64(1); i <= gf.GMask.Int64(); i++ {
	// 	g1 := uint256.NewInt(i)
	// 	// log.Printf("==>G: %v", g1)
	// 	size, rus := gf.ExtRootUnity(g1, true)
	// 	log.Printf("==>G(%v, %v)-(%v) : %v", g1, gf.Inv256(g1), size, rus)

	// }
}

func TestGF256ExtFindRootUnityAny(t *testing.T) {
	fd := 32
	gf := GF256(fd)

	for i := uint64(100); i <= 1000; i++ {
		any := ((1<<fd - 1) * i) / 3 % (1<<fd - 1)

		g1 := uint256.NewInt(any)
		// log.Printf("==>G: %v", g1)
		size, _ := gf.ExtRootUnity(g1, false)
		log.Printf("==>G(%v, %v) : %v", g1, gf.Inv256(g1), size)
	}

	// log.Println("=================================")
	// for i := int64(1); i <= gf.GMask.Int64(); i++ {
	// 	g1 := uint256.NewInt(i)
	// 	// log.Printf("==>G: %v", g1)
	// 	size, rus := gf.ExtRootUnity(g1, true)
	// 	log.Printf("==>G(%v, %v)-(%v) : %v", g1, gf.Inv256(g1), size, rus)

	// }
}

func TestGF256ExtFindRootUnityAll2(t *testing.T) {
	gf := GF256(8)

	for i := uint64(1); i <= gf.GMask.Uint64(); i++ {
		// for i := uint64(1<<16 - 100); i <= uint64(1<<16+10); i++ {
		g1 := uint256.NewInt(i)
		// log.Printf("==>G: %v", g1)
		size, _ := gf.ExtRootUnity(g1, false)
		if 0 < size {
			log.Printf("==>G(%b, %v)-(%v)", g1, gf.Inv256(g1), size)
		}
	}
}

func TestGF256FFT(t *testing.T) {
	gf := GF256(4)

	g1 := uint256.NewInt(3)
	size, xs := gf.ExtRootUnity(g1, false)
	ys := make([]*uint256.Int, 0, size)

	for j := 0; j < size; j++ {
		r := rand.Uint64() % (1 << 4)
		a := uint256.NewInt(r)
		ys = append(ys, a)
	}

	log.Printf("==>xs:%v, ys:%v", xs, ys)
	start := time.Now().Nanosecond()
	os1 := gf.LagrangeInterp(xs, ys)
	end := time.Now().Nanosecond()
	log.Printf("LagrangeInterp(%v) : f(x)=%v", (end-start)/1000, os1)

	start = time.Now().Nanosecond()
	os2 := gf.IDFT(ys, g1)
	end = time.Now().Nanosecond()
	log.Printf("IDFT(%v) : f(x)=%v", (end-start)/1000, os2)

	os3 := gf.DFT(os1, g1)
	log.Printf("DFT : %v", os3)

}

func TestGF256FFTPerf(t *testing.T) {
	gf := GF256(8)

	g1 := uint256.NewInt(3)
	size, xs := gf.ExtRootUnity(g1, false)
	log.Printf("xs(%v) :  %v", size, xs[:10])

	if size == -1 {
		return
	}

	ys := make([]*uint256.Int, 0, size)
	for j := 0; j < size; j++ {
		r := rand.Uint64() % (1 << gf.Size)
		a := uint256.NewInt(r)
		ys = append(ys, a)
	}

	log.Printf("ys(%v) :  %v", size, ys[:10])

	start := time.Now().Nanosecond()
	os2 := gf.IDFT(ys, g1)
	end := time.Now().Nanosecond()
	log.Printf("IDFT(%v) : f(x)=%v", (end-start)/1000, os2[:10])

	os3 := gf.DFT(os2, g1)
	log.Printf("DFT :  %v", os3[:10])

}
