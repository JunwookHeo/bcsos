package cipher

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"log"

	"github.com/junwookheo/bcsos/common/galois"
)

const ALIGN = 4

var GP = galois.GFN(32)

func GetHashString(buf []byte) string {
	hash := sha256.Sum256(buf)
	hash = sha256.Sum256(hash[:])
	return hex.EncodeToString(hash[:])
}

func EncryptXorWithVariableLength(key, s []byte) (string, []byte) {
	lk := len(key)
	ls := len(s)
	d := make([]byte, ls)

	if lk < ls {
		for i := 0; i < ls; i++ {
			d[i] = key[i%lk] ^ s[i]
		}
	} else if lk > ls {
		for i := 0; i < ls; i++ {
			d[i] = key[i] ^ s[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			d[i] = key[i] ^ s[i]
		}
	}

	return GetHashString(d), d
}

func EncryptXorWithVariableLength2(key, s []byte) (string, []byte) {
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

	ui_key := make([]uint32, lk)
	err := binary.Read(key_reader, binary.LittleEndian, &ui_key)
	if err != nil {
		log.Panicf("converting key to uint32 error : %v", err)
		return "", nil
	}

	ui_s := make([]uint32, ls)
	err = binary.Read(s_reader, binary.LittleEndian, &ui_s)
	if err != nil {
		log.Panicf("converting source to uint32 error : %v", err)
		return "", nil
	}

	ui_d := make([]uint32, ls)

	if lk < ls {
		for i := 0; i < ls; i++ {
			ui_d[i] = ui_key[i%lk] ^ ui_s[i]
		}
	} else if lk > ls {
		for i := 0; i < ls; i++ {
			ui_d[i] = ui_key[i] ^ ui_s[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			ui_d[i] = ui_key[i] ^ ui_s[i]
		}
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, ui_d)
	if err != nil {
		log.Panicf("convert uint32 to byte error : %v", err)
		return "", nil
	}

	return GetHashString(buf.Bytes()), buf.Bytes()
}

func DecryptXorWithVariableLength(key, s []byte) []byte {
	lk := len(key)
	ls := len(s)
	d := make([]byte, ls)

	if lk < ls {
		for i := 0; i < ls; i++ {
			d[i] = key[i%lk] ^ s[i]
		}
	} else if lk > ls {
		for i := 0; i < ls; i++ {
			d[i] = key[i] ^ s[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			d[i] = key[i] ^ s[i]
		}
	}

	return d
}

func DecryptXorWithVariableLength2(key, s []byte) []byte {
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

	ui_key := make([]uint32, lk)
	err := binary.Read(key_reader, binary.LittleEndian, &ui_key)
	if err != nil {
		log.Panicf("converting key to uint32 error : %v", err)
		return nil
	}

	ui_s := make([]uint32, ls)
	err = binary.Read(s_reader, binary.LittleEndian, &ui_s)
	if err != nil {
		log.Panicf("converting source to uint32 error : %v", err)
		return nil
	}

	ui_d := make([]uint32, ls)

	if lk < ls {
		for i := 0; i < ls; i++ {
			ui_d[i] = ui_key[i%lk] ^ ui_s[i]
		}
	} else if lk > ls {
		for i := 0; i < ls; i++ {
			ui_d[i] = ui_key[i] ^ ui_s[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			ui_d[i] = ui_key[i] ^ ui_s[i]
		}
	}

	buf := new(bytes.Buffer)
	err = binary.Write(buf, binary.LittleEndian, ui_d)
	if err != nil {
		log.Panicf("convert uint32 to byte error : %v", err)
		return nil
	}

	return buf.Bytes()
}

func GetHashforXorKey(key []byte, ls int) string {
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
