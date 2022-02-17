package network

import (
	"sync"

	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dtype"
)

type NodeInfo struct {
	local dtype.NodeInfo
	sim   dtype.NodeInfo
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

func NodeInfoInst() *NodeInfo {
	oncenodeinfo.Do(func() {
		ni = &NodeInfo{
			sim:   dtype.NodeInfo{Mode: "", SC: config.SIM_SC, IP: "", Port: 0, Hash: ""},
			local: dtype.NodeInfo{Mode: "normal", SC: 0, IP: "", Port: 0, Hash: ""},
		}
	})
	return ni
}
