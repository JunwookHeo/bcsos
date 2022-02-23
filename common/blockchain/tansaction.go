package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"math/big"
	"time"

	"github.com/junwookheo/bcsos/common/wallet"
)

type Transaction struct {
	Hash      []byte
	Timestamp int64
	Data      []byte
	Signature []byte
	PubKey    []byte
}

func (t *Transaction) GetHash() []byte {
	data := bytes.Join(
		[][]byte{
			toHex(t.Timestamp),
			t.Data[:],
			t.Signature[:],
			t.PubKey[:],
		},
		[]byte{},
	)

	hash := sha256.Sum256(data)
	return hash[:]
}

func (t *Transaction) sign(w *wallet.Wallet) bool {
	t.PubKey = w.PublicKey
	trcpy := Transaction{nil, t.Timestamp, t.Data, nil, t.PubKey}

	// dataToVerify := fmt.Sprintf("%x\n", trcpy)
	dataToVerify := trcpy.GetHash()
	// log.Printf("sign hash : %v - %v", hash, t.Hash)

	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, dataToVerify)
	if err != nil {
		log.Panicf("Signing Transaction Error : %v", err)
		return false
	}

	//signature := append(r.Bytes(), s.Bytes()...)
	buf1 := make([]byte, 32)
	buf2 := make([]byte, 32)
	signature := append(r.FillBytes(buf1), s.FillBytes(buf2)...)
	t.Signature = signature

	//log.Printf("sign sig : %v", signature)
	return true
}

func (t *Transaction) Verify() bool {
	r := big.Int{}
	s := big.Int{}

	siglen := len(t.Signature)
	r.SetBytes(t.Signature[:(siglen / 2)])
	s.SetBytes(t.Signature[(siglen / 2):])

	x := big.Int{}
	y := big.Int{}
	keylen := len(t.PubKey)

	x.SetBytes(t.PubKey[:(keylen / 2)])
	y.SetBytes(t.PubKey[(keylen / 2):])

	// log.Printf("sig : %v, key : %v", siglen, keylen)
	// log.Printf("verify sig : %v", t.Signature)
	trcpy := Transaction{nil, t.Timestamp, t.Data, nil, t.PubKey}
	curve := elliptic.P256()
	// dataToVerify := fmt.Sprintf("%x\n", trcpy)
	dataToVerify := trcpy.GetHash()
	// log.Printf("verify hash : %v - %v", hash, t.Hash)

	rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
	if !ecdsa.Verify(&rawPubKey, dataToVerify, &r, &s) {
		log.Printf("Verification fail : %v", hex.EncodeToString(t.Hash))
		return false
	}

	return true
}

func CreateTransaction(w *wallet.Wallet, d []byte) *Transaction {
	t := Transaction{nil, time.Now().UnixNano(), d[:], nil, nil}
	t.sign(w)
	t.Hash = t.GetHash()
	return &t
}
