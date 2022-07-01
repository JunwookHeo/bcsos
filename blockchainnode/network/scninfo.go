package network

import (
	"log"
	"sync"

	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/wallet"
)

type scnInfo struct {
	scnodes [][]dtype.NodeInfo
	mutex   sync.Mutex
}

func NewSCNInfo() *scnInfo {
	// return &scnInfo{local: l, scnodes: make([]scnList, config.MAX_SC), mutex: sync.Mutex{}}
	scn := scnInfo{}
	scn.mutex = sync.Mutex{}
	scn.scnodes = make([][]dtype.NodeInfo, config.MAX_SC)
	for i := 0; i < config.MAX_SC; i++ {
		scn.scnodes[i] = make([]dtype.NodeInfo, config.MAX_SC_PEER)
		for j := 0; j < config.MAX_SC_PEER; j++ {
			scn.scnodes[i][j] = zeroNode()
		}
	}
	return &scn
}

func zeroNode() dtype.NodeInfo {
	return dtype.NodeInfo{Mode: "", SC: 0, IP: "", Port: 0, Hash: ""}
}

func (c *scnInfo) AddNSCNNode(n dtype.NodeInfo) {
	ni := NodeInfoInst()
	local := ni.GetLocalddr()

	if n.SC >= config.MAX_SC || n.Hash == local.Hash {
		return
	}
	c.mutex.Lock()
	defer c.mutex.Unlock()
	pos := 0
	for _, peer := range c.scnodes[n.SC] {
		if peer.Hash == n.Hash {
			return
		}
		if peer.Hash == "" {
			break
		}
		d1 := wallet.DistanceXor(local.Hash, n.Hash)
		d2 := wallet.DistanceXor(local.Hash, peer.Hash)
		if d1 < d2 { // if new node is closer than cur node, insert new node
			break
		}
		pos++
	}

	switch pos {
	case 0:
		for i := config.MAX_SC_PEER - 1; 0 < i; i-- {
			c.scnodes[n.SC][i] = c.scnodes[n.SC][i-1]
		}
		c.scnodes[n.SC][0] = n
	case config.MAX_SC_PEER - 1:
		c.scnodes[n.SC][pos] = n
	case config.MAX_SC_PEER:
		break
	default:
		for i := config.MAX_SC_PEER - 1; pos < i; i-- {
			c.scnodes[n.SC][i] = c.scnodes[n.SC][i-1]
		}
		c.scnodes[n.SC][pos] = n

	}
}

func (c *scnInfo) DeleteSCNNode(n dtype.NodeInfo) {
	if n.SC >= config.MAX_SC {
		return
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	pos := 0
	for _, peer := range c.scnodes[n.SC] {
		if peer.Hash == n.Hash {
			break
		}
		pos++
	}
	switch pos {
	case 0:
		//c.scnodes[n.SC] = append(c.scnodes[n.SC][1:], []dtype.NodeInfo{zeroNode()}...)
		for i := 0; i < config.MAX_SC_PEER-1; i++ {
			c.scnodes[n.SC][i] = c.scnodes[n.SC][i+1]
		}
		c.scnodes[n.SC][config.MAX_SC_PEER-1] = zeroNode()
	case config.MAX_SC_PEER - 1:
		c.scnodes[n.SC][pos] = zeroNode()
	case config.MAX_SC_PEER:
		break
	default:
		// tmp := append(c.scnodes[n.SC][pos+1:], []dtype.NodeInfo{zeroNode()}...)
		// c.scnodes[n.SC] = append(c.scnodes[n.SC][:pos], tmp...)
		for i := pos; i < config.MAX_SC_PEER-1; i++ {
			c.scnodes[n.SC][i] = c.scnodes[n.SC][i+1]
		}
		c.scnodes[n.SC][config.MAX_SC_PEER-1] = zeroNode()

	}
}

// return the copy of node list due to conccurency issues
func (c *scnInfo) GetSCNNodeList(sc int, nodes *[config.MAX_SC_PEER]dtype.NodeInfo) bool {
	if sc >= config.MAX_SC {
		return false
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	//newscn := []dtype.NodeInfo{}

	pos := 0
	for _, peer := range c.scnodes[sc] {
		if peer.Hash != "" {
			//newscn = append(newscn, []dtype.NodeInfo{peer}...)
			nodes[pos] = peer
			pos++
		}
	}

	return pos > 0
}

func (c *scnInfo) GetSCNNodeListbyDistance(sc int, oid string, nodes *[config.MAX_SC_PEER]dtype.NodeInfo) bool {
	if sc >= config.MAX_SC {
		return false
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	pos := 0
	var dists [config.MAX_SC_PEER]uint64
	for _, peer := range c.scnodes[sc] {
		if peer.Hash != "" {
			dists[pos] = wallet.DistanceXor(oid, peer.Hash)
			nodes[pos] = peer
			pos++
		}
	}
	if 1 < pos {
		// Buble sort by distance
		for i := 0; i < pos; i++ {
			for j := 1; j < pos-i; j++ {
				if dists[j] < dists[j-1] {
					dists[j], dists[j-1] = dists[j-1], dists[j]
					nodes[j], nodes[j-1] = nodes[j-1], nodes[j]
				}
			}
		}
	}

	return pos > 0
}

func (c *scnInfo) GetSCNNodeListAll(nodes *[(config.MAX_SC) * config.MAX_SC_PEER]dtype.NodeInfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	//newscn := []dtype.NodeInfo{}
	pos := 0
	for _, scn := range c.scnodes {
		for _, peer := range scn {
			if peer.Hash != "" {
				//newscn = append(newscn, []dtype.NodeInfo{peer}...)
				nodes[pos] = peer
				pos++
			}
		}
	}
}

func (c *scnInfo) ShowSCNNodeList() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for i, scn := range c.scnodes {
		for _, peer := range scn {
			if peer.Hash != "" {
				log.Printf("SC:%v, node:%v", i, peer)
			}
		}
	}
}
