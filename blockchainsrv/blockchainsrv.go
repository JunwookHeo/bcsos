package main

import (
	"log"

	"github.com/junwookheo/bcsos/blockchainsrv/bcdummy"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	log.Println("Start blockchain service")
	bcdummy.Start()
}
