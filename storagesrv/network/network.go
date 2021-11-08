package network

import (
	"fmt"
	"io"
	"log"
	"net"

	"github.com/junwookheo/bcsos/common/bcapi"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/serial"
	"github.com/junwookheo/bcsos/storagesrv/testmgrcli"
)

type Network struct {
}

func Start() {
	go start()
}

func start() {
	port, _ := GetFreePort()
	addr := fmt.Sprintf(":%v", port)
	testmgrcli.TestMgrCli.Local.port = port
	l, err := net.Listen("tcp", addr)

	log.Printf("address : %v", l.Addr())
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
		//defer conn.Close()
		go connectionHandler(conn)
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

func connectionHandler(conn net.Conn) {
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

func HandleNewBlock(d []byte) {
	b := blockchain.Block{}
	serial.Deserialize(d, &b)
	bcapi.AddBlock(&b)
	for _, tr := range b.Transactions {
		log.Printf("<===%s", tr.Data)
	}
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func GetFreePort() (port int, err error) {
	getLocalIP()
	var a *net.TCPAddr
	if a, err = net.ResolveTCPAddr("tcp", "localhost:0"); err == nil {
		var l *net.TCPListener
		if l, err = net.ListenTCP("tcp", a); err == nil {
			defer l.Close()
			return l.Addr().(*net.TCPAddr).Port, nil
		}
	}
	return
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
