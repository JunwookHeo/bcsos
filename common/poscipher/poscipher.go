package poscipher

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"log"

	"github.com/junwookheo/bcsos/common/galois"
)

const ALIGN = 4
const Ix2 = 2147483648 // P is Even so P/2 = (2^32) / 2

var GF = galois.GFN(32)

func GetHashString(buf []byte) string {
	hash := sha256.Sum256(buf)
	hash = sha256.Sum256(hash[:])
	return hex.EncodeToString(hash[:])
}

func EncryptPoSWithVariableLength(key, s []byte) (string, []byte) {
	lk := len(key) // assume key is aligned
	if lk%ALIGN != 0 {
		log.Panicf("Error Key length ALIGN: %v", lk)
	}
	ls := len(s)
	if lk%ALIGN != 0 {
		log.Panicf("Error Source length ALIGN : %v", ls)
	}

	lk /= ALIGN
	ls /= ALIGN

	key_reader := bytes.NewReader(key)
	s_reader := bytes.NewReader(s)

	k := make([]uint32, lk)
	err := binary.Read(key_reader, binary.LittleEndian, &k)
	if err != nil {
		log.Panicf("converting key to uint32 error : %v", err)
		return "", nil
	}

	x := make([]uint32, ls)
	err = binary.Read(s_reader, binary.LittleEndian, &x)
	if err != nil {
		log.Panicf("converting source to uint32 error : %v", err)
		return "", nil
	}

	y := make([]uint32, ls)

	// pre := uint32(0)
	pre := uint32(1)
	if lk < ls {
		for i := 0; i < ls; i++ {
			// d := (x[i] ^ k[i%lk]) ^ pre
			d := (x[i] ^ k[i%lk])
			if pre != 0 {
				d = uint32(GF.Div(uint64(d), uint64(pre)))
			}
			y[i] = uint32(GF.Exp(uint64(d), Ix2))
			pre = y[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			// d := (x[i] ^ k[i]) ^ pre
			d := (x[i] ^ k[i])
			if pre != 0 {
				d = uint32(GF.Div(uint64(d), uint64(pre)))
			}
			y[i] = uint32(GF.Exp(uint64(d), Ix2))
			pre = y[i]
		}
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, y)
	if err != nil {
		log.Panicf("convert uint32 to byte error : %v", err)
		return "", nil
	}

	return GetHashString(buf.Bytes()), buf.Bytes()
}

func DecryptPoSWithVariableLength(key, s []byte) []byte {
	lk := len(key) // assume key is aligned
	if lk%ALIGN != 0 {
		log.Panicf("Error Key length ALIGN: %v", lk)
	}
	ls := len(s)
	if lk%ALIGN != 0 {
		log.Panicf("Error Source length ALIGN : %v", ls)
	}

	lk /= ALIGN
	ls /= ALIGN
	key_reader := bytes.NewReader(key)
	s_reader := bytes.NewReader(s)

	k := make([]uint32, lk)
	err := binary.Read(key_reader, binary.LittleEndian, &k)
	if err != nil {
		log.Panicf("converting key to uint32 error : %v", err)
		return nil
	}

	x := make([]uint32, ls)
	err = binary.Read(s_reader, binary.LittleEndian, &x)
	if err != nil {
		log.Panicf("converting source to uint32 error : %v", err)
		return nil
	}

	y := make([]uint32, ls)
	// pre := uint32(0)
	pre := uint32(1)
	if lk < ls {
		for i := 0; i < ls; i++ {
			d := uint32(GF.Exp(uint64(x[i]), 2))
			// y[i] = (d ^ pre) ^ k[i%lk]
			if pre != 0 {
				y[i] = uint32(GF.Mul(uint64(d), uint64(pre)))
			}
			y[i] = y[i] ^ k[i%lk]
			pre = x[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			d := uint32(GF.Exp(uint64(x[i]), 2))
			// y[i] = (d ^ pre) ^ k[i]
			if pre != 0 {
				y[i] = uint32(GF.Mul(uint64(d), uint64(pre)))
			}
			y[i] = y[i] ^ k[i]
			pre = x[i]
		}
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, y)
	if err != nil {
		log.Panicf("convert uint32 to byte error : %v", err)
		return nil
	}

	return buf.Bytes()
}

func GetHashforPoSKey(key []byte, ls int) string {
	lk := len(key)
	d := make([]byte, ls)

	if lk < ls {
		for i := 0; i < ls; i++ {
			d[i] = key[i%lk]
		}
	} else {
		for i := 0; i < ls; i++ {
			d[i] = key[i]
		}
	}

	return GetHashString(d)
}

func CalculateXorWithAddress(addr, s []byte) []byte {
	lk := len(addr) // assume key is aligned
	ls := len(s)
	if lk > ls {
		log.Panicf("Error Source length is smaller then addr : %v", ls)
		return nil
	}

	d := make([]byte, ls)

	for i := 0; i < ls; i++ {
		// y[i] = k[i%lk] ^ x[i]
		d[i] = s[i] ^ addr[i%lk]
	}

	return d
}
