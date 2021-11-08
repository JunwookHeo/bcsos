package shareddata

type TestNodeInfo struct {
	IP       string `json:"id"`
	Port     int    `json:"port"`
	AddrHash string `json:"addr_hash"`
}

type TestMgrCli struct {
	Server TestNodeInfo
	Local  TestNodeInfo
}

var TestMgrInfo TestMgrCli = TestMgrCli{}
