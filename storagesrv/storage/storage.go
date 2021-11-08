package storage

import (
	"github.com/junwookheo/bcsos/common/bcapi"
	"github.com/junwookheo/bcsos/storagesrv/network"
)

const DB_PATH = "./bc_storagesrv.db"

func Start() {
	bcapi.InitBC(DB_PATH)
	network.Start()
}

func Stop() {
	bcapi.Close()
}
