package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"

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

func main() {
	log.Println("Start Storage Service")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Reset()

	port, err := getFreePort()
	if err != nil {
		log.Panicf("Get free port error : %v", err)
	}

	s := storage.NewHandler(DB_PATH, port)
	s.UpdateNeighbourNodes()

	go http.ListenAndServe(fmt.Sprintf(":%v", port), s.Handler)
	//go http.ListenAndServe(":8080", s.Handler)
	log.Printf("Server start : %v", port)

	<-interrupt
	log.Println("interrupt")
}
