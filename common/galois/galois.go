package galois

import (
	"log"
)

type GF struct {
	base  uint
	prime uint64
	size  uint64
	invs  uint64
	align uint64
}

// var PRIME = []uint64{
// 	0,
// 	1,                // 1
// 	07,               // 2
// 	013,              // 3
// 	023,              // 4
// 	045,              // 5
// 	0103,             // 6
// 	0211,             // 7
// 	0435,             // 8
// 	01021,            // 9
// 	02011,            // 10
// 	04005,            // 11
// 	010123,           // 12
// 	020033,           // 13
// 	042103,           // 14
// 	0100003,          // 15
// 	0210013,          // 16
// 	0400011,          // 17
// 	01000201,         // 18
// 	02000047,         // 19
// 	04000011,         // 20
// 	010000005,        // 21
// 	020000003,        // 22
// 	040000041,        // 23
// 	0100000207,       // 24
// 	0200000011,       // 25
// 	0400000107,       // 26
// 	01000000047,      // 27
// 	02000000011,      // 28
// 	04000000005,      // 29
// 	010040000007,     // 30
// 	020000000011,     // 31
// 	040020000007,     // 32
// 	0100000020001,    // 33
// 	0201000000007,    // 34
// 	0400000000005,    // 35
// 	01000000004001,   // 36
// 	02000000012005,   // 37
// 	04000000000143,   // 38
// 	010000000000021,  // 39
// 	020000012000005,  // 40
// 	040061000000001,  // 41
// 	0100230000000001, // 42
// 	0200001020000041, // 43
// }

// http://poincare.matf.bg.ac.rs/~ezivkovm/publications/primpol1.pdf
// Irreducible Polynomial
// F(x) = x^64 + x^61 + x^34 + x^9 + 1
// var P64 = []uint{64, 61, 34, 9}
// F(x) = x^32 + x^16 + x^7 + x^2 + 1
// var P64 = []uint{32, 16, 7, 2}

var P64 = map[uint][]uint{
	3:  {3, 1},
	4:  {4, 1},
	5:  {5, 3, 2, 1},
	6:  {6, 5, 4, 1},
	7:  {7, 4, 3, 2},
	8:  {8, 7, 2, 1},
	16: {16, 15, 12, 10},
	32: {32, 16, 7, 2},
	40: {40, 29, 27, 23},
	48: {48, 19, 9, 1},
	56: {56, 41, 39, 29},
	64: {64, 61, 34, 9},
}

func GFN(base uint) *GF {
	// if base < 2 || base > uint8(len(PRIME)-1) {
	// 	log.Panicf("Greate GF(%v) Error!!!", base)
	// 	return nil
	// }
	field := P64[base]
	if field == nil {
		log.Panicf("Not support the prime field : %v", base)
		return nil
	}

	p := uint64(1)
	for _, x := range field {
		p += 1 << x
	}

	poly := GF{}
	poly.base = base
	poly.prime = p
	poly.size = (1 << base) - 1
	poly.invs = uint64(1 << (base - 1))
	poly.align = uint64((base-1)/8) + 1

	return &poly
}

func (gf *GF) GetAlign() uint64 {
	return gf.align
}

func (gf *GF) AddN(x, y uint64) uint64 {
	return (x ^ y) & uint64(gf.size)
}

func (gf *GF) SubN(x, y uint64) uint64 {
	return (x ^ y) & uint64(gf.size)
}

func (gf *GF) lShift(x uint64) uint64 {
	s := gf.base - 1
	if x&(0x01<<s) != 0 {
		x = (x << 1) ^ gf.prime // PRIME[table.base]
	} else {
		x = x << 1
	}
	return x & uint64(gf.size)
}

// x**p = x
func (gf *GF) Mul(x, y uint64) uint64 {
	r := uint64(0)
	t := x
	for i := 0; i < int(gf.base); i++ {
		if y&(0x01<<i) != 0 {
			r = r ^ t
		}
		t = gf.lShift(t)
	}
	return r
}

// x**p = x, so x**(p-2) = x**(-1)
func (gf *GF) Div(x, y uint64) uint64 {
	if y == 0 {
		log.Printf("Div by zero")
		return 0
	}
	inv := gf.Exp(y, uint64(gf.size)-1)
	return gf.Mul(x, inv)
}

func (gf *GF) Exp(x, y uint64) uint64 {
	p := x
	b := uint64(1)
	max := 64
	for i := 0; i < 64; i++ {
		if (y >> i) == 0 {
			max = i
			break
		}
	}

	for i := 0; i < max; i++ {
		if y&(0x01<<i) != 0 {
			b = gf.Mul(b, p)
		}
		p = gf.Mul(p, p)
	}

	return b
}

// Compute square root
func (gf *GF) SqrR(x uint64) uint64 {
	return gf.Exp(x, gf.invs)
}
