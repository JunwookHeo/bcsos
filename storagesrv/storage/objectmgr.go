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

func (c *ObjectMgr) AccessWithRandom(num int) *[]dbagent.RemoverbleObj {
	hashes := c.db.GetTransactionwithRandom(num)
	rethashes := []dbagent.RemoverbleObj{}
	for _, hash := range *hashes {
		if hash.HashType == 0 {
			var bh blockchain.BlockHeader
			if c.db.GetBlockHeader(hash.Hash, &bh) == 0 {
				rethashes = append(rethashes, hash)
			}
		} else {
			var tr blockchain.Transaction
			if c.db.GetTransaction(hash.Hash, &tr) == 0 {
				rethashes = append(rethashes, hash)
			}
		}
	}

	return &rethashes
}

func (c *ObjectMgr) AccessWithTimeWeight(num int) *[]dbagent.RemoverbleObj {
	hashes := c.db.GetTransactionwithTimeWeight(num)
	rethashes := []dbagent.RemoverbleObj{}
	for _, hash := range *hashes {
		if hash.HashType == 0 {
			var bh blockchain.BlockHeader
			if c.db.GetBlockHeader(hash.Hash, &bh) == 0 {
				rethashes = append(rethashes, hash)
			}
		} else {
			var tr blockchain.Transaction
			if c.db.GetTransaction(hash.Hash, &tr) == 0 {
				rethashes = append(rethashes, hash)
			}
		}
	}
	return &rethashes
}

func NewObjMgr(db dbagent.DBAgent) *ObjectMgr {
	om := ObjectMgr{db}
	return &om
}
