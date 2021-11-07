package core

import (
	"testing"

	"github.com/junwookheo/bcsos/common/bcapi"
)

const DB_PAT_TEST = "../bc_dummy.db"

func TestDBAgent(t *testing.T) {
	bcapi.InitBC(DB_PAT_TEST)
	bcapi.ShowBlockChain()
	bcapi.GetLatestHash()
}
