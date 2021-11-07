package bcapi

import (
	"encoding/hex"
	"log"
	"testing"

	"github.com/junwookheo/bcsos/common/blockchain"
)

func TestBytetoString(t *testing.T) {
	str := "test byte to string"
	h0 := []byte(str)
	log.Printf("h0 : %v", h0)
	h1 := blockchain.CalHashSha256([]byte(str))
	log.Printf("h1 : %v", h1)
	s1 := hex.EncodeToString(h1)
	log.Printf("s1 : %v", s1)
	h2, _ := hex.DecodeString(s1)
	log.Printf("h2 : %v", h2)
	log.Printf("h2 : %s", string(h2))
}
