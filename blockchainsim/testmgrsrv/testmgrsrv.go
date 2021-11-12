package testmgrsrv

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/blockchainsim/bcdummy"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Test struct {
	Start bool `json:"start"`
}

type Handler struct {
	http.Handler
	db      dbagent.DBAgent
	Nodes   map[string](dtype.NodeInfo)
	BCDummy *bcdummy.Handler
	TC      *TestConfig
	Ready   bool
}

func (h *Handler) registerHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("registerHandler", err)
		return
	}
	defer ws.Close()

	var node dtype.NodeInfo
	if err := ws.ReadJSON(&node); err != nil {
		log.Printf("Read json error : %v", err)
	}

	ip := strings.Split(ws.RemoteAddr().String(), ":")
	node.IP = ip[0]
	h.Nodes[node.Hash] = node

	ws.WriteJSON(node)
	log.Printf("From client : %v", node)
}

func (h *Handler) nodesHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("nodesHandler", err)
		return
	}
	defer ws.Close()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			var test Test
			if err := ws.ReadJSON(&test); err != nil {
				log.Printf("Read json error : %v", err)
				return
			}
			log.Printf("Test start/stop receive : %v", test)
			h.UpdateTestStatus(test.Start)
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
				var nodes []dtype.NodeInfo
				for _, n := range h.Nodes {
					nodes = append(nodes, n)
				}
				sort.Slice(nodes, func(i, j int) bool {
					return nodes[i].Hash < nodes[j].Hash
				})

				if err := ws.WriteJSON(nodes); err != nil {
					log.Printf("Write json error : %v", err)
					return
				}
			}
		}
	}()
}

func (h *Handler) versionHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("nodesHandler", err)
		return
	}
	defer ws.Close()

	var version dtype.Version
	if err := ws.ReadJSON(&version); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}
	log.Printf("receive version : %v", version)

	var nodes []dtype.NodeInfo
	for _, n := range h.Nodes {
		nodes = append(nodes, n)
	}
	if err := ws.WriteJSON(nodes); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}
}

func (h *Handler) UpdateTestStatus(ready bool) {
	h.Ready = ready
	h.BCDummy.Ready = ready
}

func (h *Handler) StartService(port int) {
	log.Println("starting http service...")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), h.Handler); err != nil {
		log.Fatal(err)
	}
}

func (h *Handler) StartDummy() {
	// Wait until clients join
	func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()
		for {
			<-ticker.C
			if h.Ready {
				return
			}
		}
	}()

	for _, n := range h.Nodes {
		log.Printf("Test Start : %v", n)
	}

	h.BCDummy.Start()
}

func NewHandler(path string) *Handler {
	m := mux.NewRouter()
	h := &Handler{
		Handler: m,
		db:      dbagent.NewDBAgent(path, 100), // simulator use storage class 100
		Nodes:   make(map[string]dtype.NodeInfo),
		BCDummy: nil,
		TC:      nil,
		Ready:   false,
	}

	m.Handle("/", http.FileServer(http.Dir("static")))
	m.HandleFunc("/register", h.registerHandler)
	m.HandleFunc("/nodes", h.nodesHandler)
	m.HandleFunc("/version", h.versionHandler)

	h.BCDummy = bcdummy.NewBCDummy(h.db, &h.Nodes)
	h.TC = NewTestConfig(h.db, &h.Nodes)

	go h.StartDummy()

	return h
}
