package datalib

import (
	"log"
	"sync"
)

type BlockData struct {
	Height    int
	Timestamp int64
	Hash      string
	Prev      string
}

type TreeNode struct {
	block   *BlockData
	parent  *TreeNode
	child   *TreeNode
	sibling *TreeNode
}

const MAX_TREE_DEPTH int = 10

func (tn *TreeNode) SetBlockData(b *BlockData) {
	tn.block = b
}

func (tn *TreeNode) GetBlockData() *BlockData {
	return tn.block
}

func (tn *TreeNode) SetChild(child *TreeNode) {
	tn.child = child
}

func (tn *TreeNode) SetParent(parent *TreeNode) {
	tn.parent = parent
}

func (tn *TreeNode) GetParent() *TreeNode {
	return tn.parent
}

func (tn *TreeNode) GetChild() *TreeNode {
	return tn.child
}

func (tn *TreeNode) SetSibling(sibling *TreeNode) {
	tn.sibling = sibling
}

func (tn *TreeNode) GetSibling() *TreeNode {
	return tn.sibling
}

func NewTreeNode(b *BlockData) *TreeNode {
	n := &TreeNode{block: b, parent: nil, child: nil, sibling: nil}
	return n
}

type ChainTree struct {
	mutex     sync.Mutex
	threshold int // When a node receives n(threshold) consecutive danglings, it resets the chaintree.
	root      *TreeNode
	hnode     *TreeNode
	danglings *BcSortedList
	newlist   *BcNewList
}

func (tc *ChainTree) newRoot(block *BlockData) {
	tc.root = NewTreeNode(block)
	tc.hnode = tc.root
}

func (tc *ChainTree) getRootTimestamp() int64 {
	if tc.root == nil || tc.root.block == nil {
		return 0
	}

	return tc.root.block.Timestamp
}

func (tc *ChainTree) setRoot(node *TreeNode) {
	tc.root = node
	tc.root.SetParent(nil)
}

func (tc *ChainTree) GetHighestBlock() *BlockData {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if tc.hnode == nil {
		return nil
	}
	return tc.hnode.GetBlockData()
}

func (tc *ChainTree) UpdateRoot() {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if tc.hnode == nil || tc.root == nil {
		return
	}
	tmp := tc.hnode.GetParent()
	for {
		if tmp == nil {
			return
		}

		if tc.hnode.block.Height-tmp.block.Height > MAX_TREE_DEPTH {
			tc.setRoot(tmp)
			return
		}

		tmp = tmp.GetParent()
	}
}
func (tc *ChainTree) AddTreeNode(block *BlockData, isnew bool) bool {

	ret := tc._addTreeNode(block, isnew)
	if ret == true {
		tc.newlist.Push(block)
		// tc.newlist.ShowAll()
	}
	return ret
}

func (tc *ChainTree) _addTreeNode(block *BlockData, isnew bool) bool {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if tc.root == nil {
		tc.threshold = 0
		tc.newRoot(block)
		return true
	} else {
		node := tc.findNode(block.Prev)
		if node == nil {
			if block.Prev == tc.root.block.Prev { // Multiple Genesis
				node = tc.root
				for {
					sibling := node.GetSibling()
					if sibling == nil {
						tc.threshold = 0
						tmp := NewTreeNode(block)
						tmp.SetParent(node.GetParent())
						node.SetSibling(tmp)
						return true
					}
					node = sibling
				}
			} else {

				log.Printf("Not found previous hash block : %v", block)
				if isnew == true { // Add block if it is a new block.
					if tc.threshold > 6 { //  more 6 consecutive danglings received
						tc.threshold = 0
						// block.Height = 0
						tc.newRoot(block)
					} else {
						tc.threshold++
						tc.danglings.Add(block)
					}
				}
				return false
			}
		}

		child := node.GetChild()
		if child == nil {
			tc.threshold = 0
			// block.Height = node.block.Height + 1
			tmp := NewTreeNode(block)
			tmp.SetParent(node)
			node.SetChild(tmp)
			tc.hnode = node.GetChild()
			return true
		}

		node = child
		for {
			sibling := node.GetSibling()
			if sibling == nil {
				tc.threshold = 0
				// block.Height = node.block.Height
				tmp := NewTreeNode(block)
				tmp.SetParent(node.GetParent())
				node.SetSibling(tmp)
				return true
			}
			node = sibling
		}
	}
}

func (tc *ChainTree) _addTreeNode2(block *BlockData, isnew bool) bool {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if tc.root == nil {
		tc.threshold = 0
		// block.Height = 0
		tc.newRoot(block)
		return true
	} else {
		node := tc.findNode(block.Prev)
		if node == nil {
			log.Printf("Not found previous hash block : %v", block)
			if isnew == true { // Add block if it is a new block.
				if tc.threshold > 6 { //  more 6 consecutive danglings received
					tc.threshold = 0
					// block.Height = 0
					tc.newRoot(block)
				} else {
					tc.threshold++
					tc.danglings.Add(block)
				}
			}
			return false
		}

		child := node.GetChild()
		if child == nil {
			tc.threshold = 0
			// block.Height = node.block.Height + 1
			tmp := NewTreeNode(block)
			tmp.SetParent(node)
			node.SetChild(tmp)
			tc.hnode = node.GetChild()
			return true
		}

		node = child
		for {
			sibling := node.GetSibling()
			if sibling == nil {
				tc.threshold = 0
				// block.Height = node.block.Height
				tmp := NewTreeNode(block)
				tmp.SetParent(node.GetParent())
				node.SetSibling(tmp)
				return true
			}
			node = sibling
		}
	}
}

func (tc *ChainTree) findChildNode(hash string, node *TreeNode) *TreeNode {
	if node == nil {
		return nil
	}

	child := node.GetChild()
	if child != nil {
		dest := tc.findChildNode(hash, child)
		if dest != nil {
			return dest
		}
	}

	block := node.GetBlockData()
	if block.Hash == hash {
		return node
	}

	node = node.GetSibling()
	dest := tc.findChildNode(hash, node)
	if dest != nil {
		return dest
	}

	return nil
}

func (tc *ChainTree) findNode(hash string) *TreeNode {
	return tc.findChildNode(hash, tc.root)
}

func (tc *ChainTree) UpdateDanglings() {
	for {
		i, d := tc.danglings.Next()
		if i < 0 {
			break
		}

		// call it with false because this block is dangling.
		if tc.AddTreeNode(d, false) {
			log.Printf("add dangling block to the list : %v", d)
		}
	}

	tc.danglings.Update(tc.getRootTimestamp())
}

func (tc *ChainTree) PrintNode(node *TreeNode) {
	if node == nil {
		return
	}

	child := node.GetChild()
	if child != nil {
		tc.PrintNode(child)
	}

	log.Printf("Node : %v", *node.GetBlockData())

	sibling := node.GetSibling()
	tc.PrintNode(sibling)
}

func (tc *ChainTree) PrintAll() {
	node := tc.root
	log.Printf("=========Display ChainTree ==========")
	tc.PrintNode(node)
}

func (tc *ChainTree) ShowDanglings() {
	tc.danglings.ShowAll()
}

func (tc *ChainTree) GetNewBlockInfo() *BcNewList {
	return tc.newlist
}

func NewChainTree() *ChainTree {
	tc := ChainTree{root: nil, threshold: 0, hnode: nil, danglings: NewBcSortedList(), newlist: NewBcNewList()}
	return &tc
}
