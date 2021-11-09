package testmgrsrv

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/blockchainsim/bcdummy"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/serial"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

type Handler struct {
	http.Handler
	db      dbagent.DBAgent
	Nodes   []dtype.NodeInfo
	BCDummy *bcdummy.Handler
	Ready   chan bool
}

func (h *Handler) resisterHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("resisterHandler", err)
		return
	}
	defer ws.Close()

	var node dtype.NodeInfo
	if err := ws.ReadJSON(&node); err != nil {
		log.Printf("Read json error : %v", err)
	}

	ip := strings.Split(ws.RemoteAddr().String(), ":")
	node.IP = ip[0]
	hash := sha256.Sum256(serial.Serialize(node))
	node.Hash = hex.EncodeToString(hash[:])

	h.Nodes = append(h.Nodes, node)

	ws.WriteJSON(node)
	log.Printf("From client : %v", node)
	//TODO: n node connection
	h.Ready <- true
}

func (h *Handler) StartService(port int) {
	http.HandleFunc("/resister", h.resisterHandler)

	log.Println("starting http service...")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), h.Handler); err != nil {
		log.Fatal(err)
	}
}

func (h *Handler) StartDummy() {
	// Wait until clients join
	for i := 0; i < 1; i++ {
		<-h.Ready
	}

	for _, n := range h.Nodes {
		log.Printf("Test Start : %v", n)
	}
	h.BCDummy.Start()
}

func NewHandler(path string) *Handler {
	m := mux.NewRouter()
	h := &Handler{
		Handler: m,
		db:      dbagent.NewDBAgent(path),
		Nodes:   nil,
		BCDummy: nil,
		Ready:   make(chan bool),
	}

	m.HandleFunc("/resister", h.resisterHandler)

	h.BCDummy = bcdummy.NewBCDummy(h.db, &h.Nodes)
	go h.StartDummy()

	return h
}
