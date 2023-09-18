package simulation

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/blockdata"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/wallet"
)

type Handler struct {
	w     *wallet.Wallet
	db    dbagent.DBAgent
	Ready bool
	Nodes *map[string]dtype.NodeInfo
}

const PATH_BTC_BLOCK = "../blocks_720.json"
const PATH_ETH_BLOCK = "../blocks_eth_720.json"

func init() {
}

// getObjectQuery queries a transaction ot other nodes with highr Storage Class
// Request : hash of transaction
// Response : transaction
func (h *Handler) getObjectQuery(reqData *dtype.ReqData, obj interface{}) bool {
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

		hop := reqData.SC
		h.db.UpdateDBNetworkDelay(int(time.Now().UnixNano()-reqData.Timestamp), hop)
		log.Printf("==>Query read reqData: %v[hop], %v", hop, reqData)
		return true
	}

	// the number of query to other nodes
	defer h.db.UpdateDBNetworkQuery(0, 1, 1)

	// var maxdist big.Int
	var maxdist uint64
	var tnode dtype.NodeInfo

	for _, node := range *h.Nodes {
		if node.SC == 0 {
			// dist := wallet.DistanceXor(reqData.ObjHash, node.Hash)
			// if maxdist.Cmp(dist) == -1 {
			// 	maxdist = *dist
			// 	tnode = node
			// }
			dist := wallet.DistanceXor(reqData.ObjHash, node.Hash)
			if maxdist < dist {
				maxdist = dist
				tnode = node
			}
		}
	}

	return queryObject(tnode.IP, tnode.Port, reqData, obj)
}

func (h *Handler) newReqData(objtype string, hash string) dtype.ReqData {
	req := dtype.ReqData{}
	// req.Addr = fmt.Sprintf("%v:%v", local.IP, local.Port)
	req.Timestamp = time.Now().UnixNano()
	req.SC = -1
	req.Hop = 0
	req.ObjType = objtype
	req.ObjHash = hash

	return req
}

func (h *Handler) getObjectByAccessPattern(num int, hashes *[]dbagent.RemoverbleObj) bool {
	if config.ACCESS_FREQUENCY_PATTERN == config.RANDOM_ACCESS_PATTERN {
		return h.db.GetTransactionwithUniform(num, hashes)
	} else {
		return h.db.GetTransactionwithExponential(num, hashes)
	}
}

func (h *Handler) SimulateAccessPattern(pid *int) bool {
	hashes := []dbagent.RemoverbleObj{}
	if !h.getObjectByAccessPattern(1, &hashes) {
		return false
	}

	for _, hash := range hashes {
		log.Printf("%v(%v) : %v", config.ACCESS_FREQUENCY_PATTERN, *pid, hashes)
		*pid++
		if hash.HashType == 0 {
			bh := blockchain.BlockHeader{}
			req := h.newReqData("blockheader", hash.Hash)
			if h.getObjectQuery(&req, &bh) {
				h.db.AddBlockHeader(hash.Hash, &bh)
				if hash.Hash != hex.EncodeToString(bh.GetHash()) {
					log.Printf("%v header Hash not equal %v", hash.Hash, hex.EncodeToString(bh.GetHash()))
				}
			}
		} else {
			tr := blockchain.Transaction{}
			req := h.newReqData("transaction", hash.Hash)
			if h.getObjectQuery(&req, &tr) {
				h.db.AddTransaction(&tr)
				if hash.Hash != hex.EncodeToString(tr.Hash) {
					log.Printf("%v Tr Hash not equal %v", hash.Hash, hex.EncodeToString(tr.Hash))
				}
			}
		}
	}

	return true
}

func (h *Handler) sendNewTrnsaction(tr *blockchain.Transaction, ip string, port int) bool {
	url := fmt.Sprintf("ws://%v:%v/broadcastransaction", ip, port)
	// log.Printf("Send new transaction to %v", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Printf("dial: %v", err)
		return false
	}
	defer ws.Close()
	// log.Printf("DefaultDialer Send new block to %v", url)

	if err := ws.WriteJSON(tr); err != nil {
		log.Printf("Write json error : %v", err)
		return false
	}

	return true
}

func (h *Handler) broadcastNewTransaction(b *blockchain.Transaction) bool {
	num := len(*h.Nodes)
	if num == 0 {
		return true
	}

	idx := rand.Intn(num)
	i := 0
	for _, node := range *h.Nodes {
		if i == idx {
			return h.sendNewTrnsaction(b, node.IP, node.Port)
		}
		i++
	}
	return false
}

func (h *Handler) SimulateTransaction(id int) *blockchain.Transaction {
	sensordata := SensorData{Id: id, Timestamp: time.Now().UnixNano(), Temperature: (rand.Float64()*80. - 30.), Condition: genRandString()}
	jstr, err := json.Marshal(&sensordata)
	if err != nil {
		log.Panicf("gen error : %v", err)
		return nil
	}

	tr := blockchain.CreateTransaction(h.w, jstr)
	// log.Printf("Creating a new tr (%v) : %v", id, hex.EncodeToString(tr.Hash))
	for {
		if h.broadcastNewTransaction(tr) {
			break
		}
	}

	return tr
}

func (h *Handler) SimulateBtcBlock(msg chan blockdata.BlockPkt) {
	switch config.BLOCK_DATA_TYPE {
	case config.BITCOIN_BLOCK:
		go LoadBtcData(PATH_BTC_BLOCK, msg)
	case config.ETHEREUM_BLOCK:
		go LoadEthData(PATH_ETH_BLOCK, msg)
	}
}

func (h *Handler) sendNewBtcBlock(b *blockdata.BlockPkt, ip string, port int) bool {
	url := fmt.Sprintf("ws://%v:%v/broadcastnewbtcblock", ip, port)
	// log.Printf("Send new BTC block to %v", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Printf("dial: %v", err)
		return false
	}
	defer ws.Close()
	// log.Printf("DefaultDialer Send new block to %v", url)

	if err := ws.WriteJSON(b); err != nil {
		log.Printf("Write json error : %v", err)
		return false
	}

	return true
}

func (h *Handler) broadcastNewBtcBlock(b *blockdata.BlockPkt) bool {
	num := len(*h.Nodes)
	if num == 0 {
		return true
	}

	idx := rand.Intn(num)
	i := 0
	for _, node := range *h.Nodes {
		if i == idx {
			return h.sendNewBtcBlock(b, node.IP, node.Port)
		}
		i++
	}
	return false
}

func (h *Handler) BroadcastBtcBlock(b *blockdata.BlockPkt) {
	for {
		if h.broadcastNewBtcBlock(b) {
			break
		}
	}
}

func NewSimAgent(w *wallet.Wallet, db dbagent.DBAgent, nodes *map[string]dtype.NodeInfo) *Handler {
	h := Handler{w, db, false, nodes}
	log.Printf("start : %v", nodes)
	return &h
}
