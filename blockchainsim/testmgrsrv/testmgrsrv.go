package testmgrsrv

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"syscall"
	"text/template"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/blockchainsim/simulation"
	"github.com/junwookheo/bcsos/common/blockchain"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/datalib"
	"github.com/junwookheo/bcsos/common/dbagent"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/listener"
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
	db    dbagent.DBAgent
	Nodes map[string](dtype.NodeInfo)
	bcsim *simulation.Handler
	TC    *TestConfig
	Ready bool
	el    *listener.EventListener
	cand  *datalib.CandidateBlocks
	mutex sync.Mutex
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

	log.Printf("ws remote addr : %v", ws.RemoteAddr().String())
	ip := strings.Split(ws.RemoteAddr().String(), ":")
	node.IP = ip[0]
	h.mutex.Lock()
	h.Nodes[node.Hash] = node
	h.mutex.Unlock()

	ws.WriteJSON(node)
	log.Printf("From client : %v", node)

	for {
		if err := ws.ReadJSON(&node); err != nil {
			// log.Printf("Node is disconnected : %v, %v", err, h.Nodes)
			h.mutex.Lock()
			delete(h.Nodes, node.Hash)
			h.mutex.Unlock()
			// log.Printf("Client node list : %v", h.Nodes)
			break
		}
	}
}

// Send nodes information connected to simulator.
// Web app will receive the information
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
		var cnt int = 0
		for {
			select {
			case <-done:
				return
			case <-ticker.C:
				var nodes []dtype.NodeInfo
				if len(h.Nodes) != cnt {
					cnt = len(h.Nodes)

					for _, n := range h.Nodes {
						nodes = append(nodes, n)
					}
					sort.Slice(nodes, func(i, j int) bool {
						return nodes[i].SC < nodes[j].SC
					})

					if err := ws.WriteJSON(nodes); err != nil {
						log.Printf("Write json error : %v", err)
						return
					}
				}
			}
		}
	}()
}

func (h *Handler) pingHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("pingHandler", err)
		return
	}
	defer ws.Close()

	peer := dtype.NodeInfo{}
	if err := ws.ReadJSON(&peer); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}
	log.Printf("receive peer : %v", peer)

	var nodes []dtype.NodeInfo
	for _, n := range h.Nodes {
		nodes = append(nodes, n)
	}

	if err := ws.WriteJSON(nodes); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}
}
func (h *Handler) commandProc(cmd *dtype.Command) {

	log.Printf("commandProc : %v", cmd)
	if cmd.Cmd == "SET" {
		switch cmd.Subcmd {
		case "Test":
			if cmd.Arg1 == "Start" {
				h.el.Notify("Start")
			} else if cmd.Arg1 == "Stop" {
				h.el.Notify("Stop")
			}
		}
	}
}

func (h *Handler) sendCommand(cmd dtype.Command, ip string, port int) {
	url := fmt.Sprintf("ws://%v:%v/command", ip, port)
	//log.Printf("Send new block to %v", url)

	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Printf("dial: %v", err)
		return
	}
	defer ws.Close()
	//log.Printf("DefaultDialer Send command to %v", url)

	if err := ws.WriteJSON(cmd); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}

	var res dtype.Command
	if err := ws.ReadJSON(&res); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}
}

func (h *Handler) broadcastCommand(cmd dtype.Command) {
	log.Printf("broadcast command : %v", h.Nodes)
	for _, node := range h.Nodes {
		h.sendCommand(cmd, node.IP, node.Port)
	}
}

// commandHandler deals with commands from web app
// Web app --> blockchain simulator --> storage nodes
func (h *Handler) commandHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("nodesHandler", err)
		return
	}
	defer ws.Close()
	//done := make(chan struct{})
	func() {
		//defer close(done)
		for {
			var cmd dtype.Command
			if err := ws.ReadJSON(&cmd); err != nil {
				log.Printf("Read json error : %v", err)
				return
			}

			h.broadcastCommand(cmd)
			log.Printf("Test command receive : %v", cmd)

			cmd.Arg2 = "OK"
			if err := ws.WriteJSON(cmd); err != nil {
				log.Printf("Write json error : %v", err)
				return
			}
			h.commandProc(&cmd)
		}
	}()
}

// newBlockHandler is called when a new block is received from miners
// When a node receive this, it stores the block on its local db
// Request : a new block
// Response : none
func (h *Handler) newBlockHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("newBlockHandler", err)
		return
	}
	defer ws.Close()

	var block blockchain.Block
	if err := ws.ReadJSON(&block); err != nil {
		log.Printf("Read json error : %v", err)
	}

	// log.Printf("Rcv new block(%v) : %v-%v", block.Header.Height, hex.EncodeToString(block.Header.Hash), hex.EncodeToString(block.Header.PrvHash))
	h.cand.PushAndSave(&block, h.db)
	h.cand.CheckFinality()
}

func (h *Handler) UpdateTestStatus(ready bool) {
	h.Ready = ready
}

func (h *Handler) StartService(port int) {
	log.Printf("starting http service... : %v", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), h.Handler); err != nil {
		log.Fatal(err)
	}
}

func (h *Handler) SimulateAccessPatternProc() {
	command := make(chan string)
	h.el.AddListener(command)
	id := 0

	go func(command <-chan string) {
		var status = "Pause"

		for {
			select {
			case cmd := <-command:
				switch cmd {
				case "Stop":
					status = "Stop"
					log.Printf("access pattern stoping")
				case "Start":
					status = "Running"
				}
			default:
				if status == "Running" {
					h.bcsim.SimulateAccessPattern(&id)
					time.Sleep(time.Second)
				} else {
					time.Sleep(time.Duration(config.BLOCK_CREATE_PERIOD) * time.Second)
				}
			}
		}
	}(command)
}

func (h *Handler) KillProcess() {
	p, err := os.FindProcess(os.Getpid())

	if err != nil {
		return
	}

	p.Signal(syscall.SIGTERM)
}

func (h *Handler) SimulateTransactionProc() {
	command := make(chan string)
	h.el.AddListener(command)
	id := 0

	go func(command <-chan string) {
		var status = "Pause"

		for {
			select {
			case cmd := <-command:
				switch cmd {
				case "Stop":
					status = "Stop"
					log.Printf("Stop running")
					// Wait untile client nodes terminate
					time.Sleep(time.Duration(20) * time.Second)
					h.db.Close()
					h.KillProcess()
				case "Start":
					status = "Running"
				}
			default:
				if status == "Running" {
					if id < config.TOTAL_TRANSACTIONS {
						h.bcsim.SimulateTransaction(id)
						id++
						time.Sleep(time.Second)
					} else {
						//Stop generating access pattern proc
						h.el.Notify("Stop")
						// Sleep for a while for client nodes
						time.Sleep(time.Duration(10) * time.Second)
						// Send test stop to client nodes
						cmd := dtype.Command{Cmd: "SET", Subcmd: "Test", Arg1: "Stop", Arg2: "", Arg3: ""}
						h.broadcastCommand(cmd)
						log.Printf("sendiing stop")
					}
				} else {
					time.Sleep(time.Duration(config.BLOCK_CREATE_PERIOD) * time.Second)
				}
			}
		}
	}(command)
}

func (h *Handler) SimulateBtcBlockProc() {
	command := make(chan string)
	h.el.AddListener(command)

	msg := make(chan string)
	h.bcsim.SimulateBtcBlock(msg)

	go func(command <-chan string) {
		var status = "Pause"

		for {
			select {
			case cmd := <-command:
				switch cmd {
				case "Stop":
					status = "Stop"
					log.Printf("Stop running")
					close(msg)

					// Wait untile client nodes terminate
					time.Sleep(time.Duration(20) * time.Second)
					h.db.Close()
					h.KillProcess()
				case "Start":
					status = "Running"
				}
			default:
				if status == "Running" {
					b, ok := <-msg
					if !ok {
						log.Println("Channle closed")
						break
					}

					if b == config.END_TEST {
						//Stop generating access pattern proc
						h.el.Notify("Stop")
						// Sleep for a while for client nodes
						time.Sleep(time.Duration(10) * time.Second)
						// Send test stop to client nodes
						cmd := dtype.Command{Cmd: "SET", Subcmd: "Test", Arg1: "Stop", Arg2: "", Arg3: ""}
						h.broadcastCommand(cmd)
						log.Printf("sendiing stop")
						break
					} else {
						log.Printf("Block : %v", b[:8])
						h.bcsim.BroadcastBtcBlock(b)
					}

					time.Sleep(time.Second * 60)
				} else {
					time.Sleep(time.Duration(config.BLOCK_CREATE_PERIOD) * time.Second)
				}
			}
		}
	}(command)

}

func NewHandler(mode string, path string) *Handler {
	m := mux.NewRouter()
	h := &Handler{
		Handler: m,
		db:      dbagent.NewDBAgent(path), // simulator use storage class 100
		Nodes:   make(map[string]dtype.NodeInfo),
		bcsim:   nil,
		TC:      nil,
		Ready:   false,
		el:      nil,
		cand:    datalib.NewCandidateBlocks(),
		mutex:   sync.Mutex{},
	}

	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/static/", http.StripPrefix("/static", fs))
	m.Handle("/", fs)
	m.Handle("/styles.css", fs)

	m.HandleFunc("/register", h.registerHandler)
	m.HandleFunc("/nodes", h.nodesHandler)
	m.HandleFunc("/ping", h.pingHandler)
	m.HandleFunc("/command", h.commandHandler)
	m.HandleFunc("/broadcastnewblock", h.newBlockHandler)

	h.el = listener.EventListenerInst()

	h.bcsim = simulation.NewSimAgent(h.db, &h.Nodes)
	h.TC = NewTestConfig(h.db, &h.Nodes)

	// BITCOIN NOT USE
	// h.SimulateTransactionProc()
	// h.SimulateAccessPatternProc()
	h.SimulateBtcBlockProc()

	return h
}

func Home(w http.ResponseWriter, r *http.Request) {
	t, _ := template.ParseFiles("/")
	err := t.Execute(w, nil)
	if err != nil {
		log.Printf("%v", err)
	}
}
