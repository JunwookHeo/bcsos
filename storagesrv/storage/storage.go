package storage

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dbagent"
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

func (h *Handler) nodeInfoHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("newBlockHandler", err)
		return
	}
	defer ws.Close()
	h.tmc.NodeInfoHandler(ws, w, r)
}

func NewHandler(path string, port int) *Handler {
	m := mux.NewRouter()
	h := &Handler{Handler: m, db: dbagent.NewDBAgent(path), tmc: nil}
	h.tmc = testmgrcli.NewTMC(h.db, port)

	m.Handle("/", http.FileServer(http.Dir("static")))
	m.HandleFunc("/newblock", h.newBlockHandler)
	m.HandleFunc("/nodeinfo", h.nodeInfoHandler)
	return h
}
