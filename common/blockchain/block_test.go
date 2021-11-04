package blockchain

import (
	"testing"

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
	assert.Greater(t, b.Header.Timestamp, int64(0))
	// TODO:
	//assert.NotEmpty(t, b.Header.Hash)
	assert.Equal(t, trs, b.Transactions)
	for i := 0; i < len(trs); i++ {
		assert.Equal(t, *trs[i], *b.Transactions[i])
	}
}

func TestSeializeBlock(t *testing.T) {
	var trs []*Transaction
	for i := 0; i < 3; i++ {
		s := "Test create block"
		tr := CreateTransaction([]byte(s))
		assert.Equal(t, []byte(s), tr.Data)
		trs = append(trs, tr)
	}

	b1 := CreateBlock(trs, nil)
	assert.Greater(t, b1.Header.Timestamp, int64(0))
	// TODO:
	//assert.NotEmpty(t, b1.Header.Hash)
	assert.Equal(t, trs, b1.Transactions)
	for i := 0; i < len(trs); i++ {
		assert.Equal(t, *trs[i], *b1.Transactions[i])
	}

	bx := b1.Serialize()
	b2 := Block{}
	b2.Deserialize(bx)
	assert.Equal(t, b1.Header, b2.Header)
	for i := 0; i < len(trs); i++ {
		assert.Equal(t, *b1.Transactions[i], *b2.Transactions[i])
	}
}
