package storage

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/storagesrv/network"
	"github.com/junwookheo/bcsos/storagesrv/testmgrcli"
)

type Handler struct {
	http.Handler
	db    dbagent.DBAgent
	sim   dtype.Simulator
	local dtype.NodeInfo
	tmc   *testmgrcli.TestMgrCli
	nm    *network.NodeMgr
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
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

func (h *Handler) versionHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("versionHandler", err)
		return
	}
	defer ws.Close()
	h.nm.VersionHandler(ws, w, r)
}

func (h *Handler) UpdateNeighbourNodes() {
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			if h.sim.IP != "" && h.sim.Port != 0 && h.local.Hash != "" {
				h.nm.Update(h.sim, h.local)
			}
		}
	}()
}

func NewHandler(path string, port int) *Handler {
	m := mux.NewRouter()
	h := &Handler{
		Handler: m,
		db:      dbagent.NewDBAgent(path),
		sim:     dtype.Simulator{IP: "", Port: 0},
		local:   dtype.NodeInfo{Type: "", IP: "", Port: port, Hash: ""},
		tmc:     nil,
		nm:      nil,
	}
	h.tmc = testmgrcli.NewTMC(h.db, &h.sim, &h.local)
	h.nm = network.NewNodeMgr()

	m.Handle("/", http.FileServer(http.Dir("static")))
	m.HandleFunc("/newblock", h.newBlockHandler)
	m.HandleFunc("/nodeinfo", h.nodeInfoHandler)
	m.HandleFunc("/version", h.versionHandler)
	return h
}
