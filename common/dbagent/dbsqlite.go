package dbagent

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"log"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/serial"
	_ "github.com/mattn/go-sqlite3"
)

type dbagent struct {
	db *sql.DB
}

func (a *dbagent) Init() {

}
func (a *dbagent) Close() {
	a.db.Close()
}

func (a *dbagent) GetObject(obj *StorageObj) bool {
	var data []byte
	switch err := a.db.QueryRow("SELECT type, hash, data FROM bcobjects WHERE hash=?", obj.Hash).Scan(&obj.Type, &obj.Hash, &data); err {
	case sql.ErrNoRows:
		log.Printf("Object Not found : %v", err)
	case nil:
		serial.Deserialize(data, obj.Data)
		return true
	default:
		log.Printf("Get object error : %v", err)
	}

	return false
}
func (a *dbagent) AddObject(obj *StorageObj) int64 {
	st, err := a.db.Prepare("INSERT INTO bcobjects (type, hash, data) VALUES (?, ?, ?)")
	if err != nil {
		log.Printf("Prepare adding object error : %v", err)
		return -1
	}
	data := serial.Serialize(obj.Data)
	rst, err := st.Exec(obj.Type, obj.Hash, data)
	if err != nil {
		log.Panicf("Exec adding object error : %v", err)
		return -1
	}

	id, _ := rst.LastInsertId()
	return id
}

func (a *dbagent) GetBlockHeader(hash string, h *blockchain.BlockHeader) bool {
	obj := StorageObj{"blockheader", hash, h}
	return a.GetObject(&obj)
}

func (a *dbagent) AddBlockHeader(hash string, h *blockchain.BlockHeader) int64 {
	obj := StorageObj{"blockheader", hash, h}
	return a.AddObject(&obj)
}

func (a *dbagent) GetTransaction(hash string, t *blockchain.Transaction) bool {
	obj := StorageObj{"transaction", hash, t}
	return a.GetObject(&obj)
}

func (a *dbagent) AddTransaction(t *blockchain.Transaction) int64 {
	obj := StorageObj{"transaction", hex.EncodeToString(t.Hash), t}
	return a.AddObject(&obj)
}

func (a *dbagent) GetBlock(hash string, b *blockchain.Block) bool {
	hashes := []string{}
	obj := StorageObj{"block", hash, &hashes}
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
	var hashes []string
	hash := sha256.Sum256(serial.Serialize(b.Header))
	shash := hex.EncodeToString(hash[:])
	a.AddBlockHeader(shash, &b.Header)

	hashes = append(hashes, shash)
	for _, t := range b.Transactions {
		if a.AddTransaction(t) > 0 {
			hashes = append(hashes, hex.EncodeToString(t.Hash))
		}
	}

	obj := StorageObj{"block", hex.EncodeToString(b.Header.Hash), hashes}
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
		rows.Scan(&id, &obj.Type, &obj.Hash, &obj.Data)
		log.Printf("id=%d : %s %s", id, obj.Type, obj.Hash)
		if obj.Type == "transaction" {
			tr := blockchain.Transaction{}
			a.GetTransaction(obj.Hash, &tr)
			log.Printf("Transaction : %s", tr.Data)
		}
	}

	return false
}

func newDBSqlite(path string) DBAgent {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Panicf("Open sqlite db error : %v", err)
	}

	create_tbl := `CREATE TABLE IF NOT EXISTS bcobjects (
		id      INTEGER  PRIMARY KEY AUTOINCREMENT,
		type 	TEXT,
		hash    TEXT,
		data	BLOB
	);`

	st, _ := db.Prepare(create_tbl)

	st.Exec()
	return &dbagent{db: db}
}
