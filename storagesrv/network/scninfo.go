package network

import (
	"log"
	"math/big"
	"sync"

	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dtype"
)

// type scnNode struct {
// 	next *scnNode
// 	pre  *scnNode
// 	node dtype.NodeInfo
// }
// type scnList struct {
// 	header *scnNode
// 	count  int
// }

// type scnInfo struct {
// 	local   *dtype.NodeInfo
// 	scnodes []scnList
// 	mutex   sync.Mutex
// }

type scnInfo struct {
	local   *dtype.NodeInfo
	scnodes [][]dtype.NodeInfo
	mutex   sync.Mutex
}

func NewSCNInfo(l *dtype.NodeInfo) *scnInfo {
	// return &scnInfo{local: l, scnodes: make([]scnList, config.MAX_SC+1), mutex: sync.Mutex{}}
	scn := scnInfo{}
	scn.local = l
	scn.mutex = sync.Mutex{}
	scn.scnodes = make([][]dtype.NodeInfo, config.MAX_SC+1)
	for i := 0; i < config.MAX_SC+1; i++ {
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

func xordistance(h1 string, h2 string) *big.Int {
	n1, _ := new(big.Int).SetString(h1, 16)
	n2, _ := new(big.Int).SetString(h2, 16)
	return new(big.Int).Xor(n1, n2)
}

func (c *scnInfo) AddNSCNNode(n dtype.NodeInfo) {
	if n.SC > config.MAX_SC || n.Hash == c.local.Hash {
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
		d1 := xordistance(c.local.Hash, n.Hash)
		d2 := xordistance(c.local.Hash, peer.Hash)
		if d1.Cmp(d2) < 0 { // if new node is closer than cur node, insert new node
			break
		}
		pos++
	}

	switch pos {
	case 0:
		//c.scnodes[n.SC] = append([]dtype.NodeInfo{n}, c.scnodes[n.SC][1:]...)
		for i := config.MAX_SC_PEER - 1; 0 < i; i-- {
			c.scnodes[n.SC][i] = c.scnodes[n.SC][i-1]
		}
		c.scnodes[n.SC][0] = n
	case config.MAX_SC_PEER - 1:
		c.scnodes[n.SC][pos] = n
	case config.MAX_SC_PEER:
		break
	default:
		//tmp := append([]dtype.NodeInfo{n}, c.scnodes[n.SC][pos+1:]...)
		//c.scnodes[n.SC] = append(c.scnodes[n.SC][:pos], tmp...)
		for i := config.MAX_SC_PEER - 1; pos < i; i-- {
			c.scnodes[n.SC][i] = c.scnodes[n.SC][i-1]
		}
		c.scnodes[n.SC][pos] = n

	}
}

func (c *scnInfo) DeleteSCNNode(n dtype.NodeInfo) {
	if n.SC > config.MAX_SC {
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
	if sc > config.MAX_SC {
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

// return the copy of node list due to conccurency issues
func (c *scnInfo) GetSCNNodeListbyDistance(sc int, oid string, nodes *[config.MAX_SC_PEER]dtype.NodeInfo) bool {
	if sc > config.MAX_SC {
		return false
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	pos := 0
	var dists [config.MAX_SC_PEER]*big.Int
	for _, peer := range c.scnodes[sc] {
		if peer.Hash != "" {
			dists[pos] = xordistance(oid, peer.Hash)
			nodes[pos] = peer
			pos++
		}
	}
	if 1 < pos {
		// Buble sort by distance
		for i := 0; i < pos; i++ {
			for j := 1; j < pos-i; j++ {
				if dists[j].Cmp(dists[j-1]) < 0 {
					dists[j], dists[j-1] = dists[j-1], dists[j]
					nodes[j], nodes[j-1] = nodes[j-1], nodes[j]
				}
			}
		}
	}

	return pos > 0
}

func (c *scnInfo) GetSCNNodeListAll(nodes *[(config.MAX_SC + 1) * config.MAX_SC_PEER]dtype.NodeInfo) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	//newscn := []dtype.NodeInfo{}
	pos := 0
	for _, scn := range c.scnodes {
		for _, peer := range scn {
			if peer.Hash != "" {
				//newscn = append(newscn, []dtype.NodeInfo{peer}...)
				nodes[pos] = peer
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
