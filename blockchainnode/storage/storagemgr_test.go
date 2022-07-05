package storage

import (
	"database/sql"
	"encoding/hex"
	"log"
	"testing"

	"github.com/junwookheo/bcsos/blockchainnode/network"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/serial"
	"github.com/stretchr/testify/assert"
)

const DB_PATH_TEST = "../db_nodes/7002.db"

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
	prv := ""
	i := 0

	for rows.Next() {
		err := rows.Scan(&data)
		if err != nil {
			break
		}

		serial.Deserialize(data, &bh)
		log.Printf("cur : %v", hex.EncodeToString(bh.Hash))
		log.Printf("pre : %v", hex.EncodeToString(bh.PrvHash))

		assert.Equal(t, prv, hex.EncodeToString(bh.PrvHash))
		prv = hex.EncodeToString(bh.Hash)
		log.Printf("%v", i)
		i++
	}
}

func TestProofStorage(t *testing.T) {
	sm := StorageMgrInst("../db_nodes/7001.db")
	req := dtype.ReqPoStorage{}
	req.Hash = "3feb6dc6e7fa909ac617fe1b98ff52860b25f0799e5bcb31118182082cc9d6e4"
	req.Timestamp = 1656640535013592000

	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()
	local.Hash = "3feb6dc6e7fa909ac617fe1b98ff52860b25f0799e5bcb31118182082cc9d6e6"

	proof := sm.ProofStorageProc(&req, local)

	log.Printf("Proof : %v", proof)
}
