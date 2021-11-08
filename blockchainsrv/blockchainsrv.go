package main

import (
	"log"

	"github.com/junwookheo/bcsos/blockchainsrv/testmgrsrv"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	log.Println("Start blockchain service")
	testmgrsrv.StartTMS()

	// Sleep forever
	select {}
}
