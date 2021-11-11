package blockchain

import (
	"testing"

	"github.com/junwookheo/bcsos/common/serial"
	"github.com/stretchr/testify/assert"
)

func TestCreateBlock(t *testing.T) {
	var trs []*Transaction
	for i := 0; i < 3; i++ {
		s := "Test create block"
		tr := CreateTransaction([]byte(s))
		assert.Equal(t, []byte(s), tr.Data)
		trs = append(trs, tr)
	}

	b := CreateBlock(trs, nil)
	//assert.Greater(t, b.Header.Timestamp, int64(0))
	assert.NotEmpty(t, b.Header.Hash)
	assert.Equal(t, trs, b.Transactions)
	for i := 0; i < len(trs); i++ {
		assert.Equal(t, *trs[i], *b.Transactions[i])
	}
}

func TestSeializeBlock(t *testing.T) {
	var trs []*Transaction
	for i := 0; i < 3; i++ {
		s := "Test Seialize Block"
		tr := CreateTransaction([]byte(s))
		assert.Equal(t, []byte(s), tr.Data)
		trs = append(trs, tr)
	}

	b1 := CreateBlock(trs, nil)
	//assert.Greater(t, b1.Header.Timestamp, int64(0))
	assert.NotEmpty(t, b1.Header.Hash)
	assert.Equal(t, trs, b1.Transactions)
	for i := 0; i < len(trs); i++ {
		assert.Equal(t, *trs[i], *b1.Transactions[i])
	}

	bx := serial.Serialize(b1)
	b2 := Block{}
	serial.Deserialize(bx, &b2)
	assert.Equal(t, b1.Header, b2.Header)
	for i := 0; i < len(trs); i++ {
		assert.Equal(t, *b1.Transactions[i], *b2.Transactions[i])
	}
}

func TestHashBlock(t *testing.T) {
	var trs []*Transaction
	sss := []string{"1111111111111111", "2222222222222222", "333333333333333333"}
	for i := 0; i < 3; i++ {
		s := sss[i]
		tr := CreateTransaction([]byte(s))
		assert.Equal(t, []byte(s), tr.Data)
		trs = append(trs, tr)
	}

	b1 := CreateBlock(trs, nil)
	difficulty, nonce, hash := ProofWork(b1)
	assert.Equal(t, difficulty, b1.Header.Difficulty)
	assert.Equal(t, nonce, b1.Header.Nonce)
	assert.Equal(t, hash, b1.Header.Hash)
	ret := Validate(b1)
	assert.Equal(t, ret, true)
}
