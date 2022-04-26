package bcdummy

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/wallet"
)

type Handler struct {
	db    dbagent.DBAgent
	Ready bool
	Nodes *map[string]dtype.NodeInfo
}

const PATH = "./iotdata/IoT_normal_fridge_1.log"

func (h *Handler) sendNewBlock(b *blockchain.Block, ip string, port int) {
	url := fmt.Sprintf("ws://%v:%v/broadcastnewblock", ip, port)
	//log.Printf("Send new block to %v", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Printf("dial: %v", err)
		return
	}
	defer ws.Close()
	// log.Printf("DefaultDialer Send new block to %v", url)

	if err := ws.WriteJSON(b); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}

	// for _, t := range b.Transactions {
	// 	log.Printf("Recevied node : %s", t.Data)
	// }

}

func (h *Handler) broadcastNewBlock(b *blockchain.Block) {
	// log.Printf("broadcast : %v", *h.Nodes)
	for _, node := range *h.Nodes {
		go h.sendNewBlock(b, node.IP, node.Port)
	}
}

func (h *Handler) sendEndTest(ip string, port int) {
	url := fmt.Sprintf("ws://%v:%v/endtest", ip, port)
	log.Printf("Send End test to %v", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Printf("dial: %v", err)
		return
	}
	defer ws.Close()
	log.Printf("DefaultDialer Send new block to %v", url)

	if err := ws.WriteJSON(config.END_TEST); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}
	log.Printf("Send end test : %v:%v", ip, port)
}

func (h *Handler) broadcastEndTest() {
	log.Printf("broadcastEndTest : %v", *h.Nodes)
	for _, node := range *h.Nodes {
		h.sendEndTest(node.IP, node.Port)
	}
}

func (h *Handler) KillProcess() {
	p, err := os.FindProcess(os.Getpid())

	if err != nil {
		return
	}

	p.Signal(syscall.SIGTERM)
}

func (h *Handler) Start() {
	// Send Genesis Block
	wallet_path := "./bc_dummy.wallet"
	w := wallet.NewWallet(wallet_path)
	log.Printf("wallet : %v", w)

	cnt := 0
	hash, height := h.db.GetLatestBlockHash()
	if len(hash) == 0 {
		log.Printf("Create Genesis due to hash : (%v) - %v", height, hash)
		b := blockchain.CreateGenesis(w)
		h.db.AddBlock(b)
		h.broadcastNewBlock(b)
		log.Printf("Broadcast a new block : %v", cnt)
		cnt++
		// time.Sleep(time.Duration(config.BLOCK_CREATE_PERIOD) * time.Second)
	}

	msg := make(chan string)
	//defer close(msg)
	//go LoadRawdata(PATH, msg)
	go LoadRawdataFromRandom(msg)

	num_tr := func() int {
		if 0 < config.NUM_TRANSACTION_BLOCK {
			return config.NUM_TRANSACTION_BLOCK
		} else {
			return rand.Intn(3) + 3
		}
	}()

	ticker := time.NewTicker(time.Second * time.Duration(config.BLOCK_CREATE_PERIOD))
	defer ticker.Stop()
	for {
		<-ticker.C
		if h.Ready {
			var trs []*blockchain.Transaction
			for i := 0; i < num_tr; i++ {
				d := <-msg
				if d == config.END_TEST {
					h.broadcastEndTest()
					h.db.Close()
					//syscall.Kill(syscall.Getpid(), syscall.SIGINT)
					h.KillProcess()
					return
				}
				// log.Printf("==>%s", d)
				tr := blockchain.CreateTransaction(w, []byte(d))
				trs = append(trs, tr)
			}
			hash, height := h.db.GetLatestBlockHash()
			b := blockchain.CreateBlock(trs, []byte(hash), height)
			h.db.AddBlock(b)
			h.broadcastNewBlock(b)
			log.Printf("Broadcast a new block : %v", cnt)
			cnt++
		}

	}
}

func (h *Handler) Stop() {
	h.Ready = false
}

func NewBCDummy(db dbagent.DBAgent, nodes *map[string]dtype.NodeInfo) *Handler {
	h := Handler{db, false, nodes}
	log.Printf("start : %v", nodes)
	return &h
}
