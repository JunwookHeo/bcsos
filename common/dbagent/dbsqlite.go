package dbagent

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/junwookheo/bcsos/blockchainnode/network"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/serial"
	_ "github.com/mattn/go-sqlite3"
)

type dbagent struct {
	db       *sql.DB
	sclass   int
	dbstatus DBStatus
	mutex    sync.Mutex
}

func (a *dbagent) Close() {
	a.db.Close()
}

func (a *dbagent) GetLatestBlockHash() (string, int) {
	var id int = 0
	var height int = -1
	var data []byte
	var dtype, hash string = "", ""
	var ts int64

	rows, err := a.db.Query(`SELECT id, type, hash, timestamp, data FROM bcobjects WHERE id = (SELECT MAX(id) FROM bcobjects WHERE type = 'block');`)

	if err != nil {
		log.Printf("Show latest objects Error : %v", err)
		return hash, -1
	}

	defer rows.Close()

	for rows.Next() {
		err := rows.Scan(&id, &dtype, &hash, &ts, &data)
		serial.Deserialize(data, &height)

		if err != nil {
			log.Printf("Read rows Error : %v", err)
			return hash, height
		}
		//log.Printf("latest block : %d, %s, %v, %s", id, dtype, ts, hash)
	}

	return hash, height
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
	rst, err := st.Exec(act, a.sclass, hash)
	if err != nil {
		log.Panicf("Update exec error id(%v): %v", hash, err)
		return false
	}

	cnt, _ := rst.RowsAffected()

	return cnt > 0
}

func (a *dbagent) getObject(obj *StorageObj) int64 {
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
	if id := a.getObject(obj); id != 0 {
		// log.Printf("Replicatoin exists : %v - %v", id, obj)
		return id
	}

	id := func() int64 {
		a.mutex.Lock()
		defer a.mutex.Unlock()
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
		return id
	}()

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
	obj := StorageBLTR{bh, index, th, time.Now().UnixNano(), a.sclass}
	a.mutex.Lock()
	defer a.mutex.Unlock()
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

// DeleteNoAccedObjects will delete transaction if there is no access more than a hour
func (a *dbagent) DeleteNoAccedObjects() {
	//log.Printf("%v", config.TSC0I)
	ts := time.Now().UnixNano() - int64(config.TSCX[a.sclass]*float32(1e9)) // no access for if one hour, delete it
	//rows, err := a.db.Query(`SELECT transactionhash FROM blocktrtbl WHERE actime >= ? AND actime < ?;`, a.latestts, ts)
	rows, err := a.db.Query(`SELECT hash FROM bcobjects WHERE type != 'block' AND hash 
								IN (SELECT transactionhash FROM blocktrtbl WHERE actime < ?) ;`, ts)
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

func (a *dbagent) GetTransactionwithUniform(num int, hashes *[]RemoverbleObj) bool {
	w := config.TOTAL_TRANSACTIONS + config.TOTAL_TRANSACTIONS/config.NUM_TRANSACTION_BLOCK

	ids := func(w int, num int) string {
		ids := []string{}
		for i := 0; i < num; i++ {
			l := rand.Intn(int(w))
			ids = append(ids, strconv.Itoa(l))
		}
		return strings.Join(ids, ", ")
	}(w, num)

	// select_hashes := fmt.Sprintf(`select transactionhash from (select *, row_number() over (order by actime desc) rownum
	// 					from blocktrtbl where idx != 0) where rownum in (%s) LIMIT 50;`, ids)
	select_hashes := fmt.Sprintf(`SELECT idx, transactionhash FROM (SELECT *, row_number() OVER (ORDER BY actime desc) rownum 
						FROM blocktrtbl) WHERE rownum IN (%s) LIMIT %d;`, ids, num)

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
	// log.Printf("GetTransactionwithUniform : %v", hashes)
	return true
}

// It is difficult to select samples with exponential distribution.
// It just select rows by exponentially generated numbers
// without considering access time.
func (a *dbagent) GetTransactionwithExponential(num int, hashes *[]RemoverbleObj) bool {
	w := float64(config.BASIC_UNIT_TIME*config.RATE_TSC*(config.NUM_TRANSACTION_BLOCK+1.)) / float64(config.BLOCK_CREATE_PERIOD)

	ids := func(w float64, num int) string {
		ids := []string{}
		for i := 0; i < num; i++ {
			f := rand.ExpFloat64() / float64(config.LAMBDA_ED)
			l := int(f * w)
			ids = append(ids, strconv.Itoa(l))
		}
		return strings.Join(ids, ", ")
	}(w, num)

	// select_hashes := fmt.Sprintf(`select transactionhash from (select *, row_number() over (order by actime desc) rownum
	// 					from blocktrtbl where idx != 0) where rownum in (%s) LIMIT 50;`, ids)
	select_hashes := fmt.Sprintf(`SELECT idx, transactionhash FROM (SELECT *, row_number() OVER (ORDER BY actime desc) rownum 
						FROM blocktrtbl) WHERE rownum IN (%s) LIMIT %d;`, ids, num)

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
	//log.Printf("GetTransactionwithExponential : %v", hashes)
	return true
}

func (a *dbagent) GetBlockHeader(hash string, h *blockchain.BlockHeader) int64 {
	obj := StorageObj{"blockheader", hash, h.Timestamp, h}
	return a.getObject(&obj)
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
	return a.getObject(&obj)
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
	id := a.getObject(&obj)
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

func (a *dbagent) AddNewBlock(block interface{}) int64 {
	// TODO : Impliment encrypting a new block and stor it for Proof of Storage
	b, ok := block.(*blockchain.Block)
	if !ok {
		log.Panicf("Type mismatch : %v", ok)
		return -1
	}

	hash := hex.EncodeToString(b.Header.Hash)
	obj := StorageObj{"block", hash, b.Header.Timestamp, b.Header.Height} // store height in data field.
	if id := a.getObject(&obj); id != 0 {
		log.Printf("Replicatoin exists : %v - %v", id, hex.EncodeToString(b.Header.Hash))
		return id
	}

	bhash := b.Header.GetHash()
	header_hash := hex.EncodeToString(bhash[:])
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
	obj = StorageObj{"block", hex.EncodeToString(b.Header.Hash), b.Header.Timestamp, b.Header.Height}

	if id := a.AddObject(&obj); id != 0 {
		a.mutex.Lock()
		defer a.mutex.Unlock()
		status := &a.dbstatus
		status.TotalBlocks += 1
		status.TotalTransactoins += cnt
		//a.updateMemDBStatus(status)

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
		err = rows.Scan(&index, &hash)
		if err != nil {
			log.Printf("ShowAllObjets error : %v", err)
			return false
		}
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
		return 0
	}
	log.Printf("size : %v", size)

	return size
}

func (a *dbagent) getLatestDBStatus(status *DBStatus) bool {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	//rows, err := a.db.Query(`SELECT id, timestamp, totalblocks, totaltransactions, headers, blocks, transactions, size, totalquery, queryfrom, queryto, totaldelay, totalhop  FROM dbstatus WHERE id = (SELECT MAX(id)  FROM dbstatus);`)
	rows, err := a.db.Query(`SELECT *  FROM dbstatus WHERE id = (SELECT MAX(id)  FROM dbstatus);`)
	if err != nil {
		log.Printf("Show latest db status Error : %v", err)
		return false
	}

	defer rows.Close()

	for rows.Next() {
		rows.Scan(&status.ID, &status.Timestamp, &status.TotalBlocks, &status.TotalTransactoins, &status.Headers, &status.Blocks, &status.Transactions,
			&status.Size, &status.TotalQuery, &status.QueryFrom, &status.QueryTo, &status.TotalDelay, &status.Hop0, &status.Hop1, &status.Hop2, &status.Hop3)

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

	a.mutex.Lock()
	defer a.mutex.Unlock()
	status := &a.dbstatus

	for rows.Next() {
		obj := StorageObj{}
		var size int
		rows.Scan(&obj.Type, &size)
		switch obj.Type {
		case "block":
			status.Blocks -= 1
			status.Size -= size
		case "transaction":
			status.Transactions -= 1
			status.Size -= size
		case "blockheader":
			status.Headers -= 1
			status.Size -= size
		default:
			log.Printf("Type error %s", obj.Type)
		}
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

	a.mutex.Lock()
	defer a.mutex.Unlock()
	status := &a.dbstatus

	for rows.Next() {
		var obj StorageObj
		var size int
		rows.Scan(&obj.Type, &size)
		switch obj.Type {
		case "block":
			status.Blocks += 1
			status.Size += size
		case "transaction":
			status.Transactions += 1
			status.Size += size
		case "blockheader":
			status.Headers += 1
			status.Size += size
		default:
			log.Printf("Type error %s", obj.Type)
		}
	}
}

// func (a *dbagent) updateMemDBStatus(status *DBStatus) {
// 	a.mutex.Lock()
// 	defer a.mutex.Unlock()
// 	a.dbstatus = *status
// }

func (a *dbagent) updateDBStatus() {
	getHash := func(status DBStatus) string {
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

			st, err := a.db.Prepare(`INSERT INTO dbstatus (timestamp, totalblocks, totaltransactions, headers, blocks, transactions, size, 
			totalquery, queryfrom, queryto, totaldelay, hop0, hop1, hop2, hop3) VALUES ( datetime('now'), ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`)
			if err != nil {
				log.Printf("Prepare adding dbstatus error : %v", err)
				return
			}
			defer st.Close()

			rst, err := st.Exec(status.TotalBlocks, status.TotalTransactoins, status.Headers, status.Blocks, status.Transactions, status.Size,
				status.TotalQuery, status.QueryFrom, status.QueryTo, status.TotalDelay, status.Hop0, status.Hop1, status.Hop2, status.Hop3)
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

func (a *dbagent) UpdateDBNetworkQuery(fromqc int, toqc int, totalqc int) {
	a.mutex.Lock()
	defer a.mutex.Unlock()
	status := &a.dbstatus
	status.QueryFrom += fromqc
	status.QueryTo += toqc
	status.TotalQuery += totalqc
	// a.updateMemDBStatus(status)
}

// addtime : nano second
// hop : 1 to 3
func (a *dbagent) UpdateDBNetworkDelay(addtime int, hop int) {
	status := &a.dbstatus
	status.TotalDelay += (addtime / 1000000) //milli second
	switch hop {
	case 0:
		status.Hop0 += 1
	case 1:
		status.Hop1 += 1
	case 2:
		status.Hop2 += 1
	case 3:
		status.Hop3 += 1
	}
}

func (a *dbagent) GetNonInteractiveProof(hash string) *dtype.NonInteractiveProof {
	return nil
}

func (a *dbagent) VerifyNonInteractiveProof(proof *dtype.NonInteractiveProof) bool {
	return false
}

// func (a *dbagent) ProofStorage2() {
// 	TargetBits := 5
// 	tidx, _ := hex.DecodeString("0ab51095bf5314967f964422f91fc6b39e7761103875eeafebe1cef430d9f531")

// 	matcht := hex.EncodeToString(tidx[:])
// 	matchs := matcht[len(matcht)-TargetBits/4:]
// 	log.Printf("matchs : %v", matchs)

// 	a.mutex.Lock()
// 	defer a.mutex.Unlock()

// 	query := fmt.Sprintf(`SELECT hash FROM bcobjects WHERE type = "transaction" and hash LIKE "%%%s";`, matchs)
// 	rows, err := a.db.Query(query)
// 	if err != nil {
// 		log.Printf("Show latest db status Error : %v", err)
// 		return
// 	}

// 	defer rows.Close()
// 	i := 0
// 	var hash string
// 	// posDiffBit := 5
// 	mask := (uint64)(0xFFFFFFFFFFFFFFFF >> (64 - TargetBits))
// 	matchi, _ := strconv.ParseUint(matcht[len(matcht)-16:], 16, 64)
// 	matchi = matchi & mask
// 	log.Printf("match : %x", matchi)

// 	for rows.Next() {
// 		err := rows.Scan(&hash)
// 		if err != nil {
// 			break
// 		}
// 		value, _ := strconv.ParseUint(hash[len(hash)-16:], 16, 64)
// 		if value&mask == matchi {
// 			log.Printf("not match %v : %x, %x, %v", i, mask, value&mask, hash)
// 			i++
// 		}

// 	}
// }

// // Proof of Storage
// func (a *dbagent) ProofStorage(tidx [32]byte, timestamp int64, tsc int) []byte {
// 	match := hex.EncodeToString(tidx[:])
// 	log.Printf("match : %v", match)

// 	a.mutex.Lock()
// 	defer a.mutex.Unlock()

// 	query := fmt.Sprintf(`SELECT hash, timestamp, data FROM bcobjects WHERE type = "transaction" and hash LIKE "%%%s";`, match[len(match)-1:])
// 	rows, err := a.db.Query(query)
// 	if err != nil {
// 		log.Printf("Show latest db status Error : %v", err)
// 		return nil
// 	}

// 	defer rows.Close()

// 	var data []byte
// 	i := 0
// 	tr := blockchain.Transaction{}
// 	var hash string
// 	var ts int64
// 	var hashes [][]byte

// 	// Guard Time : 60 sec
// 	th_lower := timestamp - int64(config.TSCX[tsc]*float32(1e9)) + int64(60*float32(1e9))
// 	th_upper := timestamp - int64(60*float32(1e9))
// 	// log.Printf("threshold : %v, %v, %v", th_lower, th_upper, timestamp)

// 	for rows.Next() {
// 		err := rows.Scan(&hash, &ts, &data)
// 		if err != nil {
// 			break
// 		}

// 		serial.Deserialize(data, &tr)
// 		if th_lower < tr.Timestamp && tr.Timestamp < th_upper {
// 			log.Printf("match %v : %v - %v", i, time.Unix((ts/1000)/1e6, ((ts/1000)%1e6)*1e3), hex.EncodeToString(tr.Hash))
// 			tr.Timestamp = timestamp
// 			hashes = append(hashes, tr.GetHash())
// 			i++
// 		} else {
// 			log.Printf("not match %v : %v - %v", i, time.Unix((ts/1000)/1e6, ((ts/1000)%1e6)*1e3), hex.EncodeToString(tr.Hash))

// 		}
// 	}
// 	log.Printf("MerkleRoot : %v - %v", i, hex.EncodeToString(blockchain.CalMerkleRootHash(hashes)))
// 	return blockchain.CalMerkleRootHash(hashes)
// }

func (a *dbagent) GetDBStatus() *DBStatus {
	a.dbstatus.Timestamp = time.Now()
	return &a.dbstatus
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

	// totalquery : query objects including local storage
	// queryfrom : the number of received queries to get deleted transactions
	// queryto : the number of send queries to get deleted transactions
	create_statustlb := `CREATE TABLE IF NOT EXISTS dbstatus (
		id      			INTEGER  PRIMARY KEY AUTOINCREMENT,
		timestamp			DATETIME,
		totalblocks			INTEGER,
		totaltransactions 	INTEGER,
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

	dba := dbagent{db: db, sclass: local.SC, dbstatus: DBStatus{Timestamp: time.Now()}, mutex: sync.Mutex{}}
	dba.getLatestDBStatus(&dba.dbstatus)
	go dba.updateDBStatus()
	return &dba
}
