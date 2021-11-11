package bcdummy

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
)

type Handler struct {
	db    dbagent.DBAgent
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

	// var block blockchain.Block
	// if err := ws.ReadJSON(&block); err != nil {
	// 	log.Printf("Read json error : %v", err)
	// 	return
	// }

	for _, t := range b.Transactions {
		log.Printf("Recevied node : %s", t.Data)
	}

}

func (h *Handler) broadcastNewBlock(b *blockchain.Block) {
	log.Printf("broadcast : %v", *h.Nodes)
	for _, node := range *h.Nodes {
		ip := node.IP
		if node.Mode == "docker" {
			ip = "localhost"
		}
		go h.sendNewBlock(b, ip, node.Port)
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
		time.Sleep(5 * time.Second)
	}

	msg := make(chan string)
	go LoadRawdata(PATH, msg)

	for {
		num := rand.Intn(3) + 1
		var trs []*blockchain.Transaction
		for i := 0; i < num; i++ {
			d := <-msg
			log.Printf("==>%s", d)
			tr := blockchain.CreateTransaction([]byte(d))
			trs = append(trs, tr)
		}
		b := blockchain.CreateBlock(trs, []byte(h.db.GetLatestBlockHash()))
		h.db.AddBlock(b)
		h.broadcastNewBlock(b)
		time.Sleep(5 * time.Second)
	}
}

func Stop() {

}

func NewBCDummy(db dbagent.DBAgent, nodes *map[string]dtype.NodeInfo) *Handler {
	h := Handler{db, nodes}
	log.Printf("start : %v", nodes)
	return &h
}
