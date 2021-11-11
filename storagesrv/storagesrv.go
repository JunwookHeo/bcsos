package main

import (
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/serial"
	"github.com/junwookheo/bcsos/storagesrv/storage"
)

var db_path string = "./bc_dev.db"

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (port int, err error) {
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
}

func flagParse() (string, string, int) {
	pmode := flag.String("mode", "dev", "Operation mode : dev or pan")
	ptype := flag.String("type", "0", "Storage class : 0 to 4")
	pport := flag.Int("port", 0, "Port number of local if 0, it will use a free port")
	flag.Parse()
	if *pmode == "pan" && *pport == 0 {
		log.Panicf("This is not allowed : %v, %v", *pmode, *pport)
	}
	return *pmode, *ptype, *pport
}

func GetLocalAddress() dtype.NodeInfo {
	var err error
	local := dtype.NodeInfo{Mode: "normal", Type: "", IP: "", Port: 0, Hash: ""}
	mode, stype, port := flagParse()
	local.Mode = mode
	local.Type = stype
	if mode == "dev" && port == 0 {
		port, err = getFreePort()
		if err != nil {
			log.Panicf("Get free port error : %v", err)
		}
	} else {
		db_path = fmt.Sprintf("./db_nodes/%v.db", port)
	}

	local.Port = port
	hash := sha256.Sum256(serial.Serialize(local))
	local.Hash = hex.EncodeToString(hash[:])
	return local
}

func main() {
	log.Println("Start Storage Service")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Reset()

	local := GetLocalAddress()
	s := storage.NewHandler(db_path, local)
	s.UpdateNeighbourNodes()

	log.Printf("Server start : %v", local.Port)
	go http.ListenAndServe(fmt.Sprintf(":%v", local.Port), s.Handler)
	//go http.ListenAndServe(":8080", s.Handler)

	<-interrupt
	log.Println("interrupt")
}
