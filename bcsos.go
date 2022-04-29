package main

import (
	"fmt"
	"log"
	"net"
	"strings"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}
func localAddresses() {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
		return
	}
	for _, i := range ifaces {
		addrs, err := i.Addrs()
		if err != nil {
			log.Print(fmt.Errorf("localAddresses: %+v\n", err.Error()))
			continue
		}
		log.Printf("iface : %v", i)
		for _, a := range addrs {
			ips := strings.Split(a.String(), "/")
			nip := net.ParseIP(ips[0])

			log.Printf("a : %v, ,nip : %v", a, nip.String())
			log.Printf("To4 : %v, To16 : %v", nip.To4(), nip.To16())

		}
	}
}

func main() {
	localAddresses()
}
