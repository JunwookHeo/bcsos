package storage

import (
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dbagent"
)

type ObjectMgr struct {
	db dbagent.DBAgent
}

func (c *ObjectMgr) DeleteNoAccedObjects() {
	c.db.DeleteNoAccedObjects()
}

func (c *ObjectMgr) AccessWithRandom(num int) []string {
	hashes := c.db.GetTransactionwithRandom(num)
	var rethashes []string
	for _, hash := range hashes {
		var tr blockchain.Transaction
		if c.db.GetTransaction(hash, &tr) == 0 {
			rethashes = append(rethashes, hash)
		}
	}
	return rethashes
}

func (c *ObjectMgr) AccessWithTimeWeight(num int) []string {
	hashes := c.db.GetTransactionwithTimeWeight(num)
	var rethashes []string
	for _, hash := range hashes {
		var tr blockchain.Transaction
		if c.db.GetTransaction(hash, &tr) == 0 {
			rethashes = append(rethashes, hash)
		}
	}
	return rethashes
}

func NewObjMgr(db dbagent.DBAgent) *ObjectMgr {
	om := ObjectMgr{db}
	return &om
}
