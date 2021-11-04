package blockchain

import (
	"log"
	"testing"
)

func TestTransaction(t *testing.T) {
	tx := Transaction{Id: []byte("1234"), Data: []byte("Test serialize")}
	b := Serialize(tx)
	log.Printf("==>%v", tx)
	var rx Transaction
	Deserialize(b, &rx)
	log.Printf("<==%v", rx)
}

func TestTransaction2(t *testing.T) {
	tx := Transaction{Id: []byte("5647"), Data: []byte("Test serialize")}
	b := SerializeTransaction(tx)
	log.Printf("==>%s", tx)

	rx := DeserializeTransaction(b)
	log.Printf("<==%s", rx)
}
