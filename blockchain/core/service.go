package core

import (
	"io"
	"log"
	"net"
	"time"

	"github.com/junwookheo/bcsos/common/serial"
)

type Service struct {
}

func sendHandler(conn net.Conn) {
	for {
		send := time.Now().String()
		btx := serial.Serialize([]byte(send))
		if _, err := conn.Write(btx); err != nil {
			log.Printf("send data error : %v", err)
			return
		}
		time.Sleep(3 * time.Second)
	}
}

func receiveHandler(conn net.Conn) {
	rxbuf := make([]byte, 4096)
	for {
		n, err := conn.Read(rxbuf)
		if err != nil {
			if err == io.EOF {
				log.Println("Connection closed")
				return
			}
			log.Printf("Fail to receive data : %v", err)
			return
		}
		if n > 0 {
			brx := serial.Deserialize(rxbuf[:n])
			rx := string(brx)
			log.Printf("<=%s", rx)
		}
	}
}

func Start() {
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		log.Printf("Start blockchain service error : %v", err)
		return
	}

	defer conn.Close()
	go sendHandler(conn)
	go receiveHandler(conn)
	for {
		time.Sleep(1 * time.Second)
	}
}

func Stop() {

}
