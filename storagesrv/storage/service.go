package storage

import (
	"log"

	"github.com/junwookheo/bcsos/common/bcapi"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/serial"
)

const DB_PATH = "./bc_storagesrv.db"

type STGMSG struct {
	cmd  int32
	data []byte
}

type StorageSrv struct {
	rxmsg chan STGMSG
	txmsg chan STGMSG
}

var storagesrv StorageSrv = StorageSrv{make(chan STGMSG), make(chan STGMSG)}

func Start() {
	bcapi.InitBC(DB_PATH)
	for {
		msg := <-storagesrv.rxmsg
		msgHandler(msg)
	}
}

func msgHandler(sm STGMSG) {
	switch sm.cmd {
	case int32(config.NEWBLOCK):
		HandleAddBlock(sm.data)
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
	var sm STGMSG = STGMSG{int32(config.NEWBLOCK), d}
	storagesrv.rxmsg <- sm
}

func Stop() {
	bcapi.Close()
}
