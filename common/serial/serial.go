package serial

import (
	"bytes"
	"encoding/gob"
	"log"
)

func Serialize(data []byte) []byte {
	var buffer bytes.Buffer
	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(data)
	handle(err)
	return buffer.Bytes()
}

func Deserialize(data []byte) []byte {
	var outputs []byte
	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(&outputs)
	handle(err)
	return outputs
}

func handle(err error) {
	if err != nil {
		log.Fatalf("Failed to serialize/deserialize data : %v", err)
	}
}
