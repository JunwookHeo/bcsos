package network

import (
	"io"
	"log"
	"net"

	"github.com/junwookheo/bcsos/common/blockchain"
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
			b := blockchain.Block{}
			serial.Deserialize(rxbuf[:n], &b)

			if _, err := conn.Write(serial.Serialize(b)); err != nil {
				log.Printf("Write data error : %v", err)
				return
			}

			for _, tr := range b.Transactions {
				log.Printf("<===%s", tr.Data)
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
