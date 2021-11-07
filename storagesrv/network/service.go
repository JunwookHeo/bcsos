package network

import (
	"io"
	"log"
	"net"

	"github.com/junwookheo/bcsos/common/bcapi"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
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

func readConn(conn net.Conn) (*config.TcpPacket, []byte) {
	var body []byte
	rxbuf := make([]byte, config.TCPSIZE)
	n, err := conn.Read(rxbuf)
	if nil != err {
		if io.EOF == err {
			log.Printf("connection closed : %v", conn.RemoteAddr().String())
			return nil, nil
		}
		log.Printf("Fail to receive data with err : %v", err)
		return nil, nil
	}

	if n > 0 {
		head := config.BytesToTcpHeader(rxbuf[:12])
		if head.LastFlag == 0 {
			head, body = readConn(conn)
			body = append(rxbuf[12:], body...)
			return head, body
		} else {
			body = rxbuf[12:]
			return head, body
		}
	}
	return nil, nil
}

func HandleNewBlock(d []byte) {
	// storage.AddBlock(d)
	b := blockchain.Block{}
	serial.Deserialize(d, &b)
	bcapi.AddBlock(&b)
	for _, tr := range b.Transactions {
		log.Printf("<===%s", tr.Data)
	}
}

func connHandler(conn net.Conn) {
	for {
		head, body := readConn(conn)
		if head != nil {
			switch head.MsgType {
			case uint32(config.NEWBLOCK):
				HandleNewBlock(body)
			}

		} else {
			break
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
