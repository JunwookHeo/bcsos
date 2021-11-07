package storage

import (
	"log"

	"github.com/junwookheo/bcsos/common/bcapi"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/serial"
)

const DB_PATH = "./bc_storagesrv.db"

type StorageSrv struct {
	rxmsg chan []byte
	txmsg chan []byte
}

var storagesrv StorageSrv = StorageSrv{make(chan []byte), make(chan []byte)}

func Start() {
	bcapi.InitBC(DB_PATH)
	for {
		msg := <-storagesrv.rxmsg
		HandleAddBlock(msg)
	}
}

func HandleAddBlock(d []byte) {
	b := blockchain.Block{}
	serial.Deserialize(d, &b)
	log.Printf("%v", b)
	bcapi.AddBlock(&b)

	for _, tr := range b.Transactions {
		log.Printf("<===%s", tr.Data)
	}
}

func AddBlock(d []byte) {
	storagesrv.rxmsg <- d
}

func Stop() {
	bcapi.Close()
}
