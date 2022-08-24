package mining

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/serial"
	"github.com/stretchr/testify/assert"
)

const DB_PATH_TEST = "../db_nodes/7001.db"

//const DB_PATH_TEST = "../../blockchainsim/bc_sim.db"

func TestBlockchainConsistency(t *testing.T) {
	db, err := sql.Open("sqlite3", DB_PATH_TEST)
	if err != nil {
		log.Panicf("Open sqlite db error : %v", err)
	}

	rows, err := db.Query(`SELECT data FROM bcobjects WHERE type = "blockheader";`)
	if err != nil {
		log.Printf("Show latest db status Error : %v", err)
		return
	}

	defer rows.Close()

	var data []byte
	bh := blockchain.BlockHeader{}
	i := 0
	prev := ""

	for rows.Next() {
		err := rows.Scan(&data)
		if err != nil {
			break
		}

		serial.Deserialize(data, &bh)
		log.Printf("add block : %v, %v", i, bh.Height)
		log.Printf("%v - %v", hex.EncodeToString(bh.Hash), hex.EncodeToString(bh.PrvHash))
		assert.Equal(t, prev, hex.EncodeToString(bh.PrvHash))
		prev = hex.EncodeToString(bh.Hash)

		i++
		// if i == 243 {
		// 	return
		// }
	}
}

func TestBlockchainPoStorage(t *testing.T) {
	db, err := sql.Open("sqlite3", DB_PATH_TEST)
	if err != nil {
		log.Panicf("Open sqlite db error : %v", err)
	}

	rows1, err1 := db.Query(`SELECT data FROM bcobjects WHERE type = "blockheader" ORDER BY ROWID DESC LIMIT 1;`)
	if err1 != nil {
		log.Printf("Show latest db status Error : %v", err)
		return
	}

	defer rows1.Close()

	var data []byte
	bh := blockchain.BlockHeader{}
	i := 0

	tth := int64(0)
	for rows1.Next() {
		err := rows1.Scan(&data)
		if err != nil {
			break
		}

		serial.Deserialize(data, &bh)
		tth = bh.Timestamp - int64(config.TSCX[0]*float32(1e9))
		log.Printf("add block : %v, %v, %v", i, bh.Height, time.Unix((tth/1000)/1e6, ((tth/1000)%1e6)*1e3))
		log.Printf("%v - %v", hex.EncodeToString(bh.Hash), hex.EncodeToString(bh.PrvHash))
		i++
	}

	addr := "000dc82e66b0465fe0d9021bffe6e4092969526e8ee50f6cc7355feb81ebe699"
	baddr, _ := hex.DecodeString(addr)
	ridx := sha256.Sum256(append(bh.Hash, baddr...))
	log.Printf("ridx : %v", hex.EncodeToString(ridx[:]))
	match := hex.EncodeToString(ridx[:])
	log.Printf("ridx : %v", match[len(match)-3:])

	query := fmt.Sprintf(`SELECT hash, timestamp, data FROM bcobjects WHERE type = "transaction" and hash LIKE "%%%s";`, match[len(match)-2:])
	// query := `SELECT hash, timestamp, data FROM bcobjects WHERE type = "transaction" and hash LIKE "%%4947";`
	rows2, err2 := db.Query(query)
	if err2 != nil {
		log.Printf("Show latest db status Error : %v", err)
		return
	}

	defer rows2.Close()

	i = 0
	tr := blockchain.Transaction{}
	var hash string
	var ts int64
	var hashes [][]byte

	for rows2.Next() {
		err := rows2.Scan(&hash, &ts, &data)
		if err != nil {
			break
		}
		serial.Deserialize(data, &tr)
		if tth < tr.Timestamp {
			log.Printf("match %v : %v - %v", i, time.Unix((tth/1000)/1e6, ((tth/1000)%1e6)*1e3), hex.EncodeToString(tr.Hash))
			tr.Timestamp = bh.Timestamp
			hashes = append(hashes, tr.GetHash())
			i++
		}
	}
	log.Printf("Hashes : %v", hex.EncodeToString(blockchain.CalMerkleRootHash(hashes)))
}
