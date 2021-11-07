package storage

import (
	"log"
	"testing"

	"github.com/junwookheo/bcsos/common/bcapi"
	"github.com/junwookheo/bcsos/common/dbagent"
)

const DB_PATH_TEST = "../bc_storagesrv.db"

func TestDBAgent(t *testing.T) {
	bcapi.InitBC(DB_PATH_TEST)
	bcapi.ShowBlockChain()
	bcapi.GetLatestHash()
	status := dbagent.DBStatus{}
	bcapi.GetDBStatus(&status)
	log.Printf("DB Status : %v", status)
}
