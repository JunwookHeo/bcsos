package mining

import (
	"log"
	"math/big"
	"testing"
)

func TestSimulateTransaction(t *testing.T) {
	var wallet_path string = "../bc_dev.wallet"
	wm := WalletMgrInst(wallet_path)
	log.Printf("===start TestSimulateTransaction : %v", wm)
	for i := 0; i < 10000; i++ {
		tr := generateTransactionFromRandom(i)
		if tr.Verify() == false {
			log.Printf("verify fail : %d", i)
			break
		}
	}

}

func TestBigInt(t *testing.T) {
	r := big.NewInt(2)
	buf := make([]byte, 128)
	rb := r.FillBytes(buf)
	log.Printf(" %v : %v", rb, len(rb))
}
