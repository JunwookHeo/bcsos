package dbagent

import (
	"time"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dtype"
)

type DBAgent interface {
	Close()
	GetLatestBlockHash() (string, int)
	RemoveObject(hash string) bool
	AddBlockHeader(hash string, h *blockchain.BlockHeader) int64
	GetBlockHeader(hash string, h *blockchain.BlockHeader) int64
	AddTransaction(t *blockchain.Transaction) int64
	GetTransaction(hash string, t *blockchain.Transaction) int64
	AddNewBlock(block interface{}) int64
	GetBlock(hash string, b *blockchain.Block) int64
	ShowAllObjets() bool
	GetDBDataSize() uint64
	GetDBStatus() *DBStatus
	GetTransactionwithUniform(num int, hashes *[]RemoverbleObj) bool
	GetTransactionwithExponential(num int, hashes *[]RemoverbleObj) bool
	DeleteNoAccedObjects()
	UpdateDBNetworkQuery(fromqc int, toqc int, totalqc int)
	UpdateDBNetworkDelay(addtime int, hop int)
	GetNonInteractiveStarksProof(hash string) *dtype.NonInteractiveProof
	VerifyInterActiveProofStorage(proof *dtype.PoSProof) bool
	VerifyNonInterActiveProofStorage(proof *dtype.NonInteractiveProof) bool
	GetInteractiveProof(height int) *dtype.PoSProof
	GetRandomHeightForNConsecutiveBlocks(hash string) int
	GetLastBlockTime() int64
	// ProofStorage(tidx [32]byte, timestamp int64, tsc int) []byte
	// ProofStorage2()
}

type StorageObj struct {
	Type      string
	Hash      string
	Timestamp int64
	Data      interface{}
}

type StorageBLTR struct {
	Blockhash       string
	index           int
	Transactionhash string
	ACTime          int64
	AFLever         int
}

type DBStatus struct {
	Timestamp         time.Time
	ID                int
	TotalBlocks       int
	TotalTransactoins int
	Headers           int
	Blocks            int
	Transactions      int
	Size              int
	TotalQuery        int // the number of query including local storage
	QueryFrom         int // the number of received query
	QueryTo           int // the number of send query
	TotalDelay        int
	Hop0              int
	Hop1              int
	Hop2              int
	Hop3              int
}

type RemoverbleObj struct {
	HashType int // blockheader == 0 otherwise transaction
	Hash     string
}

func NewDBAgent(path string) DBAgent {
	// return newDBSqlite(path)
	return newDBBtcSqlite(path)
}
