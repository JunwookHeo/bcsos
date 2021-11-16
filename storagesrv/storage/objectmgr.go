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

func (c *ObjectMgr) AccessWithRandom(num int, rethashes *[]dbagent.RemoverbleObj) bool {
	hashes := []dbagent.RemoverbleObj{}
	ret := c.db.GetTransactionwithRandom(num, &hashes)
	if !ret {
		return false
	}

	for _, hash := range hashes {
		if hash.HashType == 0 {
			var bh blockchain.BlockHeader
			if c.db.GetBlockHeader(hash.Hash, &bh) == 0 {
				*rethashes = append(*rethashes, hash)
				ret = true
			}
		} else {
			var tr blockchain.Transaction
			if c.db.GetTransaction(hash.Hash, &tr) == 0 {
				*rethashes = append(*rethashes, hash)
				ret = true
			}
		}
	}

	return ret
}

func (c *ObjectMgr) AccessWithTimeWeight(num int, rethashes *[]dbagent.RemoverbleObj) bool {
	hashes := []dbagent.RemoverbleObj{}
	ret := c.db.GetTransactionwithTimeWeight(num, &hashes)
	if !ret {
		return false
	}

	for _, hash := range hashes {
		if hash.HashType == 0 {
			var bh blockchain.BlockHeader
			if c.db.GetBlockHeader(hash.Hash, &bh) == 0 {
				*rethashes = append(*rethashes, hash)
				ret = true
			}
		} else {
			var tr blockchain.Transaction
			if c.db.GetTransaction(hash.Hash, &tr) == 0 {
				*rethashes = append(*rethashes, hash)
				ret = true
			}
		}
	}
	return ret
}

func NewObjMgr(db dbagent.DBAgent) *ObjectMgr {
	om := ObjectMgr{db}
	return &om
}
