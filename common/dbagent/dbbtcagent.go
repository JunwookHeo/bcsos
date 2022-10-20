package dbagent

import (
	"database/sql"
	"log"
	"sync"
	"time"

	"github.com/junwookheo/bcsos/blockchainnode/network"
)

func newDBBtcSqlite(path string) DBAgent {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		log.Panicf("Open sqlite db error : %v", err)
	}

	create_objtlb := `CREATE TABLE IF NOT EXISTS btcblock (
		id      	INTEGER  PRIMARY KEY AUTOINCREMENT,
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

	dba := dbagent{db: db, SClass: local.SC, dbstatus: DBStatus{Timestamp: time.Now()}, mutex: sync.Mutex{}}
	dba.getLatestDBStatus(&dba.dbstatus)
	go dba.updateDBStatus()
	return &dba
}
