package bcdummy

import (
	"log"
	"testing"

	"github.com/junwookheo/bcsos/common/bcapi"
	"github.com/junwookheo/bcsos/common/dbagent"
)

const DB_PAT_TEST = "../bc_dummy.db"

func TestDBAgent(t *testing.T) {
	bcapi.InitBC(DB_PAT_TEST)
	bcapi.ShowBlockChain()
	bcapi.GetLatestHash()
	status := dbagent.DBStatus{}
	bcapi.GetDBStatus(&status)
	log.Printf("DB Status : %v", status)
}
