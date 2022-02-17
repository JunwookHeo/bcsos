package mining

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/blockchainnode/network"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dtype"
)

type SensorData struct {
	Id          int     `json:"id"`
	Timestamp   int64   `json:"Timestamp"`
	Temperature float64 `json:"Fridge_Temperature"`
	Condition   string  `json:"Temp_Condition"`
}

var letterRunes = []rune("0123456789 abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func genRandString() string {
	n := rand.Intn(70) + 30
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func generateTransactionFromRandom(id int) *blockchain.Transaction {
	wm := WalletMgrInst("")
	w := wm.GetWallet()

	sensordata := SensorData{Id: id, Timestamp: time.Now().UnixNano(), Temperature: (rand.Float64()*80. - 30.), Condition: genRandString()}
	jstr, err := json.Marshal(&sensordata)
	if err != nil {
		log.Panicf("gen error : %v", err)
		return nil
	}

	tr := blockchain.CreateTransaction(w, jstr)

	return tr
}

// sendTransactionwithLocal
func sendTransactionwithLocal(t *blockchain.Transaction) {
	sendTransaction := func(node *dtype.NodeInfo) {
		url := fmt.Sprintf("ws://%v:%v/broadcastransaction", node.IP, node.Port)
		log.Printf("Send sendTransactionwithLocal with local info : %v", url)

		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			log.Printf("sendTransactionwithLocal error : %v", err)
			return
		}
		defer ws.Close()

		if err := ws.WriteJSON(t); err != nil {
			log.Printf("Write json error : %v", err)
			return
		}
	}

	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()
	sendTransaction(local)
}

func SimulateTransaction(id int) {
	t := rand.Intn(config.BLOCK_CREATE_PERIOD * 1000)
	log.Printf("time : %v", t)
	time.Sleep(time.Duration(t) * time.Millisecond)

	if rand.Intn(2) != 0 {
		tr := generateTransactionFromRandom(id)
		log.Printf("===Local : %v", hex.EncodeToString(tr.Hash))
		sendTransactionwithLocal(tr)
	}

	time.Sleep(time.Duration(config.BLOCK_CREATE_PERIOD*1000-t) * time.Millisecond)
}
