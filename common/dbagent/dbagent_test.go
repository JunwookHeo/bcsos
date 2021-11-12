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
	dba := NewDBAgent(path, 0)
	assert.FileExists(t, path)
	dba.Close()
	os.Remove(path)
}

func TestDBSqliteAdd(t *testing.T) {
	path := "test2.db"
	dba := NewDBAgent(path, 0)
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

	log.Println(dba.GetDBDataSize())
	dba.RemoveObject(hex.EncodeToString(b2.Transactions[0].Hash))
	dba.GetDBStatus(&status)
	log.Println(dba.GetDBDataSize())
	dba.RemoveObject(hex.EncodeToString(b2.Transactions[1].Hash))
	dba.GetDBStatus(&status)
	log.Println(dba.GetDBDataSize())
	dba.RemoveObject(hex.EncodeToString(b2.Transactions[2].Hash))
	dba.GetDBStatus(&status)
	log.Println(dba.GetDBDataSize())
	log.Println(status)

	//dba.ShowAllObjets()
	dba.Close()
	os.Remove(path)
}

// func TestDBSqliteRandom(t *testing.T) {
// 	dba := NewDBAgent("../../storagesrv/bc_storagesrv.db", 0)
// 	hashes := dba.GetTransactionwithRandom()
// 	for _, h := range hashes {
// 		var tr blockchain.Transaction
// 		dba.GetTransaction(h, &tr)
// 		assert.Equal(t, hex.EncodeToString(tr.Hash), h)
// 	}
// 	for i := 0; i < 100; i++ {
// 		expdist := rand.ExpFloat64() / 0.5
// 		log.Printf(", %v", expdist)
// 	}
// }

func TestDBAgentReplicatoin(t *testing.T) {
	path := "../../storagesrv/bc_dev.db"
	if !assert.FileExistsf(t, path, "no file exist"){
		return
	}

	dba := NewDBAgent(path, 0)
	dba.ShowAllObjets()
	dba.GetLatestBlockHash()
	status := DBStatus{}
	dba.GetDBStatus(&status)
	log.Printf("DB Status : %v", status)
	dba.GetDBDataSize()

	// hash1, _ := hex.DecodeString("0007c6e53cff577e2b87ea385541acb3872d10874eb3f2cc438b37c5f0683f93")
	// obj1 := blockchain.Block{}
	// obj1.Header.Hash = hash1
	// dba.AddBlock(&obj1)

	// hash2, _ := hex.DecodeString("10fdff4e973df14173d6ebb66605717cd5fa46f2f78861b590de549a1ffefcd5")
	// obj2 := blockchain.Transaction{}
	// obj2.Hash = hash2
	// dba.AddTransaction(&obj2)
}
