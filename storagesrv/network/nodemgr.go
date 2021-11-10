package network

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/common/dtype"
)

type NodeMgr struct {
	Neighbours map[string]dtype.NodeInfo
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
				delete(n.Neighbours, hash)
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

		for _, l := range nodes {
			if l.Hash != local.Hash {
				n.Neighbours[l.Hash] = l
			}
		}

		log.Printf("Update neighbours : %v", n.Neighbours)
	}
	for _, node := range n.Neighbours {
		checkVer(node.IP, node.Port, node.Hash)
	}

	if len(n.Neighbours) == 0 {
		checkVer(sim.IP, sim.Port, "")
	}
}

func (n *NodeMgr) VersionHandler(ws *websocket.Conn, w http.ResponseWriter, r *http.Request) {

	var version dtype.Version
	if err := ws.ReadJSON(&version); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}
	log.Printf("receive version : %v", version)

	var nodes []dtype.NodeInfo
	for _, n := range n.Neighbours {
		nodes = append(nodes, n)
	}
	if err := ws.WriteJSON(nodes); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}
}

func NewNodeMgr() *NodeMgr {
	nm := NodeMgr{Neighbours: make(map[string]dtype.NodeInfo)}

	return &nm
}
