package simulation

import (
	"bufio"
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/junwookheo/bcsos/common/bitcoin"
	"github.com/junwookheo/bcsos/common/dbagent"
)

// const PATH_TEST = "../iotdata/IoT_normal_fridge_1.log"
const PATH_TEST = "../../blocks.json"
const DB_PATH_TEST = "../bc_dummy.db"

const SIZE_BH = 80

type btcblheader struct {
	Version    uint32
	PreHash    [32]byte
	MerkelRoot [32]byte
	Timestamp  uint32
	Difficulty uint32
	Nonce      uint32
}

func TestLoadBlockFromJson(t *testing.T) {
	f, err := os.Open(PATH_TEST)
	if err != nil {
		log.Printf("Open error : %v", err)
		return
	}

	scanner := bufio.NewScanner(f)
	const maxCapacity = 16 * 1025 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		var rec map[string]interface{}
		err := json.Unmarshal([]byte(scanner.Text()), &rec)
		if err != nil {
			panic(err)
		}

		// var b map[string]interface{}
		b, ok := rec["data"].(map[string]interface{})
		if !ok {
			log.Panicln("Read block type error")
		}

		for key := range b {
			hash := b[key].(map[string]interface{})["decoded_raw_block"].(map[string]interface{})["hash"]
			// time := b[key].(map[string]interface{})["decoded_raw_block"].(map[string]interface{})["time"]
			raw := b[key].(map[string]interface{})["raw_block"].(string)
			bh := btcblheader{}
			rawb, _ := hex.DecodeString(raw)

			// Parse Block Header
			err = binary.Read(bytes.NewBuffer(rawb), binary.LittleEndian, &bh)
			if err != nil {
				log.Panicln(err)
			}
			// Reverse
			reverse := func(arr []byte) {
				n := len(arr)
				for i := 0; i < n/2; i++ {
					arr[i], arr[n-1-i] = arr[n-1-i], arr[i]
				}
			}

			reverse(bh.PreHash[:])
			reverse(bh.MerkelRoot[:])
			hash2 := sha256.Sum256(rawb[0:80])
			hash2 = sha256.Sum256(hash2[:])
			reverse(hash2[:])
			log.Printf("Block : %v", key)
			log.Printf("Block : %v", hash)
			log.Printf("Block : %v", hex.EncodeToString(hash2[:]))
			log.Printf("Block : %x-%v-%v", bh.Version, bh.Timestamp, hex.EncodeToString(bh.PreHash[:]))
			log.Printf("Block : %v-%v-%v", hex.EncodeToString(bh.MerkelRoot[:]), bh.Difficulty, bh.Nonce)

			// Parse Transactions
			b := byte(0)
			err = binary.Read(bytes.NewBuffer(rawb[80:]), binary.LittleEndian, &b)
			if err != nil {
				log.Panicln(err)
			}
			log.Printf("Count : %x", b)

		}

	}
}

func TestLoadTransactionFromJson(t *testing.T) {
	f, err := os.Open(PATH_TEST)
	if err != nil {
		log.Printf("Open error : %v", err)
		return
	}

	scanner := bufio.NewScanner(f)
	const maxCapacity = 16 * 1025 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		var rec map[string]interface{}
		err := json.Unmarshal([]byte(scanner.Text()), &rec)
		if err != nil {
			panic(err)
		}

		// var b map[string]interface{}
		b, ok := rec["data"].(map[string]interface{})
		if !ok {
			log.Panicln("Read block type error")
		}

		for key := range b {
			raw := b[key].(map[string]interface{})["raw_block"].(string)
			rawb, _ := hex.DecodeString(raw)

			// Parse Transactions
			pos := SIZE_BH
			b := byte(0)
			err = binary.Read(bytes.NewBuffer(rawb[pos:]), binary.LittleEndian, &b)
			if err != nil {
				log.Panicln(err)
			}
			log.Printf("Count : %x", b)
			pos += 1

			get_tx_count := func(buf interface{}) {
				err = binary.Read(bytes.NewBuffer(rawb[pos:]), binary.LittleEndian, buf)
				if err != nil {
					log.Panicln(err)
				}
			}
			numtx := uint64(0)
			if b == 0xfd {
				buf_s := uint16(0)
				get_tx_count(&buf_s)
				numtx = uint64(buf_s)
				pos += 2
			} else if b == 0xfe {
				buf_i := uint32(0)
				get_tx_count(buf_i)
				numtx = uint64(buf_i)
				pos += 4
			} else if b == 0xff {
				buf_l := uint64(0)
				get_tx_count(buf_l)
				numtx = uint64(buf_l)
				pos += 8
			} else {
				buf_b := uint8(0)
				get_tx_count(buf_b)
				numtx = uint64(buf_b)
				pos += 1
			}

			log.Printf("Transaction count : %v", numtx)
			version := uint32(0)
			err = binary.Read(bytes.NewBuffer(rawb[pos:]), binary.LittleEndian, &version)
			if err != nil {
				log.Panicln(err)
			}
			log.Printf("Version : %x", version)
			pos += 4

			numipn := uint64(0)
			if b == 0xfd {
				buf_s := uint16(0)
				get_tx_count(&buf_s)
				numipn = uint64(buf_s)
				pos += 2
			} else if b == 0xfe {
				buf_i := uint32(0)
				get_tx_count(buf_i)
				numipn = uint64(buf_i)
				pos += 4
			} else if b == 0xff {
				buf_l := uint64(0)
				get_tx_count(buf_l)
				numipn = uint64(buf_l)
				pos += 8
			} else {
				buf_b := uint8(0)
				get_tx_count(buf_b)
				numipn = uint64(buf_b)
				pos += 1
			}
			log.Printf("Input count : %d", numipn)

		}

	}
}

func TestLoadBlockFromJson2(t *testing.T) {
	f, err := os.Open(PATH_TEST)
	if err != nil {
		log.Printf("Open error : %v", err)
		return
	}

	scanner := bufio.NewScanner(f)
	const maxCapacity = 16 * 1025 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		var rec map[string]interface{}
		err := json.Unmarshal([]byte(scanner.Text()), &rec)
		if err != nil {
			panic(err)
		}

		// var b map[string]interface{}
		b, ok := rec["data"].(map[string]interface{})
		if !ok {
			log.Panicln("Read block type error")
		}

		for key := range b {
			raw := b[key].(map[string]interface{})["raw_block"].(string)
			rb := bitcoin.NewRawBlock(raw)
			version := rb.ReadUint32()
			prehash := rb.ReverseBuf(rb.ReadBytes(32))
			merkleroot := rb.ReverseBuf(rb.ReadBytes(32))
			timestamp := rb.ReadUint32()
			difficulty := rb.ReadUint32()
			nonce := rb.ReadUint32()
			txcount := rb.ReadVariant()
			log.Printf("Version : %x", version)
			log.Printf("Pev Hash : %v", hex.EncodeToString(prehash))
			log.Printf("Merkle Root : %v", hex.EncodeToString(merkleroot))
			log.Printf("Timestamp : %d", timestamp)
			log.Printf("Dificulty : %d", difficulty)
			log.Printf("Nonce : %d", nonce)
			log.Printf("Tx Count : %d", txcount)

			for i := 0; i < int(txcount); i++ {
				log.Printf("Tx(%v) =================================", i)
				version := rb.ReadUint32()
				witflag := rb.ReadOptional(2)
				if witflag != nil && witflag[0] == 0x00 && witflag[1] == 0x01 {
					rb.IncreasePos(2)
				}
				incount := rb.ReadVariant()
				log.Printf("Version : %x", version)
				log.Printf("Witness Flag : %v", witflag)
				log.Printf("Input count : %v", incount)

				for j := 0; j < int(incount); j++ {
					preout := rb.ReadBytes(32)
					locktime := rb.ReadUint32()
					scrlen := rb.ReadVariant()

					log.Printf("Previous Output : %v", preout)
					log.Printf("Lock Time : %v", locktime)
					log.Printf("Script Length : %v", scrlen)

					if scrlen > 0 {
						script := rb.ReadBytes(uint32(scrlen))
						log.Printf("Unlock Script : %x", script[0])
					}

					seqnumber := rb.ReadUint32()
					log.Printf("Sequence Number : %v", seqnumber)
				}

				outlen := rb.ReadVariant()
				log.Printf("Output count : %v", outlen)
				for k := 0; k < int(outlen); k++ {
					txvalue := rb.ReadUint64()
					scrlen := rb.ReadVariant()
					script := rb.ReadBytes(uint32(scrlen))
					log.Printf("%v Transaction Value : %v", k, txvalue)
					log.Printf("%v scrlen : %v", k, scrlen)
					log.Printf("%v Lock Script : %x", k, script[0])

				}

				if witflag != nil && witflag[0] == 0x00 && witflag[1] == 0x01 {
					for j := 0; j < int(incount); j++ {
						witcnt := rb.ReadVariant()
						log.Printf("Witness count : %v", witcnt)
						for l := 0; l < int(witcnt); l++ {
							witlen := rb.ReadVariant()
							log.Printf("Witness length : %v", witlen)
							witness := rb.ReadBytes(uint32(witlen))
							log.Printf("Witness : %v", witness)
						}
					}
				}

				locktime := rb.ReadUint32()
				log.Printf("Lock Time : %v", locktime)

			}
			return
		}
	}
}

func TestDBAgent(t *testing.T) {
	dba := dbagent.NewDBAgent(DB_PATH_TEST)
	dba.ShowAllObjets()
	dba.GetLatestBlockHash()
	status := dba.GetDBStatus()
	log.Printf("DB Status : %v", status)
}
