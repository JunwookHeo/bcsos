package main

import (
	"log"

	"github.com/junwookheo/bcsos/blockchain/core"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func main() {
	log.Println("Start blockchain service")
	core.Start()
}
