package wallet

import (
	"testing"

	log "github.com/sirupsen/logrus"

	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	w := NewWallet()
	log.Infof("Private Key : 0x%X\n", w.PrivateKey.D)
	log.Infof("Public Key (0x%X, 0x%X)\n", w.PrivateKey.X, w.PrivateKey.Y)

	address := w.getAddress()
	log.Infof("Address : %s\n", address)

	assert.True(t, ValidateAddress(address))

	ValidateAddress([]byte("5J3mBbAH58CpQ3Y5RNJpUKPE62SQ5tfcvU2JpbnkeyhfsYB1Jcn"))
}
