package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/junwookheo/bcsos/blockchainnode/mining"
	"github.com/junwookheo/bcsos/blockchainnode/network"
	"github.com/junwookheo/bcsos/blockchainnode/storage"
	"github.com/junwookheo/bcsos/blockchainnode/testmgrcli"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dtype"
	"github.com/junwookheo/bcsos/common/listener"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

var DATA_DIR string = "./db_nodes"
var db_path string = DATA_DIR + "/dev.db"
var wallet_path string = DATA_DIR + "/dev.wallet"

var (
	ni  *network.NodeInfo
	nm  *network.NodeMgr
	wm  *mining.WalletMgr
	mi  *mining.Mining
	sm  *storage.StorageMgr
	tmc *testmgrcli.TestMgrCli
	el  *listener.EventListener
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)

	config.DB_PATH = db_path
	config.WALLET_PATH = wallet_path
}

// GetFreePort asks the kernel for a free open port that is ready to use.
func getFreePort() (port int, err error) {
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

func flagParse() (string, int, int) {
	pmode := flag.String("mode", "ST", "ST: Test storage (Server generates tr and ap object), MI: Test Miner(generate tr and access object in local)")
	psc := flag.Int("sc", 0, "Storage class : 0 to 4")
	pport := flag.Int("port", 0, "Port number of local if 0, it will use a free port")
	flag.Parse()
	if *pport == 0 {
		port, err := getFreePort()
		if err != nil {
			log.Panicf("Get free port error : %v", err)
		}
		*pport = port
	}
	return *pmode, *psc, *pport
}

func initNode() {
	mode, sc, port := flagParse()
	mode = strings.ToUpper(mode)

	if _, err := os.Stat(DATA_DIR); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(DATA_DIR, os.ModePerm)
		if err != nil {
			log.Panicln(err)
			return
		}
	}

	config.DB_PATH = fmt.Sprintf("./db_nodes/%v.db", port)
	config.WALLET_PATH = fmt.Sprintf("./db_nodes/%v.wallet", port)

	// init wallet Manager
	wm = mining.WalletMgrInst(config.WALLET_PATH)
	w := wm.GetWallet()
	//w := wallet.NewWallet(config.WALLET_PATH)

	// hash := hex.EncodeToString(w.GetAddress()[:])
	// Hash of Bitcoin address is used for content-addressing
	bhash := sha256.Sum256(w.GetAddress()[:])
	hash := hex.EncodeToString(bhash[:])
	log.Printf("==>%v", hash)

	// init nodeInfo
	ni = network.NodeInfoInst()
	ni.SetLocalddrParam(mode, sc, port, hash)

	// init testmgrcli
	tmc = testmgrcli.TestMgrCliInst()

	// init network manager
	nm = network.NodeMgrInst()

	// init EventListener
	el = listener.EventListenerInst()
}

func commandHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("commandHandler", err)
		return
	}
	defer ws.Close()
	var cmd dtype.Command
	if err := ws.ReadJSON(&cmd); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}
	// log.Printf("Test command receive : %v", cmd)

	commandProc(&cmd)

	cmd.Arg2 = "OK"
	if err := ws.WriteJSON(cmd); err != nil {
		log.Printf("Write json error : %v", err)
		return
	}
}

func commandProc(cmd *dtype.Command) {
	// log.Printf("commandProc : %v", cmd)
	if cmd.Cmd == "SET" {
		switch cmd.Subcmd {
		case "Test":
			if cmd.Arg1 == "Start" {
				el.Notify("Start")
			} else if cmd.Arg1 == "Stop" {
				el.Notify("Stop")
			} else if cmd.Arg1 == "Pause" {
				el.Notify("Pause")
			} else if cmd.Arg1 == "Resume" {
				el.Notify("Resume")
			}
		}
	}
}

func KillProcess() {
	p, err := os.FindProcess(os.Getpid())

	if err != nil {
		return
	}

	p.Signal(syscall.SIGTERM)
}

// Response to web app with dbstatus information
// keep sending dbstatus to the web app
func endTestHandler(w http.ResponseWriter, r *http.Request) {
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("endTestHandler", err)
		return
	}
	defer ws.Close()
	var endtest string
	if err := ws.ReadJSON(&endtest); err != nil {
		log.Printf("Read json error : %v", err)
		return
	}

	if endtest == config.END_TEST {
		log.Println("Received End test")
		sm.Stop()
		time.Sleep(3 * time.Second)
		//syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		KillProcess()
	}
}

func EndTestProc() {
	command := make(chan string)
	el.AddListener(command)

	go func(command <-chan string) {
		for {
			select {
			case cmd := <-command:
				switch cmd {
				case "Stop":
					log.Println("Received End test")
					sm.Stop()
					time.Sleep(3 * time.Second)
					//syscall.Kill(syscall.Getpid(), syscall.SIGINT)
					KillProcess()
					return
				}
			default:
				// log.Println("=========EndTestProc")
				time.Sleep(time.Duration(config.TIME_AP_GEN) * time.Second)
			}
		}
	}(command)
}

func PeerListProc() {
	command := make(chan string)
	el.AddListener(command)

	go func(command <-chan string) {
		var status = "Pause"
		for {
			select {
			case cmd := <-command:
				switch cmd {
				case "Stop":
					return
				case "Pause":
					status = "Pause"
				case "Resume":
					status = "Running"
				case "Start":
					status = "Running"
				}
			default:
				if status == "Running" {
					ni := network.NodeInfoInst()
					local := ni.GetLocalddr()
					sim := ni.GetSimAddr()
					nm := network.NodeMgrInst()

					if sim.IP != "" && sim.Port != 0 && local.Hash != "" {
						nm.UpdatePeerList(sim, local)
					}
					// log.Println("=========PeerListProc")
					time.Sleep(time.Duration(config.TIME_UPDATE_NEITHBOUR) * time.Second)
				} else {
					time.Sleep(time.Second)
				}
			}
		}

	}(command)
}

func TransactionProc() {
	rand.Seed(time.Now().UnixNano())

	command := make(chan string)
	el.AddListener(command)
	id := 0

	go func(command <-chan string) {
		var status = "Pause"
		for {
			select {
			case cmd := <-command:
				switch cmd {
				case "Stop":
					status = "Stop"
					//mining.SimulateTransaction(-1)
					return
				case "Pause":
					status = "Pause"
				case "Resume":
					status = "Running"
				case "Start":
					if status != "Running" {
						status = "Running"
						log.Println("start mining ===")
						go mi.StartMiningNewBlock(&status)
					}
				}
			default:
				if status == "Running" {
					ni := network.NodeInfoInst()
					local := ni.GetLocalddr()
					if strings.ToUpper(local.Mode) == "MI" {
						mining.SimulateTransaction(id)
						id++
					} else {
						time.Sleep(time.Duration(config.BLOCK_CREATE_PERIOD) * time.Second)
						// log.Printf("Mode : %v", local.Mode)
					}
				} else {
					time.Sleep(time.Second)
				}
			}
		}

	}(command)
}

func main() {
	log.Println("Start Storage Service")
	rand.Seed(time.Now().UnixNano())
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Reset()

	m := mux.NewRouter()
	initNode()
	sm = storage.StorageMgrInst(config.DB_PATH)
	mi = mining.MiningInst()

	m.Handle("/", http.FileServer(http.Dir("static")))
	m.HandleFunc("/command", commandHandler)
	m.HandleFunc("/endtest", endTestHandler)

	sm.SetHttpRouter(m)
	nm.SetHttpRouter(m)
	mi.SetHttpRouter(m)

	sm.ObjectbyAccessPatternProc()
	PeerListProc()
	// No need for BTC PoS
	// TransactionProc()
	EndTestProc()

	local := ni.GetLocalddr()
	log.Printf("Server start : %v", local.Port)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%v", local.Port),
		Handler: m,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	log.Println("End wait")
	<-interrupt
	log.Println("interrupt")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		// extra handling here
		cancel()
	}()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
	log.Print("Server Exited Properly")
}
