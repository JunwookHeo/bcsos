package main

import (
	"log"
	"os"
	"os/signal"

	"github.com/grandcat/zeroconf"
	"github.com/junwookheo/bcsos/blockchainsim/testmgrsrv"
)

const DB_PATH = "./bc_dummy.db"
const PORT = 8082

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	log.Println("Start blockchain simulator")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Reset()

	s := testmgrsrv.NewHandler(DB_PATH)
	go s.StartService(PORT)
	//go bcdummy.Start()

	// Extra information about our service
	meta := []string{
		"version=0.1.0",
		"bctestmgr",
	}

	service, err := zeroconf.Register(
		"bcsos-tms",       // service instance name
		"_omxremote._tcp", // service type and protocl
		"local.",          // service domain
		PORT,              // service port
		meta,              // service metadata
		nil,               // register on all network interfaces
	)

	if err != nil {
		log.Fatal(err)
	}

	defer service.Shutdown()

	<-interrupt
	log.Println("interrupt")
}
