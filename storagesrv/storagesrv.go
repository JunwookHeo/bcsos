package main

import (
	"log"

	"github.com/junwookheo/bcsos/common/bcapi"
	"github.com/junwookheo/bcsos/storagesrv/storage"
	"github.com/junwookheo/bcsos/storagesrv/testmgrcli"
)

const DB_PATH = "./bc_storagesrv.db"

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	log.Println("Start Storage Service")

	bcapi.InitBC(DB_PATH)
	go storage.Start()
	go testmgrcli.StartTMC()

	select {}
}
