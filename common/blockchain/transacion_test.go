package blockchain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreateTransaction(t *testing.T) {
	s := "test creating transaction"
	tr := CreateTransaction([]byte(s))
	assert.Equal(t, []byte(s), tr.Data)
	assert.NotEmpty(t, tr.Id)
}

func TestSerializeTransaction(t *testing.T) {
	s := "test creating transaction"
	tr := CreateTransaction([]byte(s))
	b := tr.Serialize()
	tr2 := Transaction{}
	tr2.Deserialize(b)
	assert.Equal(t, *tr, tr2)
}
