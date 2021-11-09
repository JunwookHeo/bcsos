package testmgrcli

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"github.com/junwookheo/bcsos/common/dtype"
)

type Simulator struct {
	IP   string
	Port int
}

type TestMgrCli struct {
	Sim   Simulator
	Local dtype.NodeInfo
}

func (t *TestMgrCli) resisterNode(ip string, port int) {
	url := fmt.Sprintf("ws://%v:%v/resister", ip, port)
	log.Println("Making call to", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer ws.Close()

	if err := ws.WriteJSON(t.Local); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}

	var node dtype.NodeInfo
	if err := ws.ReadJSON(&node); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}

	t.Local.IP = node.IP
	t.Local.Hash = node.Hash
	log.Printf("Got response: %v\n", t.Local)
	log.Printf("Recevied node : %v", node)
}

func (t *TestMgrCli) setServerInfo(ip string, port int) {
	t.Sim.IP = ip
	t.Sim.Port = port
}

func (t *TestMgrCli) startResolver() {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("TestMgr Info : %v", t.Local)

	// Channel to receive discovered service entries
	entries := make(chan *zeroconf.ServiceEntry)

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Println("Found service:", entry.ServiceInstanceName(), entry.Text)
			names := strings.Split(entry.ServiceInstanceName(), ".")
			if names[0] == "bcsos-tms" {
				t.setServerInfo(entry.AddrIPv4[0].String(), entry.Port)
				t.resisterNode(entry.AddrIPv4[0].String(), entry.Port)
			}
		}
	}(entries)

	ctx := context.Background()
	err = resolver.Browse(ctx, "_omxremote._tcp", "local.", entries)
	if err != nil {
		log.Fatalln("Failed to browse:", err.Error())
	}

	<-ctx.Done()
	log.Println("==========Found service: done")
}

func NewTMC(port int) *TestMgrCli {
	log.Println("start Testmgr Client")
	tmc := TestMgrCli{Sim: Simulator{}, Local: dtype.NodeInfo{Type: "", IP: "", Port: port, Hash: ""}}
	go tmc.startResolver()
	return &tmc
}

// Example of websocket
// https://github.com/gorilla/websocket/blob/master/examples/echo/client.go
// func wsTestMgrHandler(ip string, port int) {
// 	interrupt := make(chan os.Signal, 1)
// 	signal.Notify(interrupt, os.Interrupt)
// 	defer signal.Reset()

// 	url := fmt.Sprintf("ws://%v:%v/ws", ip, port)
// 	log.Printf("connecting to %s", url)

// 	c, _, err := websocket.DefaultDialer.Dial(url, nil)
// 	if err != nil {
// 		log.Fatal("dial:", err)
// 	}
// 	defer c.Close()

// 	done := make(chan struct{})

// 	go func() {
// 		defer close(done)
// 		for {
// 			_, message, err := c.ReadMessage()
// 			if err != nil {
// 				log.Println("read:", err)
// 				return
// 			}
// 			log.Printf("recv: %s", message)
// 		}
// 	}()

// 	ticker := time.NewTicker(time.Second)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-done:
// 			return
// 		case t := <-ticker.C:
// 			err := c.WriteMessage(websocket.TextMessage, []byte(t.String()))
// 			if err != nil {
// 				log.Println("write:", err)
// 				return
// 			}
// 		case <-interrupt:
// 			log.Println("interrupt")

// 			// Cleanly close the connection by sending a close message and then
// 			// waiting (with timeout) for the server to close the connection.
// 			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
// 			if err != nil {
// 				log.Println("write close:", err)
// 				return
// 			}
// 			select {
// 			case <-done:
// 			case <-time.After(time.Second):
// 			}
// 			return
// 		}
// 	}
// }
