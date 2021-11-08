package main

import (
	"log"

	"github.com/junwookheo/bcsos/storagesrv/storage"
	"github.com/junwookheo/bcsos/storagesrv/testmgrcli"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	log.Println("Start Storage Service")
	storage.Start()
	testmgrcli.StartTMC()

	select {}
}
