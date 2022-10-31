package dbagent

import (
	"crypto/sha256"
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
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/poscipher"
	"github.com/junwookheo/bcsos/common/serial"
	"github.com/junwookheo/bcsos/common/wallet"
)

type btcBlock struct {
	timestamp int64
	height    int
	hashprev  string
	hash      string // hash of plain data
	hashenc   string // hash of encrypted data
	hashkey   string // hash of key data of the previous encrypted block used when enctypting
	encblock  []byte // encrypted block to encrypt the next block
}

type btcDBStatus struct {
	Timestamp      time.Time
	ID             int
	TotalBlocks    int
	TotalSize      int
	TimeEncryptAcc int
	TimeDecryptAcc int
	NumPoSAcc      int
	TimePosAcc     int
}

type btcdbagent struct {
	db        *sql.DB
	sclass    int
	dbstatus  btcDBStatus
	dirpath   string
	lastblock btcBlock
	mutex     sync.Mutex
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

func (a *btcdbagent) addBtcBlocktoList(b *btcBlock) int64 {
	// if id := a.GetObject(obj); id != 0 {
	// 	// log.Printf("Replicatoin exists : %v - %v", id, obj)
	// 	return id
	// }

	id := func() int64 {
		a.mutex.Lock()
		defer a.mutex.Unlock()
		st, err := a.db.Prepare("INSERT INTO btcblocklist (timestamp, height, hashprev, hash, hashenc, hashkey) VALUES (?, ?, ?, ?, ?, ?)")
		if err != nil {
			log.Printf("Prepare update BTC Block object error : %v", err)

			return -1
		}
		defer st.Close()

		rst, err := st.Exec(b.timestamp, b.height, b.hashprev, b.hash, b.hashenc, b.hashkey)
		if err != nil {
			log.Panicf("Exec adding object error : %v", err)
			return -1
		}

		id, _ := rst.LastInsertId()
		return id
	}()

	// Update db status after adding object
	// a.updateAddDBStatus(id)
	return id
}

func (a *btcdbagent) getEncryptKeyforGenesis() []byte {
	w := wallet.NewWallet(config.WALLET_PATH)
	return w.PublicKey //GetAddress()
}

func (a *btcdbagent) encryptPoSWithVariableLength(key, s []byte) (string, []byte) {
	start := time.Now().UnixNano()

	hash, d := poscipher.EncryptPoSWithVariableLength2(key, s)
	gap := int(time.Now().UnixNano() - start)
	{
		a.mutex.Lock()
		defer a.mutex.Unlock()
		status := &a.dbstatus
		status.TimeEncryptAcc += gap
	}

	return hash, d
}

func (a *btcdbagent) decryptPoSWithVariableLength(key, s []byte) []byte {
	start := time.Now().UnixNano()
	d := poscipher.DecryptPoSWithVariableLength(key, s)

	gap := int(time.Now().UnixNano() - start)
	{
		a.mutex.Lock()
		defer a.mutex.Unlock()
		status := &a.dbstatus
		status.TimeEncryptAcc += gap
	}

	return d
}

func (a *btcdbagent) getHashforPoSKey(key []byte, ls int) string {
	return poscipher.GetHashforPoSKey(key, ls)
}

func (a *btcdbagent) AddNewBlock(ib interface{}) int64 {
	sb, ok := ib.(*bitcoin.BlockPkt)
	if !ok {
		log.Panicf("Type mismatch : %v", ok)
		return -1
	}

	block := bitcoin.NewBlock()
	rb := bitcoin.NewRawBlock(sb.Block)
	hashprev := ""
	if a.lastblock.height > -1 { // if it is not the first block
		_ = rb.ReadUint32()
		hbuf := rb.ReverseBuf(rb.ReadBytes(32))
		hashprev = hex.EncodeToString(hbuf)
	}

	block.SetHash(rb.GetRawBytes(0, 80))
	hash := block.GetHashString()
	s := rb.GetBlockBytes()
	size := len(s)

	// Enctypting a new block
	key := a.lastblock.encblock
	hashenc, encblock := a.encryptPoSWithVariableLength(key, s)
	a.lastblock.timestamp = sb.Timestamp
	a.lastblock.hash = hash
	a.lastblock.height += 1
	a.lastblock.hashprev = hashprev
	a.lastblock.hashenc = hashenc
	a.lastblock.encblock = encblock
	a.lastblock.hashkey = a.getHashforPoSKey(key, size)

	// Add a new btc block to list in db
	a.addBtcBlocktoList(&a.lastblock)

	err := ioutil.WriteFile(filepath.Join(a.dirpath, hash), encblock, 0777)
	if err != nil {
		log.Panicf("Wrinting block err : %v", err)
		return -1
	}

	{
		a.mutex.Lock()
		defer a.mutex.Unlock()
		status := &a.dbstatus
		status.TotalBlocks += 1
		status.TotalSize += size
	}

	return int64(a.lastblock.height)
}

func (a *btcdbagent) initLastBlock() {
	a.lastblock.timestamp = time.Now().UnixNano()
	a.lastblock.encblock = a.getEncryptKeyforGenesis() // encblock is data to encrypt the next block
	a.lastblock.hash = poscipher.GetHashString(a.lastblock.encblock)
	a.lastblock.hashenc = a.lastblock.hash // This block does not need to be encrypted
	a.lastblock.hashkey = ""               // There is no key
	a.lastblock.height = -1                // This is key for the first block(B0)
	a.lastblock.hashprev = ""              // No previous block

	// TODO : Add to it to db
	a.addBtcBlocktoList(&a.lastblock)
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

func (a *btcdbagent) getLatestDBStatus(status *btcDBStatus) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	rows, err := a.db.Query(`SELECT *  FROM dbstatus WHERE id = (SELECT MAX(id)  FROM dbstatus);`)
	if err != nil {
		log.Printf("Show latest db status Error : %v", err)
		return false
	}

	defer rows.Close()

	for rows.Next() {
		rows.Scan(&status.ID, &status.Timestamp, &status.TotalBlocks, &status.TotalSize, &status.TimeEncryptAcc, &status.TimeDecryptAcc, &status.NumPoSAcc, &status.TimePosAcc)

		return true
	}

	return false
}

func (a *btcdbagent) updateDBStatus() {
	getHash := func(status btcDBStatus) string {
		status.ID = 0
		status.Timestamp = time.Time{}
		byte_status := sha256.Sum256(serial.Serialize(status))
		return hex.EncodeToString(byte_status[:])
	}

	last_hash := getHash(a.dbstatus)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	for {
		<-ticker.C
		func() {
			a.mutex.Lock()
			defer a.mutex.Unlock()
			status := &a.dbstatus
			hash_status := getHash(*status)
			if last_hash == hash_status {
				return
			}

			last_hash = hash_status

			st, err := a.db.Prepare(`INSERT INTO dbstatus (timestamp, totalblocks, totalsize, timeencryptacc, timedecryptacc, numposacc, timeposacc) 
					VALUES ( datetime('now'), ?, ?, ?, ?, ?, ?)`)
			if err != nil {
				log.Printf("Prepare adding dbstatus error : %v", err)
				return
			}
			defer st.Close()

			rst, err := st.Exec(status.TotalBlocks, status.TotalSize, status.TimeEncryptAcc, status.TimeDecryptAcc, status.NumPoSAcc, status.TimePosAcc)
			if err != nil {
				log.Panicf("Exec adding dbstatus error : %v", err)
				return
			}

			id, _ := rst.LastInsertId()
			status.ID = int(id)
			// log.Printf("Update dbstatus : %v", *status)
		}()
	}
}

func newDBBtcSqlite(path string) DBAgent {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Panicf("Open sqlite db error : %v", err)
	}

	create_objtlb := `CREATE TABLE IF NOT EXISTS btcblocklist (
		id      	INTEGER  PRIMARY KEY AUTOINCREMENT,
		timestamp	INTEGER,
		height		INTEGER,
		hashprev   	TEXT,
		hash    	TEXT,
		hashenc		TEXT,
		hashkey		TEXT
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
		totalsize			INTEGER,
		timeencryptacc		INTEGER,
		timedecryptacc		INTEGER,
		numposacc			INTEGER,
		timeposacc			INTEGER
	);`

	st, err = db.Prepare(create_statustlb)
	if err != nil {
		log.Panicf("create_statustlb error %v", err)
	}
	defer st.Close()

	st.Exec()

	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()

	dba := btcdbagent{db: db, sclass: local.SC, dbstatus: btcDBStatus{Timestamp: time.Now()}, dirpath: "", lastblock: btcBlock{}, mutex: sync.Mutex{}}
	dba.getLatestDBStatus(&dba.dbstatus)
	go dba.updateDBStatus()

	dba.initLastBlock()

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
