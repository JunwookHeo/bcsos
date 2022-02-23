package mining

import (
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
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/datalib"
	"github.com/junwookheo/bcsos/common/dtype"
)

type Mining struct {
	tp    map[string]*blockchain.Transaction
	st    *datalib.BcQueue // list of broadcast new transactions
	sb    *datalib.BcQueue // list of broadcast new blocks
	cm    *ChainMgr
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

func (mi *Mining) StartMiningNewBlock(preblock *blockchain.Block) {
	// Remove transactions in the thransaction pool
	if preblock != nil {
		var trhashes []string
		for _, tr := range preblock.Transactions {
			trhashes = append(trhashes, hex.EncodeToString(tr.Hash))
		}

		mi.DeleteTransactionsFromPool(trhashes)
	}

	sm := storage.StorageMgrInst("")
	//prehash := sm.GetLatestBlockHash()
	prehash := mi.cm.GetHighestBlockHash()

	// waithing some time instead of using high difficulty
	t := rand.Intn(2000)
	ticker := time.NewTicker(time.Millisecond * time.Duration(config.BLOCK_CREATE_PERIOD*1000+t))
	defer ticker.Stop()

	<-ticker.C

	if prehash != mi.cm.GetHighestBlockHash() {
		//if prehash != sm.GetLatestBlockHash() {
		log.Printf("===============Terminating mining")
		return
	}

	log.Printf("===============Start mining")
	// Statrt PoW
	trs := mi.GetTransactionsFromPool()

	if len(trs) == 0 {
		log.Printf("===============Start mining : trs : %v", trs)
		mi.StartMiningNewBlock(preblock)
	}
	hash, _ := hex.DecodeString(sm.GetLatestBlockHash())
	b := blockchain.CreateBlock(trs, hash)

	log.Printf("===============create block : %v", b)
	// Send block to local node
	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()
	mi.sendBlock(b, local)
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

	log.Printf("===FWD bc block : %v", hex.EncodeToString(block.Header.Hash))
	go mi.BroadcastNewBlock(&block)

	sm := storage.StorageMgrInst("")
	mi.cm.AddedNewBlock(&block.Header)
	sm.AddNewBlock(&block)

	go mi.StartMiningNewBlock(&block)
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

func (mi *Mining) SetHttpRouter(m *mux.Router) {
	m.HandleFunc("/broadcastnewblock", mi.newBlockHandler)
	m.HandleFunc("/broadcastransaction", mi.broadcastTrascationHandler)
}

func MiningInst() *Mining {
	oncemining.Do(func() {
		mi = &Mining{
			tp:    make(map[string]*blockchain.Transaction),
			st:    datalib.NewBcQueue(config.BLOCK_CREATE_PERIOD * 2),
			sb:    datalib.NewBcQueue(6 * 2), // Light nodes in Bitcoin has 6
			cm:    NewChainMgr(),
			mutex: sync.Mutex{},
		}
	})
	return mi
}
