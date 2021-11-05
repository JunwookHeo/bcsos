package blockchain

import (
	"log"
	"testing"
)

func TestInit(t *testing.T) {
	InitBC()
	for k, v := range BC {
		log.Println(k, v)
	}
}
