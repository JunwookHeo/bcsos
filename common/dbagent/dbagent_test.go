package dbagent

import (
	"encoding/hex"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"testing"

	"github.com/junwookheo/bcsos/common/bitcoin"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/wallet"
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
	wallet_path := "./wallet_test.wallet"
	dba := NewDBAgent(path)
	assert.FileExists(t, path)
	w := wallet.NewWallet(wallet_path)

	crbl := func(pre string, height int) *blockchain.Block {
		var trs []*blockchain.Transaction
		sss := []string{pre + "11111111111111111111", pre + "22222222222222222222222", pre + "333333333333333333"}
		for i := 0; i < 3; i++ {
			s := sss[i]
			tr := blockchain.CreateTransaction(w, []byte(s))
			assert.Equal(t, []byte(s), tr.Data)
			trs = append(trs, tr)
		}

		return blockchain.CreateBlock(trs, nil, height)
	}

	dba.GetLatestBlockHash()
	b1 := crbl("aaaaa-", 0)
	dba.AddNewBlock(b1)
	status := dba.GetDBStatus()
	log.Println(status)
	phash, height := dba.GetLatestBlockHash()
	log.Println(phash, height)

	hash := hex.EncodeToString(b1.Header.Hash)
	b2 := blockchain.Block{}
	dba.GetBlock(hash, &b2)
	assert.Equal(t, b1.Header, b2.Header)
	for i := 0; i < len(b2.Transactions); i++ {
		assert.Equal(t, b1.Transactions[i], b2.Transactions[i])
		log.Printf("==> %v, %v", b1.Transactions[i].Signature, b2.Transactions[i].Signature)
	}

	b1 = crbl("bbbbb-", 1)
	dba.AddNewBlock(b1)
	status = dba.GetDBStatus()
	log.Println(status)
	phash, height = dba.GetLatestBlockHash()
	log.Println(phash, height)
	//dba.ShowAllObjets()

	log.Println(dba.GetDBDataSize())
	dba.RemoveObject(hex.EncodeToString(b2.Transactions[0].Hash))
	status = dba.GetDBStatus()
	log.Println(status)
	log.Println(dba.GetDBDataSize())
	dba.RemoveObject(hex.EncodeToString(b2.Transactions[1].Hash))
	status = dba.GetDBStatus()
	log.Println(status)
	log.Println(dba.GetDBDataSize())
	dba.RemoveObject(hex.EncodeToString(b2.Transactions[2].Hash))
	status = dba.GetDBStatus()
	log.Println(dba.GetDBDataSize())
	log.Println(status)

	//dba.ShowAllObjets()
	dba.Close()
	os.Remove(path)
	os.Remove(wallet_path)
}

func TestDBSqliteRandom(t *testing.T) {
	dba := NewDBAgent("../../storagesrv/bc_dev.db")
	//hashes := dba.GetTransactionwithUniform(50)
	dba.DeleteNoAccedObjects()
	// csvfile, _ := os.Create("../../rs.csv")
	// csvwriter := csv.NewWriter(csvfile)

	// row := []string{"Hash", "Time"}
	// csvwriter.Write(row)
	// for _, h := range hashes {
	// 	var tr blockchain.Transaction
	// 	dba.GetTransaction(h, &tr)
	// 	log.Printf("timestamp : %v", tr.Timestamp)
	// 	row := []string{h, fmt.Sprintf("%v", tr.Timestamp)}
	// 	csvwriter.Write(row)
	// 	assert.Equal(t, h, hex.EncodeToString(tr.Hash))
	// }

	// csvwriter.Flush()
	// csvfile.Close()
}

func TestDBAgentReplicatoin(t *testing.T) {
	// path := "../../storagesrv/bc_dev.db"
	// if !assert.FileExistsf(t, path, "no file exist"){
	// 	return
	// }

	// dba := NewDBAgent(path, 0)
	// dba.ShowAllObjets()
	// dba.GetLatestBlockHash()
	// status := DBStatus{}
	// dba.GetDBStatus(&status)
	// log.Printf("DB Status : %v", status)
	// dba.GetDBDataSize()

	// hash1, _ := hex.DecodeString("0007c6e53cff577e2b87ea385541acb3872d10874eb3f2cc438b37c5f0683f93")
	// obj1 := blockchain.Block{}
	// obj1.Header.Hash = hash1
	// dba.AddBlock(&obj1)

	// hash2, _ := hex.DecodeString("10fdff4e973df14173d6ebb66605717cd5fa46f2f78861b590de549a1ffefcd5")
	// obj2 := blockchain.Transaction{}
	// obj2.Hash = hash2
	// dba.AddTransaction(&obj2)
}

// func TestDBAgentProofStorage(t *testing.T) {
// 	path := "../../blockchainnode/db_nodes/7001.db"
// 	if !assert.FileExistsf(t, path, "no file exist") {
// 		return
// 	}

// 	dba := NewDBAgent(path)
// 	tt, _ := hex.DecodeString("0ab51095bf5314967f964422f91fc6b39e7761103875eeafebe1cef430d9f531")
// 	var tid [32]byte
// 	copy(tid[:], tt)
// 	ts := int64(1661228540001266000)

// 	dba.ProofStorage(tid, ts, 0)
// }

// func TestDBAgentTestQuery(t *testing.T) {
// 	path := "../../blockchainnode/db_nodes/7001.db"
// 	if !assert.FileExistsf(t, path, "no file exist") {
// 		return
// 	}

// 	dba := NewDBAgent(path)

// 	dba.ProofStorage2()
// }

func TestBtcDBAgent(t *testing.T) {
	path := "../../blockchainnode/db_nodes/7001.db" + ".blocks"
	fenc2 := "0000000000000000000027895a1788f2339b84a4f365c0accb95be3d406726fb"
	enc2, err := ioutil.ReadFile(filepath.Join(path, fenc2))
	if err != nil {
		log.Panicf("Read 2 block err : %v", err)
		return
	}
	fenc1 := "00000000000000000005a72e37590b534da3667ae2da19979e28a1229ebf94f0"
	enc1, err := ioutil.ReadFile(filepath.Join(path, fenc1))
	if err != nil {
		log.Panicf("Read 2 block err : %v", err)
		return
	}

	sb := hex.EncodeToString(DecryptXorWithVariableLength(enc1, enc2))
	block := bitcoin.NewBlock()
	rb := bitcoin.NewRawBlock(sb)

	block.Header.Version = rb.ReadUint32()
	block.Header.PreHash = rb.ReverseBuf(rb.ReadBytes(32))
	block.Header.MerkelRoot = rb.ReverseBuf(rb.ReadBytes(32))
	block.Header.Timestamp = rb.ReadUint32()
	block.Header.Difficulty = rb.ReadUint32()
	block.Header.Nonce = rb.ReadUint32()
	log.Printf("Header : %x", block.Header)
	block.SetHash(rb.GetRawBytes(0, 80))
	log.Printf("Hash : %v", block.GetHashString())

	txcount := rb.ReadVariant()
	log.Printf("Tx Count : %d", txcount)

}
