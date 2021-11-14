package testmgrcli

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/grandcat/zeroconf"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
)

type TestMgrCli struct {
	sim   *dtype.NodeInfo
	local *dtype.NodeInfo
	db    dbagent.DBAgent
}

func (t *TestMgrCli) registerNode(ip string, port int) {
	url := fmt.Sprintf("ws://%v:%v/register", ip, port)
	log.Println("Making call to", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer ws.Close()

	if err := ws.WriteJSON(t.local); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}

	var node dtype.NodeInfo
	if err := ws.ReadJSON(&node); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}

	t.local.IP = node.IP
	t.local.Hash = node.Hash
	log.Printf("Got response: %v\n", t.local)
	log.Printf("Recevied node : %v", node)
}

type Test struct {
	Start bool `json:"start"`
}

func (t *TestMgrCli) NodeInfoHandler(ws *websocket.Conn, w http.ResponseWriter, r *http.Request) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var test Test
			if err := ws.ReadJSON(&test); err != nil {
				log.Printf("Read json error : %v", err)
				return
			}
			log.Printf("receive : %v", test)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	func() {
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				//var status dbagent.DBStatus
				status := t.db.GetDBStatus()
				if err := ws.WriteJSON(status); err != nil {
					log.Printf("Write json error : %v", err)
					return
				}
			}
		}
	}()
}

func (t *TestMgrCli) setServerInfo(ip string, port int) {
	t.sim.IP = ip
	t.sim.Port = port
}

func (t *TestMgrCli) startResolver() {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("TestMgr Info : %v", t.local)

	// Channel to receive discovered service entries
	entries := make(chan *zeroconf.ServiceEntry)

	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			log.Println("Found service:", entry.ServiceInstanceName(), entry.Text)
			names := strings.Split(entry.ServiceInstanceName(), ".")
			if names[0] == "bcsos-tms" {
				t.setServerInfo(entry.AddrIPv4[0].String(), entry.Port)
				t.registerNode(entry.AddrIPv4[0].String(), entry.Port)
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

func NewTMC(db dbagent.DBAgent, sim *dtype.NodeInfo, local *dtype.NodeInfo) *TestMgrCli {
	log.Println("start Testmgr Client")
	tmc := TestMgrCli{sim: sim, local: local, db: db}
	go tmc.startResolver()
	return &tmc
}
