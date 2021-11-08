package storage

import (
	"fmt"
	"log"
	"net"

	"github.com/junwookheo/bcsos/common/bcapi"
	"github.com/junwookheo/bcsos/common/shareddata"
	"github.com/junwookheo/bcsos/storagesrv/network"
)

func init() {
	port, err := getFreePort()
	if err != nil {
		log.Panicf("Searching free port fail : %v", err)
	}

	// Set Local node info
	shareddata.TestMgrInfo.Local.Port = port
	shareddata.TestMgrInfo.Local.IP = getLocalIP()
	shareddata.TestMgrInfo.Local.AddrHash = ""
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
	return ""
}

func Start() {
	network.Start()
}

func Stop() {
	bcapi.Close()
}
