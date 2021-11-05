package blockchain

import (
	"testing"

	"github.com/junwookheo/bcsos/common/serial"
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
	b := serial.Serialize(tr)
	tr2 := Transaction{}
	serial.Deserialize(b, &tr2)
	assert.Equal(t, *tr, tr2)
}
