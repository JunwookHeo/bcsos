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

type Command struct {
	Cmd    string `json:"cmd"`
	Subcmd string `json:"subcmd"`
	Arg1   string `json:"arg1"`
	Arg2   string `json:"arg2"`
	Arg3   string `json:"arg3"`
}

type ReqPoStorage struct {
	Hash      string `json:"Hash"`
	Timestamp int64  `json:"Timestamp"`
}

type ResPoStorage struct {
	Addr  string `json:"Address"`
	SC    int    `json:"storage_class"`
	Proof string `json:"Proof"`
}
