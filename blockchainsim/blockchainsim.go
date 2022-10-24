package main

import (
	"flag"
	"log"
	"net"
	"os"
	"os/signal"
	"strings"

	"github.com/grandcat/zeroconf"
	"github.com/junwookheo/bcsos/blockchainsim/testmgrsrv"
	"github.com/junwookheo/bcsos/common/config"
)

const DB_PATH = "./bc_sim.db"
const WALLET_PATH = "./bc_sim.wallet"

const PORT = 8080

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
	config.DB_PATH = DB_PATH
	config.WALLET_PATH = WALLET_PATH
}

func flagParse() (string, string) {
	pmode := flag.String("mode", "ST", "ST: Test storage (Server generates tr and ap object), MI: Test Miner(generate tr and access object in local)")
	ip := flag.String("ip", "", "IP for simulation server")
	iface := flag.String("iface", "", "IP Interface for simulation server, 'eth0', 'wi-fi'")
	flag.Parse()

	log.Printf("=== ip : %v", *ip)
	if *ip == "" && *iface != "" {
		*ip = localAddresses(iface)
	}
	log.Printf("=== ip : %v", *ip)
	return *pmode, *ip
}

func localAddresses(target *string) string {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Printf("iface error : %v", err)
		return ""
	}
	for _, i := range ifaces {
		if strings.ToLower(i.Name) == strings.ToLower(*target) {
			addrs, err := i.Addrs()
			if err != nil {
				log.Printf("Net Addrs error : %v", err)
				break
			}
			log.Printf("iface : %v", i)
			for _, a := range addrs {
				ips := strings.Split(a.String(), "/")
				nip := net.ParseIP(ips[0])

				if nip.To4() != nil {
					log.Printf("IP : %v", nip.String())
					return nip.String()
				}
			}
		}
	}

	return ""
}

func main() {
	log.Println("Start blockchain simulator")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	defer signal.Reset()

	mode, ip := flagParse()
	s := testmgrsrv.NewHandler(mode, config.DB_PATH)
	go s.StartService(PORT)
	//go bcdummy.Start()

	// Extra information about our service
	meta := []string{
		"version=0.1.0",
		"bctestmgr",
		"sim_ip=",
	}

	service, err := zeroconf.Register(
		"mldc_sim:"+ip,    // service instance name
		"_omxremote._tcp", // service type and protocl
		"local.",          // service domain
		PORT,              // service port
		meta,              // service metadata
		nil,               // register on all network interfaces
	)

	if err != nil {
		log.Fatal(err)
	}

	defer service.Shutdown()

	<-interrupt
	log.Println("interrupt finish")
}
