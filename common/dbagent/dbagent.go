package dbagent

import (
	"time"

	"github.com/junwookheo/bcsos/common/blockchain"
)

type DBAgent interface {
	Init()
	Close()
	RemoveObject(hash string) bool
	AddBlock(b *blockchain.Block) int64
	GetBlock(hash string, b *blockchain.Block) int64
	ShowAllObjets() bool
	GetDBSize() uint64
	GetDBStatus(status *DBStatus) bool
}

type StorageObj struct {
	Type string
	Hash string
	Data interface{}
}

type DBStatus struct {
	ID           int
	Headers      int
	Blocks       int
	Transactions int
	Size         int
	Timestamp    time.Time
}

func NewDBAgent(path string) DBAgent {
	return newDBSqlite(path)
}
