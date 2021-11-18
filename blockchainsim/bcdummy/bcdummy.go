package bcdummy

import (
	"fmt"
	"log"
	"math/rand"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
)

type Handler struct {
	db    dbagent.DBAgent
	Ready bool
	Nodes *map[string]dtype.NodeInfo
}

const PATH = "./iotdata/IoT_normal_fridge_1.log"

func (h *Handler) sendNewBlock(b *blockchain.Block, ip string, port int) {
	url := fmt.Sprintf("ws://%v:%v/newblock", ip, port)
	log.Printf("Send new block to %v", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Printf("dial: %v", err)
		return
	}
	defer ws.Close()
	log.Printf("DefaultDialer Send new block to %v", url)

	if err := ws.WriteJSON(b); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}

	for _, t := range b.Transactions {
		log.Printf("Recevied node : %s", t.Data)
	}

}

func (h *Handler) broadcastNewBlock(b *blockchain.Block) {
	log.Printf("broadcast : %v", *h.Nodes)
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

func (h *Handler) Start() {
	// Send Genesis Block
	hash := h.db.GetLatestBlockHash()
	if len(hash) == 0 {
		log.Printf("Create Genesis due to hash : %v", hash)
		b := blockchain.CreateGenesis()
		h.db.AddBlock(b)
		h.broadcastNewBlock(b)
		time.Sleep(time.Duration(config.BLOCK_CREATE_PERIOD) * time.Second)
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

	ticker := time.NewTicker(time.Duration(config.BLOCK_CREATE_PERIOD) * time.Second)
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
					syscall.Kill(syscall.Getpid(), syscall.SIGINT)
					return
				}
				log.Printf("==>%s", d)
				tr := blockchain.CreateTransaction([]byte(d))
				trs = append(trs, tr)
			}
			b := blockchain.CreateBlock(trs, []byte(h.db.GetLatestBlockHash()))
			h.db.AddBlock(b)
			h.broadcastNewBlock(b)
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
