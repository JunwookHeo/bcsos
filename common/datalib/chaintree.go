package datalib

import (
	"log"
	"sync"
)

type BlockData struct {
	Height int
	Hash   string
	Prev   string
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
	root      *TreeNode
	hnode     *TreeNode
	danglings []BlockData
}

func (tc *ChainTree) newRoot(block *BlockData) {
	tc.root = NewTreeNode(block)
	tc.hnode = tc.root
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
func (tc *ChainTree) AddTreeNode(block *BlockData) {
	tc.mutex.Lock()
	defer tc.mutex.Unlock()

	if tc.root == nil {
		block.Height = 0
		tc.newRoot(block)
		return
	} else {
		node := tc.findNode(block.Prev)
		if node == nil {
			log.Panicf("Not found previous hash block")
			return
		}

		child := node.GetChild()
		if child == nil {
			block.Height = node.block.Height + 1
			tmp := NewTreeNode(block)
			tmp.SetParent(node)
			node.SetChild(tmp)
			tc.hnode = node.GetChild()
			return
		}

		node = child
		for {
			sibling := node.GetSibling()
			if sibling == nil {
				block.Height = node.block.Height
				tmp := NewTreeNode(block)
				tmp.SetParent(node.GetParent())
				node.SetSibling(tmp)

				return
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

func NewChainTree() *ChainTree {
	tc := ChainTree{root: nil, hnode: nil, danglings: make([]BlockData, 0)}
	return &tc
}
