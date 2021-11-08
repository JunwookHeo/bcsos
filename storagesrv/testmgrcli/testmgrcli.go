package testmgrcli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"github.com/junwookheo/bcsos/common/shareddata"
)

func postTestClientInfo(ip string, port int) {
	url := fmt.Sprintf("http://%v:%v/clientNotify", ip, port)
	log.Println("Making call to", url)

	pbytes, _ := json.Marshal(shareddata.TestMgrInfo.Local)
	buff := bytes.NewBuffer(pbytes)

	resp, err := http.Post(url, "application/json", buff)
	if err != nil {
		log.Printf("Connecting test server error : %v", err)
		return
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	var local shareddata.TestNodeInfo
	json.Unmarshal(data, &local)
	shareddata.TestMgrInfo.Local.AddrHash = local.AddrHash
	log.Printf("Got response: %v\n", shareddata.TestMgrInfo.Local)
}

// Example of websocket
// https://github.com/gorilla/websocket/blob/master/examples/echo/client.go
func wsTestMgrHandler(ip string, port int) {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Reset()

	url := fmt.Sprintf("ws://%v:%v/ws", ip, port)
	log.Printf("connecting to %s", url)

	c, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case t := <-ticker.C:
			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func setServerInfo(ip string, port int, addhash string) {
	shareddata.TestMgrInfo.Server.IP = ip
	shareddata.TestMgrInfo.Server.Port = port
	shareddata.TestMgrInfo.Server.AddrHash = ""
}

func startResolver() {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("TestMgr Info : %v", shareddata.TestMgrInfo)

	// Channel to receive discovered service entries
	entries := make(chan *zeroconf.ServiceEntry)

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Println("Found service:", entry.ServiceInstanceName(), entry.Text)
			names := strings.Split(entry.ServiceInstanceName(), ".")
			if names[0] == "bcsos-tms" {
				setServerInfo(entry.AddrIPv4[0].String(), entry.Port, entry.Text[0])
				postTestClientInfo(entry.AddrIPv4[0].String(), entry.Port)
			}
		}
	}(entries)

	ctx := context.Background()

	err = resolver.Browse(ctx, "_omxremote._tcp", "local.", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()
}

func StartTMC() {
	log.Println("start Testmgr Client")
	startResolver()
}
