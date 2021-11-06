package dbagent

import "github.com/junwookheo/bcsos/common/blockchain"

type DBAgent interface {
	Init()
	Close()
	AddBlock(b *blockchain.Block) int64
	GetBlock(hash string, b *blockchain.Block) bool
	ShowAllObjets() bool
}

type StorageObj struct {
	Type string
	Hash string
	Data interface{}
}

func NewDBAgent(path string) DBAgent {
	return newDBSqlite(path)
}
