package testmgrcli

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"github.com/junwookheo/bcsos/blockchainnode/network"
	"github.com/junwookheo/bcsos/common/dtype"
)

type TestMgrCli struct {
}

var (
	tmc  *TestMgrCli
	once sync.Once
)

func (t *TestMgrCli) checkPortAvailavle(ips []net.IP, p int) string {
	port := fmt.Sprintf("%v", p)
	for _, ip := range ips {
		log.Printf("...connecting to %v", ip)
		conn, _ := net.DialTimeout("tcp", net.JoinHostPort(ip.String(), port), time.Second)
		if conn != nil {
			conn.Close()
			return ip.String()
		}
	}
	return ""
}

func (t *TestMgrCli) registerNode(ip string, port int) {
	url := fmt.Sprintf("ws://%v:%v/register", ip, port)
	log.Println("Making call to", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer ws.Close()

	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()

	if err := ws.WriteJSON(local); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}

	var node dtype.NodeInfo
	if err := ws.ReadJSON(&node); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}

	ni.SetLocalddrIP(node.IP)
	log.Printf("Got response: %v\n", local)
	log.Printf("Recevied node : %v", node)
}

func (t *TestMgrCli) startResolver() {
	ni := network.NodeInfoInst()
	local := ni.GetLocalddr()

	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("TestMgr Info : %v", local)

	// Channel to receive discovered service entries
	entries := make(chan *zeroconf.ServiceEntry)

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Println("Found service:", entry.ServiceInstanceName(), entry.Text)
			names := strings.Split(entry.Instance, ":")
			if names[0] == "mldc_sim" {
				log.Printf("entry.Domain : %v", entry.Domain)
				log.Printf("entry.HostName : %v", entry.HostName)
				log.Printf("entry.Instance : %v", entry.Instance)
				log.Printf("entry.Service : %v", entry.Service)
				log.Printf("entry.Text : %v", entry.Text)
				log.Printf("ip addrs : %v", entry.AddrIPv4)
				var ip string = ""
				if names[1] != "" {
					ip = names[1]
				} else {
					ip = t.checkPortAvailavle(entry.AddrIPv4, entry.Port)
				}
				log.Printf("Sim Server IP : %v", ip)
				ni.SetSimAddr(ip, entry.Port)
				t.registerNode(ip, entry.Port)

				//ni.SetSimAddr(entry.AddrIPv4[0].String(), entry.Port)
				//t.setServerInfo(entry.AddrIPv4[0].String(), entry.Port)
				//t.registerNode(entry.AddrIPv4[0].String(), entry.Port)
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

// func NewTMC(db dbagent.DBAgent, sim *dtype.NodeInfo, local *dtype.NodeInfo) *TestMgrCli {
// 	log.Println("start Testmgr Client")
// 	tmc := TestMgrCli{sim: sim, local: local, db: db}
// 	go tmc.startResolver()
// 	return &tmc
// }

func TestMgrCliInst() *TestMgrCli {
	once.Do(func() {
		tmc = &TestMgrCli{}
		log.Println("start Testmgr Client")
		go tmc.startResolver()
	})

	return tmc
}
