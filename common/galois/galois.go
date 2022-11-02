package galois

import (
	"log"
)

type gf struct {
	base uint8
	size uint64
}

var PRIME = []uint64{
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

func GFN(base uint8) *gf {
	if base < 2 || base > uint8(len(PRIME)-1) {
		log.Panicf("Greate GF(%v) Error!!!", base)
		return nil
	}

	poly := gf{}
	poly.base = base
	poly.size = 1 << base
	return &poly
}

func (table *gf) AddN(x, y uint64) uint64 {
	return x ^ y
}

func (table *gf) SubN(x, y uint64) uint64 {
	return x ^ y
}

func (table *gf) lShift(x uint64) uint64 {
	s := table.base - 1
	if x&(0x01<<s) != 0 {
		x = (x << 1) ^ PRIME[table.base]
	} else {
		x = x << 1
	}
	return x & uint64(table.size-1)
}

// x**p = x
func (table *gf) Mul(x, y uint64) uint64 {
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
func (table *gf) Div(x, y uint64) uint64 {
	if y == 0 {
		log.Printf("Div by zero")
		return 0
	}
	inv := table.Exp(y, uint64(table.size)-2)
	return table.Mul(x, inv)
}

func (table *gf) Exp(x, y uint64) uint64 {
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
			b = table.Mul(b, p)
		}
		p = table.Mul(p, p)
	}

	return b
}
