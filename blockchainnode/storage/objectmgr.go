package storage

import (
	"log"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dbagent"
)

type ObjectMgr struct {
	db dbagent.DBAgent
}

func (c *ObjectMgr) DeleteNoAccedObjects() {
	c.db.DeleteNoAccedObjects()
}

func (c *ObjectMgr) AccessWithUniform(num int, rethashes *[]dbagent.RemoverbleObj) bool {
	hashes := []dbagent.RemoverbleObj{}
	ret := c.db.GetTransactionwithUniform(num, &hashes)
	if !ret {
		return false
	}

	cnt := 0
	for _, hash := range hashes {
		if hash.HashType == 0 {
			var bh blockchain.BlockHeader
			if c.db.GetBlockHeader(hash.Hash, &bh) == 0 {
				*rethashes = append(*rethashes, hash)
				ret = true
			} else {
				cnt++ //count if local access
			}
		} else {
			var tr blockchain.Transaction
			if c.db.GetTransaction(hash.Hash, &tr) == 0 {
				*rethashes = append(*rethashes, hash)
				ret = true
			} else {
				cnt++ //count if local access
			}
		}
	}
	c.db.UpdateDBNetworkQuery(0, 0, cnt) // local access
	log.Printf("===> number of gen : %v, %v", len(hashes), cnt)
	return ret
}

func (c *ObjectMgr) AccessWithExponential(num int, rethashes *[]dbagent.RemoverbleObj) bool {
	hashes := []dbagent.RemoverbleObj{}
	ret := c.db.GetTransactionwithExponential(num, &hashes)
	if !ret {
		return false
	}

	cnt := 0
	for _, hash := range hashes {
		if hash.HashType == 0 {
			var bh blockchain.BlockHeader
			if c.db.GetBlockHeader(hash.Hash, &bh) == 0 {
				*rethashes = append(*rethashes, hash)
				ret = true
			} else {
				cnt++ //count if local access
			}
		} else {
			var tr blockchain.Transaction
			if c.db.GetTransaction(hash.Hash, &tr) == 0 {
				*rethashes = append(*rethashes, hash)
				ret = true
			} else {
				cnt++ //count if local access
			}
		}
	}
	c.db.UpdateDBNetworkQuery(0, 0, cnt) // local
	return ret
}

func NewObjMgr(db dbagent.DBAgent) *ObjectMgr {
	om := ObjectMgr{db}
	return &om
}
