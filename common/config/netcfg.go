package config

import (
	"bytes"
	"encoding/binary"
)

type MSGTYPE int

/* Definition og Message Type for TCP Packet
|MSGTYPE(4)|LASTFLAG(1)|RESERVED(3)|Length(4)|BODY(MAX - 12)|
*/
const TCPSIZE = 4096

const MAXBDYLEN = TCPSIZE - 12

const (
	NEWBLOCK MSGTYPE = iota
	QRYBLOCK
	QRYTRANSACTION
	QRYBLOCKHEADER
)

type TcpPacket struct {
	MsgType  uint32
	LastFlag uint8
	Reserved [3]uint8
	Length   uint32
}

func SplitTcp(buf []byte) [][]byte {
	var chunk []byte
	chunks := make([][]byte, 0, len(buf)/MAXBDYLEN+1)
	for len(buf) >= MAXBDYLEN {
		chunk, buf = buf[:MAXBDYLEN], buf[MAXBDYLEN:]
		chunks = append(chunks, chunk)
	}
	if len(buf) > 0 {
		chunks = append(chunks, buf[:])
	}
	return chunks
}

func TcpHeaderToBytes(h *TcpPacket) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, *h)
	return buf.Bytes()
}

func BytesToTcpHeader(b []byte) *TcpPacket {
	var h TcpPacket
	buf := bytes.NewReader(b)
	binary.Read(buf, binary.BigEndian, &h)
	return &h
}
