package network

import (
	"log"
	"testing"

	"github.com/junwookheo/bcsos/common/dtype"
)

func TestXOR(t *testing.T) {
	d1 := xordistance("aaf0aa63ef4fb4d8933524ebcfd97c33e7c2cd7c31ccbda894aa792a7d63cc95", "1e78c84fce7b59fe96f8267989cb96a1ebd70dc83111eecfd73b4c21ab5d84e3")
	d2 := xordistance("aaf0aa63ef4fb4d8933524ebcfd97c33e7c2cd7c31ccbda894aa792a7d63cc95", "ac15c0aa170e705b104dd2b1000f8650a4255f6b813cc9559843a3d580454048")
	log.Printf("distance1 : %v", d1)
	log.Printf("distance2 : %v", d2)
	log.Printf("d1 - d2 : %v", d2.Cmp(d1))
}

func TestAddNSCNNode(t *testing.T) {
	node1 := dtype.NodeInfo{Mode: "mode", SC: 1, IP: "ip", Port: 1, Hash: "c900bd2b85fc79461560dc90854c01dc9861690c1d1171ca4c040e936a3e32df"}
	node2 := dtype.NodeInfo{Mode: "mode", SC: 1, IP: "ip", Port: 2, Hash: "8ccc5577008159b4a86800dd83de4494c78ee7caa364bb0561fa544dacde53a3"}
	node3 := dtype.NodeInfo{Mode: "mode", SC: 1, IP: "ip", Port: 3, Hash: "8ccc5577008159b4a86800dd83de4494c78ee7caa364bb0561fa544dacd6981c"}

	local := dtype.NodeInfo{Mode: "mode", SC: 1, IP: "ip", Port: 4, Hash: "c900bd2b85fc79461560d0dd83de4494c78ee7caa364bb0561fa544dacde53a3"}
	ap := NewSCNInfo(&local)

	ap.AddNSCNNode(node1)
	ap.AddNSCNNode(node2)
	ap.AddNSCNNode(node3)
	ap.ShowSCNNodeList()
}
