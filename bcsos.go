package main

import (
	"fmt"
	"log"
	"math/big"
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
	// localAddresses()
	h1 := "0009ad6df4a2a0eb5f34f3d728044ce1fec0b05894f425f0625e935589195fd7"
	n1, _ := new(big.Int).SetString(h1[6:], 16)
	log.Printf("n1 =  %v", n1)

	h2 := "000c96e9ff47879778c43a593e7cafbd31bdd4104e5578b05368505727a4c9fa"
	n2, _ := new(big.Int).SetString(h2[6:], 16)
	log.Printf("n2 =  %v", n2)

}
