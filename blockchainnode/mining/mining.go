package mining

import (
	"encoding/hex"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/blockchainnode/network"
	"github.com/junwookheo/bcsos/blockchainnode/storage"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dtype"
)

type Mining struct {
	tp    map[string]*blockchain.Transaction
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

func (mi *Mining) GetTransactionFromPool(key string) *blockchain.Transaction {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()
	tr, ok := mi.tp[key]
	if !ok {
		return nil
	}
	return tr
}

func (mi *Mining) DeleteTransactionFromPool(key string) bool {
	mi.mutex.Lock()
	defer mi.mutex.Unlock()
	_, ok := mi.tp[key]
	if !ok {
		return false
	}

	delete(mi.tp, key)

	return true
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

	sm := storage.StorageMgrInst("")
	sm.AddNewBlock(&block)
	// ws.WriteJSON(block)
	// for _, t := range block.Transactions {
	// 	log.Printf("From client : %s", t.Data)
	// }
}

// Update peers list
func (mi *Mining) BroadcasTransaction(t *blockchain.Transaction) {
	sendTransaction := func(node *dtype.NodeInfo) {
		url := fmt.Sprintf("ws://%v:%v/broadcastransaction", node.IP, node.Port)
		log.Printf("Send ping with local info : %v", url)

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
		log.Printf("===Verification failed : %v", tr.Hash)
		return
	}

	log.Printf("===RCV : %v", hex.EncodeToString(tr.Hash))
	ret := mi.AddTransactionToPool(hex.EncodeToString(tr.Hash), &tr)
	log.Printf("%v", mi.tp)

	if ret {
		log.Printf("===FWD : %v", hex.EncodeToString(tr.Hash))
		mi.BroadcasTransaction(&tr)
	}

}

func (mi *Mining) SetHttpRouter(m *mux.Router) {
	m.HandleFunc("/newblock", mi.newBlockHandler)
	m.HandleFunc("/broadcastransaction", mi.broadcastTrascationHandler)
}

func MiningInst() *Mining {
	oncemining.Do(func() {
		mi = &Mining{
			tp:    make(map[string]*blockchain.Transaction),
			mutex: sync.Mutex{},
		}
	})
	return mi
}
