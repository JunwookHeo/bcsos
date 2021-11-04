package blockchain

import (
	"bytes"
	"encoding/gob"
	"log"
)

func Serialize(v interface{}) []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)
	err := encoder.Encode(v)
	ErrorHanlder(err)

	return res.Bytes()
}

func Deserialize(d []byte, v interface{}) {
	decoder := gob.NewDecoder(bytes.NewReader(d))
	err := decoder.Decode(v)
	ErrorHanlder(err)
}

func ErrorHanlder(e error) {
	if e != nil {
		log.Panic(e)
	}
}
