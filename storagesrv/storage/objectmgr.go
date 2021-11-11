package storage

import (
	"log"

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

func (c *ObjectMgr) AccessWithRandom() {
	hashes := c.db.GetTransactionwithRandom()
	log.Printf("Randomly selected transactions %v", hashes)
}

func (c *ObjectMgr) AccessWithTimeWeight() {
	hashes := c.db.GetTransactionwithTimeWeight()
	log.Printf("Time weight selected transactions %v", hashes)
}

func NewObjMgr(db dbagent.DBAgent) *ObjectMgr {
	om := ObjectMgr{db}
	return &om
}
