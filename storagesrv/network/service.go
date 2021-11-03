package network

import (
	"io"
	"io/ioutil"
	"log"
	"net"

	"github.com/junwookheo/bcsos/common/serial"
)

type Network struct {
}

func Start() {
	go start()
}

func start() {
	l, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Printf("Start network error : %v", err)
		return
	}
	defer l.Close()

	log.Printf("Start network : %v", l)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Printf("Accept error : %v", err)
			continue
		}
		defer conn.Close()

		log.Printf("Accept from : %v", conn)
		go connHandler(conn)
	}
}

func connHandler(conn net.Conn) {
	for {
		rxbuf, err := ioutil.ReadAll(conn)
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
			if _, err := conn.Write(brx); err != nil {
				log.Printf("Write data error : %v", err)
				return
			}
		}
	}
}

func Stop() bool {
	return false
}

func Send(dst []byte, data []byte) int {
	return -1
}

func Receive(buf *[]byte, size int) int {
	return -1
}
