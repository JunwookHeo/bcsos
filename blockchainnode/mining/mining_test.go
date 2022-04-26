package mining

import (
	"database/sql"
	"encoding/hex"
	"log"
	"testing"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/serial"
)

const DB_PATH_TEST = "../db_nodes/7001.db"

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
	cm := NewChainMgr()

	for rows.Next() {
		err := rows.Scan(&data)
		if err != nil {
			break
		}

		serial.Deserialize(data, &bh)
		log.Printf("add block : %v, %v", i, bh.Height)
		log.Printf("%v - %v", hex.EncodeToString(bh.Hash), hex.EncodeToString(bh.PrvHash))
		cm.AddedNewBlock(&bh)
		i++
		if i == 243 {
			return
		}
	}
}
