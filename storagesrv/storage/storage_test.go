package storage

import (
	"log"
	"testing"

	"github.com/junwookheo/bcsos/common/dbagent"
)

const DB_PATH_TEST = "../bc_storagesrv.db"

func TestDBAgent(t *testing.T) {
	dba := dbagent.NewDBAgent(DB_PATH_TEST, 0)
	dba.ShowAllObjets()
	dba.GetLatestBlockHash()
	status := dba.GetDBStatus()
	log.Printf("DB Status : %v", status)
}
