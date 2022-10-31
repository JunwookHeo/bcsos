package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"log"
	"time"

	"github.com/junwookheo/bcsos/common/galois"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}
func test_gf_8() {
	gfpoly, err := galois.GF(8)
	if err == nil {
		log.Println("GF(1) should rise error")
	}

	start := time.Now().UnixNano()
	var cal uint32
	for i := 0; i < 256; i++ {
		cal, _ = gfpoly.Expon(uint32(i), 7)
	}
	end := time.Now().UnixNano()
	log.Printf("Time for FWD : %v", end-start)
	log.Printf("Time for FWD : %v", cal)

	start = time.Now().UnixNano()
	for i := 0; i < 256; i++ {
		cal, _ = gfpoly.Expon(uint32(i), 73)
	}
	end = time.Now().UnixNano()
	log.Printf("Time for REV : %v", end-start)
	log.Printf("Time for FWD : %v", cal)
}

func test_gf_16() {
	gfpoly, err := galois.GF(16)
	if err == nil {
		log.Println("GF(1) should rise error")
	}
	// 11 <-> 23831
	// 2 <-> 98303
	start := time.Now().UnixNano()
	var cal uint32
	for i := 0; i < 65536; i++ {
		cal, _ = gfpoly.Expon(uint32(i), 11)
	}
	end := time.Now().UnixNano()
	log.Printf("Time for FWD : %v", end-start)
	log.Printf("Time for FWD : %v", cal)

	start = time.Now().UnixNano()
	for i := 0; i < 65536; i++ {
		cal, _ = gfpoly.Expon(uint32(i), 23831)
	}
	end = time.Now().UnixNano()
	log.Printf("Time for REV : %v", end-start)
	log.Printf("Time for REV : %v", cal)
}

func test_gf() {
	gfpoly, err := galois.GF(16)
	if err == nil {
		log.Println("GF(1) should rise error")
	}

	start := time.Now().UnixNano()
	var cal uint32
	for i := 0; i < 65536; i++ {
		cal, _ = gfpoly.Expon(uint32(i), 2)
		cal, _ = gfpoly.Expon(cal, 98303)
		log.Printf("%v <==> %v", i, cal)
	}
	end := time.Now().UnixNano()
	log.Printf("Time for REV : %v", end-start)
}

func test_gf2() {
	gfpoly, err := galois.GF(16)
	if err != nil {
		log.Println("GF(1) should rise error")
	}

	start := time.Now().UnixNano()
	var cal uint32
	max_len := 65536
	for i := 0; i < max_len; i++ {
		cal, _ = gfpoly.Expon(uint32(i), 2)
	}
	end := time.Now().UnixNano()
	log.Printf("Time for EXP1 : %v", (end-start)/1000)
	log.Printf("<==> %v", cal)

	start = time.Now().UnixNano()
	for i := 0; i < max_len; i++ {
		cal, _ = gfpoly.Expon2(uint32(i), 98303)
	}
	end = time.Now().UnixNano()
	log.Printf("Time for EXP2 : %v", (end-start)/1000)
	log.Printf("<==> %v", cal)
}

// func test_gf_mult() {
// 	gfpoly, err := galois.GF(16)
// 	if err != nil {
// 		log.Println("GF(1) should rise error")
// 	}

// 	start := time.Now().UnixNano()
// 	var cal1 uint32
// 	var cal2 uint64
// 	max_len := 65536
// 	for i := 0; i < max_len; i++ {
// 		for j := 0; j < max_len; j++ {
// 			cal1, _ = gfpoly.Mul(uint32(i), uint32(j))
// 			cal2 = gfpoly.MulN(uint64(i), uint64(j))
// 			if cal1 != uint32(cal2) {
// 				log.Printf("Fail (%v, %v) : %v - %v", i, j, cal1, cal2)
// 				return
// 			}
// 		}
// 	}
// 	end := time.Now().UnixNano()
// 	log.Printf("Time for EXP1 : %v", (end-start)/1000)
// }

// func test_gf_div() {
// 	gfpoly, err := galois.GF(16)
// 	if err != nil {
// 		log.Println("GF(1) should rise error")
// 	}

// 	start := time.Now().UnixNano()
// 	var cal1 uint32
// 	var cal2 uint64
// 	max_len := 65536
// 	for i := 65536 / 2; i < max_len; i++ {
// 		for j := 0; j < max_len; j++ {
// 			cal1, _ = gfpoly.Div(uint32(i), uint32(j))
// 			cal2 = gfpoly.DivN(uint64(i), uint64(j))
// 			if cal1 != uint32(cal2) {
// 				log.Printf("Fail (%v, %v) : %v - %v", i, j, cal1, cal2)
// 				return
// 			}
// 		}
// 	}
// 	end := time.Now().UnixNano()
// 	log.Printf("Time for EXP1 : %v", (end-start)/1000)
// }

// func test_gf_exp() {
// 	gfpoly, err := galois.GF(16)
// 	if err != nil {
// 		log.Println("GF(1) should rise error")
// 	}

// 	start := time.Now().UnixNano()
// 	var cal1 uint32
// 	var cal2 uint64
// 	max_len := 65536
// 	for i := 0; i < max_len; i++ {
// 		for j := 0; j < max_len; j++ {
// 			cal1, _ = gfpoly.Expon(uint32(i), uint32(j))
// 			cal2 = gfpoly.Exp(uint64(i), uint64(j))
// 			if cal1 != uint32(cal2) {
// 				log.Printf("Fail (%v, %v) : %v - %v", i, j, cal1, cal2)
// 				return
// 			}
// 		}
// 	}
// 	end := time.Now().UnixNano()
// 	log.Printf("Time for EXP1 : %v", (end-start)/1000)
// }

// 2^32 :
//   - 2P -1 :  7 * 1227133513
//   - 3P -2 :  2 * 6442450943
func test_gf_exp2() {
	gfpoly := galois.GFN(32)
	if gfpoly == nil {
		log.Println("GF(1) should rise error")
		return
	}

	start := time.Now().UnixNano()
	var cal1 uint64
	var cal2 uint64
	max_len := 4294967296
	for i := 0; i < max_len; i++ {
		cal1 = gfpoly.Exp(uint64(i), 2)
		cal2 = gfpoly.Exp(uint64(cal1), 6442450943)
		if uint64(i) != cal2 {
			log.Printf("ERROR   %v : %v <==> %v", i, cal1, cal2)
			break
		}
	}
	end := time.Now().UnixNano()
	log.Printf("Time for EXP1 : %v", (end-start)/1000)
}

// 2^32 :
//   - 2P -1 :  7 * 1227133513
//   - 3P -2 :  2 * 6442450943
func test_gf_exp3() {
	gfpoly := galois.GFN(32)
	if gfpoly == nil {
		log.Println("GF(1) should rise error")
		return
	}

	start := time.Now().UnixNano()
	var cal uint64
	max_len := 4294967296
	for i := 1; i < max_len; i += 4096 {
		cal = gfpoly.Exp(uint64(i), 2)
	}
	end := time.Now().UnixNano()
	log.Printf("Time for EXP1 : %v", (end-start)/1000000)
	log.Printf("<==> %v", cal)

	start = time.Now().UnixNano()
	for i := 1; i < max_len; i += 4096 {
		cal = gfpoly.Exp(uint64(i), 6442450943)
	}
	end = time.Now().UnixNano()
	log.Printf("Time for EXP2 : %v", (end-start)/1000000)
	log.Printf("<==> %v", cal)
}

func test_gf_div2() {
	gfpoly := galois.GFN(32)
	if gfpoly == nil {
		log.Println("GF(1) should rise error")
		return
	}

	start := time.Now().UnixNano()
	var cal1 uint64
	var cal2 uint64
	max_len := 4294967296
	for i := 1; i < max_len; i += 3333 {
		for j := 1; j < max_len; j += 3333 {
			cal1 = gfpoly.MulN(uint64(i), uint64(j))
			cal2 = gfpoly.DivN(uint64(cal1), uint64(j))
			if uint64(i) != cal2 {
				log.Printf("ERROR   %v : %v <==> %v", i, cal1, cal2)
				log.Printf("ERROR %v <==> %v", i, j)
				return
			}
		}
		log.Printf("PROCESS %v", i)
	}
	end := time.Now().UnixNano()
	log.Printf("Time for EXP1 : %v", (end-start)/1000)
}

func main() {
	// test_gf_8()
	// test_gf_16()
	// test_gf()
	// test_gf_div()
	// test_gf_exp3()
	// test_gf_div2()

	buf := new(bytes.Buffer)
	source := []uint32{1, 2, 3}
	err := binary.Write(buf, binary.LittleEndian, source)
	if err != nil {
		fmt.Println("binary.Write failed:", err)
	}
	fmt.Printf("Encoded: % x\n", buf.Bytes())

	check := make([]uint32, 3)
	rbuf := bytes.NewReader(buf.Bytes())
	err = binary.Read(rbuf, binary.LittleEndian, &check)
	if err != nil {
		fmt.Println("binary.Read failed:", err)
	}
	fmt.Printf("Decoded: %v\n", check)
}
