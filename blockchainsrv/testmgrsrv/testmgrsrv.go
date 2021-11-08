package testmgrsrv

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"github.com/junwookheo/bcsos/common/bcapi"
)

const DB_PATH = "./bc_dummy.db"
const PORT = 8081

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	defer ws.Close()

	if err != nil {
		log.Println("endpoint", err)
		return
	}

	log.Println("Client Connected!!!")
	reader(ws)
}

func reader(conn *websocket.Conn) {
	for {
		mt, p, err := conn.ReadMessage()
		if err != nil {
			log.Println("reader", err)
			return
		}

		log.Println(string(p))

		if err := conn.WriteMessage(mt, p); err != nil {
			log.Println("writer", err)
			return
		}
	}
}

// Our fake service.
// This could be a HTTP/TCP service or whatever you want.
func startService() {
	http.HandleFunc("/", func(rw http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(rw, "Hello world!")
	})

	http.HandleFunc("/ws", wsEndpoint)

	log.Println("starting http service...")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil); err != nil {
		log.Fatal(err)
	}
}

func StartTMS() {
	bcapi.InitBC(DB_PATH)
	// Start out http service
	go startService()
	//go bcdummy.Start()

	// Extra information about our service
	meta := []string{
		"version=0.1.0",
		"hello=world",
	}

	service, err := zeroconf.Register(
		"awesome-sauce",   // service instance name
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
}
