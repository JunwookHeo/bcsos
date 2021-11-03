package network

import (
	"io"
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
	rxbuf := make([]byte, 4096)
	for {
		n, err := conn.Read(rxbuf)
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
			log.Printf("<==%s", rx)
			send := serial.Serialize([]byte(rx))
			if _, err := conn.Write(send); err != nil {
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
