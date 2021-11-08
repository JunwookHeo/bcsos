package main

import (
	"log"

	"github.com/junwookheo/bcsos/blockchainsrv/testmgrsrv"
	"github.com/junwookheo/bcsos/common/bcapi"
)

const DB_PATH = "./bc_dummy.db"

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	log.Println("Start blockchain service")
	bcapi.InitBC(DB_PATH)
	go testmgrsrv.StartTMS()
	//go bcdummy.Start()

	select {}
}
