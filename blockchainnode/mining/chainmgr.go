package mining

import (
	"encoding/hex"

	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/datalib"
)

type ChainMgr struct {
	tree *datalib.ChainTree
}

func (cm *ChainMgr) AddedNewBlock(h *blockchain.BlockHeader) {
	block := datalib.BlockData{Height: h.Height, Timestamp: h.Timestamp, Hash: hex.EncodeToString(h.Hash), Prev: hex.EncodeToString(h.PrvHash)}
	if cm.tree.AddTreeNode(&block, true) {
		cm.tree.UpdateDanglings() // update dangling blocks
	}
	cm.tree.UpdateRoot()
	// cm.tree.PrintAll()
	// cm.tree.ShowDanglings()
}

func (cm *ChainMgr) GetHighestBlockHash() (int, string) {
	block := cm.tree.GetHighestBlock()
	if block == nil {
		return -1, ""
	}

	return block.Height, block.Hash
}

func (cm *ChainMgr) GetNewBlockInfo() *datalib.BcNewList {
	if cm.tree == nil {
		return nil
	}

	return cm.tree.GetNewBlockInfo()
}

func NewChainMgr() *ChainMgr {
	return &ChainMgr{
		tree: datalib.NewChainTree(),
	}
}
