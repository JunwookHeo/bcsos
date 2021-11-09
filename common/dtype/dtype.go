package dtype

type NodeInfo struct {
	Type string `json:"type"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
	Hash string `json:"hash"`
}
