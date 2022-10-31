package bitcoin

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"log"

	"github.com/junwookheo/bcsos/common/poscipher"
)

const ALIGN = poscipher.ALIGN

type RawBlock struct {
	rawBuf []byte
	pos    uint32
}

func NewRawBlock(raw string) *RawBlock {
	rawb, err := hex.DecodeString(raw)
	if err != nil {
		log.Panicf("NewRawBlock Error : %v", err)
		return nil
	}
	// log.Printf("===> leng block : %v", len(rawb))

	lp := (ALIGN - len(rawb)%ALIGN) % ALIGN // length of padding
	pad := make([]byte, lp)

	return &RawBlock{rawBuf: append(rawb, pad...), pos: 0}
}

func (rb *RawBlock) GetBlockBytes() []byte {
	return rb.rawBuf
}

func (rb *RawBlock) ReadBytes(len uint32) []byte {
	buf := make([]byte, len)
	err := binary.Read(bytes.NewBuffer(rb.rawBuf[rb.pos:]), binary.LittleEndian, &buf)
	if err != nil {
		log.Panicln(err)
		return nil
	}

	rb.pos += len
	return buf
}

func (rb *RawBlock) ReadUint8() uint8 {
	buf := make([]byte, 1)
	err := binary.Read(bytes.NewBuffer(rb.rawBuf[rb.pos:]), binary.LittleEndian, &buf)
	if err != nil {
		log.Panicln(err)
		return 0
	}

	rb.pos += 1
	return buf[0]
}

func (rb *RawBlock) ReadUint16() uint16 {
	buf := make([]byte, 2)
	err := binary.Read(bytes.NewBuffer(rb.rawBuf[rb.pos:]), binary.LittleEndian, &buf)
	if err != nil {
		log.Panicln(err)
		return 0
	}

	rb.pos += 2
	return binary.LittleEndian.Uint16(buf)
}

func (rb *RawBlock) ReadUint32() uint32 {
	buf := make([]byte, 4)
	err := binary.Read(bytes.NewBuffer(rb.rawBuf[rb.pos:]), binary.LittleEndian, &buf)
	if err != nil {
		log.Panicln(err)
		return 0
	}

	rb.pos += 4
	return binary.LittleEndian.Uint32(buf)
}

func (rb *RawBlock) ReadUint64() uint64 {
	buf := make([]byte, 8)
	err := binary.Read(bytes.NewBuffer(rb.rawBuf[rb.pos:]), binary.LittleEndian, &buf)
	if err != nil {
		log.Panicln(err)
		return 0
	}

	rb.pos += 8
	return binary.LittleEndian.Uint64(buf)
}

func (rb *RawBlock) ReadVariant() uint64 {
	t := rb.ReadUint8()
	var val uint64
	if t == 0xfd {
		val = uint64(rb.ReadUint16())
	} else if t == 0xfe {
		val = uint64(rb.ReadUint32())
	} else if t == 0xff {
		val = uint64(rb.ReadUint64())
	} else {
		val = uint64(t)
	}
	return val
}

func (rb *RawBlock) ReadOptional(len uint32) []byte {
	buf := make([]byte, len)
	err := binary.Read(bytes.NewBuffer(rb.rawBuf[rb.pos:]), binary.LittleEndian, &buf)
	if err != nil {
		log.Panicln(err)
		return nil
	}

	return buf
}

func (rb *RawBlock) IncreasePos(len uint32) {
	rb.pos += len
}

func (rb *RawBlock) ReverseBuf(buf []byte) []byte {
	n := len(buf)
	for i := 0; i < n/2; i++ {
		buf[i], buf[n-1-i] = buf[n-1-i], buf[i]
	}
	return buf
}

func (rb *RawBlock) GetRawBytes(start uint32, len uint32) []byte {
	return rb.rawBuf[start : start+len]
}

func (rb *RawBlock) GetPosition() uint32 {
	return rb.pos
}
