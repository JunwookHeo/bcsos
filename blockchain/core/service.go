package core

import (
	"bytes"
	"io"
	"log"
	"net"

	"github.com/junwookheo/bcsos/common/serial"
)

type Service struct {
}

func Start() {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		log.Printf("Start blockchain service error : %v", err)
		return
	}

	defer conn.Close()

	btx := serial.Serialize([]byte("Start blockchain service"))
	if _, err := io.Copy(conn, bytes.NewReader(btx)); err != nil {
		log.Printf("Send data error : %v", err)
		return
	}

	rxbuf, err := io.ReadAll(conn)
	n := len(rxbuf)
	if nil != err {
		if io.EOF == err {
			log.Printf("connection closed : %v", conn.RemoteAddr().String())
			return
		}
		log.Printf("Fail to receive data with err : %v", err)
		return
	}

	if n > 0 {
		brx := serial.Deserialize(rxbuf[:n])
		rx := string(brx)
		log.Printf("Received data length : %s", rx)
	}
}

func Stop() {

}
