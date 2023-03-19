package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/junwookheo/bcsos/blockchainsim/simulation"
	"github.com/junwookheo/bcsos/common/bitcoin"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/poscipher"
	"github.com/junwookheo/bcsos/common/wallet"
)

var IV = []byte("1234567812345678")
var TAU = 347

const PATH_TEST = "../blocks_720.json"
const PATH_WALLET = "blocks.json.wallet"

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func flagParse() string {
	ptest := flag.String("test", "ASYM", "AES: Test AES-256, ASYM: Test Asymetric")
	flag.Parse()

	return *ptest
}

func EncryptAES(key []byte, pt []byte) []byte {
	c, err := aes.NewCipher(key)
	CheckError(err)

	out := make([]byte, len(pt))

	for i := 0; i < TAU; i++ {
		stream := cipher.NewCTR(c, IV)
		stream.XORKeyStream(out, pt)
		pt = out
	}

	return out
}

func DecryptAES(key []byte, ct []byte) []byte {
	c, err := aes.NewCipher(key)
	CheckError(err)
	out := make([]byte, len(ct))

	for i := 0; i < TAU; i++ {
		stream := cipher.NewCTR(c, IV)
		stream.XORKeyStream(out, ct)
		ct = out
	}

	return out
}

func CheckError(err error) {
	if err != nil {
		panic(err)
	}
}

func test_aes_cbc() {
	w := wallet.NewWallet(PATH_WALLET)
	key := w.PublicKey[0:32]

	msg := make(chan bitcoin.BlockPkt)
	go simulation.LoadBtcData(PATH_TEST, msg)

	tenc := int64(0)
	tdec := int64(0)
	size := int64(0)
	log.Println("AES-256 Encrypt/Decrypt")
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
		size += int64(len(x))
		// log.Printf("Block : %v", x[:80])

		// Start Encryption
		start := time.Now().UnixNano()
		y := EncryptAES(key, x)
		tenc += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Encryption Time : %v", tenc)
		log.Printf("Enc x:%x", y[0:80])
		// Start Decryption
		start = time.Now().UnixNano()
		x_t := DecryptAES(key, y)
		tdec += (time.Now().UnixNano() - start) / 1000000 // msec
		log.Printf("Decryption Time : %v", tdec)

		log.Printf("Org x:%v", x[0:80])
		log.Printf("New x:%v", x_t[0:80])

		log.Printf("Size : %v", size)
	}
	close(msg)
}

func test_asymm_ppos() {
	w := wallet.NewWallet(PATH_WALLET)
	key := w.PublicKey
	addr := w.PublicKey

	msg := make(chan bitcoin.BlockPkt)
	go simulation.LoadBtcData(PATH_TEST, msg)

	tenc := int64(0)
	tdec := int64(0)
	size := int64(0)
	log.Println("Asymetric Encrypt/Decrypt")

	fcsv, err := os.Create("ppos.csv")
	if err != nil {
		log.Fatalf("failed creating file: %s", err)
	}
	defer fcsv.Close()
	csvwriter := csv.NewWriter(fcsv)
	defer csvwriter.Flush()

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
		size += int64(len(x))
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

		row := [3]string{fmt.Sprintf("%v", tenc), fmt.Sprintf("%v", tdec), fmt.Sprintf("%v", size)}
		csvwriter.Write(row[:])

		log.Printf("Org x:%v", x[0:80])
		log.Printf("New x:%v", x_t[0:80])
		log.Printf("Size : %v", size)

		key = y
	}
	close(msg)
}
func main() {
	test := flagParse()
	if test == "AES" {
		test_aes_cbc()
	} else {
		test_asymm_ppos()
	}
}
