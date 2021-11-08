package testmgrsrv

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"github.com/junwookheo/bcsos/common/serial"
	"github.com/junwookheo/bcsos/common/shareddata"
)

const PORT = 8082

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

func clientNotifyHandler(w http.ResponseWriter, r *http.Request) {
	var node shareddata.TestNodeInfo

	json.NewDecoder(r.Body).Decode(&node)
	log.Printf("From client : %v", node)

	enc := json.NewEncoder(w)
	w.Header().Set("Content-Type", "application/json")
	hash := sha256.Sum256(serial.Serialize(node))
	node.AddrHash = hex.EncodeToString(hash[:])
	enc.Encode(node)
}

// Our fake service.
// This could be a HTTP/TCP service or whatever you want.
func startService() {
	http.HandleFunc("/clientNotify", clientNotifyHandler)

	http.HandleFunc("/ws", wsEndpoint)

	log.Println("starting http service...")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", PORT), nil); err != nil {
		log.Fatal(err)
	}
}

func StartTMS() {
	// Start out http service
	go startService()

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
	// Sleep forever
	select {}
}
