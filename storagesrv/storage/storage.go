package storage

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/storagesrv/network"
	"github.com/junwookheo/bcsos/storagesrv/testmgrcli"
)

type Handler struct {
	http.Handler
	db  dbagent.DBAgent
	tmc *testmgrcli.TestMgrCli
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
}

func Start() {
	network.Start()
}

func (h *Handler) Stop() {
	h.db.Close()
}

func (h *Handler) newBlockHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("newBlockHandler", err)
		return
	}
	defer ws.Close()

	var block blockchain.Block
	if err := ws.ReadJSON(&block); err != nil {
		log.Printf("Read json error : %v", err)
	}

	h.db.AddBlock(&block)
	// ws.WriteJSON(block)
	for _, t := range block.Transactions {
		log.Printf("From client : %s", t.Data)
	}
}

func NewHandler(path string, port int) *Handler {
	m := mux.NewRouter()
	h := &Handler{Handler: m, db: dbagent.NewDBAgent(path), tmc: testmgrcli.NewTMC(port)}

	m.HandleFunc("/newblock", h.newBlockHandler)
	return h
}
