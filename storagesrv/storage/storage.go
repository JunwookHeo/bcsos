package storage

import (
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/storagesrv/network"
	"github.com/junwookheo/bcsos/storagesrv/testmgrcli"
)

type Handler struct {
	http.Handler
	db    dbagent.DBAgent
	sim   dtype.NodeInfo
	local dtype.NodeInfo
	tmc   *testmgrcli.TestMgrCli
	nm    *network.NodeMgr
	om    *ObjectMgr
	mutex sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func (h *Handler) Stop() {
	h.db.Close()
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
	// for _, t := range block.Transactions {
	// 	log.Printf("From client : %s", t.Data)
	// }
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

	defer h.db.UpdateDBNetworkOverhead(1, 0)
	var hash string
	if err := ws.ReadJSON(&hash); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}

	transaction := blockchain.Transaction{}
	if h.db.GetTransaction(hash, &transaction) == 0 {
		// log.Printf("Not having it, so request the transaction to other node")
		if h.getTransactionQuery(hash, &transaction) {
			h.db.AddTransaction(&transaction)
		}
	}

	ws.WriteJSON(transaction)
}

// getTransactionQuery queries a transaction ot other nodes with highr Storage Class
// Request : hash of transaction
// Response : transaction
func (h *Handler) getTransactionQuery(hash string, tr *blockchain.Transaction) bool {
	queryTransaction := func(ip string, port int, hash string, tr *blockchain.Transaction) bool {
		url := fmt.Sprintf("ws://%v:%v/gettransaction", ip, port)
		//log.Printf("getTransactionQuery : %v", url)

		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("getTransactionQuery Dial error : %v", err)
			return false
		}
		defer ws.Close()

		// the number of query to other nodes
		defer h.db.UpdateDBNetworkOverhead(0, 1)

		if err := ws.WriteJSON(hash); err != nil {
			log.Printf("Write json error : %v", err)
			return false
		}

		if err := ws.ReadJSON(tr); err != nil {
			log.Printf("Read json error : %v", err)
			return false
		}

		return true
	}

	for i := h.local.SC + 1; i <= config.MAX_SC; i++ {
		var nodes [config.MAX_SC_PEER]dtype.NodeInfo
		if h.nm.GetSCNNodeListbyDistance(i, hash, &nodes) {
			for _, node := range nodes {
				if queryTransaction(node.IP, node.Port, hash, tr) {
					return true
				}
				//time.Sleep(time.Duration(200 * time.Microsecond.Seconds()))
				log.Printf("queryTransaction fail : query other nodes")
			}
		}
	}

	return false
}

// getBlockHeaderHandler is called when block header query from other nodes is received
// if the node does not have the block header, the node will query it to other nodes with highr SC
// Request : hash of block header
// Response : block header
func (h *Handler) getBlockHeaderHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("getBlockHeaderHandler transaction error : ", err)
		return
	}
	defer ws.Close()

	var hash string
	if err := ws.ReadJSON(&hash); err != nil {
		log.Printf("Read json error : %v", err)
	}

	bh := blockchain.BlockHeader{}
	if h.db.GetBlockHeader(hash, &bh) == 0 {
		//log.Printf("Not having it, so request the transaction to other node")
		if h.getBlockHeaderQuery(hash, &bh) {
			h.db.AddBlockHeader(string(bh.Hash), &bh)
		}
	}

	ws.WriteJSON(bh)
}

// getBlockHeaderQuery queries a block header ot other nodes with highr Storage Class
// Request : hash of block header
// Response : block header
func (h *Handler) getBlockHeaderQuery(hash string, bh *blockchain.BlockHeader) bool {
	queryBlockHeader := func(ip string, port int, hash string, bh *blockchain.BlockHeader) bool {
		url := fmt.Sprintf("ws://%v:%v/getblockheader", ip, port)
		//log.Printf("getBlockHeaderQuery : %v", url)

		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("getBlockHeaderQuery Dial error : %v", err)
			return false
		}
		defer ws.Close()

		if err := ws.WriteJSON(hash); err != nil {
			log.Printf("Write json error : %v", err)
			return false
		}

		if err := ws.ReadJSON(bh); err != nil {
			log.Printf("Read json error : %v", err)
			return false
		}

		return true
	}

	for i := h.local.SC + 1; i <= config.MAX_SC; i++ {
		var nodes [config.MAX_SC_PEER]dtype.NodeInfo
		if h.nm.GetSCNNodeListbyDistance(i, hash, &nodes) {
			for _, node := range nodes {
				if queryBlockHeader(node.IP, node.Port, hash, bh) {
					return true
				}
				//time.Sleep(time.Duration(100 * time.Microsecond.Seconds()))
				log.Printf("queryBlockHeader fail : query other nodes")
			}
		}
	}

	return false
}

// Response to web app with dbstatus information
// keep sending dbstatus to the web app
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

// Send response to connector with its local information
func (h *Handler) pingHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("pingHandler", err)
		return
	}
	defer ws.Close()

	peer := dtype.NodeInfo{}
	if err := ws.ReadJSON(&peer); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}
	//log.Printf("receive peer addr : %v", peer)

	// Received peer info so add it to peer list
	if peer.Hash != "" && peer.Hash != h.local.Hash {
		h.nm.AddNSCNNode(peer)
	}

	// Send peers info to the connector
	var nodes [(config.MAX_SC + 1) * config.MAX_SC_PEER]dtype.NodeInfo
	h.nm.GetSCNNodeListAll(&nodes)
	if err := ws.WriteJSON(nodes); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}
}

func (h *Handler) ObjectbyAccessPatternProc() {
	go func() {
		ticker := time.NewTicker(time.Duration(config.TIME_AP_GEN) * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			hashes := []dbagent.RemoverbleObj{}
			ret := false

			if config.ACCESS_FREQUENCY_PATTERN == config.RANDOM_ACCESS_PATTERN {
				ret = h.om.AccessWithRandom(config.NUM_AP_GEN, &hashes)
			} else {
				ret = h.om.AccessWithTimeWeight(config.NUM_AP_GEN, &hashes)
			}

			if ret {
				for _, hash := range hashes {
					if hash.HashType == 0 {
						bh := blockchain.BlockHeader{}
						if h.getBlockHeaderQuery(hash.Hash, &bh) {
							h.db.AddBlockHeader(hash.Hash, &bh)
							// log.Printf("add transaction from other node %v", hex.EncodeToString(tr.Hash))
						}
					} else {
						tr := blockchain.Transaction{}
						if h.getTransactionQuery(hash.Hash, &tr) {
							h.db.AddTransaction(&tr)
							// log.Printf("add transaction from other node %v", hex.EncodeToString(tr.Hash))
						}
					}
				}

				if h.local.SC < config.MAX_SC {
					h.om.DeleteNoAccedObjects()
				}
			}

			status := h.om.db.GetDBStatus()
			log.Printf("Status : %v", status)
		}
	}()
}

func (h *Handler) PeerListProc() {
	go func() {
		ticker := time.NewTicker(time.Duration(config.TIME_UPDATE_NEITHBOUR) * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			if h.sim.IP != "" && h.sim.Port != 0 && h.local.Hash != "" {
				h.nm.UpdatePeerList(h.sim, h.local)
			}
		}
	}()
}

func NewHandler(path string, local dtype.NodeInfo) *Handler {
	m := mux.NewRouter()
	h := &Handler{
		Handler: m,
		db:      dbagent.NewDBAgent(path, local.SC),
		sim:     dtype.NodeInfo{Mode: "", SC: config.SIM_SC, IP: "", Port: 0, Hash: ""},
		local:   local,
		tmc:     nil,
		nm:      nil,
		om:      nil,
		mutex:   sync.Mutex{},
	}

	h.tmc = testmgrcli.NewTMC(h.db, &h.sim, &h.local)
	h.nm = network.NewNodeMgr(&h.local)
	h.om = NewObjMgr(h.db)

	m.Handle("/", http.FileServer(http.Dir("static")))
	m.HandleFunc("/newblock", h.newBlockHandler)
	m.HandleFunc("/gettransaction", h.getTransactionHandler)
	m.HandleFunc("/getblockheader", h.getBlockHeaderHandler)
	m.HandleFunc("/nodeinfo", h.nodeInfoHandler)
	m.HandleFunc("/ping", h.pingHandler)
	return h
}
