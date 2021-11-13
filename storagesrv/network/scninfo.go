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

// func (c *scnInfo) newSCNNode(n dtype.NodeInfo) *scnNode {
// 	return &scnNode{nil, nil, n}
// }

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
		c.scnodes[n.SC] = append([]dtype.NodeInfo{n}, c.scnodes[n.SC][1:]...)
	case config.MAX_SC_PEER - 1:
		c.scnodes[n.SC][pos] = n
	case config.MAX_SC_PEER:
		break
	default:
		tmp := append([]dtype.NodeInfo{n}, c.scnodes[n.SC][pos+1:]...)
		c.scnodes[n.SC] = append(c.scnodes[n.SC][:pos], tmp...)

	}

	// sl := &c.scnodes[n.SC]
	// if sl.header == nil {
	// 	sl.header = c.newSCNNode(n)
	// 	sl.count = 1
	// 	return
	// } else {
	// 	cur := sl.header
	// 	pre := (*scnNode)(nil)
	// 	d1 := xordistance(c.local.Hash, n.Hash)
	// 	for {
	// 		if cur.node.Hash == n.Hash {
	// 			return
	// 		}

	// 		d2 := xordistance(c.local.Hash, cur.node.Hash)
	// 		if d1.Cmp(d2) < 0 { // if new node is closer than cur node, insert new node
	// 			item := c.newSCNNode(n)
	// 			item.next = cur
	// 			item.pre = pre
	// 			cur.pre = item
	// 			if pre == nil {
	// 				sl.header = item
	// 			} else {
	// 				pre.next = item
	// 			}
	// 			sl.count++
	// 			return
	// 		} else if cur.next == nil { // new node has the largest distance
	// 			item := c.newSCNNode(n)
	// 			cur.next = item
	// 			item.pre = cur
	// 			sl.count++
	// 			return
	// 		}
	// 		cur = cur.next
	// 		pre = cur
	// 	}
	// }
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
		c.scnodes[n.SC] = append(c.scnodes[n.SC][1:], []dtype.NodeInfo{zeroNode()}...)
	case config.MAX_SC_PEER - 1:
		c.scnodes[n.SC][pos] = zeroNode()
	case config.MAX_SC_PEER:
		break
	default:
		tmp := append(c.scnodes[n.SC][pos+1:], []dtype.NodeInfo{zeroNode()}...)
		c.scnodes[n.SC] = append(c.scnodes[n.SC][:pos], tmp...)

	}

	// c.mutex.Lock()
	// defer c.mutex.Unlock()
	// sl := &c.scnodes[n.SC]
	// if sl.header == nil {
	// 	return
	// } else {
	// 	cur := sl.header
	// 	for {
	// 		if cur == nil {
	// 			return
	// 		}
	// 		if cur.node.Hash == n.Hash {
	// 			if cur.pre == nil {
	// 				sl.header = cur.next
	// 			} else {
	// 				cur.pre.next = cur.next
	// 			}
	// 			sl.count--
	// 			return
	// 		}
	// 		cur = cur.next
	// 	}
	// }
}

// return the copy of node list due to conccurency issues
func (c *scnInfo) GetSCNNodeList(sc int) []dtype.NodeInfo {
	if sc > config.MAX_SC {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()
	newscn := []dtype.NodeInfo{}

	for _, peer := range c.scnodes[sc] {
		if peer.Hash != "" {
			newscn = append(newscn, []dtype.NodeInfo{peer}...)
		}
	}

	return newscn
	// scn := &c.scnodes[sc]
	// if scn.header == nil {
	// 	return nil
	// }
	// c.mutex.Lock()
	// defer c.mutex.Unlock()

	// newnl := []dtype.NodeInfo{}
	// node := scn.header

	// for {
	// 	if node == nil {
	// 		break
	// 	}
	// 	newnl = append(newnl, []dtype.NodeInfo{node.node}...)
	// 	node = node.next
	// }
	// return newnl
}

func (c *scnInfo) GetSCNNodeListAll() []dtype.NodeInfo {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	newscn := []dtype.NodeInfo{}
	for _, scn := range c.scnodes {
		for _, peer := range scn {
			if peer.Hash != "" {
				newscn = append(newscn, []dtype.NodeInfo{peer}...)
			}
		}
	}
	return newscn
	// for _, scn := range c.scnodes {
	// 	node := scn.header
	// 	for {
	// 		if node == nil {
	// 			break
	// 		}
	// 		newscn = append(newscn, []dtype.NodeInfo{node.node}...)

	// 		node = node.next
	// 	}
	// }
	// return newscn
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

	// for i, scn := range c.scnodes {
	// 	node := scn.header
	// 	for {
	// 		if node == nil {
	// 			break
	// 		}
	// 		log.Printf("SC:%v, node:%v", i, node.node)
	// 		node = node.next
	// 	}
	// }
}
