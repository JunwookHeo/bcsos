package core

import (
	"encoding/hex"
	"io"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/junwookheo/bcsos/common/blockchain"
)

type bcsrv struct {
	done chan bool
}

const PATH = "./iotdata/IoT_normal_fridge_1.log"

var gbcsrv bcsrv = bcsrv{make(chan bool)}

func sendHandler(conn net.Conn) {
	msg := make(chan string)
	go LoadRawdata(PATH, msg)

	for {
		num := rand.Intn(3) + 1
		var trs []*blockchain.Transaction
		for i := 0; i < num; i++ {
			d := <-msg
			log.Println(d)
			tr := blockchain.CreateTransaction([]byte(d))
			trs = append(trs, tr)
		}
		b := blockchain.CreateBlock(trs, blockchain.GetLatestHash())
		blockchain.AddBlock(b)
		if _, err := conn.Write(b.Serialize()); err != nil {
			log.Printf("Send block error : %v", err)
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
			b := blockchain.Block{}
			b.Deserialize(rxbuf[:n])
			//brx := serial.Deserialize(rxbuf[:n])
			rx := string(hex.EncodeToString(b.Header.Hash))
			log.Printf("<=%s", rx)
		}
	}
}

func Start() {
	blockchain.InitBC()
	conn, err := net.Dial("tcp", ":8080")
	if err != nil {
		log.Printf("Start blockchain service error : %v", err)
		return
	}

	defer conn.Close()
	go sendHandler(conn)
	go receiveHandler(conn)

	<-gbcsrv.done
}

func Stop() {
	gbcsrv.done <- true
}
