package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"

	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/serial"
	"github.com/junwookheo/bcsos/storagesrv/storage"
)

const DB_PATH = "./bc_storagesrv.db"

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

func GetLocalAddress() dtype.NodeInfo {
	var port int
	var err error
	local := dtype.NodeInfo{Mode: "normal", Type: "", IP: "", Port: 0, Hash: ""}
	sport := os.Getenv("BCPORT")
	mode := os.Getenv("BCMODE")
	if mode != "" {
		local.Mode = mode
	}

	if sport == "" {
		port, err = getFreePort()
	} else {
		port, err = strconv.Atoi(sport)
	}
	if err != nil {
		log.Panicf("Get free port error : %v", err)
		return local
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
	s := storage.NewHandler(DB_PATH, local)
	s.UpdateNeighbourNodes()

	log.Printf("Server start : %v", local.Port)
	go http.ListenAndServe(fmt.Sprintf(":%v", local.Port), s.Handler)
	//go http.ListenAndServe(":8080", s.Handler)

	<-interrupt
	log.Println("interrupt")
}
