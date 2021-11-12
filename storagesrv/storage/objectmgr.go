package storage

import (
	"log"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dbagent"
)

type ObjectMgr struct {
	db dbagent.DBAgent
}

func Read() {

}

func Write() {

}

func Delete() {

}

func Update() {

}

func (c *ObjectMgr) DeleteNoAccedObject() {
	c.db.DeleteNoAccedObject()
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

func (c *ObjectMgr) AccessWithTimeWeight() {
	hashes := c.db.GetTransactionwithTimeWeight()
	log.Printf("Time weight selected transactions %v", hashes)
}

func NewObjMgr(db dbagent.DBAgent) *ObjectMgr {
	om := ObjectMgr{db}
	return &om
}
