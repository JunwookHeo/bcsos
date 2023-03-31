package main

import (
	"encoding/hex"
	"flag"
	"log"
	"math/big"
	"math/rand"
	"os"
	"runtime/pprof"
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
		x := rb.GetBlockBytes(0)
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
		x := rb.GetBlockBytes(0)
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
	length := 65536 /// 4

	f := starks.NewStarks(length / 8)

	tm1 := int64(0)
	tm2 := int64(0)

	rand.Seed(time.Now().UnixNano())
	N := 1
	for i := 0; i < N; i++ {
		g := f.GFP.Prime.Clone()
		g.Sub(g, uint256.NewInt(1))
		g.Div(g, uint256.NewInt(uint64(length)))
		g2 := f.GFP.Exp(uint256.NewInt(7), g)
		log.Printf("g1 : %v", g2)

		yss := make([]*uint256.Int, length/8)
		yss2 := make([]*uint256.Int, length/8)
		for j := 0; j < len(yss); j++ {
			r := rand.Int63()
			yss[j] = uint256.NewInt(uint64(r))
			r = rand.Int63()
			yss2[j] = uint256.NewInt(uint64(r))
		}

		ys := f.GFP.DFT(yss, g2)
		ys2 := f.GFP.DFT(yss2, g2)
		for j := 0; j < len(ys); j++ {
			// r := rand.Int63()
			ys[j] = f.GFP.Div(ys[j], ys2[j])
			// ys[j] = f.GFP.Sub(ys[j], uint256.NewInt(uint64(600)))
		}

		start := time.Now().UnixNano()
		proof := f.ProveLowDegree(ys, g2)
		end := time.Now().UnixNano()
		tm1 += end - start
		log.Printf("size of Proof : %v, %v", len(proof), tm1/1000000)

		m1 := f.Merklize(ys)

		start = time.Now().UnixNano()
		eval := f.VerifyLowDegreeProof(m1[1], proof, g2)
		end = time.Now().UnixNano()
		tm2 += end - start
		log.Printf("Eval : %v", eval)
		log.Printf("Verify: %v", tm2/1000000)

		// Test Fake Data
		tfake := false
		if tfake == true {
			findx := rand.Int() % len(ys)
			ys[findx] = f.GFP.Add(ys[findx], uint256.NewInt(1))

			m2 := f.Merklize(ys)
			eval = f.VerifyLowDegreeProof(m2[1], proof, g2)
			log.Printf("Eval Fake : %v", eval)

			proof = f.ProveLowDegree(ys, g2)
			eval = f.VerifyLowDegreeProof(m1[1], proof, g2)
			log.Printf("Eval Fake : %v", eval)
		}
	}

	log.Printf("Avg Proof : %v, Verify : %v", tm1/int64(N)/1000000, tm2/int64(N)/1000000)
}

func test_starks_prime() {
	const PATH_TEST = "./blocks.json"
	w := wallet.NewWallet("blocks.json.wallet")
	addr := w.PublicKey
	key := make([]byte, 0, len(addr)*32)
	for i := 0; i < len(addr); i++ {
		t := uint256.NewInt(uint64(addr[i])).Bytes32()
		key = append(key, t[:]...)
	}

	msg := make(chan bitcoin.BlockPkt)
	go simulation.LoadBtcData(PATH_TEST, msg)

	tenc := int64(0)
	tdec := int64(0)
	tpro := int64(0)
	tver := int64(0)
	loop := 0

	f := starks.NewStarks(65536 / 16 / 4)

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
		x := rb.GetBlockBytes(f.GetSteps() * 31)

		// log.Printf("Block len : %v", len(x))
		// if len(x) < 64000 {
		// 	continue
		// }

		// Start Encryption
		start := time.Now().UnixNano()
		vis := poscipher.CalculateXorWithAddress(addr, x)
		_, y := poscipher.EncryptPoSWithPrimeField(key, vis)
		tenc += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Encryption Time : %v", tenc)
		// y[100] += 100
		// y[9] += 100

		// Start generating proof
		start = time.Now().UnixNano()
		proof := f.GenerateStarksProof(vis, y, key)
		tpro += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Generating Proof Time : %v, Merkle Root : %v", tpro, hex.EncodeToString(proof.MerkleRoot))

		// Start verification
		start = time.Now().UnixNano()
		ret := f.VerifyStarksProof(vis, proof)
		tver += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Verifying Proof Time : %v, %v", tver, ret)
		if !ret {
			log.Panicf("Verification Fail : %v", ret)
		}

		start = time.Now().UnixNano()
		x_t := poscipher.DecryptPoSWithPrimeField(key, y)
		x_o := poscipher.CalculateXorWithAddress(addr, x_t[:len(x)])
		tdec += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Decryption Time : %v", tdec)

		for i := 0; i < 10; i++ {
			r := rand.Int() % len(x)
			if x_o[r] != x[r] {
				log.Panicf("Decryption Fail : %v", x_o[r])
			}
		}

		key = y
		loop++
		// return
	}
	close(msg)
}

func test_starks_prime_prekey() {
	const PATH_TEST = "./blocks_720.json"
	w := wallet.NewWallet("blocks.json.wallet")
	addr := w.PublicKey
	key := make([]byte, 0, len(addr)*32)
	for i := 0; i < len(addr); i++ {
		t := uint256.NewInt(uint64(addr[i])).Bytes32()
		key = append(key, t[:]...)
	}

	msg := make(chan bitcoin.BlockPkt)
	go simulation.LoadBtcData(PATH_TEST, msg)

	tenc := int64(0)
	tdec := int64(0)
	tpro := int64(0)
	tver := int64(0)
	loop := 0

	f := starks.NewStarks(65536 / 16 / 4)

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
		x := rb.GetBlockBytes(f.GetSteps() * 31)

		block := bitcoin.NewBlock()
		block.SetHash(rb.GetRawBytes(0, 80))
		hash := block.GetHashString()

		// log.Printf("Block : %v", x[:80])

		// Start Encryption
		start := time.Now().UnixNano()
		vis := poscipher.CalculateXorWithAddress(addr, x)
		_, y := poscipher.EncryptPoSWithPrimeFieldPreKey(key, vis)
		tenc += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Encryption Time : %v", tenc)

		// Start generating proof
		start = time.Now().UnixNano()
		proof := f.GenerateStarksProofPreKey(hash, vis, y, key)
		tpro += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Generating Proof Time : %v, Merkle Root : %v", tpro, hex.EncodeToString(proof.MerkleRoot))

		// Start verification
		start = time.Now().UnixNano()
		ret := f.VerifyStarksProofPreKey(vis, proof)
		tver += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Verifying Proof Time : %v, %v", tver, ret)
		if !ret {
			log.Panicf("Verification Fail : %v", ret)
		}

		proof_size := f.GetSizeStarksProofPreKey(proof)
		log.Printf("Proof Size : %v", proof_size)

		start = time.Now().UnixNano()
		x_t := poscipher.DecryptPoSWithPrimeFieldPreKey(key, y)
		x_o := poscipher.CalculateXorWithAddress(addr, x_t[:len(x)])
		tdec += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Decryption Time : %v", tdec)

		for i := 0; i < 10; i++ {
			r := rand.Int() % len(x)
			if x_o[r] != x[r] {
				log.Panicf("Decryption Fail : %v", x_o[r])
			}
		}

		key = y
		loop++
		// return
	}
	close(msg)
}

func test_error_1byte_detect_starks(target string) {
	const PATH_TEST = "./blocks_720.json"
	if target == "" {
		target = PATH_TEST
	}
	w := wallet.NewWallet("blocks.json.wallet")
	addr := w.PublicKey
	key := make([]byte, 0, len(addr)*32)
	for i := 0; i < len(addr); i++ {
		t := uint256.NewInt(uint64(addr[i])).Bytes32()
		key = append(key, t[:]...)
	}

	rand.Seed(time.Now().Local().UnixNano())

	msg := make(chan bitcoin.BlockPkt)
	go simulation.LoadBtcData(target, msg)

	step := 65536 / 16 / 4
	f := starks.NewStarks(step)

	det := 0
	inc := 0
	sizex := 0
	sizey := 0
	loop := 0
	bsizes := make([]int, 0)
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
		x := rb.GetBlockBytes(f.GetSteps() * 31)

		block := bitcoin.NewBlock()
		block.SetHash(rb.GetRawBytes(0, 80))
		hash := block.GetHashString()

		// log.Printf("Block : %v", x[:80])
		sizex += len(x)
		bsizes = append(bsizes, len(x))

		// Start Encryption
		vis := poscipher.CalculateXorWithAddress(addr, x)
		_, y := poscipher.EncryptPoSWithPrimeFieldPreKey(key, vis)

		sizey += len(y)

		// Start generating proof
		pos := rand.Intn(len(y))
		y[pos] += 1

		proof := f.GenerateStarksProofPreKey(hash, vis, y, key)

		// Start verification
		ret := f.VerifyStarksProofPreKey(vis, proof)
		if !ret {
			log.Printf("Verification Fail : %v", ret)
		}

		proof_size := f.GetSizeStarksProofPreKey(proof)
		log.Printf("Proof Size : %v", proof_size)

		// x_t := poscipher.DecryptPoSWithPrimeFieldPreKey(key, y)
		// x_o := poscipher.CalculateXorWithAddress(addr, x_t[:len(x)])
		// log.Printf("Decryption Time : %v", tdec)

		log.Printf("Block # %v", loop)
		log.Printf("Error position : %v-%v", pos, len(y))
		si := f.GetStartingPos()
		visu := f.GFP.LoadUint256FromStream31(vis)
		// vosu := f.GFP.LoadUint256FromStream32(y)
		if len(visu) > f.GetSteps() {
			si = si % (len(visu) - f.GetSteps())
		} else {
			si = 0
		}

		log.Printf("Starks position : %v", si)
		included := false
		if pos >= si*32 && pos < (si+step)*32 {
			included = true
			inc += 1
		}
		if !ret {
			det += 1
		}
		log.Printf("Verification : %v-%v", included, ret)

		log.Printf("Verification(%v) : det %v- inc %v", loop+1, det, inc)
		key = y
		loop++
		// break
	}

	log.Printf("Verification : det %v- inc %v", float64(det)/float64(loop), float64(inc)/float64(loop))
	log.Printf("Avg. Size X %v, Probability X %v,", float64(sizex)/float64(loop), float64((31*2048))/float64(sizex)*float64(loop))
	log.Printf("Avg. Size Y %v, Probability Y %v,", float64(sizey)/float64(loop), float64((32*2048))/float64(sizey)*float64(loop))
	log.Printf("Block Size : %v", bsizes)
	close(msg)
}

func float(det int) {
	panic("unimplemented")
}

func test_prime_field() {
	g := galois.NewGFP()

	length := 65536
	xs := make([]*uint256.Int, length)
	ys := make([]*uint256.Int, length)
	for i := 0; i < len(ys); i++ {
		r1 := rand.Int63()
		xs[i] = uint256.NewInt(uint64(r1))
		r2 := rand.Int63()
		ys[i] = uint256.NewInt(uint64(r2))
	}

	tm1 := int64(0)
	start := time.Now().UnixNano()
	for i := 0; i < length/16; i++ {
		for j := 0; j < length; j++ {
			g.Add(xs[j], ys[j])
		}
	}
	end := time.Now().UnixNano()
	tm1 = end - start
	log.Printf("Mul : %v", tm1/1000)

	tm2 := int64(0)
	start = time.Now().UnixNano()
	for i := 0; i < length/16; i++ {
		for j := 0; j < length; j++ {
			g.Add(xs[j], ys[j])
		}
	}
	end = time.Now().UnixNano()
	tm2 = end - start
	log.Printf("Mul : %v", tm2/1000)
}

func test_fft() {
	gf := galois.NewGFP()

	length := 65536
	g := gf.Prime.Clone()
	g.Sub(g, uint256.NewInt(1))
	g.Div(g, uint256.NewInt(uint64(length)))
	g1 := gf.Exp(uint256.NewInt(7), g)

	size, xs := gf.ExtRootUnity(g1, false)
	xs = xs[:size-1]
	// size -= 1
	ys := make([]*uint256.Int, 0, size)

	for j := 0; j < size; j++ {
		r := rand.Uint64() % gf.Prime.Uint64()
		a := uint256.NewInt(r)
		ys = append(ys, a)
	}
	log.Printf("==>xs:%v", xs[:10])
	log.Printf("==>ys:%v", ys[:10])

	lag := false
	if lag {
		start := time.Now().UnixNano()
		os1 := gf.LagrangeInterp(xs, ys)
		end := time.Now().UnixNano()
		log.Printf("LagrangeInterp(%v) : f(x)=%v", (end-start)/1000, os1[:10])
	}

	start := time.Now().UnixNano()
	os2 := gf.IDFT(ys, g1)
	end := time.Now().UnixNano()
	log.Printf("IDFT(%v) : f(x)=%v", (end-start)/1000, len(os2))
	log.Printf("IDFT : %v", os2[:10])

	start = time.Now().UnixNano()
	os3 := gf.DFT(os2, g1)
	end = time.Now().UnixNano()
	log.Printf("DFT(%v) : %v", (end-start)/1000, len(os3))
	log.Printf("DFT : %v", os3[:10])
}

// running with cpu profiling
// eg) go run bcsos.go -cpuprofile=cpu.prof
var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	target := flag.String("bfile", "", "Blocks.Json")
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	log.Printf("Target : %v", *target)

	// test_encypt_2()
	// test_encypt_decrypt()
	// test_fri_prove_low_degree()
	// test_encypt_decrypt_prime()
	// test_starks_prime()
	test_starks_prime_prekey()
	// test_error_1byte_detect_starks(*target)
	// test_prime_field()
	// test_fft()
}
