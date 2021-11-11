package dbagent

import (
	"database/sql"
	"encoding/hex"
	"log"
	"math/rand"
	"time"
	"unsafe"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/serial"
	_ "github.com/mattn/go-sqlite3"
)

type dbagent struct {
	db *sql.DB
}

func (a *dbagent) Close() {
	a.db.Close()
}

func (a *dbagent) GetLatestBlockHash() string {
	var id int = 0
	var dtype, hash string = "", ""
	var ts time.Time
	rows, err := a.db.Query(`SELECT id, type, hash, timestamp FROM bcobjects WHERE id = (SELECT MAX(id) FROM bcobjects WHERE type = 'block');`)

	if err != nil {
		log.Printf("Show latest objects Error : %v", err)
		return hash
	}

	defer rows.Close()

	for rows.Next() {
		rows.Scan(&id, &dtype, &ts, &hash)
		log.Printf("latest block : %d, %s, %v, %s", id, dtype, ts, hash)
	}

	return hash
}

func (a *dbagent) RemoveObject(hash string) bool {
	// Before removing data, update status first.
	a.updateRemoveDBStatus(hash)

	st, err := a.db.Prepare("DELETE FROM bcobjects WHERE hash=?")
	if err != nil {
		log.Printf("Preparing removig object error : %v", err)
		return false
	}
	rst, err := st.Exec(hash)
	if err != nil {
		log.Printf("Exec removing object error : %v", err)
	}
	cnt, _ := rst.RowsAffected()
	return cnt > 0
}

func (a *dbagent) GetObject(obj *StorageObj) int64 {
	var data []byte
	var id int64 = 0
	switch err := a.db.QueryRow("SELECT id, type, hash, timestamp, data FROM bcobjects WHERE hash=?", obj.Hash).Scan(&id, &obj.Type, &obj.Hash, &obj.Timestamp, &data); err {
	case sql.ErrNoRows:
		log.Printf("Object Not found : %v", err)
	case nil:
		serial.Deserialize(data, obj.Data)
		return id
	default:
		log.Printf("Get object error : %v", err)
	}

	return id
}

func (a *dbagent) AddObject(obj *StorageObj) int64 {
	if a.getObjectCount(obj.Hash) > 0 {
		log.Panicf("Replicatoin exists : %v", obj)
		return -1
	}

	st, err := a.db.Prepare("INSERT INTO bcobjects (type, hash, timestamp, data) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Printf("Prepare adding object error : %v", err)
		return -1
	}
	data := serial.Serialize(obj.Data)
	rst, err := st.Exec(obj.Type, obj.Hash, obj.Timestamp, data)
	if err != nil {
		log.Panicf("Exec adding object error : %v", err)
		return -1
	}

	id, _ := rst.LastInsertId()

	// Update db status after adding object
	a.updateAddDBStatus(id)
	return id
}

func (a *dbagent) GetTransactionwithRandom() []string {
	selhashes := []string{}
	// Randomly select 10 blocks in the ledger
	rows, err := a.db.Query(`SELECT type, hash, timestamp, data FROM bcobjects WHERE type = 'block' ORDER BY RANDOM() LIMIT 10;`)
	if err != nil {
		log.Printf("Object Not found : %v", err)
		return selhashes
	}

	defer rows.Close()

	// Select a transaction in each block
	for rows.Next() {
		var data []byte
		hashes := []string{}
		obj := StorageObj{"block", "", 0, &hashes}
		rows.Scan(&obj.Type, &obj.Hash, &obj.Timestamp, &data)
		serial.Deserialize(data, obj.Data)
		//log.Printf("selected item : %s, %s", dtype, hash)
		num := rand.Intn(len(hashes)-1) + 1 // because the first is a hash of header
		selhashes = append(selhashes, hashes[num])
	}

	//log.Printf("hash : %v", selhashes)
	return selhashes
}

func (a *dbagent) GetTransactionwithTimeWeight() []string {
	selhashes := []string{}
	// Randomly select 10 blocks in the ledger
	rows, err := a.db.Query(`SELECT type, hash, timestamp, data FROM bcobjects WHERE type = 'block' ORDER BY RANDOM() LIMIT 10;`)
	if err != nil {
		log.Printf("Object Not found : %v", err)
		return selhashes
	}

	defer rows.Close()

	// Select a transaction in each block
	for rows.Next() {
		var data []byte
		hashes := []string{}
		obj := StorageObj{"block", "", 0, &hashes}
		rows.Scan(&obj.Type, &obj.Hash, &obj.Timestamp, &data)
		serial.Deserialize(data, obj.Data)
		//log.Printf("selected item : %s, %s", dtype, hash)
		num := rand.Intn(len(hashes)-1) + 1 // because the first is a hash of header
		selhashes = append(selhashes, hashes[num])
	}

	//log.Printf("hash : %v", selhashes)
	return selhashes
}

func (a *dbagent) getObjectCount(hash string) int {
	var count int = 0

	err := a.db.QueryRow("SELECT COUNT(*) FROM bcobjects WHERE hash=?", hash).Scan(&count)
	switch {
	case err != nil:
		log.Fatal(err)
	default:
		log.Printf("Number of rows are %d", count)
	}
	return count
}

func (a *dbagent) GetBlockHeader(hash string, h *blockchain.BlockHeader) int64 {
	obj := StorageObj{"blockheader", hash, h.Timestamp, h}
	return a.GetObject(&obj)
}

func (a *dbagent) AddBlockHeader(hash string, h *blockchain.BlockHeader) int64 {
	obj := StorageObj{"blockheader", hash, h.Timestamp, h}
	return a.AddObject(&obj)
}

func (a *dbagent) GetTransaction(hash string, t *blockchain.Transaction) int64 {
	obj := StorageObj{"transaction", hash, t.Timestamp, t}
	return a.GetObject(&obj)
}

func (a *dbagent) AddTransaction(t *blockchain.Transaction) int64 {
	obj := StorageObj{"transaction", hex.EncodeToString(t.Hash), t.Timestamp, t}
	return a.AddObject(&obj)
}

func (a *dbagent) GetBlock(hash string, b *blockchain.Block) int64 {
	hashes := []string{}
	obj := StorageObj{"block", hash, b.Header.Timestamp, &hashes}
	a.GetObject(&obj)

	h := blockchain.BlockHeader{}
	a.GetBlockHeader(hashes[0], &h)
	b.Header = h

	for i := 1; i < len(hashes); i++ {
		tr := blockchain.Transaction{}
		a.GetTransaction(hashes[i], &tr)
		b.Transactions = append(b.Transactions, &tr)
	}

	return a.GetObject(&obj)
}

func (a *dbagent) AddBlock(b *blockchain.Block) int64 {
	if a.getObjectCount(hex.EncodeToString(b.Header.Hash)) > 0 {
		log.Panicf("Replicatoin exists : %v", hex.EncodeToString(b.Header.Hash))
		return 0
	}

	hash := b.Header.GetHash()
	shash := hex.EncodeToString(hash[:])
	a.AddBlockHeader(shash, &b.Header)

	var hashes []string
	hashes = append(hashes, shash)
	for _, t := range b.Transactions {
		if a.AddTransaction(t) > 0 {
			hashes = append(hashes, hex.EncodeToString(t.Hash))
		}
	}

	obj := StorageObj{"block", hex.EncodeToString(b.Header.Hash), b.Header.Timestamp, hashes}
	return a.AddObject(&obj)
}

func (a *dbagent) ShowAllObjets() bool {
	rows, err := a.db.Query("SELECT * FROM bcobjects")
	if err != nil {
		log.Printf("Show all objects Error : %v", err)
		return false
	}

	defer rows.Close()
	for rows.Next() {
		var obj StorageObj
		var id int
		rows.Scan(&id, &obj.Type, &obj.Hash, &obj.Timestamp, &obj.Data)
		log.Printf("id=%d : %s %s", id, obj.Type, obj.Hash)
		if obj.Type == "transaction" {
			tr := blockchain.Transaction{}
			a.GetTransaction(obj.Hash, &tr)
			log.Printf("Transaction : %s", tr.Data)
		}
	}

	return false
}

func (a *dbagent) GetDBSize() uint64 {
	var size uint64 = 0
	rows, err := a.db.Query("SELECT type, hash, timestamp, data FROM bcobjects")
	if err != nil {
		log.Printf("Show db size Error : %v", err)
		return 0
	}

	defer rows.Close()
	for rows.Next() {
		var obj StorageObj
		rows.Scan(&obj.Type, &obj.Hash, &obj.Timestamp, &obj.Data)
		size += uint64(unsafe.Sizeof(obj.Type)) + uint64(unsafe.Sizeof(obj.Hash)) + uint64(unsafe.Sizeof(obj.Timestamp)) + uint64(len(obj.Data.([]byte)))
		//log.Printf("size : %d %d %d %d", unsafe.Sizeof(obj.Type), unsafe.Sizeof(obj.Hash), unsafe.Sizeof(obj.Timestamp), len(obj.Data.([]byte)))
	}
	return size
}

func (a *dbagent) getLatestDBStatus(status *DBStatus) bool {
	rows, err := a.db.Query("SELECT id, headers, blocks, transactions, size, timestamp FROM dbstatus WHERE id = (SELECT MAX(id)  FROM dbstatus)")
	if err != nil {
		log.Printf("Show latest db status Error : %v", err)
		return false
	}

	defer rows.Close()

	for rows.Next() {
		rows.Scan(&status.ID, &status.Headers, &status.Blocks, &status.Transactions, &status.Size, &status.Timestamp)
		//log.Printf("DB Status : %v", status)
		return true
	}

	return false
}

func (a *dbagent) updateRemoveDBStatus(hash string) {
	rows, err := a.db.Query("SELECT type, hash, timestamp, data FROM bcobjects WHERE hash=?", hash)
	if err != nil {
		log.Printf("update remove db status error : %v", err)
		return
	}

	defer rows.Close()
	status := DBStatus{}
	a.getLatestDBStatus(&status)
	update := false

	for rows.Next() {
		var obj StorageObj
		rows.Scan(&obj.Type, &obj.Hash, &obj.Timestamp, &obj.Data)
		size := int(unsafe.Sizeof(obj.Type)) + int(unsafe.Sizeof(obj.Hash)) + int(unsafe.Sizeof(obj.Timestamp)) + int(len(obj.Data.([]byte)))
		switch obj.Type {
		case "block":
			status.Blocks -= 1
			status.Size -= size
			update = true
		case "transaction":
			status.Transactions -= 1
			status.Size -= size
			update = true
		case "blockheader":
			status.Headers -= 1
			status.Size -= size
			update = true
		default:
			log.Printf("Type error %s", obj.Type)
		}
	}

	if update == true {
		a.updateDBStatus(&status)
	}
}

func (a *dbagent) updateAddDBStatus(id int64) {
	rows, err := a.db.Query("SELECT type, hash, timestamp, data FROM bcobjects WHERE id=?", id)
	if err != nil {
		log.Printf("update remove db status error : %v", err)
		return
	}

	defer rows.Close()
	status := DBStatus{}
	a.getLatestDBStatus(&status)
	update := false

	for rows.Next() {
		var obj StorageObj
		rows.Scan(&obj.Type, &obj.Hash, &obj.Timestamp, &obj.Data)
		size := int(unsafe.Sizeof(obj.Type)) + int(unsafe.Sizeof(obj.Hash)) + int(unsafe.Sizeof(obj.Timestamp)) + int(len(obj.Data.([]byte)))
		switch obj.Type {
		case "block":
			status.Blocks += 1
			status.Size += size
			update = true
		case "transaction":
			status.Transactions += 1
			status.Size += size
			update = true
		case "blockheader":
			status.Headers += 1
			status.Size += size
			update = true
		default:
			log.Printf("Type error %s", obj.Type)
		}
	}

	if update == true {
		a.updateDBStatus(&status)
	}
}

func (a *dbagent) updateDBStatus(status *DBStatus) int64 {
	st, err := a.db.Prepare("INSERT INTO dbstatus (headers, blocks, transactions, size, timestamp) VALUES (?, ?, ?, ?,  datetime('now'))")
	if err != nil {
		log.Printf("Prepare adding dbstatus error : %v", err)
		return -1
	}
	rst, err := st.Exec(status.Headers, status.Blocks, status.Transactions, status.Size)
	if err != nil {
		log.Panicf("Exec adding dbstatus error : %v", err)
		return -1
	}

	id, _ := rst.LastInsertId()
	return id
}

func (a *dbagent) GetDBStatus(status *DBStatus) bool {
	return a.getLatestDBStatus(status)
}

func newDBSqlite(path string) DBAgent {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Panicf("Open sqlite db error : %v", err)
	}

	create_objtlb := `CREATE TABLE IF NOT EXISTS bcobjects (
		id      	INTEGER  PRIMARY KEY AUTOINCREMENT,
		type 		TEXT,
		hash    	TEXT,
		timestamp	INTEGER,
		data		BLOB
	);`

	st, err := db.Prepare(create_objtlb)
	if err != nil {
		log.Panicf("create_objtlb error %v", err)
	}
	st.Exec()

	create_statustlb := `CREATE TABLE IF NOT EXISTS dbstatus (
		id      		INTEGER  PRIMARY KEY AUTOINCREMENT,
		headers			INTEGER,
		blocks 			INTEGER,
		transactions    INTEGER,
		size			INTEGER,
		timestamp		DATETIME
	);`

	st, err = db.Prepare(create_statustlb)
	if err != nil {
		log.Panicf("create_statustlb error %v", err)
	}
	st.Exec()

	return &dbagent{db: db}
}
