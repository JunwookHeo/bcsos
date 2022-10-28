package main

import (
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

func main() {
	// test_gf_8()
	// test_gf_16()
	// test_gf()
	test_gf2()
}
