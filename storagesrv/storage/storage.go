package storage

import (
	"time"

	"github.com/junwookheo/bcsos/common/bcapi"
	"github.com/junwookheo/bcsos/storagesrv/network"
)

const DB_PATH = "./bc_storagesrv.db"

type STGMSG struct {
	cmd  int32
	data []byte
}

func Start() {
	bcapi.InitBC(DB_PATH)
	network.Start()
	for {
		time.Sleep(10 * time.Second)
		// msg := <-storagesrv.rxmsg
		// msgHandler(msg)
	}
}

func Stop() {
	bcapi.Close()
}
