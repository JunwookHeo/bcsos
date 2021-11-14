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

func (c *ObjectMgr) AccessWithRandom(num int) *dbagent.RemoverbleObj {
	hashes := c.db.GetTransactionwithRandom(num)
	rethashes := dbagent.RemoverbleObj{}
	for _, hash := range hashes.TransactionHash {
		var tr blockchain.Transaction
		if c.db.GetTransaction(hash, &tr) == 0 {
			rethashes.TransactionHash = append(rethashes.TransactionHash, hash)
		}
	}

	for _, hash := range hashes.BlockHeaderHash {
		var bh blockchain.BlockHeader
		if c.db.GetBlockHeader(hash, &bh) == 0 {
			rethashes.BlockHeaderHash = append(rethashes.BlockHeaderHash, hash)
		}
	}
	return &rethashes
}

func (c *ObjectMgr) AccessWithTimeWeight(num int) *dbagent.RemoverbleObj {
	hashes := c.db.GetTransactionwithTimeWeight(num)
	rethashes := dbagent.RemoverbleObj{}
	for _, hash := range hashes.TransactionHash {
		var tr blockchain.Transaction
		if c.db.GetTransaction(hash, &tr) == 0 {
			rethashes.TransactionHash = append(rethashes.TransactionHash, hash)
		}
	}
	for _, hash := range hashes.BlockHeaderHash {
		var bh blockchain.BlockHeader
		if c.db.GetBlockHeader(hash, &bh) == 0 {
			rethashes.BlockHeaderHash = append(rethashes.BlockHeaderHash, hash)
		}
	}
	return &rethashes
}

func NewObjMgr(db dbagent.DBAgent) *ObjectMgr {
	om := ObjectMgr{db}
	return &om
}
