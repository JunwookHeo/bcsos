package dtype

import "github.com/holiman/uint256"

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

type ReqConsecutiveHashes struct {
	Height int `json:"Height"`
	Count  int `json:"Count"`
}

type ResConsecutiveHashes struct {
	Hashes string `json:"Hashes"`
}

type ReqEncryptedBlock struct {
	Hash string `json:"Hash"`
}

type ResEncryptedBlock struct {
	Block string `json:"Block"`
}

type PoSProof struct {
	Timestamp int64
	Address   []byte
	Root      []byte
	HashEncs  [][]byte
	HashKeys  [][]byte
	Selected  int
	Hash      string
	EncBlock  []byte
}
type FriProofElement struct {
	Root2   [][]byte
	CBranch [][][]byte
	PBranch [][][]byte
}

type StarksProof struct {
	RandomHash string
	MerkleRoot []byte
	TreeRoots  [][]byte
	TreeVosu   [][][]byte
	TreeKey    [][][]byte
	TreeD      [][][]byte
	TreeB      [][][]byte
	TreeL      [][][]byte
	VosuFl     []*uint256.Int
	FriProof   []*FriProofElement
}

type NonInteractiveProof struct {
	Address []byte
	Hash    string

	Starks *StarksProof
}
