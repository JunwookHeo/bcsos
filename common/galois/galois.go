/*
**
// cmehay/galois
// This is from the linke of https://sourcegraph.com/github.com/cmehay/galois
**
*/
package galois

import (
	"errors"
	"log"
)

// MaxGF is the maximum GF(x)
// default 24, up to 31 if you have petabytes of ram :p
const MaxGF = 24

var primPoly = []uint32{
	0,
	/*  1 */ 1,
	/*  2 */ 07,
	/*  3 */ 013,
	/*  4 */ 023,
	/*  5 */ 045,
	/*  6 */ 0103,
	/*  7 */ 0211,
	/*  8 */ 0435,
	/*  9 */ 01021,
	/* 10 */ 02011,
	/* 11 */ 04005,
	/* 12 */ 010123,
	/* 13 */ 020033,
	/* 14 */ 042103,
	/* 15 */ 0100003,
	/* 16 */ 0210013,
	/* 17 */ 0400011,
	/* 18 */ 01000201,
	/* 19 */ 02000047,
	/* 20 */ 04000011,
	/* 21 */ 010000005,
	/* 22 */ 020000003,
	/* 23 */ 040000041,
	/* 24 */ 0100000207,
	/* 25 */ 0200000011,
	/* 26 */ 0400000107,
	/* 27 */ 01000000047,
	/* 28 */ 02000000011,
	/* 29 */ 04000000005,
	/* 30 */ 010040000007,
	/* 31 */ 020000000011,
	/* 32 */ /* 040020000007, overflow */
}

// GfPoly is Polynomial struct
type GfPoly struct {
	base   uint8
	NW     uint32
	gflog  []uint32
	gfilog []uint32
}

var gfPolyInstance = make([]*GfPoly, MaxGF+1)

func newGF(base uint8) (*GfPoly, error) {

	var b, log uint32
	var poly *GfPoly

	poly = new(GfPoly)
	if base < 2 || base > MaxGF {
		return nil, errors.New("Prim polynomial out of range")
	}
	poly.base = base
	poly.NW = 1 << base
	poly.gflog = make([]uint32, poly.NW)
	poly.gfilog = make([]uint32, poly.NW)
	b = 1

	for log = 0; log < poly.NW; log++ {
		poly.gflog[b] = log
		poly.gfilog[log] = b
		b = b << 1
		if b&poly.NW != 0 {
			b = b ^ primPoly[base]
		}
	}
	return poly, nil
}

// GF is a singleton getter for new GfPoly struct
func GF(base uint8) (*GfPoly, error) {
	var error error

	if base < 2 || base > MaxGF {
		return nil, errors.New("Prim polynomial out of range")
	}
	if gfPolyInstance[base] == nil {
		gfPolyInstance[base], error = newGF(base)
	}
	return gfPolyInstance[base], error
}

// Mul is GfPoly struct method for multiplication
func (table *GfPoly) Mul(a, b uint32) (uint32, error) {
	if a == 0 || b == 0 {
		return 0, nil
	}
	if a >= table.NW || b >= table.NW {
		return 0, errors.New("mul: polynomial out of range")
	}
	sumLog := table.gflog[a] + table.gflog[b]
	if sumLog >= table.NW-1 {
		sumLog -= table.NW - 1
	}
	return table.gfilog[sumLog], nil
}

// Div is GfPoly struct method for division
func (table *GfPoly) Div(a, b uint32) (uint32, error) {
	var diffLog int64
	if b == 0 {
		return 0, errors.New("div: division by 0 :/")
	}
	if a == 0 {
		return 0, nil
	}
	if a >= table.NW || b >= table.NW {
		return 0, errors.New("div: polynomial out of range")
	}
	diffLog = int64(table.gflog[a]) - int64(table.gflog[b])
	if diffLog < 0 {
		diffLog += int64(table.NW) - 1
	}
	return table.gfilog[diffLog], nil
}

// Expon is GfPoly struct method for exponential
func (table *GfPoly) Expon(a, e uint32) (uint32, error) {

	var err error
	var i uint32

	b := a
	if e == 0 {
		return 1, nil
	}
	if e == 1 {
		return a, nil
	}
	for i = 1; i < e; i++ {
		b, err = table.Mul(b, a)
		if err != nil {
			return 0, err
		}
	}
	return b, nil
}

func (table *GfPoly) Expon2(a, e uint32) (uint32, error) {

	var err error
	var i uint32

	// b := a
	if e == 0 {
		return 1, nil
	}
	if e == 1 {
		return a, nil
	}

	p := a
	b := uint32(1)

	for i = 0; i < 32; i++ {
		if err != nil {
			return 0, err
		}

		if e&(0x01<<i) != 0 {
			b, err = table.Mul(b, p)
			if err != nil {
				return 0, err
			}
		}
		p, err = table.Mul(p, p)
	}

	return b, nil
}

type GfPolyN struct {
	base uint8
	NW   uint64
}

var primPoly2 = []uint64{
	0,
	1,               // 1
	07,              // 2
	013,             // 3
	023,             // 4
	045,             // 5
	0103,            // 6
	0211,            // 7
	0435,            // 8
	01021,           // 9
	02011,           // 10
	04005,           // 11
	010123,          // 12
	020033,          // 13
	042103,          // 14
	0100003,         // 15
	0210013,         // 16
	0400011,         // 17
	01000201,        // 18
	02000047,        // 19
	04000011,        // 20
	010000005,       // 21
	020000003,       // 22
	040000041,       // 23
	0100000207,      // 24
	0200000011,      // 25
	0400000107,      // 26
	01000000047,     // 27
	02000000011,     // 28
	04000000005,     // 29
	010040000007,    // 30
	020000000011,    // 31
	040020000007,    // 32
	0100000020001,   // 33
	0201000000007,   // 34
	0400000000005,   // 35
	01000000004001,  // 36
	02000000012005,  // 37
	04000000000143,  // 38
	010000000000021, // 39
	020000012000005, // 40
}

func GFN(base uint8) *GfPolyN {
	if base < 2 || base > 40 {
		log.Panicf("Greate GF(%v) Error!!!", base)
		return nil
	}

	poly := GfPolyN{}
	poly.base = base
	poly.NW = 1 << base
	return &poly
}

func (table *GfPolyN) AddN(x, y uint64) uint64 {
	return x ^ y
}

func (table *GfPolyN) SubN(x, y uint64) uint64 {
	return x ^ y
}

func (table *GfPolyN) lShift(x uint64) uint64 {
	s := table.base - 1
	if x&(0x01<<s) != 0 {
		x = (x << 1) ^ primPoly2[table.base]
	} else {
		x = x << 1
	}
	return x & uint64(table.NW-1)
}

// x**p = x
func (table *GfPolyN) MulN(x, y uint64) uint64 {
	r := uint64(0)
	t := x
	for i := 0; i < int(table.base); i++ {
		if y&(0x01<<i) != 0 {
			r = r ^ t
		}
		t = table.lShift(t)
	}
	return r
}

// x**p = x, so x**(p-2) = x**(-1)
func (table *GfPolyN) DivN(x, y uint64) uint64 {
	if y == 0 {
		log.Printf("Div by zero")
		return 0
	}
	inv := table.Exp(y, uint64(table.NW)-2)
	return table.MulN(x, inv)
}

func (table *GfPolyN) Exp(x, y uint64) uint64 {
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
			b = table.MulN(b, p)
		}
		p = table.MulN(p, p)
	}

	return b
}
