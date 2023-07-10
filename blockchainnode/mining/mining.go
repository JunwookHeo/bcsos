package mining

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/blockchainnode/network"
	"github.com/junwookheo/bcsos/blockchainnode/storage"
	"github.com/junwookheo/bcsos/common/bitcoin"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/datalib"
	"github.com/junwookheo/bcsos/common/dtype"
)

const TPERIOD int = config.BLOCK_CREATE_PERIOD * 1000000000

type Mining struct {
	tp    map[string]*blockchain.Transaction
	st    *datalib.BcQueue // list of broadcast new transactions
	sb    *datalib.BcQueue // list of broadcast new blocks
	mutex sync.Mutex
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var (
	mi         *Mining
	oncemining sync.Once
)

type ChainInfoCmd struct {
	Cmd string `json:"cmd"`
}

// Add transaction to the pool
// true : If new transaction
// false : the transaction already exists in the pool
func (mi *Mining) AddTransactionToPool(key string, data *blockchain.Transaction) bool {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()

	_, ok := mi.tp[key]
	if !ok {
		mi.tp[key] = data
		return true
	}

	// Exist
	return false
}

func (mi *Mining) GetTransactionsFromPool() []*blockchain.Transaction {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()
	var trs []*blockchain.Transaction
	for _, tr := range mi.tp {
		trs = append(trs, tr)
	}

	return trs
}

func (mi *Mining) DeleteTransactionsFromPool(keys []string) {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()
	for _, key := range keys {
		_, ok := mi.tp[key]
		if ok {
			delete(mi.tp, key)
		}
	}
}

func (mi *Mining) ShowTransactionsFromPool() {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()
	// log.Printf("Tr Pool : %v", mi.tp)
}

func (mi *Mining) sendBlock(b *blockchain.Block, node *dtype.NodeInfo) {
	url := fmt.Sprintf("ws://%v:%v/broadcastnewblock", node.IP, node.Port)
	// log.Printf("Send ping with local info : %v", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Printf("BroadcasNewBlock error : %v", err)
		return
	}
	defer ws.Close()

	if err := ws.WriteJSON(b); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}
}

func (mi *Mining) BroadcastNewBlock(b *blockchain.Block) {
	var nodes [(config.MAX_SC) * config.MAX_SC_PEER]dtype.NodeInfo
	nm := network.NodeMgrInst()
	nm.GetSCNNodeListAll(&nodes)
	for _, node := range nodes {
		if node.IP != "" {
			mi.sendBlock(b, &node)
			// log.Printf("Broadcast Transaction : %v", node)
		}
	}
}

// func (mi *Mining) sendBtcBlock(b *bitcoin.BlockPkt, node *dtype.NodeInfo) {
// 	url := fmt.Sprintf("ws://%v:%v/broadcastnewbtcblock", node.IP, node.Port)
// 	// log.Printf("Send new btc block with local info : %v", url)

// 	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
// 	if err != nil {
// 		log.Printf("BroadcasNewBlock error : %v", err)
// 		return
// 	}
// 	defer ws.Close()

// 	if err := ws.WriteJSON(b); err != nil {
// 		log.Printf("Write json error : %v", err)
// 		return
// 	}
// }

func (mi *Mining) BroadcastNewBtcBlock(b *bitcoin.BlockPkt) {
	var wg sync.WaitGroup

	sendBtcBlock := func(node dtype.NodeInfo) {
		defer wg.Done()

		url := fmt.Sprintf("ws://%v:%v/broadcastnewbtcblock", node.IP, node.Port)
		// log.Printf("Send new btc block with local info : %v", url)

		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("BroadcasNewBlock error : %v", err)
			return
		}
		defer ws.Close()

		if err := ws.WriteJSON(b); err != nil {
			log.Printf("Write json error : %v", err)
			return
		}
	}

	var nodes [(config.MAX_SC) * config.MAX_SC_PEER]dtype.NodeInfo
	nm := network.NodeMgrInst()
	nm.GetSCNNodeListAll(&nodes)
	for _, node := range nodes {
		if node.IP != "" {
			wg.Add(1)
			go sendBtcBlock(node)
			// log.Printf("Broadcast Transaction : %v", node)
		}
	}
	wg.Wait()
	log.Println("Broadcast BTC Bloc : Done")
}

func (mi *Mining) UpdateTransactionPool(block *blockchain.Block) {
	// Remove transactions in the thransaction pool
	if block != nil {
		var trhashes []string
		for _, tr := range block.Transactions {
			trhashes = append(trhashes, hex.EncodeToString(tr.Hash))
		}

		mi.DeleteTransactionsFromPool(trhashes)
		mi.ShowTransactionsFromPool()
	}
}

func (mi *Mining) StartMiningNewBlock(status *string) {
	for {
		// This sleep is needed for updating a new block after sending the mining block
		time.Sleep(time.Nanosecond * time.Duration(TPERIOD/10))

		// _, prehash := mi.cm.GetHighestBlockHash()
		sm := storage.StorageMgrInst("")
		_, prehash := sm.GetHighestBlockHash()

		delay := TPERIOD - int(time.Now().UnixNano())%TPERIOD
		time.Sleep(time.Nanosecond * time.Duration(delay))

		if *status == "Stop" {
			log.Println("StartMiningNewBlock() : end")
			return
		}
		// height, curhash := mi.cm.GetHighestBlockHash()
		height, curhash := sm.GetHighestBlockHash()

		if prehash != curhash {
			continue
		}

		trs := mi.GetTransactionsFromPool()

		if len(trs) != 0 {
			hash, _ := hex.DecodeString(curhash)
			b := blockchain.CreateBlock(trs, hash, height+1)

			// Too much forks happen so add random delay
			ms := rand.Intn(100) * 10
			time.Sleep(time.Millisecond * time.Duration(ms))
			// Send block to local node
			ni := network.NodeInfoInst()
			local := ni.GetLocalddr()
			server := ni.GetSimAddr()
			// height, curhash = mi.cm.GetHighestBlockHash()
			height, curhash = sm.GetHighestBlockHash()

			if prehash != curhash {
				continue
			}

			log.Printf("==>mining a new block(%v):%v %v", height+1, hex.EncodeToString(b.Header.Hash), curhash)
			mi.sendBlock(b, local)
			mi.sendBlock(b, server) // Send a new block to simulation server
		}

	}
}

// newBlockHandler is called when a new block is received from miners
// When a node receive this, it stores the block on its local db
// Request : a new block
// Response : none
func (mi *Mining) newBlockHandler(w http.ResponseWriter, r *http.Request) {
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

	mi.mutex.Lock()
	if mi.sb.Find(hex.EncodeToString(block.Header.Hash)) {
		// log.Printf("===END bc Block: %v", hex.EncodeToString(block.Header.Hash))
		mi.mutex.Unlock()
		return
	}

	mi.sb.Push(hex.EncodeToString(block.Header.Hash))
	mi.mutex.Unlock()

	mi.UpdateTransactionPool(&block)

	// log.Printf("===FWD bc block : %v", hex.EncodeToString(block.Header.Hash))
	go mi.BroadcastNewBlock(&block)

	sm := storage.StorageMgrInst("")
	sm.AddNewBlock(&block)
}

// Update peers list
func (mi *Mining) BroadcasTransaction(t *blockchain.Transaction) {
	sendTransaction := func(node *dtype.NodeInfo) {
		url := fmt.Sprintf("ws://%v:%v/broadcastransaction", node.IP, node.Port)
		//log.Printf("BroadcasTransaction to : %v", url)

		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("Broadcastransaction error : %v", err)
			return
		}
		defer ws.Close()

		if err := ws.WriteJSON(t); err != nil {
			log.Printf("Write json error : %v", err)
			return
		}
	}

	var nodes [(config.MAX_SC) * config.MAX_SC_PEER]dtype.NodeInfo
	nm := network.NodeMgrInst()
	nm.GetSCNNodeListAll(&nodes)
	for _, node := range nodes {
		if node.IP != "" {
			sendTransaction(&node)
			// log.Printf("Broadcast Transaction : %v", node)
		}

	}
}

// broadcastTrascationHandler is called when a new transaction is received from nodes
// When a node receive this, it verifies and stores the transaction into the transaction pool
// Request : a new transaction
// Response : none
func (mi *Mining) broadcastTrascationHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("broadcastTrascationHandler", err)
		return
	}
	defer ws.Close()

	var tr blockchain.Transaction
	if err := ws.ReadJSON(&tr); err != nil {
		log.Printf("Read json error : %v", err)
	}

	if !tr.Verify() {
		log.Printf("===Verification failed : %v", hex.EncodeToString(tr.Hash))
		log.Panicln("===Verification failed")
		return
	}

	mi.mutex.Lock()
	if mi.st.Find(hex.EncodeToString(tr.Hash)) {
		// log.Printf("===END TR : %v", hex.EncodeToString(tr.Hash))
		mi.mutex.Unlock()
		return
	}

	mi.st.Push(hex.EncodeToString(tr.Hash))
	mi.mutex.Unlock()

	mi.AddTransactionToPool(hex.EncodeToString(tr.Hash), &tr)
	// log.Printf("===FWD TR : %v", hex.EncodeToString(tr.Hash))
	mi.BroadcasTransaction(&tr)
}

// newBlockHandler is called when a new block is received from miners
// When a node receive this, it stores the block on its local db
// Request : a new block
// Response : none
func (mi *Mining) newBtcBlockHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("newBtcBlockHandler", err)
		return
	}
	defer ws.Close()

	var buf bitcoin.BlockPkt
	if err := ws.ReadJSON(&buf); err != nil {
		log.Printf("Read json error : %v", err)
	}

	block := bitcoin.NewBlock()
	rb := bitcoin.NewRawBlock(buf.Block)
	block.SetHash(rb.GetRawBytes(0, 80))
	hash := block.GetHashString()

	mi.mutex.Lock()
	if mi.sb.Find(hash) {
		// log.Printf("===END bc Block: %v", hex.EncodeToString(block.Header.Hash))
		mi.mutex.Unlock()
		return
	}

	mi.sb.Push(hash)
	mi.mutex.Unlock()

	log.Printf("Received New BTC Block : %v", hash)
	go mi.BroadcastNewBtcBlock(&buf)

	mi.requestProof(hash)

	sm := storage.StorageMgrInst("")
	sm.AddNewBtcBlock(&buf, hash)
}

func (mi *Mining) broadcastNonInteractiveProof(p *dtype.NonInteractiveProof) {
	var wg sync.WaitGroup

	sendProof := func(node dtype.NodeInfo) {
		defer wg.Done()
		url := fmt.Sprintf("ws://%v:%v/noninteractiveproof", node.IP, node.Port)
		//log.Printf("BroadcasTransaction to : %v", url)

		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("noninteractiveproof error : %v", err)
			return
		}
		defer ws.Close()

		if err := ws.WriteJSON(p); err != nil {
			log.Printf("Write json error : %v", err)
			return
		}
	}

	var nodes [(config.MAX_SC) * config.MAX_SC_PEER]dtype.NodeInfo
	nm := network.NodeMgrInst()
	nm.GetSCNNodeListAll(&nodes)
	for _, node := range nodes {
		if node.IP != "" {
			wg.Add(1)
			go sendProof(node)
			// log.Printf("Broadcast Transaction : %v", node)
		}

	}
	wg.Wait()
	log.Println("Broadcast Transaction : Done")
}

// Request Proof of Storage
func (mi *Mining) requestProof(hash string) {
	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()

	tmp, _ := hex.DecodeString(hash)
	h1 := sha256.Sum256(tmp)
	h2, _ := hex.DecodeString(local.Hash)

	h2[len(h2)-1] = byte(local.Port & 0xFF)

	// Check address and hash
	mask := byte(config.MASK_SELECT_POS_NODE) // compare 4-bit, so pos is performed avg. 16*T(block generation time)
	log.Printf("check hash for Pos : %v-%v", h1[len(h1)-1]&mask, h2[len(h2)-1]&mask)
	if h1[len(h1)-1]&mask != h2[len(h2)-1]&mask {
		return
	}

	if config.PROOFSTORAGE_METHOD == "INTERACTIVE" {
		mi.requestInteractiveProof(hash)
	} else {
		mi.nonInteractiveProofStorage(hash)
	}
}

// Non-interactive Proof of Storage
func (mi *Mining) nonInteractiveProofStorage(hash string) {
	// Create a Proof
	sm := storage.StorageMgrInst("")
	proof := sm.GetNonInteractiveStarksProof(hash)

	// Broadcast the Proof
	if proof != nil {
		log.Printf("Non-interactive ProofStorage hash : %v", proof.Hash)
		mi.broadcastNonInteractiveProof(proof)
	}
}

func (mi *Mining) requestInteractiveProof(hash string) {
	var nodes [(config.MAX_SC) * config.MAX_SC_PEER]dtype.NodeInfo
	nm := network.NodeMgrInst()
	nm.GetSCNNodeListAll(&nodes)

	// Get target node for Proof of Storage
	node := mi.GetTargetNodePoS(nodes)
	if node == nil {
		return
	}
	log.Printf("check node for Pos : %v-%v", node.IP, node.Port)

	url := fmt.Sprintf("ws://%v:%v/interactiveproof", node.IP, node.Port)
	// log.Printf("Send ping with local info : %v", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Printf("BroadcasNewBlock error : %v", err)
		return
	}
	defer ws.Close()

	sm := storage.StorageMgrInst("")
	height := sm.GetRandomHeightForNConsecutiveBlocks(hash)

	if height == -1 {
		return
	}

	req := dtype.ReqConsecutiveHashes{Height: height, Count: config.NUM_CONSECUTIVE_HASHES}
	if err := ws.WriteJSON(req); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}
	start := time.Now().UnixNano()

	var proof dtype.PoSProof
	if err := ws.ReadJSON(&proof); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}

	if proof.Timestamp > time.Now().UnixNano() {
		log.Printf("Verify Proof : Time error %v", proof.Timestamp)
		return
	}

	if (proof.Timestamp-start)/1000000 > int64(config.MAX_PROOF_TIME_MSEC) {
		log.Printf("Verify Proof : Time Exceed %v", (proof.Timestamp-start)/1000000)
		return
	}

	sm.VerifyInterActiveProofStorage(&proof)
}

func (mi *Mining) GetTargetNodePoS(nodes [config.MAX_SC * config.MAX_SC_PEER]dtype.NodeInfo) *dtype.NodeInfo {
	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()

	if config.MAX_SC < local.SC {
		return nil
	}

	var t_nodes []dtype.NodeInfo
	cnt := 0
	for _, node := range nodes {
		if node.Hash != "" && node.Hash != local.Hash {
			t_nodes = append(t_nodes, node)
			cnt++
		}
	}

	if cnt == 0 {
		return nil
	}

	t := rand.Intn(cnt)
	return &t_nodes[t]
}

// chainInfoHandler sends blockchain connection information to the webapp.
// So, the webapp displays the connection of chain.
// Request : a new transaction
// Response : none
// func (mi *Mining) chainInfoHandler(w http.ResponseWriter, r *http.Request) {
// 	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
// 	ws, err := upgrader.Upgrade(w, r, nil)
// 	if err != nil {
// 		log.Println("chainInfoHandler", err)
// 		return
// 	}

// 	log.Println("chainInfoHandler")

// 	defer ws.Close()
// 	done := make(chan struct{})
// 	var cmd ChainInfoCmd

// 	go func() {
// 		defer close(done)
// 		for {
// 			if err := ws.ReadJSON(&cmd); err != nil {
// 				log.Printf("Read json error : %v", err)
// 				return
// 			}
// 			log.Printf("Test resume/pause receive : %v", cmd)
// 		}
// 	}()

// 	ticker := time.NewTicker(time.Second)
// 	defer ticker.Stop()

// 	func() {
// 		var block *datalib.BlockData
// 		for {
// 			select {
// 			case <-done:
// 				return
// 			case <-ticker.C:
// 				newlist := mi.cm.GetNewBlockInfo()

// 				if newlist != nil {
// 					start := newlist.Find(block)
// 					for {
// 						i, b := newlist.Next()
// 						if b == nil {
// 							break
// 						}
// 						if start < i {
// 							log.Printf("new block %v, %v", i, b)
// 							block = b
// 							if err := ws.WriteJSON(block); err != nil {
// 								log.Printf("Write json error : %v", err)
// 								return
// 							}
// 						}
// 					}

// 				}

// 			}
// 		}
// 	}()
// }

func (mi *Mining) SetHttpRouter(m *mux.Router) {
	m.HandleFunc("/broadcastnewblock", mi.newBlockHandler)
	m.HandleFunc("/broadcastransaction", mi.broadcastTrascationHandler)
	// m.HandleFunc("/chaininfo", mi.chainInfoHandler)

	m.HandleFunc("/broadcastnewbtcblock", mi.newBtcBlockHandler)
}

func MiningInst() *Mining {
	oncemining.Do(func() {
		mi = &Mining{
			tp:    make(map[string]*blockchain.Transaction),
			st:    datalib.NewBcQueue(config.BLOCK_CREATE_PERIOD * 2),
			sb:    datalib.NewBcQueue(6 * 2), // Light nodes in Bitcoin has 6
			mutex: sync.Mutex{},
		}
	})
	return mi
}
