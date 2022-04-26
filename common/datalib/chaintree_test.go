package datalib

import (
	"log"
	"testing"
)

func TestNewChainTree(t *testing.T) {
	tc := NewChainTree()
	log.Printf("Start TestNewChainTree")
	b1 := BlockData{0, 1, "a", ""}
	tc.AddTreeNode(&b1, true)
	//tc.PrintAll()
	b2 := BlockData{1, 2, "b", "a"}
	tc.AddTreeNode(&b2, true)
	//tc.PrintAll()
	b3 := BlockData{2, 3, "c", "b"}
	tc.AddTreeNode(&b3, true)
	//tc.PrintAll()
	b4 := BlockData{3, 4, "c2", "b"}
	tc.AddTreeNode(&b4, true)
	//tc.PrintAll()
	b5 := BlockData{4, 5, "c3", "b"}
	tc.AddTreeNode(&b5, true)
	log.Printf("Highest Node : %v", *tc.GetHighestBlock())

	b6 := BlockData{5, 6, "d", "c2"}
	tc.AddTreeNode(&b6, true)
	b7 := BlockData{6, 7, "e", "d"}
	tc.AddTreeNode(&b7, true)
	b8 := BlockData{7, 8, "f", "e"}
	tc.AddTreeNode(&b8, true)
	tc.PrintAll()
	log.Printf("Highest Node : %v", *tc.GetHighestBlock())
	tc.UpdateRoot()
	tc.PrintAll()
}
