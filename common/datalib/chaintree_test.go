package datalib

import (
	"log"
	"testing"
)

func TestNewTreeChain(t *testing.T) {
	tc := NewTreeChain()
	log.Printf("Start TestNewTreeChain")
	b1 := BlockData{0, "a", ""}
	tc.AddTreeNode(&b1)
	//tc.PrintAll()
	b2 := BlockData{0, "b", "a"}
	tc.AddTreeNode(&b2)
	//tc.PrintAll()
	b3 := BlockData{0, "c", "b"}
	tc.AddTreeNode(&b3)
	//tc.PrintAll()
	b4 := BlockData{0, "c2", "b"}
	tc.AddTreeNode(&b4)
	//tc.PrintAll()
	b5 := BlockData{0, "c3", "b"}
	tc.AddTreeNode(&b5)
	log.Printf("Highest Node : %v", *tc.GetHighestBlock())

	b6 := BlockData{0, "d", "c2"}
	tc.AddTreeNode(&b6)	
	b7 := BlockData{0, "e", "d"}
	tc.AddTreeNode(&b7)
	b8 := BlockData{0, "f", "e"}
	tc.AddTreeNode(&b8)
	tc.PrintAll()
	log.Printf("Highest Node : %v", *tc.GetHighestBlock())
	tc.UpdateRoot()
	tc.PrintAll()
}
