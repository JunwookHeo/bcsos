package config

import (
	"log"
	"testing"
	"unsafe"
)

func TestNetConfig(t *testing.T) {
	var i uint32 = 1
	var b uint8 = 1
	log.Printf("int : %d, bool : %d", unsafe.Sizeof(i), unsafe.Sizeof(b))
}

func TestReadSlice(t *testing.T) {
	b := []byte("12345678901234567890")
	cs := SplitTcp(b)
	log.Printf("b-%d: %v", len(b), b)
	log.Printf("cs-%d: %v", len(cs), cs[len(cs)-1])
}

func TestTcpPacket(t *testing.T) {
	tp := TcpPacket{5, 10, [3]byte{}, 111115678}
	log.Printf("TCP Packet Size : %d, %v", unsafe.Sizeof(tp), tp)
	b := TcpHeaderToBytes(&tp)
	log.Printf("arry %d - %v", len(b), b)
	tp2 := BytesToTcpHeader(b)
	log.Printf("tp2 : %v", *tp2)
}
