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

	pre := uint32(0)
	if lk < ls {
		for i := 0; i < ls; i++ {
			// y[i] = k[i%lk] ^ x[i]
			d := (x[i] ^ k[i%lk]) ^ pre
			y[i] = uint32(GF.Exp(uint64(d), 6442450943))
			pre = y[i]
		}
	} else if lk > ls {
		for i := 0; i < ls; i++ {
			// y[i] = k[i] ^ x[i]
			d := (x[i] ^ k[i]) ^ pre
			y[i] = uint32(GF.Exp(uint64(d), 6442450943))
			pre = y[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			// y[i] = k[i] ^ x[i]
			d := (x[i] ^ k[i]) ^ pre
			y[i] = uint32(GF.Exp(uint64(d), 6442450943))
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
	pre := uint32(0)
	if lk < ls {
		for i := 0; i < ls; i++ {
			// ui_d[i] = ui_key[i%lk] ^ ui_s[i]
			d := uint32(GF.Exp(uint64(x[i]), 2))
			y[i] = (d ^ pre) ^ k[i%lk]
			pre = x[i]
		}
	} else if lk > ls {
		for i := 0; i < ls; i++ {
			// y[i] = k[i] ^ x[i]
			d := uint32(GF.Exp(uint64(x[i]), 2))
			y[i] = (d ^ pre) ^ k[i]
			pre = x[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			// y[i] = k[i] ^ x[i]
			d := uint32(GF.Exp(uint64(x[i]), 2))
			y[i] = (d ^ pre) ^ k[i]
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
	} else if lk > ls {
		for i := 0; i < ls; i++ {
			d[i] = key[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			d[i] = key[i]
		}
	}

	return GetHashString(d)
}
