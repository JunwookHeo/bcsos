package dbagent

import (
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/serial"
	_ "github.com/mattn/go-sqlite3"
)

type dbagent struct {
	db       *sql.DB
	AFLevel  int
	dbstatus DBStatus
	mutex    sync.Mutex
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
		err := rows.Scan(&id, &dtype, &ts, &hash)
		if err != nil {
			log.Printf("Read rows Error : %v", err)
			return hash
		}
		log.Printf("latest block : %d, %s, %v, %s", id, dtype, ts, hash)
	}

	return hash
}

func (a *dbagent) RemoveObject(hash string) bool {
	//Before removing data, update status first.
	a.updateRemoveDBStatus(hash)
	// log.Printf("RemoveObject starting %v", hash)
	// defer log.Printf("RemoveObject finished %v", hash)
	a.mutex.Lock()
	defer a.mutex.Unlock()

	st, err := a.db.Prepare("DELETE FROM bcobjects WHERE hash=?")
	if err != nil {
		log.Printf("Preparing removig object error : %v", err)
		return false
	}
	defer st.Close()

	rst, err := st.Exec(hash)
	if err != nil {
		log.Printf("Exec removing object error : %v", err)
		return false
	}
	cnt, _ := rst.RowsAffected()
	return cnt > 0
}

// func (a *dbagent) RemoveObjectMultiple(hashes []string) bool {
// 	//Before removing data, update status first.
// 	//a.updateRemoveDBStatus(hash)
// 	hashes_str := ""
// 	hashes_str += fmt.Sprintf("\"%s\"", strings.Join(hashes, `", "`))

// 	a.mutex.Lock()
// 	defer a.mutex.Unlock()
// 	delete_strings := fmt.Sprintf("DELETE FROM bcobjects WHERE hash IN (%v);", hashes_str)
// 	//log.Printf("delete : %v", delete_strings)
// 	st, err := a.db.Prepare(delete_strings)
// 	if err != nil {
// 		log.Printf("Preparing removig object error : %v", err)
// 		return false
// 	}
// 	defer st.Close()

// 	rst, err := st.Exec()
// 	if err != nil {
// 		log.Printf("Exec removing object error : %v", err)
// 		return false
// 	}
// 	cnt, _ := rst.RowsAffected()
// 	log.Printf("delete multiple : %v", cnt)
// 	return cnt > 0
// }

func (a *dbagent) updateACTimeObject(hash string) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	st, err := a.db.Prepare("UPDATE blocktrtbl SET actime=?, aflevel=? WHERE transactionhash=?")
	if err != nil {
		log.Panicf("Update error id(%v) : %v", hash, err)
		return false
	}
	defer st.Close()

	act := time.Now().UnixNano()
	rst, err := st.Exec(act, a.AFLevel, hash)
	if err != nil {
		log.Panicf("Update exec error id(%v): %v", hash, err)
		return false
	}

	cnt, _ := rst.RowsAffected()

	return cnt > 0
}

func (a *dbagent) GetObject(obj *StorageObj) int64 {
	var data []byte
	var id int64 = 0
	switch err := a.db.QueryRow("SELECT id, type, hash, timestamp, data FROM bcobjects WHERE hash=?",
		obj.Hash).Scan(&id, &obj.Type, &obj.Hash, &obj.Timestamp, &data); err {
	case sql.ErrNoRows:
		//log.Printf("Object Not found : %v", err)
		break
	case nil:
		serial.Deserialize(data, obj.Data)
		a.updateACTimeObject(obj.Hash)
		return id
	default:
		log.Printf("Get object error : %v", err)
	}

	return id
}

func (a *dbagent) AddObject(obj *StorageObj) int64 {
	if id := a.GetObject(obj); id != 0 {
		log.Printf("Replicatoin exists : %v - %v", id, obj)
		return id
	}

	st, err := a.db.Prepare("INSERT INTO bcobjects (type, hash, timestamp, data) VALUES (?, ?, ?, ?)")
	if err != nil {
		log.Printf("Prepare adding object error : %v", err)
		return -1
	}
	defer st.Close()

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

func (a *dbagent) GetBlockTransactionMatching(bh string, hashes *[]string) int {
	var th string
	var index int64 = 0
	rows, err := a.db.Query("SELECT idx, transactionhash FROM blocktrtbl WHERE blockhash=? ORDER BY idx ASC", bh)
	if err != nil {
		log.Printf("GetBlockTransactionMatching get transaction hashes error : %v", err)
		return 0
	}
	defer rows.Close()

	cnt := 0
	for rows.Next() {
		err := rows.Scan(&index, &th)
		if err != nil {
			log.Printf("Read rows Error : %v", err)
			return 0
		}
		*hashes = append(*hashes, th)
		cnt++
		log.Printf("transactions : %d, %v", index, th)
	}

	return cnt
}

func (a *dbagent) AddBlockTransactionMatching(bh string, index int, th string) int64 {
	obj := StorageBLTR{bh, index, th, time.Now().UnixNano(), a.AFLevel}
	st, err := a.db.Prepare("INSERT INTO blocktrtbl (blockhash, idx, transactionhash, actime, aflevel) VALUES (?, ?, ?, ?, ?)")
	if err != nil {
		log.Printf("AddBlockTransactionMatching adding object error : %v", err)
		return 0
	}
	defer st.Close()

	rst, err := st.Exec(obj.Blockhash, obj.index, obj.Transactionhash, obj.ACTime, obj.AFLever)
	if err != nil {
		log.Panicf("Exec adding object error : %v", err)
		return 0
	}

	id, _ := rst.LastInsertId()

	return id
}

//DeleteNoAccedObjects will delete transaction if there is no access more than a hour
func (a *dbagent) DeleteNoAccedObjects() {
	//log.Printf("%v", config.TSC0F)
	ts := time.Now().UnixNano() - int64(config.TSC0F*(a.AFLevel+1)*1e9) // no access for if one hour, delete it
	//rows, err := a.db.Query(`SELECT transactionhash FROM blocktrtbl WHERE actime >= ? AND actime < ?;`, a.latestts, ts)
	rows, err := a.db.Query(`SELECT hash FROM bcobjects WHERE type != 'block' AND hash 
								IN (SELECT transactionhash FROM blocktrtbl WHERE actime < ?) LIMIT 50;`, ts)
	if err != nil {
		log.Printf("Object Not found : %v", err)
		return
	}

	defer rows.Close()
	for rows.Next() {
		var hash string
		err := rows.Scan(&hash)
		if err != nil {
			log.Printf("Delete no access transaction error : %v", err)
			return
		}
		//log.Printf("Delete no access transaction : %v", hash)
		go a.RemoveObject(hash)
	}
}

func (a *dbagent) GetTransactionwithRandom(num int, hashes *[]RemoverbleObj) bool {
	// Randomly select 10 blocks in the ledger
	//rows, err := a.db.Query(`SELECT transactionhash FROM blocktrtbl WHERE idx != 0 ORDER BY RANDOM() LIMIT ?;`, num)
	rows, err := a.db.Query(`SELECT idx, transactionhash FROM blocktrtbl ORDER BY RANDOM() LIMIT ?;`, num)
	if err != nil {
		log.Printf("Object Not found : %v", err)
		return false
	}

	defer rows.Close()

	// Select a transaction in each block
	for rows.Next() {
		var hash string
		var idx int
		err := rows.Scan(&idx, &hash)
		if err != nil {
			log.Printf("Read rows Error : %v", err)
			return false
		}
		*hashes = append(*hashes, RemoverbleObj{idx, hash})
		//log.Printf("Random choose hash : %v", hash)
	}
	log.Printf("Random choose hash : %v", *hashes)
	return true
}

func (a *dbagent) GetTransactionwithTimeWeight(num int, hashes *[]RemoverbleObj) bool {
	w := config.BASIC_UNIT_TIME * config.RATE_TSC0
	ids := func(w int, umn int) string {
		ids := []string{}
		cnt := 0
		for {
			if cnt > num {
				break
			}

			f := rand.ExpFloat64() / float64(config.LAMBDA_ED)
			l := int(f) * w
			flg := false
			for i := 0; i < cnt; i++ {
				if ids[i] == strconv.Itoa(l) {
					flg = true
					break
				}
			}
			if !flg {
				ids = append(ids, strconv.Itoa(l))
				cnt++
			}
		}
		return strings.Join(ids, ", ")
	}(w, num)

	// select_hashes := fmt.Sprintf(`select transactionhash from (select *, row_number() over (order by actime desc) rownum
	// 					from blocktrtbl where idx != 0) where rownum in (%s) LIMIT 50;`, ids)
	select_hashes := fmt.Sprintf(`SELECT idx, transactionhash FROM (SELECT *, row_number() OVER (ORDER BY actime desc) rownum 
						FROM blocktrtbl) WHERE rownum IN (%s) LIMIT 50;`, ids)

	//log.Printf("exponential items: %v", select_hashes)
	rows, err := a.db.Query(select_hashes)
	if err != nil {
		log.Printf("Object Not found : %v", err)
		return false
	}

	defer rows.Close()
	for rows.Next() {
		var hash string
		var idx int
		err := rows.Scan(&idx, &hash)
		if err != nil {
			log.Printf("Read rows Error : %v", err)
			return false
		}
		*hashes = append(*hashes, RemoverbleObj{idx, hash})
	}
	//log.Printf("GetTransactionwithTimeWeight : %v", hashes)
	return true
}

func (a *dbagent) GetBlockHeader(hash string, h *blockchain.BlockHeader) int64 {
	obj := StorageObj{"blockheader", hash, h.Timestamp, h}
	return a.GetObject(&obj)
}

func (a *dbagent) AddBlockHeader(hash string, h *blockchain.BlockHeader) int64 {
	if hash == "" {
		return 0
	}
	obj := StorageObj{"blockheader", hash, h.Timestamp, h}
	return a.AddObject(&obj)
}

func (a *dbagent) GetTransaction(hash string, t *blockchain.Transaction) int64 {
	obj := StorageObj{"transaction", hash, t.Timestamp, t}
	return a.GetObject(&obj)
}

func (a *dbagent) AddTransaction(t *blockchain.Transaction) int64 {
	if hex.EncodeToString(t.Hash) == "" {
		return 0
	}
	obj := StorageObj{"transaction", hex.EncodeToString(t.Hash), t.Timestamp, t}
	return a.AddObject(&obj)
}

func (a *dbagent) GetBlock(hash string, b *blockchain.Block) int64 {
	if hash == "" {
		return 0
	}

	obj := StorageObj{}
	obj.Type, obj.Hash = "block", hash
	id := a.GetObject(&obj)
	if id == 0 {
		log.Printf("Not found block : %v", hash)
		return 0
	}

	var hashes []string = []string{}
	cnt := a.GetBlockTransactionMatching(hash, &hashes)
	if cnt == 0 {
		log.Printf("Not found matching transactions : %v", hash)
		return 0
	}
	a.GetBlockHeader(hashes[0], &b.Header)

	for i := 1; i < len(hashes); i++ {
		tr := blockchain.Transaction{}
		a.GetTransaction(hashes[i], &tr)
		b.Transactions = append(b.Transactions, &tr)
	}

	return id
}

func (a *dbagent) AddBlock(b *blockchain.Block) int64 {
	hash := hex.EncodeToString(b.Header.Hash)
	obj := StorageObj{"block", hash, b.Header.Timestamp, nil}
	if id := a.GetObject(&obj); id != 0 {
		log.Printf("Replicatoin exists : %v - %v", id, hex.EncodeToString(b.Header.Hash))
		return id
	}

	data := b.Header.GetHash()
	header_hash := hex.EncodeToString(data[:])
	a.AddBlockHeader(header_hash, &b.Header)

	// Add block - transactions list in the table
	a.AddBlockTransactionMatching(hash, 0, header_hash)
	cnt := 0
	for i, t := range b.Transactions {
		a.AddBlockTransactionMatching(hash, i+1, hex.EncodeToString(t.Hash))
		a.AddTransaction(t)
		cnt++
	}

	// Add only block information without data, the data is stored in block-transaction matching table
	obj = StorageObj{"block", hex.EncodeToString(b.Header.Hash), b.Header.Timestamp, []byte{}}
	if id := a.AddObject(&obj); id != 0 {
		status := &a.dbstatus
		status.TotalBlocks += 1
		status.TotalTransactoins += cnt
		a.updateDBStatus(status)
		return id
	}

	return 0
}

func (a *dbagent) ShowAllObjets() bool {
	rows, err := a.db.Query("SELECT idx, transactionhash FROM blocktrtbl")
	if err != nil {
		log.Printf("Show all objects Error : %v", err)
		return false
	}

	defer rows.Close()
	for rows.Next() {
		var index int
		var hash string
		rows.Scan(&index, &hash)
		//log.Printf("%v : %v", index, hash)
		if index == 0 { // header
			bh := blockchain.BlockHeader{}
			a.GetBlockHeader(hash, &bh)
			log.Printf("Block Header %v : %v", bh.Hash, bh.Timestamp)
		} else {
			tr := blockchain.Transaction{}
			a.GetTransaction(hash, &tr)
			log.Printf("Transaction : %s", tr.Data)
		}
	}

	return false
}

func (a *dbagent) GetDBDataSize() uint64 {
	var size uint64 = 0
	err := a.db.QueryRow(`SELECT sum(length(type)) + sum(length(hash)) + sum(length(timestamp)) + sum(length(data)) AS size FROM bcobjects;`).Scan(&size)
	if err != nil {
		log.Printf("get size error : %v", err)
	}
	log.Printf("size : %v", size)

	return size
}

func (a *dbagent) getLatestDBStatus(status *DBStatus) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	rows, err := a.db.Query(`SELECT id, totalblocks, totaltransactions, headers, blocks, transactions, size, overheadfrom, overheadto, timestamp 
				FROM dbstatus WHERE id = (SELECT MAX(id)  FROM dbstatus);`)
	if err != nil {
		log.Printf("Show latest db status Error : %v", err)
		return false
	}

	defer rows.Close()

	for rows.Next() {
		rows.Scan(&status.ID, &status.TotalBlocks, &status.TotalTransactoins, &status.Headers, &status.Blocks, &status.Transactions,
			&status.Size, &status.Overheadfrom, status.Overheadto, &status.Timestamp)
		//log.Printf("DB Status : %v", status)
		return true
	}

	return false
}

func (a *dbagent) updateRemoveDBStatus(hash string) {
	// TODO: How to calculate the size of data
	// Include meta data like timestamp
	//rows, err := a.db.Query("SELECT type, length(hash) + length(timestamp) + length(data) FROM bcobjects WHERE hash=?", hash)
	rows, err := a.db.Query("SELECT type, length(hash) + length(data) FROM bcobjects WHERE hash=?", hash)
	if err != nil {
		log.Printf("update remove db status error : %v", err)
		return
	}

	defer rows.Close()

	status := &a.dbstatus
	update := false

	for rows.Next() {
		obj := StorageObj{}
		var size int
		rows.Scan(&obj.Type, &size)
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

	if update {
		a.updateDBStatus(status)
	}
}

func (a *dbagent) updateAddDBStatus(id int64) {
	// TODO: How to calculate the size of data
	// Include meta data like timestamp
	//rows, err := a.db.Query("SELECT type, length(hash) + length(timestamp) + length(data) FROM bcobjects WHERE id=?", id)
	rows, err := a.db.Query("SELECT type, length(hash) + length(data) FROM bcobjects WHERE id=?", id)
	if err != nil {
		log.Printf("update remove db status error : %v", err)
		return
	}

	defer rows.Close()

	status := &a.dbstatus
	update := false

	for rows.Next() {
		var obj StorageObj
		var size int
		rows.Scan(&obj.Type, &size)
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

	if update {
		a.updateDBStatus(status)
	}
}

func (a *dbagent) updateDBStatus(status *DBStatus) int64 {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	st, err := a.db.Prepare("INSERT INTO dbstatus (totalblocks, totaltransactions, headers, blocks, transactions, size, overheadfrom, overheadto, timestamp) VALUES (?, ?, ?, ?, ?, ?, ?, ?,  datetime('now'))")
	if err != nil {
		log.Printf("Prepare adding dbstatus error : %v", err)
		return -1
	}
	defer st.Close()

	rst, err := st.Exec(status.TotalBlocks, status.TotalTransactoins, status.Headers, status.Blocks, status.Transactions, status.Size, status.Overheadfrom, status.Overheadto)
	if err != nil {
		log.Panicf("Exec adding dbstatus error : %v", err)
		return -1
	}

	id, _ := rst.LastInsertId()
	status.ID = int(id)
	return id
}

func (a *dbagent) UpdateDBNetworkOverhead(fromqc int, toqc int) {
	status := &a.dbstatus
	status.Overheadfrom += fromqc
	status.Overheadto += toqc
	a.updateDBStatus(status)
}

func (a *dbagent) GetDBStatus() *DBStatus {
	return &a.dbstatus
}

func newDBSqlite(path string, afl int) DBAgent {
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
	defer st.Close()

	st.Exec()

	// block - transaction matching table
	// idx : 0-th is header, n-th transaction string from 0
	create_blocktrtbl := `CREATE TABLE IF NOT EXISTS blocktrtbl (
		id      		INTEGER  PRIMARY KEY AUTOINCREMENT,
		blockhash 		TEXT,
		idx				INTEGER,
		transactionhash TEXT,
		actime			INTEGER,
		aflevel			INTEGER
	);`

	st, err = db.Prepare(create_blocktrtbl)
	if err != nil {
		log.Panicf("create_objtlb error %v", err)
	}
	defer st.Close()

	st.Exec()

	// overheadfrom : the number of received queries to get deleted transactions
	// overheadto : the number of send queries to get deleted transactions
	create_statustlb := `CREATE TABLE IF NOT EXISTS dbstatus (
		id      			INTEGER  PRIMARY KEY AUTOINCREMENT,
		totalblocks			INTEGER,
		totaltransactions 	INTERGER,
		headers				INTEGER,
		blocks 				INTEGER,
		transactions    	INTEGER,
		size				INTEGER,
		overheadfrom		INTEGR,
		overheadto			INTEGR,
		timestamp			DATETIME
	);`

	st, err = db.Prepare(create_statustlb)
	if err != nil {
		log.Panicf("create_statustlb error %v", err)
	}
	defer st.Close()

	st.Exec()

	dba := dbagent{db: db, AFLevel: afl, dbstatus: DBStatus{}, mutex: sync.Mutex{}}
	dba.getLatestDBStatus(&dba.dbstatus)
	return &dba
}
