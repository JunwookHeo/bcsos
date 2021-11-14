package network

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dtype"
)

type NodeMgr struct {
	scn   scnInfo //[]dtype.NodeInfo
	mutex sync.Mutex
}

// Update peers list
func (n *NodeMgr) UpdatePeerList(sim dtype.NodeInfo, local dtype.NodeInfo) {
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
	for i := 0; i <= config.MAX_SC; i++ {
		nodes := n.scn.GetSCNNodeList(i)
		for _, node := range nodes {
			sendPing(node)
			checked = true
		}
	}

	if !checked {
		sendPing(sim)
	}

	//n.scn.ShowSCNNodeList()
}

func (n *NodeMgr) AddNSCNNode(node dtype.NodeInfo) {
	n.scn.AddNSCNNode(node)
}

func (n *NodeMgr) GetSCNNodeListAll() []dtype.NodeInfo {
	return n.scn.GetSCNNodeListAll()
}

func (n *NodeMgr) GetSCNNodeList(sc int) []dtype.NodeInfo {
	return n.scn.GetSCNNodeList(sc)
}

func NewNodeMgr(local *dtype.NodeInfo) *NodeMgr {
	nm := NodeMgr{scn: *NewSCNInfo(local), mutex: sync.Mutex{}}

	return &nm
}
