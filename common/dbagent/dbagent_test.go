package dbagent

import (
	"encoding/hex"
	"log"
	"os"
	"testing"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/stretchr/testify/assert"
)

func TestDBSqlite(t *testing.T) {
	path := "test1.db"
	dba := NewDBAgent(path)
	assert.FileExists(t, path)
	dba.Close()
	os.Remove(path)
}

func TestDBSqliteAdd(t *testing.T) {
	path := "test2.db"
	dba := NewDBAgent(path)
	assert.FileExists(t, path)

	crbl := func(pre string) *blockchain.Block {
		var trs []*blockchain.Transaction
		sss := []string{pre + "11111111111111111111", pre + "22222222222222222222222", pre + "333333333333333333"}
		for i := 0; i < 3; i++ {
			s := sss[i]
			tr := blockchain.CreateTransaction([]byte(s))
			assert.Equal(t, []byte(s), tr.Data)
			trs = append(trs, tr)
		}

		return blockchain.CreateBlock(trs, nil)
	}

	status := DBStatus{}

	dba.GetLatestBlockHash()
	b1 := crbl("aaaaa-")
	dba.AddBlock(b1)
	dba.GetDBStatus(&status)
	dba.GetLatestBlockHash()

	hash := hex.EncodeToString(b1.Header.Hash)
	b2 := blockchain.Block{}
	dba.GetBlock(hash, &b2)
	assert.Equal(t, b1.Header, b2.Header)
	for i := 0; i < len(b2.Transactions); i++ {
		assert.Equal(t, b1.Transactions[i], b2.Transactions[i])
	}

	b1 = crbl("bbbbb-")
	dba.AddBlock(b1)
	dba.GetDBStatus(&status)
	dba.GetLatestBlockHash()
	//dba.ShowAllObjets()

	log.Println(dba.GetDBSize())
	dba.RemoveObject(hex.EncodeToString(b2.Transactions[0].Hash))
	dba.GetDBStatus(&status)
	log.Println(dba.GetDBSize())
	dba.RemoveObject(hex.EncodeToString(b2.Transactions[1].Hash))
	dba.GetDBStatus(&status)
	log.Println(dba.GetDBSize())
	dba.RemoveObject(hex.EncodeToString(b2.Transactions[2].Hash))
	dba.GetDBStatus(&status)
	log.Println(dba.GetDBSize())
	log.Println(status)

	//dba.ShowAllObjets()
	dba.Close()
	os.Remove(path)
}
