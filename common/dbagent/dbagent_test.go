package dbagent

import (
	"encoding/hex"
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

	var trs []*blockchain.Transaction
	sss := []string{"11111111111111111111", "22222222222222222222222", "333333333333333333"}
	for i := 0; i < 3; i++ {
		s := sss[i]
		tr := blockchain.CreateTransaction([]byte(s))
		assert.Equal(t, []byte(s), tr.Data)
		trs = append(trs, tr)
	}

	b1 := blockchain.CreateBlock(trs, nil)

	dba.AddBlock(b1)
	hash := hex.EncodeToString(b1.Header.Hash)
	b2 := blockchain.Block{}
	dba.GetBlock(hash, &b2)
	assert.Equal(t, b1.Header, b2.Header)
	for i := 0; i < len(b2.Transactions); i++ {
		assert.Equal(t, b1.Transactions[i], b2.Transactions[i])
	}

	dba.AddBlock(&b2)
	dba.ShowAllObjets()

	dba.Close()
	os.Remove(path)
}
