package main

import (
	"log"
	"math/big"
	"math/rand"
	"time"

	"github.com/holiman/uint256"
	"github.com/junwookheo/bcsos/blockchainsim/simulation"
	"github.com/junwookheo/bcsos/common/bitcoin"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/galois"
	"github.com/junwookheo/bcsos/common/poscipher"
	"github.com/junwookheo/bcsos/common/starks"
	"github.com/junwookheo/bcsos/common/wallet"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func test_encypt_decrypt() {
	const PATH_TEST = "blocks_360.json"
	w := wallet.NewWallet("blocks.json.wallet")
	key := w.PublicKey
	addr := w.PublicKey

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
		_, y := poscipher.EncryptPoSWithVariableLength(key, poscipher.CalculateXorWithAddress(addr, x))
		tenc += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Encryption Time : %v", tenc)
		log.Printf("Enc x:%x", y[0:80])
		// Start Decryption
		start = time.Now().UnixNano()
		x_t := poscipher.DecryptPoSWithVariableLength(key, y)
		x_t = poscipher.CalculateXorWithAddress(addr, x_t)
		tdec += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Decryption Time : %v", tdec)

		log.Printf("Org x:%v", x[0:80])
		log.Printf("New x:%v", x_t[0:80])
		key = y
	}
	close(msg)
}

func test_encypt_2() {
	gf := galois.NewGFP()
	if gf == nil {
		log.Println("GF(1) should rise error")
		return
	}

	p := gf.Prime.ToBig()
	p = p.Mul(p, big.NewInt(2))
	p = p.Sub(p, big.NewInt(1))
	p = p.Div(p, big.NewInt(3))
	p = p.Mod(p, gf.Prime.ToBig())
	I, _ := uint256.FromBig(p)
	log.Printf("Inv : %x", I)

	tenc := int64(0)
	tdec := int64(0)

	for k := 0; k < 10000; k++ {
		size := 100
		x := make([]uint64, size)
		k := make([]uint64, len(x))
		y := make([]*uint256.Int, len(x))
		for i := 0; i < len(x); i++ {
			x[i] = rand.Uint64()
			k[i] = rand.Uint64()
		}

		pre := uint256.NewInt(0)
		start := time.Now().UnixNano()
		for i := 0; i < len(x); i++ {
			xu := uint256.NewInt(x[i])
			ku := uint256.NewInt(k[i])
			d := gf.Add(xu, ku)
			d = gf.Add(d, pre)
			y[i] = gf.Exp(d, I)

			pre = y[i]
			// log.Printf("X:%v, d:%v, y:%v", x[i], d, y[i])
		}
		tenc += (time.Now().UnixNano() - start) / 1000000 // msec

		pre = uint256.NewInt(0)
		start = time.Now().UnixNano()
		for i := 0; i < len(y); i++ {
			ku := uint256.NewInt(k[i])
			d := gf.Exp(y[i], uint256.NewInt(3))
			d = gf.Sub(d, pre)
			d = gf.Sub(d, ku)
			pre = y[i]
			if x[i] != d.Uint64() {
				log.Printf("%v, X:%v, y:%v", i, x[i], d)
				break
			}

		}
		tdec += (time.Now().UnixNano() - start) / 1000000 // msec
	}

	log.Printf("enc : %v, dec : %v", tenc, tdec)
}

func test_encypt_decrypt_prime() {
	const PATH_TEST = "blocks_360.json"
	w := wallet.NewWallet("blocks.json.wallet")
	key := w.PublicKey
	addr := w.PublicKey

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
		_, y := poscipher.EncryptPoSWithPrimeField(key, poscipher.CalculateXorWithAddress(addr, x))
		tenc += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Encryption Time : %v", tenc)
		log.Printf("Enc x %v :%x", len(y), y[0:80])
		// Start Decryption
		start = time.Now().UnixNano()
		x_t := poscipher.DecryptPoSWithPrimeField(key, y)
		x_t = poscipher.CalculateXorWithAddress(addr, x_t[:len(x)])
		tdec += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Decryption Time : %v", tdec)

		log.Printf("Org x %v :%v", len(x), x[len(x)-80:])
		log.Printf("New x %v :%v", len(x_t), x_t[len(x)-80:len(x)])
		key = y
	}
	close(msg)
}

func test_fri_prove_low_degree() {
	f := starks.NewStarks()
	length := 65536
	ys := make([]*uint256.Int, length)
	for i := 0; i < len(ys); i++ {
		r := rand.Int63()
		ys[i] = uint256.NewInt(uint64(r))
	}

	g := f.GFP.Prime.Clone()
	g.Sub(g, uint256.NewInt(1))
	g.Div(g, uint256.NewInt(uint64(length)))
	g1 := f.GFP.Exp(uint256.NewInt(7), g)

	tm1 := int64(0)
	start := time.Now().UnixNano()
	proof := f.ProveLowDegree(ys, g1)
	end := time.Now().UnixNano()
	tm1 = end - start
	log.Printf("size of Proof : %v, %v", len(proof), tm1/1000000)

	m1 := f.Merklize(ys)
	tm2 := int64(0)
	start = time.Now().UnixNano()
	eval := f.VerifyLowDegreeProof(m1[1], proof, g1)
	end = time.Now().UnixNano()
	tm2 = end - start
	log.Printf("Eval : %v", eval)
	log.Printf("Verify: %v", tm2/1000000)
	// for i:=0; i<len(proof); i++{
	// 	log.Printf("Proof ======= %v", i)
	// 	log.Printf("Proof : %v", proof[i])
	// }

}

func test_starks_prime() {
	const PATH_TEST = "blocks_360.json"
	w := wallet.NewWallet("blocks.json.wallet")
	key := w.PublicKey
	addr := w.PublicKey

	msg := make(chan bitcoin.BlockPkt)
	go simulation.LoadBtcData(PATH_TEST, msg)

	tenc := int64(0)
	tdec := int64(0)
	f := starks.NewStarks()

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
		vis := poscipher.CalculateXorWithAddress(addr, x)
		_, y := poscipher.EncryptPoSWithPrimeField(key, vis)
		start := time.Now().UnixNano()
		f.GenerateStarksProof(vis, y, key)
		tenc += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Encryption Time : %v", tenc)
		log.Printf("Enc y %v :%v", len(y), y[len(y)-80:])
		// Start Decryption
		start = time.Now().UnixNano()
		x_t := poscipher.DecryptPoSWithPrimeField(key, y)
		x_t = poscipher.CalculateXorWithAddress(addr, x_t[:len(x)])
		tdec += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Decryption Time : %v", tdec)

		log.Printf("Org x %v :%v", len(x), x[len(x)-80:])
		log.Printf("New x %v :%v", len(x_t), x_t[len(x_t)-80:])
		key = y
		break
	}
	close(msg)
}

func main() {
	// test_encypt_2()
	// test_encypt_decrypt()
	// test_fri_prove_low_degree()
	// test_encypt_decrypt_prime()
	test_starks_prime()
}
