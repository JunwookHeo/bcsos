package network

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dtype"
)

type NodeMgr struct {
	scn   scnInfo //[]dtype.NodeInfo
	mutex sync.Mutex
}

var (
	nm   *NodeMgr
	once sync.Once
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// Update peers list
func (n *NodeMgr) UpdatePeerList(sim *dtype.NodeInfo, local *dtype.NodeInfo) {
	sendPing := func(node dtype.NodeInfo) {
		url := fmt.Sprintf("ws://%v:%v/ping", node.IP, node.Port)
		//log.Printf("Send ping with local info : %v", url)

		ws, _, err := websocket.DefaultDialer.Dial(url, nil)
		if err != nil {
			if node.Hash != "" {
				n.scn.DeleteSCNNode(node)
				log.Printf("Remove node because ping error : %v", err)
			}
			return
		}
		defer ws.Close()

		if err := ws.WriteJSON(local); err != nil {
			log.Printf("Write json error : %v", err)
			return
		}

		var nodes []dtype.NodeInfo
		if err := ws.ReadJSON(&nodes); err != nil {
			log.Printf("Read json error : %v", err)
			return
		}
		//log.Printf("Read json nodes : %v", nodes)
		for _, nh := range nodes {
			if nh.Hash != "" && nh.Hash != local.Hash {
				n.scn.AddNSCNNode(nh)
			}
		}
	}

	checked := false
	for i := 0; i < config.MAX_SC; i++ {
		var nodes [config.MAX_SC_PEER]dtype.NodeInfo
		if n.scn.GetSCNNodeList(i, &nodes) {
			for _, node := range nodes {
				sendPing(node)
				checked = true
			}
		}
	}

	if !checked {
		sendPing(*sim)
	}

	//n.scn.ShowSCNNodeList()
}

// Send response to connector with its local information
func (n *NodeMgr) pingHandler(w http.ResponseWriter, r *http.Request) {
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
	ni := NodeInfoInst()
	local := ni.GetLocalddr()
	nm := NodeMgrInst()

	if peer.Hash != "" && peer.Hash != local.Hash {
		nm.AddNSCNNode(peer)
	}

	// Send peers info to the connector
	var nodes [(config.MAX_SC) * config.MAX_SC_PEER]dtype.NodeInfo
	nm.GetSCNNodeListAll(&nodes)
	if err := ws.WriteJSON(nodes); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}
}

func (n *NodeMgr) AddNSCNNode(node dtype.NodeInfo) {
	n.scn.AddNSCNNode(node)
}

func (n *NodeMgr) GetSCNNodeListbyDistance(sc int, oid string, nodes *[config.MAX_SC_PEER]dtype.NodeInfo) bool {
	return n.scn.GetSCNNodeListbyDistance(sc, oid, nodes)
}

func (n *NodeMgr) GetSCNNodeListAll(nodes *[(config.MAX_SC) * config.MAX_SC_PEER]dtype.NodeInfo) {
	n.scn.GetSCNNodeListAll(nodes)
}

func (n *NodeMgr) SetHttpRouter(m *mux.Router) {
	m.HandleFunc("/ping", n.pingHandler)
}

func NodeMgrInst() *NodeMgr {
	once.Do(func() {
		nm = &NodeMgr{
			scn:   *NewSCNInfo(),
			mutex: sync.Mutex{},
		}
	})

	return nm
}
