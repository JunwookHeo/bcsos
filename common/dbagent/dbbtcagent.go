package dbagent

import (
	"database/sql"
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/junwookheo/bcsos/blockchainnode/network"
	"github.com/junwookheo/bcsos/common/bitcoin"
	"github.com/junwookheo/bcsos/common/blockchain"
)

type btcdbagent struct {
	db       *sql.DB
	sclass   int
	dbstatus DBStatus
	dirpath  string
	mutex    sync.Mutex
}

func (a *btcdbagent) getLatestDBStatus(status *DBStatus) bool {
	// TODO:
	return true
}

func (a *btcdbagent) updateDBStatus() {
	// TODO:
}

func (a *btcdbagent) Close() {
	a.db.Close()
}

func (a *btcdbagent) GetLatestBlockHash() (string, int) {
	log.Panicln("GetLatestBlockHash")
	return "", -1
}

func (a *btcdbagent) RemoveObject(hash string) bool {
	log.Panicln("RemoveObject")
	return false
}

func (a *btcdbagent) AddBlockHeader(hash string, h *blockchain.BlockHeader) int64 {
	log.Panicln("AddBlockHeader")
	return -1
}

func (a *btcdbagent) GetBlockHeader(hash string, h *blockchain.BlockHeader) int64 {
	log.Panicln("GetBlockHeader")
	return -1
}

func (a *btcdbagent) AddTransaction(t *blockchain.Transaction) int64 {
	log.Panicln("AddTransaction")
	return -1
}

func (a *btcdbagent) GetTransaction(hash string, t *blockchain.Transaction) int64 {
	log.Panicln("GetTransaction")
	return -1
}

func encryptXorWithFixedLength(key, s []byte) []byte {
	if len(key) < len(s) {
		d := make([]byte, len(s))
		m := len(key)
		for i := 0; i < len(s); i++ {
			d[i] = key[i%m] ^ s[i]
		}
		return d
	} else if len(key) > len(s) {
		d := make([]byte, len(key))
		m := len(s)
		for i := 0; i < len(key); i++ {
			d[i] = key[i] ^ s[i%m]
		}
		return d
	} else {
		d := make([]byte, len(s))
		for i := 0; i < len(s); i++ {
			d[i] = key[i] ^ s[i]
		}
		return d
	}
}

func (a *btcdbagent) AddNewBlock(ib interface{}) int64 {
	sb, ok := ib.(string)
	if !ok {
		log.Panicf("Type mismatch : %v", ok)
		return -1
	}

	block := bitcoin.NewBlock()
	rb := bitcoin.NewRawBlock(sb)
	block.SetHash(rb.GetRawBytes(0, 80))
	hash := block.GetHashString()

	// TODO : Encryptions

	buf, err := hex.DecodeString(sb)
	if err != nil {
		log.Panicf("Decoding string block err : %v", err)
	}

	err = ioutil.WriteFile(filepath.Join(a.dirpath, hash), buf, 0777)
	if err != nil {
		log.Panicf("Wrinting block err : %v", err)
	}

	buf2, err := ioutil.ReadFile(filepath.Join(a.dirpath, hash))
	if err != nil {
		log.Panicf("Wrinting block err : %v", err)
	}
	sb2 := hex.EncodeToString(buf2)
	log.Printf("encode(%v) : %v", len(sb2), sb2[:80])

	return -1
}

func (a *btcdbagent) GetBlock(hash string, b *blockchain.Block) int64 {
	log.Panicln("GetBlock")
	return -1
}

func (a *btcdbagent) ShowAllObjets() bool {
	log.Panicln("ShowAllObjets")
	return false
}

func (a *btcdbagent) GetDBDataSize() uint64 {
	log.Panicln("GetDBDataSize")
	return 0
}

func (a *btcdbagent) GetDBStatus() *DBStatus {
	log.Panicln("GetDBStatus")
	return nil
}

func (a *btcdbagent) GetTransactionwithUniform(num int, hashes *[]RemoverbleObj) bool {
	log.Panicln("GetTransactionwithUniform")
	return false
}

func (a *btcdbagent) GetTransactionwithExponential(num int, hashes *[]RemoverbleObj) bool {
	log.Panicln("GetTransactionwithExponential")
	return false
}

func (a *btcdbagent) DeleteNoAccedObjects() {
	// No Need this function for PoS
}

func (a *btcdbagent) UpdateDBNetworkQuery(fromqc int, toqc int, totalqc int) {
	log.Panicln("UpdateDBNetworkQuery")
}

func (a *btcdbagent) UpdateDBNetworkDelay(addtime int, hop int) {
	log.Panicln("UpdateDBNetworkDelay")
}

func newDBBtcSqlite(path string) DBAgent {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Panicf("Open sqlite db error : %v", err)
	}

	create_objtlb := `CREATE TABLE IF NOT EXISTS btcblock (
		id      	INTEGER  PRIMARY KEY AUTOINCREMENT,
		height		INTEGER,
		hash    	TEXT,
		enchash		TEXT
	);`

	st, err := db.Prepare(create_objtlb)
	if err != nil {
		log.Panicf("create_objtlb error %v", err)
	}
	defer st.Close()

	st.Exec()

	// totalquery : query objects including local storage
	// queryfrom : the number of received queries to get deleted transactions
	// queryto : the number of send queries to get deleted transactions
	create_statustlb := `CREATE TABLE IF NOT EXISTS dbstatus (
		id      			INTEGER  PRIMARY KEY AUTOINCREMENT,
		timestamp			DATETIME,
		totalblocks			INTEGER,
		totaltransactions 	INTERGER,
		headers				INTEGER,
		blocks 				INTEGER,
		transactions    	INTEGER,
		size				INTEGER,
		totalquery			INTEGER,
		queryfrom			INTEGER,
		queryto				INTEGER,
		totaldelay			INTEGER,
		hop0				INTEGER,
		hop1				INTEGER,
		hop2				INTEGER,
		hop3				INTEGER
	);`

	st, err = db.Prepare(create_statustlb)
	if err != nil {
		log.Panicf("create_statustlb error %v", err)
	}
	defer st.Close()

	st.Exec()

	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()

	dba := btcdbagent{db: db, sclass: local.SC, dbstatus: DBStatus{Timestamp: time.Now()}, dirpath: "", mutex: sync.Mutex{}}
	dba.getLatestDBStatus(&dba.dbstatus)
	go dba.updateDBStatus()

	dba.dirpath = path + ".blocks"
	err = os.RemoveAll(dba.dirpath)
	if err != nil {
		log.Panicf("Error Remove Dir : %v", err)
	}

	err = os.MkdirAll(dba.dirpath, 0777)
	if err != nil {
		log.Panicf("Erro Make Dir %v: %v", path, err)
	}

	return &dba
}
