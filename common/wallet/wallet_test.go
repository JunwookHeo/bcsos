package wallet

import (
	"encoding/binary"
	"fmt"
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWallet(t *testing.T) {
	log.Println("TestNewWallet ==>")
	w := NewWallet()
	log.Printf("Private Key : 0x%X", w.PrivateKey.D)
	log.Printf("Public Key (0x%X, 0x%X)", w.PrivateKey.X, w.PrivateKey.Y)

	address := w.getAddress()
	log.Printf("Address : %s", address)

	assert.True(t, ValidateAddress(address))

}

func TestValidateAddress(t *testing.T) {
	log.Println("TestValidateAddress ==>")
	version, payload, checksum := getPayload([]byte("5J3mBbAH58CpQ3Y5RNJpUKPE62SQ5tfcvU2JpbnkeyhfsYB1Jcn"))
	assert.Equal(t, byte(128), version)
	assert.Equal(t, "1e99423a4ed27608a15a2616a2b0e9e52ced330ac530edcc32c8ffc6a526aedd", fmt.Sprintf("%x", payload))
	assert.Equal(t, uint32(4286807748), binary.LittleEndian.Uint32(checksum))

	version, payload, checksum = getPayload([]byte("KxFC1jmwwCoACiCAWZ3eXa96mBM6tb3TYzGmf6YwgdGWZgawvrtJ"))
	assert.Equal(t, byte(128), version)
	assert.Equal(t, "1e99423a4ed27608a15a2616a2b0e9e52ced330ac530edcc32c8ffc6a526aedd01", fmt.Sprintf("%x", payload))
	assert.Equal(t, uint32(2339607926), binary.LittleEndian.Uint32(checksum))
}
