package storage

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/blockchainnode/network"
	"github.com/junwookheo/bcsos/common/bitcoin"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/datalib"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/listener"
)

type RcvBlockTime struct {
	GenTime int64
	RcvTime int64
	Hash    string
}

type StorageMgr struct {
	db   dbagent.DBAgent
	om   *ObjectMgr
	cand *datalib.CandidateBlocks
	ltb  RcvBlockTime // Latest Time receiving a new block
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	sm   *StorageMgr
	once sync.Once
)

func (h *StorageMgr) Stop() {
	h.db.Close()
}

// getTransactionHandler is called when transaction query from other nodes is received
// if the node does not have the transaction, the node will query it to other nodes with highr SC
// Request : hash of transaction
// Response : transaction
func (h *StorageMgr) getObjectHandler(w http.ResponseWriter, r *http.Request) {
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

	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()
	reqData.SC = local.SC
	var obj interface{}
	if reqData.ObjType == "transaction" {
		tr := blockchain.Transaction{}
		if h.db.GetTransaction(reqData.ObjHash, &tr) == 0 {
			if h.getObjectQuery(local.SC+1, &reqData, &tr) {
				h.db.AddTransaction(&tr)
			}
		} else {
			h.db.UpdateDBNetworkQuery(0, 0, 1)
		}
		obj = tr
	} else if reqData.ObjType == "blockheader" {
		bh := blockchain.BlockHeader{}
		if h.db.GetBlockHeader(reqData.ObjHash, &bh) == 0 {
			if h.getObjectQuery(local.SC+1, &reqData, &bh) {
				h.db.AddBlockHeader(reqData.ObjHash, &bh)
			}
		} else {
			h.db.UpdateDBNetworkQuery(0, 0, 1)
		}
		obj = bh
	} else {
		log.Panicf("Not support object type")
	}

	reqData.Addr = fmt.Sprintf("%v:%v", local.IP, local.Port)
	reqData.Hop += 1

	ws.WriteJSON(reqData)
	ws.WriteJSON(obj)
	log.Printf("<==Query write reqData: %v", reqData)
}

func (h *StorageMgr) newReqData(objtype string, hash string) dtype.ReqData {
	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()

	req := dtype.ReqData{}
	req.Addr = fmt.Sprintf("%v:%v", local.IP, local.Port)
	req.Timestamp = time.Now().UnixNano()
	req.SC = local.SC
	req.Hop = 0
	req.ObjType = objtype
	req.ObjHash = hash

	return req
}

// getTransactionQuery queries a transaction ot other nodes with highr Storage Class
// Request : hash of transaction
// Response : transaction
func (h *StorageMgr) getObjectQuery(startSC int, reqData *dtype.ReqData, obj interface{}) bool {
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

		ni := network.NodeInfoInst()
		local := ni.GetLocalddr()

		hop := reqData.SC - local.SC
		h.db.UpdateDBNetworkDelay(int(time.Now().UnixNano()-reqData.Timestamp), hop)
		log.Printf("==>Query read reqData: %v[hop], %v", hop, reqData)
		return true
	}

	// the number of query to other nodes
	defer h.db.UpdateDBNetworkQuery(0, 1, 1)

	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()
	nm := network.NodeMgrInst()

	for i := startSC; i < config.MAX_SC; i++ {
		var nodes [config.MAX_SC_PEER]dtype.NodeInfo
		if nm.GetSCNNodeListbyDistance(i, reqData.ObjHash, &nodes) {
			for _, node := range nodes {
				if node.IP == "" {
					continue
				}
				if node.Hash == local.Hash { // If the node is itself, skip
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

type Test struct {
	Start bool `json:"start"`
}

func (h *StorageMgr) statusInfo(ws *websocket.Conn, w http.ResponseWriter, r *http.Request) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var test Test
			if err := ws.ReadJSON(&test); err != nil {
				log.Printf("Read json error : %v", err)
				return
			}
			log.Printf("receive : %v", test)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				//var status dbagent.DBStatus
				status := h.db.GetDBStatus()
				if err := ws.WriteJSON(status); err != nil {
					log.Printf("Write json error : %v", err)
					return
				}
			}
		}
	}()
}

// Response to web app with dbstatus information
// keep sending dbstatus to the web app
func (h *StorageMgr) statusInfoHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("newBlockHandler", err)
		return
	}
	defer ws.Close()
	h.statusInfo(ws, w, r)
}

func (h *StorageMgr) ObjectbyAccessPattern() {
	hashes := []dbagent.RemoverbleObj{}
	ret := false

	if config.ACCESS_FREQUENCY_PATTERN == config.RANDOM_ACCESS_PATTERN {
		ret = h.om.AccessWithUniform(config.NUM_AP_GEN, &hashes)
	} else {
		ret = h.om.AccessWithExponential(config.NUM_AP_GEN, &hashes)
	}

	if ret {
		ni := network.NodeInfoInst()
		local := ni.GetLocalddr()

		for _, hash := range hashes {
			if hash.HashType == 0 {
				bh := blockchain.BlockHeader{}
				req := h.newReqData("blockheader", hash.Hash)
				if h.getObjectQuery(local.SC, &req, &bh) {
					h.db.AddBlockHeader(hash.Hash, &bh)
					if hash.Hash != hex.EncodeToString(bh.GetHash()) {
						log.Panicf("%v header Hash not equal %v", hash.Hash, hex.EncodeToString(bh.GetHash()))
					}
				}
			} else {
				tr := blockchain.Transaction{}
				req := h.newReqData("transaction", hash.Hash)
				if h.getObjectQuery(local.SC, &req, &tr) {
					h.db.AddTransaction(&tr)
					if hash.Hash != hex.EncodeToString(tr.Hash) {
						log.Panicf("%v Tr Hash not equal %v", hash.Hash, hex.EncodeToString(tr.Hash))
					}
				}
			}
		}

		h.RemoveNoAccessObjects()
	}

	// status := h.om.db.GetDBStatus()
	// log.Printf("Status : %v", status)
}

func (h *StorageMgr) RemoveNoAccessObjects() {
	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()

	if local.SC < config.MAX_SC-1 {
		h.om.DeleteNoAccedObjects()
	}
}

// interactiveProofHandler handles the request of Proof of Storage
// Request : Hash of block and timestamp
// Response : Merkel root of transactions to be proven
func (h *StorageMgr) interactiveProofHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("proofStorageHandler", err)
		return
	}
	defer ws.Close()

	var reqHash dtype.ReqConsecutiveHashes
	if err := ws.ReadJSON(&reqHash); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}

	proof := h.GetInteractiveProof(reqHash.Height)

	if err := ws.WriteJSON(proof); err != nil {
		log.Printf("Write proof error : %v", err)
		return
	}
}

// nonInteractiveProofHandler handles the request of Non-interactive Proof of Storage
// Request : Non-interactive Proof
// Response : No Response
func (h *StorageMgr) nonInteractiveProofHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("nonInteractiveProofHandler", err)
		return
	}
	defer ws.Close()

	// var proof dtype.PoSProof
	var proof dtype.NonInteractiveProof
	if err := ws.ReadJSON(&proof); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}

	if proof.Starks.RandomHash == h.ltb.Hash {
		start := h.ltb.RcvTime //h.db.GetLastBlockTime()
		current := time.Now().UnixNano()
		log.Printf("Event Proof Time : T1:%v, T2:%v, T3:%v", (h.ltb.RcvTime-h.ltb.GenTime)/1000000, (current-h.ltb.GenTime)/1000000, (current-h.ltb.RcvTime)/1000000)
		if (current-start)/1000000 > int64(config.MAX_PROOF_TIME_MSEC) {
			log.Printf("Verify Proof : Time Exceed %v", (current-start)/1000000)
			return
		} else {
			log.Printf("Receive Verify Proof : Time %v", (current-start)/1000000)
		}

		h.VerifyNonInteractiveProofStorage(h.ltb.GenTime, start, current, &proof)
	} else {
		log.Panicf("hash mismatching, %v, %v", proof.Starks.RandomHash, h.ltb.Hash)
	}

}

func (h *StorageMgr) GetRandomHeightForNConsecutiveBlocks(hash string) int {
	return h.db.GetRandomHeightForNConsecutiveBlocks(hash)
}

func (h *StorageMgr) GetNonInteractiveStarksProof(hash string) *dtype.NonInteractiveProof {
	return h.db.GetNonInteractiveStarksProof(hash)
}

func (h *StorageMgr) GetInteractiveProof(height int) *dtype.PoSProof {
	return h.db.GetInteractiveProof(height)
}

func (h *StorageMgr) VerifyInterActiveProofStorage(proof *dtype.PoSProof) bool {
	return h.db.VerifyInterActiveProofStorage(proof)
}

func (h *StorageMgr) VerifyNonInteractiveProofStorage(tlb, trb, trp int64, proof *dtype.NonInteractiveProof) bool {
	return h.db.VerifyNonInterActiveProofStorage(tlb, trb, trp, proof)
}

func (h *StorageMgr) ObjectbyAccessPatternProc() {
	command := make(chan string)
	el := listener.EventListenerInst()
	el.AddListener(command)

	go func(command <-chan string) {
		var status = "Pause"
		for {
			select {
			case cmd := <-command:
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
					ni := network.NodeInfoInst()
					local := ni.GetLocalddr()
					if strings.ToUpper(local.Mode) == "MI" {
						h.ObjectbyAccessPattern()
						// log.Println("=========ObjectbyAccessPatternProc")
						time.Sleep(time.Duration(config.TIME_AP_GEN) * time.Second)
					} else {
						h.RemoveNoAccessObjects()
						time.Sleep(time.Duration(config.BLOCK_CREATE_PERIOD*2) * time.Second)
						// log.Printf("Mode : %v", local.Mode)
					}
				} else {
					time.Sleep(time.Second)
				}
			}
		}

	}(command)
}

func (h *StorageMgr) AddNewBlock(b *blockchain.Block) {
	// log.Printf("Rcv new block(%v) : %v-%v", b.Header.Height, hex.EncodeToString(b.Header.Hash), hex.EncodeToString(b.Header.PrvHash))
	h.cand.PushAndSave(b, h.db)
	// h.cand.ShowAll()
}

func (h *StorageMgr) AddNewBtcBlock(b *bitcoin.BlockPkt, hash string) {
	h.ltb.RcvTime = time.Now().UnixNano()
	h.ltb.GenTime = b.Timestamp
	h.ltb.Hash = hash

	log.Printf("Received Time new block : %v ms", (h.ltb.RcvTime-b.Timestamp)/1000000)

	// Update time first and then add block
	h.db.AddNewBlock(b)
}

func (h *StorageMgr) GetHighestBlockHash() (int, string) {
	return h.cand.GetHighestBlockHash()
}

func (h *StorageMgr) GetLatestBlockHash() (string, int) {
	return h.db.GetLatestBlockHash()
}

func (sm *StorageMgr) SetHttpRouter(m *mux.Router) {
	m.HandleFunc("/getobject", sm.getObjectHandler)
	m.HandleFunc("/statusinfo", sm.statusInfoHandler)
	m.HandleFunc("/interactiveproof", sm.interactiveProofHandler)
	m.HandleFunc("/noninteractiveproof", sm.nonInteractiveProofHandler)
}

func StorageMgrInst(db_path string) *StorageMgr {
	if db_path == "" {
		return sm
	}

	ltb := RcvBlockTime{0, time.Now().UnixNano(), ""}

	once.Do(func() {
		sm = &StorageMgr{
			db:   dbagent.NewDBAgent(db_path),
			om:   nil,
			cand: datalib.NewCandidateBlocks(),
			ltb:  ltb,
		}
		sm.om = NewObjMgr(sm.db)
	})

	return sm
}
