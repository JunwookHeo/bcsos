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

	"github.com/junwookheo/bcsos/common/dbagent"
)

// const PATH_TEST = "../iotdata/IoT_normal_fridge_1.log"
const PATH_TEST = "../../blocks.json"
const DB_PATH_TEST = "../bc_dummy.db"

const SIZE_BH = 80

type bitcoinheader struct {
	Version    uint32
	PreHash    [32]byte
	MerkelRoot [32]byte
	Timestamp  uint32
	Difficulty uint32
	Nonce      uint32
}

func TestLoadFromJson(t *testing.T) {
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
			bh := bitcoinheader{}
			rawb, _ := hex.DecodeString(raw)
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
