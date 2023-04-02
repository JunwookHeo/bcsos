package network

import (
	"sync"

	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dtype"
)

type NodeInfo struct {
	local dtype.NodeInfo
	sim   dtype.NodeInfo
	OAM   bool // Outsourcing attack mode
}

var (
	ni           *NodeInfo
	oncenodeinfo sync.Once
)

func (ni *NodeInfo) SetSimAddr(ip string, port int) {
	ni.sim.IP = ip
	ni.sim.Port = port
}

func (ni *NodeInfo) GetSimAddr() *dtype.NodeInfo {
	return &ni.sim
}

func (ni *NodeInfo) GetLocalddr() *dtype.NodeInfo {
	return &ni.local
}

func (ni *NodeInfo) SetLocalddrParam(mode string, sc int, port int, hash string) {
	ni.local.Mode = mode
	ni.local.SC = sc
	ni.local.Port = port
	ni.local.Hash = hash
}

func (ni *NodeInfo) SetLocalddrIP(ip string) {
	ni.local.IP = ip
}

func (ni *NodeInfo) SetOoutSourcingAttackMode(hash string) {
	if ni.local.Hash == hash {
		ni.OAM = true
	} else {
		ni.OAM = false
	}
}

func (ni *NodeInfo) GetOoutSourcingAttackMode() bool {
	return ni.OAM
}

func NodeInfoInst() *NodeInfo {
	oncenodeinfo.Do(func() {
		ni = &NodeInfo{
			sim:   dtype.NodeInfo{Mode: "ST", SC: config.SIM_SC, IP: "", Port: 0, Hash: ""},
			local: dtype.NodeInfo{Mode: "ST", SC: 0, IP: "", Port: 0, Hash: ""},
			OAM:   false,
		}
	})
	return ni
}
