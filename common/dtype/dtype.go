package dtype

type NodeInfo struct {
	Mode string `json:"mode"`
	SC   int    `json:"storage_class"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
	Hash string `json:"hash"`
}

type ReqData struct {
	Addr      string `json:"Addr"`
	Timestamp int64  `json:"Timestamp"`
	SC        int    `json:"storage_class"`
	Hop       int    `json:"Hop"`
	ObjType   string `json:"ObjType"`
	ObjHash   string `json:"ObjHash"`
}
