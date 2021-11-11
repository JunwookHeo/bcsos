package storage

import (
	"fmt"
	"log"
	"net/http"
	"sort"
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

func (h *Handler) sortbyTypeNeighbourMap() *[]dtype.NodeInfo {
	var nodes []dtype.NodeInfo
	for _, n := range h.nm.Neighbours {
		if h.local.Type < n.Type {
			nodes = append(nodes, n)
		}
	}
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Type < nodes[j].Type
	})

	log.Printf("Sorted Heighbours %v, %v", h.local.Type, nodes)
	return &nodes
}

// newBlockHandler is called when a new block is received from miners
// When a node receive this, it stores the block on its local db
// Request : a new block
// Response : none
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

// getTransactionHandler is called when transaction query from other nodes is received
// if the node does not have the transaction, the node will query it to other nodes with highr SC
// Request : hash of transaction
// Response : transaction
func (h *Handler) getTransactionHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("getTransactionHandler transaction error : ", err)
		return
	}
	defer ws.Close()

	var hash string
	if err := ws.ReadJSON(&hash); err != nil {
		log.Printf("Read json error : %v", err)
	}

	var transaction blockchain.Transaction
	if h.db.GetTransaction(hash, &transaction) == 0 {
		// TODO:
		log.Printf("Not having it, so request the transaction to other node")
	}

	ws.WriteJSON(transaction)
}

func (h *Handler) getTransactionQuery(hash string) *blockchain.Transaction {
	queryTransaction := func(ip string, port int, hash string) *blockchain.Transaction {
		url := fmt.Sprintf("ws://%v:%v/version", ip, port)
		log.Printf("getTransactionQuery : %v", url)

		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("getTransactionQuery Dial error : %v", err)
			return nil
		}
		defer ws.Close()

		if err := ws.WriteJSON(hash); err != nil {
			log.Printf("Write json error : %v", err)
			return nil
		}

		var transaction blockchain.Transaction
		if err := ws.ReadJSON(&transaction); err != nil {
			log.Printf("Read json error : %v", err)
			return nil
		}

		return &transaction
	}

	neighs := h.sortbyTypeNeighbourMap()
	for _, node := range *neighs {
		transaction := queryTransaction(node.IP, node.Port, hash)
		if transaction != nil {
			return transaction
		}
	}

	return nil
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

	var version dtype.Version
	if err := ws.ReadJSON(&version); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}
	log.Printf("receive version : %v", version)

	var nodes []dtype.NodeInfo
	for _, n := range h.nm.Neighbours {
		nodes = append(nodes, n)
	}
	if err := ws.WriteJSON(nodes); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}

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

func NewHandler(path string, local dtype.NodeInfo) *Handler {
	m := mux.NewRouter()
	h := &Handler{
		Handler: m,
		db:      dbagent.NewDBAgent(path),
		sim:     dtype.Simulator{IP: "", Port: 0},
		local:   local,
		tmc:     nil,
		nm:      nil,
	}

	h.tmc = testmgrcli.NewTMC(h.db, &h.sim, &h.local)
	h.nm = network.NewNodeMgr()

	m.Handle("/", http.FileServer(http.Dir("static")))
	m.HandleFunc("/newblock", h.newBlockHandler)
	m.HandleFunc("/gettransaction", h.getTransactionHandler)
	m.HandleFunc("/nodeinfo", h.nodeInfoHandler)
	m.HandleFunc("/version", h.versionHandler)
	return h
}
