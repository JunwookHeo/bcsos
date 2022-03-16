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
	block := datalib.BlockData{Height: 0, Hash: hex.EncodeToString(h.Hash), Prev: hex.EncodeToString(h.PrvHash)}
	cm.tree.AddTreeNode(&block)
	cm.tree.UpdateRoot()
	// cm.tree.PrintAll()
}

func (cm *ChainMgr) GetHighestBlockHash() (int, string) {
	block := cm.tree.GetHighestBlock()
	if block == nil {
		return 0, ""
	}

	return block.Height, block.Hash
}

func NewChainMgr() *ChainMgr {
	return &ChainMgr{
		tree: datalib.NewChainTree(),
	}
}
