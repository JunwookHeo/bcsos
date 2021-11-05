package serial

import (
	"bytes"
	"encoding/gob"
	"log"
)

func Serialize(v interface{}) []byte {
	var buffer bytes.Buffer
	encode := gob.NewEncoder(&buffer)
	err := encode.Encode(v)
	handle(err)
	return buffer.Bytes()
}

func Deserialize(data []byte, out interface{}) {
	decode := gob.NewDecoder(bytes.NewReader(data))
	err := decode.Decode(out)
	handle(err)
}

func handle(err error) {
	if err != nil {
		log.Panicf("Failed to serialize/deserialize data : %v", err)
	}
}
