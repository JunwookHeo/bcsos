package main

import (
	"log"

	"github.com/junwookheo/bcsos/storagesrv/storage"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	log.Println("Start Storage Service")
	storage.Start()
}
