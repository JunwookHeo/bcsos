package bcapi

import (
	"encoding/hex"
	"log"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dbagent"
)

var dba dbagent.DBAgent

func InitBC(path string) {
	dba = dbagent.NewDBAgent(path)
	//log.Printf("InitBC : %v", dba)
}

func CreateGenesis() *blockchain.Block {
	tr := blockchain.Transaction{Hash: nil, Data: []byte("This is Genesis Block")}
	return blockchain.Genesis(&tr)
}

func AddBlock(b *blockchain.Block) {
	dba.AddBlock(b)
}

func GetLatestHash() []byte {
	hash := dba.GetLatestBlockHash()
	lh, err := hex.DecodeString(hash)
	if err != nil {
		log.Panicf("Get LatestHash error : %v", err)
		return nil
	}

	return lh
}

func ShowBlockChain() {
	dba.ShowAllObjets()
}

func Close() {
	dba.Close()
}
