package testmgrcli

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
)

type TestNodeInfo struct {
	IP       string
	Port     int
	AddrHash string
}

type TestMgrCli struct {
	Server TestNodeInfo
	Local  TestNodeInfo
}

var testmgrcli TestMgrCli = TestMgrCli{}

func init() {
	testmgrcli.Local.IP = getLocalIP()
	testmgrcli.Local.Port = 0
	testmgrcli.Local.AddrHash = "Test Manager Server"
}

func getLocalIP() string {
	addrs, _ := net.InterfaceAddrs()
	fmt.Printf("%v\n", addrs)
	for _, addr := range addrs {
		if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String()
			}
		}
	}
}

func serviceCall(ip string, port int) {
	testmgrcli.Server.IP = ip
	testmgrcli.Server.Port = port
	testmgrcli.Server.AddrHash = "Test Manager Server"

	url := fmt.Sprintf("http://%v:%v/", ip, port)

	log.Println("Making call to", url)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Connecting test server error : %v", err)
		return
	}
	defer resp.Body.Close()

	data, _ := ioutil.ReadAll(resp.Body)
	log.Printf("Got response: %s\n", data)
}

// Example of websocket
// https://github.com/gorilla/websocket/blob/master/examples/echo/client.go
func serviceCall2(ip string, port int) {
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

func connectionHandler() {
	for {
		log.Printf("Server Info : %v", testmgrcli)
		serviceCall(testmgrcli.Server.IP, testmgrcli.Server.Port)
		time.Sleep(3 * time.Second)
	}
}

func startResolver() {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatal(err)
	}

	// Channel to receive discovered service entries
	entries := make(chan *zeroconf.ServiceEntry)

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Println("Found service:", entry.ServiceInstanceName(), entry.Text)
			serviceCall(entry.AddrIPv4[0].String(), entry.Port)
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
	go startResolver()
	//connectionHandler()
}
