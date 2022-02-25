package datalib

import "log"

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

type TreeChain struct {
	root  *TreeNode
	hnode *TreeNode
}

func (tc *TreeChain) NewRoot(block *BlockData) {
	tc.root = NewTreeNode(block)
	tc.hnode = tc.root
}

func (tc *TreeChain) SetRoot(node *TreeNode) {
	tc.root = node
	tc.root.SetParent(nil)
}

func (tc *TreeChain) GetHighestBlock() *BlockData {
	if tc.hnode == nil {
		return nil
	}
	return tc.hnode.GetBlockData()
}

func (tc *TreeChain) UpdateRoot() {
	if tc.hnode == nil || tc.root == nil {
		return
	}
	tmp := tc.hnode.GetParent()
	for {
		if tmp == nil {
			return
		}

		if tc.hnode.block.Height-tmp.block.Height > MAX_TREE_DEPTH {
			tc.SetRoot(tmp)
			return
		}

		tmp = tmp.GetParent()
	}
}
func (tc *TreeChain) AddTreeNode(block *BlockData) {
	if tc.root == nil {
		block.Height = 0
		tc.NewRoot(block)
		return
	} else {
		node := tc.FindNode(block.Prev)
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

func (tc *TreeChain) FindChildNode(hash string, node *TreeNode) *TreeNode {
	if node == nil {
		return nil
	}

	child := node.GetChild()
	if child != nil {
		dest := tc.FindChildNode(hash, child)
		if dest != nil {
			return dest
		}
	}

	block := node.GetBlockData()
	if block.Hash == hash {
		return node
	}

	node = node.GetSibling()
	dest := tc.FindChildNode(hash, node)
	if dest != nil {
		return dest
	}

	return nil
}

func (tc *TreeChain) FindNode(hash string) *TreeNode {
	return tc.FindChildNode(hash, tc.root)
}

func (tc *TreeChain) PrintNode(node *TreeNode) {
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

func (tc *TreeChain) PrintAll() {
	node := tc.root
	log.Printf("=========Display TreeChain ==========")
	tc.PrintNode(node)
}

func NewTreeChain() *TreeChain {
	tc := TreeChain{root: nil, hnode: nil}
	return &tc
}
