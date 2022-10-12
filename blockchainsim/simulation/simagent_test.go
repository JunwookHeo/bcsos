package simulation

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"testing"

	"github.com/junwookheo/bcsos/common/dbagent"
)

// const PATH_TEST = "../iotdata/IoT_normal_fridge_1.log"
const PATH_TEST = "../../transactions.json"
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
	f, err := os.Open(PATH_TEST)
	if err != nil {
		log.Printf("Open error : %v", err)
		return
	}

	fscanner := bufio.NewScanner(f)
	for fscanner.Scan() {
		var rec map[string]interface{}
		err := json.Unmarshal([]byte(fscanner.Text()), &rec)
		if err != nil {
			panic(err)
		}

		log.Printf("===> %v", rec["hash"])
	}
}

func TestDBAgent(t *testing.T) {
	dba := dbagent.NewDBAgent(DB_PATH_TEST)
	dba.ShowAllObjets()
	dba.GetLatestBlockHash()
	status := dba.GetDBStatus()
	log.Printf("DB Status : %v", status)
}
