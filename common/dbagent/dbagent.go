package dbagent

import (
	"time"

	"github.com/junwookheo/bcsos/common/blockchain"
)

type DBAgent interface {
	Close()
	GetLatestBlockHash() string
	RemoveObject(hash string) bool
	AddBlockHeader(hash string, h *blockchain.BlockHeader) int64
	GetBlockHeader(hash string, h *blockchain.BlockHeader) int64
	AddTransaction(t *blockchain.Transaction) int64
	GetTransaction(hash string, t *blockchain.Transaction) int64
	AddBlock(b *blockchain.Block) int64
	GetBlock(hash string, b *blockchain.Block) int64
	ShowAllObjets() bool
	GetDBDataSize() uint64
	GetDBStatus(status *DBStatus) bool
	GetTransactionwithRandom(num int) []string
	GetTransactionwithTimeWeight() []string
	DeleteNoAccedObject()
	UpdateDBNetworkOverhead(qc int)
}

type StorageObj struct {
	Type      string
	Hash      string
	Timestamp int64
	ACTime    int64
	AFLevel   int64 // access frequency level
	Data      interface{}
}

type DBStatus struct {
	ID                int
	TotalBlocks       int
	TotalTransactoins int
	Headers           int
	Blocks            int
	Transactions      int
	Size              int
	Overhead          int // the number of additional query
	Timestamp         time.Time
}

func NewDBAgent(path string, afl int) DBAgent {
	return newDBSqlite(path, afl)
}
