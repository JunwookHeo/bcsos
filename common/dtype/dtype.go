package dtype

type NodeInfo struct {
	Mode string `json:"mode"`
	SC   int    `json:"storage_class"`
	IP   string `json:"ip"`
	Port int    `json:"port"`
	Hash string `json:"hash"`
}

type Simulator struct {
	IP   string
	Port int
}

type Version struct {
	Ver int
}
