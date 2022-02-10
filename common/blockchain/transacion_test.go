package blockchain

import (
	"log"
	"os"
	"testing"
	"time"
	"unsafe"

	"github.com/junwookheo/bcsos/common/serial"
	"github.com/junwookheo/bcsos/common/wallet"
	"github.com/stretchr/testify/assert"
)

func TestCreateTransaction(t *testing.T) {
	wallet_path := "./wallet_test.wallet"
	w := wallet.NewWallet(wallet_path)

	s := "test creating transaction"
	tr := CreateTransaction(w, []byte(s))
	assert.Equal(t, []byte(s), tr.Data)
	assert.NotEmpty(t, tr.Hash)
	os.Remove(wallet_path)
}

func TestSerializeTransaction(t *testing.T) {
	wallet_path := "./wallet_test.wallet"
	w := wallet.NewWallet(wallet_path)

	s := "test creating transaction"
	tr := CreateTransaction(w, []byte(s))
	b := serial.Serialize(tr)
	tr2 := Transaction{}
	serial.Deserialize(b, &tr2)
	assert.Equal(t, *tr, tr2)
	os.Remove(wallet_path)
}

func TestCheckSize(t *testing.T) {
	ts := time.Now()
	var s uint64 = 1
	var loc *Transaction
	log.Printf("time : %v - %v - %v", unsafe.Sizeof(ts), unsafe.Sizeof(s), unsafe.Sizeof(loc))
}

func TestSignVerifyTransaction(t *testing.T) {
	wallet_path := "./wallet_test.wallet"
	w := wallet.NewWallet(wallet_path)

	s := "test sign/verify transaction"
	tr := CreateTransaction(w, []byte(s))
	assert.Equal(t, []byte(s), tr.Data)
	assert.NotEmpty(t, tr.Hash)

	tr.sign(w)
	tr.verify()

	os.Remove(wallet_path)
}
