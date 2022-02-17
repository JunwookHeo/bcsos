package bcdummy

import (
	"log"
	"testing"

	"github.com/junwookheo/bcsos/common/dbagent"
)

const PATH_TEST = "../iotdata/IoT_normal_fridge_1.log"
const DB_PATH_TEST = "../bc_dummy.db"

func TestLoadFromJson(t *testing.T) {
	// msg := make(chan string)
	// go LoadRawdata(PATH_TEST, msg)
	// for {
	// 	tr := <-msg
	// 	assert.NotEmpty(t, tr)
	// 	log.Printf("%v", tr)
	// 	time.Sleep(1 * time.Second)
	// }
}

func TestDBAgent(t *testing.T) {
	dba := dbagent.NewDBAgent(DB_PATH_TEST)
	dba.ShowAllObjets()
	dba.GetLatestBlockHash()
	status := dba.GetDBStatus()
	log.Printf("DB Status : %v", status)
}
