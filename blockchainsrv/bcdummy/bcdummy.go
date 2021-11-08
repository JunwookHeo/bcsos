package bcdummy

import (
	"encoding/hex"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/junwookheo/bcsos/common/bcapi"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/serial"
)

type bcsrv struct {
	done chan bool
}

const PATH = "./iotdata/IoT_normal_fridge_1.log"

var gbcsrv bcsrv = bcsrv{make(chan bool)}

func sendHandler(msgtype uint32, d []byte, conn net.Conn) {
	pkt := config.TcpPacket{}
	chunks := config.SplitTcp(d)
	for _, c := range chunks[:len(chunks)-1] {
		pkt.MsgType = msgtype
		pkt.LastFlag = 0
		pkt.Length = uint32(len(c))
		buf := config.TcpHeaderToBytes(&pkt)
		buf = append(buf, c...)

		if _, err := conn.Write(buf); err != nil {
			log.Printf("Send block error : %v", err)
			return
		}
	}

	// Send last chunck
	c := chunks[len(chunks)-1]
	pkt.MsgType = msgtype
	pkt.LastFlag = 1
	pkt.Length = uint32(len(c))
	buf := config.TcpHeaderToBytes(&pkt)
	buf = append(buf, c...)

	if _, err := conn.Write(buf); err != nil {
		log.Printf("Send block error : %v", err)
		return
	}
}

func receiveHandler(conn net.Conn) {
	rxbuf := make([]byte, config.TCPSIZE)
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
			b := blockchain.Block{}
			serial.Deserialize(rxbuf[:n], &b)
			rx := string(hex.EncodeToString(b.Header.Hash))
			log.Printf("<=%s", rx)
			break
		}
	}
}
func getNodeIP() []string {
	ips := []string{":8080"}
	return ips
}

func sendBlock(b *blockchain.Block, ip string) {
	conn, err := net.Dial("tcp", ip)
	if err != nil {
		log.Printf("Send block error : %v", err)
		return
	}

	defer conn.Close()

	btx := serial.Serialize(b)
	sendHandler(uint32(config.NEWBLOCK), btx, conn)
	//receiveHandler(conn) // no need to get response
}

func broadcastBlock(b *blockchain.Block) {
	ips := getNodeIP()
	for _, ip := range ips {
		go sendBlock(b, ip)
	}
}

func Start() {
	// Send Genesis Block
	hash := bcapi.GetLatestHash()
	if len(hash) == 0 {
		log.Printf("Create Genesis due to hash : %v", hash)
		b := bcapi.CreateGenesis()
		bcapi.AddBlock(b)
		broadcastBlock(b)
		time.Sleep(3 * time.Second)
	}

	msg := make(chan string)
	go LoadRawdata(PATH, msg)

	for {
		num := rand.Intn(3) + 1
		var trs []*blockchain.Transaction
		for i := 0; i < num; i++ {
			d := <-msg
			log.Printf("==>%s", d)
			tr := blockchain.CreateTransaction([]byte(d))
			trs = append(trs, tr)
		}
		b := blockchain.CreateBlock(trs, bcapi.GetLatestHash())
		bcapi.AddBlock(b)
		broadcastBlock(b)
		time.Sleep(3 * time.Second)
	}

	//<-gbcsrv.done
}

func Stop() {
	gbcsrv.done <- true
}
