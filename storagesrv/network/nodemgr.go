package network

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/dtype"
)

type NodeMgr struct {
	Neighbours []dtype.NodeInfo
	mutex      sync.Mutex
}

var version dtype.Version = dtype.Version{Ver: 1}

func (n *NodeMgr) Synbc() bool {
	return false
}

func (n *NodeMgr) Update(sim dtype.Simulator, local dtype.NodeInfo) {
	checkVer := func(ip string, port int, hash string) {
		url := fmt.Sprintf("ws://%v:%v/version", ip, port)
		log.Printf("Update neighbour checking version : %v", url)

		ws, _, err := websocket.DefaultDialer.Dial(url, nil)

		if err != nil {
			if hash != "" {
				n.mutex.Lock()
				for i, nei := range n.Neighbours {
					if nei.Hash == hash {
						n.Neighbours = append(n.Neighbours[:i], n.Neighbours[i+1:]...)
					}
				}
				n.mutex.Unlock()
				log.Printf("Remove node because checking version error : %v", err)
			}
			return
		}
		defer ws.Close()

		if err := ws.WriteJSON(version); err != nil {
			log.Printf("Write json error : %v", err)
			return
		}

		var nodes []dtype.NodeInfo
		if err := ws.ReadJSON(&nodes); err != nil {
			log.Printf("Read json error : %v", err)
			return
		}

		for _, nh := range nodes {
			if nh.Hash != local.Hash {
				n.mutex.Lock()
				find := false
				for _, nn := range n.Neighbours {
					if nh == nn {
						find = true
					}
				}
				if !find {
					n.Neighbours = append(n.Neighbours, nh)
				}
				n.mutex.Unlock()
			}
		}
		// for _, nn := range n.Neighbours {
		// 	log.Printf("Update neighbours : %v", nn)
		// }
	}

	n.mutex.Lock()
	neighbours := make([]dtype.NodeInfo, len(n.Neighbours))
	copy(neighbours, n.Neighbours)
	n.mutex.Unlock()

	for _, node := range neighbours {
		checkVer(node.IP, node.Port, node.Hash)
	}

	if len(neighbours) == 0 {
		checkVer(sim.IP, sim.Port, "")
	}
}

func NewNodeMgr() *NodeMgr {
	nm := NodeMgr{Neighbours: []dtype.NodeInfo{}, mutex: sync.Mutex{}}

	return &nm
}
