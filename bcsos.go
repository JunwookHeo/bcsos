package main

import (
	"log"
	"time"

	"github.com/junwookheo/bcsos/blockchainsim/simulation"
	"github.com/junwookheo/bcsos/common/bitcoin"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/galois"
	"github.com/junwookheo/bcsos/common/poscipher"
	"github.com/junwookheo/bcsos/common/wallet"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
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
			cal1 = gfpoly.Mul(uint64(i), uint64(j))
			cal2 = gfpoly.Div(uint64(cal1), uint64(j))
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

func test_encypt_1() {
	x := []uint32{21, 32, 43, 54, 65, 76, 87, 98, 19, 20, 31}
	k := []uint32{19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 19}
	y := make([]uint32, len(x))

	gfpoly := galois.GFN(32)
	if gfpoly == nil {
		log.Println("GF(1) should rise error")
		return
	}

	pre := uint32(0)
	for i := 0; i < len(x); i++ {
		d := (x[i] ^ k[i]) ^ pre
		y[i] = uint32(gfpoly.Exp(uint64(d), 6442450943))
		pre = y[i]
		log.Printf("X:%v, d:%v, y:%v", x[i], d, y[i])
	}

	pre = 0
	for i := 0; i < len(y); i++ {
		d := uint32(gfpoly.Exp(uint64(y[i]), 2))
		x[i] = (d ^ pre) ^ k[i]
		pre = y[i]
		log.Printf("X:%v, d:%v, y:%v", x[i], d, y[i])
	}
}

func test_encypt_2() {
	x := []uint32{21, 32, 43, 54, 65, 76, 87, 98, 19, 20, 31}
	k := []uint32{19, 18, 17, 16, 15, 14, 13, 12, 11, 10, 19}
	y := make([]uint32, len(x))

	gfpoly := galois.GFN(32)
	if gfpoly == nil {
		log.Println("GF(1) should rise error")
		return
	}

	pre := uint32(0)
	for i := 0; i < len(x); i++ {
		d := (x[i] ^ k[i]) ^ pre
		y[i] = uint32(gfpoly.Exp(uint64(d), 6442450943))
		pre = y[i]
		log.Printf("X:%v, d:%v, y:%v", x[i], d, y[i])
	}

	pre = 0
	for i := 0; i < len(y); i++ {
		d := uint32(gfpoly.Exp(uint64(y[i]), 2))
		x[i] = (d ^ pre) ^ k[i]
		pre = y[i]
		log.Printf("X:%v, d:%v, y:%v", x[i], d, y[i])
	}
}

func test_encypt_decrypt() {
	const PATH_TEST = "blocks.json"
	w := wallet.NewWallet("blocks.json.wallet")
	key := w.PublicKey

	msg := make(chan bitcoin.BlockPkt)
	go simulation.LoadBtcData(PATH_TEST, msg)

	tenc := int64(0)
	tdec := int64(0)
	for {
		d, ok := <-msg
		if !ok {
			log.Println("Channle closed")
			break
		}

		if d.Block == config.END_TEST {
			break
		}

		rb := bitcoin.NewRawBlock(d.Block)
		x := rb.GetBlockBytes()
		// log.Printf("Block : %v", x[:80])

		// Start Encryption
		start := time.Now().UnixNano()
		_, y := poscipher.EncryptPoSWithVariableLength(key, x)
		tenc += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Encryption Time : %v", tenc)

		// Start Decryption
		start = time.Now().UnixNano()
		x_t := poscipher.DecryptPoSWithVariableLength(key, y)
		tdec += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Decryption Time : %v", tdec)

		log.Printf("Org x:%v", x[0:80])
		log.Printf("New x:%v", x_t[0:80])
		key = y
	}
	close(msg)
}
func main() {
	// test_gf_8()
	// test_gf_16()
	// test_gf()
	// test_gf_div()
	// test_gf_exp3()
	// test_gf_div2()
	// test_encypt_2()
	test_encypt_decrypt()
}
