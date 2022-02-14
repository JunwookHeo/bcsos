package storage

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/listener"
	"github.com/junwookheo/bcsos/common/wallet"
	"github.com/junwookheo/bcsos/storagesrv/network"
	"github.com/junwookheo/bcsos/storagesrv/testmgrcli"
)

type Handler struct {
	http.Handler
	wallet    *wallet.Wallet
	db        dbagent.DBAgent
	sim       dtype.NodeInfo
	local     dtype.NodeInfo
	tmc       *testmgrcli.TestMgrCli
	nm        *network.NodeMgr
	om        *ObjectMgr
	mutex     sync.Mutex
	listeners *listener.EventListener
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
func (h *Handler) getObjectHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("getTransactionHandler transaction error : ", err)
		return
	}
	defer ws.Close()

	var reqData dtype.ReqData
	if err := ws.ReadJSON(&reqData); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}

	h.db.UpdateDBNetworkQuery(1, 0, 0)
	reqData.SC = h.local.SC
	var obj interface{}
	if reqData.ObjType == "transaction" {
		tr := blockchain.Transaction{}
		if h.db.GetTransaction(reqData.ObjHash, &tr) == 0 {
			if h.getObjectQuery(h.local.SC+1, &reqData, &tr) {
				h.db.AddTransaction(&tr)
			}
		} else {
			h.db.UpdateDBNetworkQuery(0, 0, 1)
		}
		obj = tr
	} else if reqData.ObjType == "blockheader" {
		bh := blockchain.BlockHeader{}
		if h.db.GetBlockHeader(reqData.ObjHash, &bh) == 0 {
			if h.getObjectQuery(h.local.SC+1, &reqData, &bh) {
				h.db.AddBlockHeader(reqData.ObjHash, &bh)
			}
		} else {
			h.db.UpdateDBNetworkQuery(0, 0, 1)
		}
		obj = bh
	} else {
		log.Panicf("Not support object type")
	}

	reqData.Addr = fmt.Sprintf("%v:%v", h.local.IP, h.local.Port)
	reqData.Hop += 1

	ws.WriteJSON(reqData)
	ws.WriteJSON(obj)
	log.Printf("<==Query write reqData: %v", reqData)
}

func (h *Handler) newReqData(objtype string, hash string) dtype.ReqData {
	req := dtype.ReqData{}
	req.Addr = fmt.Sprintf("%v:%v", h.local.IP, h.local.Port)
	req.Timestamp = time.Now().UnixNano()
	req.SC = h.local.SC
	req.Hop = 0
	req.ObjType = objtype
	req.ObjHash = hash

	return req
}

// getTransactionQuery queries a transaction ot other nodes with highr Storage Class
// Request : hash of transaction
// Response : transaction
func (h *Handler) getObjectQuery(startSC int, reqData *dtype.ReqData, obj interface{}) bool {
	queryObject := func(ip string, port int, reqData *dtype.ReqData, obj interface{}) bool {
		url := fmt.Sprintf("ws://%v:%v/getobject", ip, port)
		//log.Printf("getTransactionQuery : %v", url)

		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("getTransactionQuery Dial error : %v", err)
			return false
		}
		defer ws.Close()

		if err := ws.WriteJSON(*reqData); err != nil {
			log.Printf("Write json error : %v", err)
			return false
		}

		if err := ws.ReadJSON(reqData); err != nil {
			log.Printf("Read json error : %v", err)
			return false
		}
		if err := ws.ReadJSON(obj); err != nil {
			log.Printf("Read json error : %v", err)
			return false
		}

		hop := reqData.SC - h.local.SC
		h.db.UpdateDBNetworkDelay(int(time.Now().UnixNano()-reqData.Timestamp), hop)
		log.Printf("==>Query read reqData: %v[hop], %v", hop, reqData)
		return true
	}

	// the number of query to other nodes
	defer h.db.UpdateDBNetworkQuery(0, 1, 1)

	for i := startSC; i < config.MAX_SC; i++ {
		var nodes [config.MAX_SC_PEER]dtype.NodeInfo
		if h.nm.GetSCNNodeListbyDistance(i, reqData.ObjHash, &nodes) {
			for _, node := range nodes {
				if node.IP == "" {
					continue
				}
				if node.Hash == h.local.Hash { // If the node is itself, skip
					continue
				}

				if queryObject(node.IP, node.Port, reqData, obj) {
					return true
				}
				//time.Sleep(time.Duration(200 * time.Microsecond.Seconds()))
				log.Printf("queryObject fail : query other nodes")
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
	var nodes [(config.MAX_SC) * config.MAX_SC_PEER]dtype.NodeInfo
	h.nm.GetSCNNodeListAll(&nodes)
	if err := ws.WriteJSON(nodes); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}
}

// Response to web app with dbstatus information
// keep sending dbstatus to the web app
func (h *Handler) endTestHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("endTestHandler", err)
		return
	}
	defer ws.Close()
	var endtest string
	if err := ws.ReadJSON(&endtest); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}

	if endtest == config.END_TEST {
		log.Println("Received End test")
		h.db.Close()
		time.Sleep(3 * time.Second)
		syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}
}

func (h *Handler) EndTestProc() {
	command := make(chan string)
	h.listeners.AddListener(command)

	go func(command <-chan string) {
		for {
			select {
			case cmd := <-command:
				log.Println(cmd)
				switch cmd {
				case "Stop":
					log.Println("Received End test")
					h.db.Close()
					time.Sleep(3 * time.Second)
					syscall.Kill(syscall.Getpid(), syscall.SIGINT)
					return
				}
			default:
				log.Println("=========EndTestProc")
				time.Sleep(time.Duration(config.TIME_AP_GEN) * time.Second)
			}
		}
	}(command)
}

func (h *Handler) commandProc(cmd *dtype.Command) {
	log.Printf("commandProc : %v", cmd)
	if cmd.Cmd == "SET" {
		switch cmd.Subcmd {
		case "Test":
			if cmd.Arg1 == "Start" {
				h.listeners.Notify("Start")
			} else if cmd.Arg1 == "Stop" {
				h.listeners.Notify("Stop")
			} else if cmd.Arg1 == "Pause" {
				h.listeners.Notify("Pause")
			} else if cmd.Arg1 == "Resume" {
				h.listeners.Notify("Resume")
			}
		}
	}
}

func (h *Handler) commandHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("nodesHandler", err)
		return
	}
	defer ws.Close()
	var cmd dtype.Command
	if err := ws.ReadJSON(&cmd); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}
	log.Printf("Test command receive : %v", cmd)

	h.commandProc(&cmd)

	cmd.Arg2 = "OK"
	if err := ws.WriteJSON(cmd); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}
}

func (h *Handler) ObjectbyAccessPattern() {
	hashes := []dbagent.RemoverbleObj{}
	ret := false

	if config.ACCESS_FREQUENCY_PATTERN == config.RANDOM_ACCESS_PATTERN {
		ret = h.om.AccessWithUniform(config.NUM_AP_GEN, &hashes)
	} else {
		ret = h.om.AccessWithExponential(config.NUM_AP_GEN, &hashes)
	}

	if ret {
		for _, hash := range hashes {
			if hash.HashType == 0 {
				bh := blockchain.BlockHeader{}
				req := h.newReqData("blockheader", hash.Hash)
				if h.getObjectQuery(h.local.SC, &req, &bh) {
					h.db.AddBlockHeader(hash.Hash, &bh)
					if hash.Hash != hex.EncodeToString(bh.GetHash()) {
						log.Panicf("%v header Hash not equal %v", hash.Hash, hex.EncodeToString(bh.GetHash()))
					}
				}
			} else {
				tr := blockchain.Transaction{}
				req := h.newReqData("transaction", hash.Hash)
				if h.getObjectQuery(h.local.SC, &req, &tr) {
					h.db.AddTransaction(&tr)
					if hash.Hash != hex.EncodeToString(tr.Hash) {
						log.Panicf("%v Tr Hash not equal %v", hash.Hash, hex.EncodeToString(tr.Hash))
					}
				}
			}
		}

		if h.local.SC < config.MAX_SC-1 {
			h.om.DeleteNoAccedObjects()
		}
	}

	status := h.om.db.GetDBStatus()
	log.Printf("Status : %v", status)
}

func (h *Handler) ObjectbyAccessPatternProc() {
	command := make(chan string)
	h.listeners.AddListener(command)

	go func(command <-chan string) {
		var status = "Pause"
		for {
			select {
			case cmd := <-command:
				log.Println(cmd)
				switch cmd {
				case "Stop":
					return
				case "Pause":
					status = "Pause"
				case "Resume":
					status = "Running"
				case "Start":
					status = "Running"
				}
			default:
				if status == "Running" {
					h.ObjectbyAccessPattern()
					log.Println("=========ObjectbyAccessPatternProc")
					time.Sleep(time.Duration(config.TIME_AP_GEN) * time.Second)
				}
			}
		}

	}(command)
}

func (h *Handler) PeerListProc() {
	command := make(chan string)
	h.listeners.AddListener(command)

	go func(command <-chan string) {
		var status = "Pause"
		for {
			select {
			case cmd := <-command:
				log.Println(cmd)
				switch cmd {
				case "Stop":
					return
				case "Pause":
					status = "Pause"
				case "Resume":
					status = "Running"
				case "Start":
					status = "Running"
				}
			default:
				if status == "Running" {
					if h.sim.IP != "" && h.sim.Port != 0 && h.local.Hash != "" {
						h.nm.UpdatePeerList(h.sim, h.local)
					}
					log.Println("=========PeerListProc")
					time.Sleep(time.Duration(config.TIME_UPDATE_NEITHBOUR) * time.Second)
				}
			}
		}

	}(command)
}

func NewHandler(db_path string, wallet_path string, local dtype.NodeInfo) *Handler {
	m := mux.NewRouter()
	w, _ := wallet.LoadFile(wallet_path)
	h := &Handler{
		Handler:   m,
		wallet:    w,
		db:        dbagent.NewDBAgent(db_path, local.SC),
		sim:       dtype.NodeInfo{Mode: "", SC: config.SIM_SC, IP: "", Port: 0, Hash: ""},
		local:     local,
		tmc:       nil,
		nm:        nil,
		om:        nil,
		mutex:     sync.Mutex{},
		listeners: &listener.EventListener{},
	}

	h.tmc = testmgrcli.NewTMC(h.db, &h.sim, &h.local)
	h.nm = network.NewNodeMgr(&h.local)
	h.om = NewObjMgr(h.db)

	m.Handle("/", http.FileServer(http.Dir("static")))
	m.HandleFunc("/newblock", h.newBlockHandler)
	m.HandleFunc("/getobject", h.getObjectHandler)
	m.HandleFunc("/nodeinfo", h.nodeInfoHandler)
	m.HandleFunc("/ping", h.pingHandler)
	m.HandleFunc("/endtest", h.endTestHandler)
	m.HandleFunc("/command", h.commandHandler)
	return h
}
