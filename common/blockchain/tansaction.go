package blockchain

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
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
			t.Data,
			t.Signature,
			t.PubKey,
		},
		[]byte{},
	)

	hash := sha256.Sum256(data)
	return hash[:]
}

func (t *Transaction) sign(w *wallet.Wallet) bool {
	t.PubKey = w.PublicKey
	trcpy := Transaction{t.Hash, t.Timestamp, t.Data, nil, t.PubKey}

	hash := trcpy.GetHash()
	log.Printf("sign hash : %v", hash)

	r, s, err := ecdsa.Sign(rand.Reader, w.PrivateKey, hash)
	if err != nil {
		log.Panicf("Signing Transaction Error : %v", err)
		return false
	}

	signature := append(r.Bytes(), s.Bytes()...)
	t.Signature = signature

	return true
}

func (t *Transaction) verify() bool {
	r := big.Int{}
	s := big.Int{}

	siglen := len(t.Signature)
	r.SetBytes(t.Signature[:(siglen / 2)])
	s.SetBytes(t.Signature[(siglen / 2):])

	x := big.Int{}
	y := big.Int{}
	keylen := len(t.PubKey)

	x.SetBytes(t.PubKey[:keylen/2])
	y.SetBytes(t.PubKey[keylen/2:])

	trcpy := Transaction{t.Hash, t.Timestamp, t.Data, nil, t.PubKey}
	curve := elliptic.P256()
	hash := trcpy.GetHash()
	log.Printf("verify hash : %v", hash)

	rawPubKey := ecdsa.PublicKey{Curve: curve, X: &x, Y: &y}
	if ecdsa.Verify(&rawPubKey, hash, &r, &s) == false {
		log.Printf("Verification fail")
		return false
	}

	return true
}

func CreateTransaction(w *wallet.Wallet, d []byte) *Transaction {
	t := Transaction{nil, time.Now().UnixNano(), d[:], nil, nil}
	t.sign(w)
	t.verify()
	t.Hash = t.GetHash()
	return &t
}
