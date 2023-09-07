package poscipher

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"log"

	"github.com/junwookheo/bcsos/common/galois"
)

// var ALIGN = 4

// // const Ix2 = 2147483648 // P is Even so P/2 = (2^32) / 2

// var GF = galois.GFN(32)

type Encode struct {
	Align int
	gf    *galois.GF
}

// func init() {
// 	GF = galois.GFN(config.GF_FIELD_SIZE)
// 	ALIGN = int(GF.GetAlign())
// }

func NewEncoder(size uint) *Encode {
	en := Encode{}
	en.gf = galois.GFN(size)
	en.Align = int(en.gf.GetAlign())

	return &en
}

func (en *Encode) GetHashString(buf []byte) string {
	hash := sha256.Sum256(buf)
	hash = sha256.Sum256(hash[:])
	return hex.EncodeToString(hash[:])
}

func (en *Encode) EncryptPoSWithVariableLength(key, s []byte) (string, []byte) {
	lk := len(key) // assume key is aligned
	for i := 0; i < lk%en.Align; i++ {
		key = append(key, 0x00)
	}
	ls := len(s)
	for i := 0; i < ls%en.Align; i++ {
		s = append(s, 0x00)
	}

	lk = len(key) / en.Align
	ls = len(s) / en.Align

	k := make([]uint64, lk)
	for i := 0; i < lk; i++ {
		buf := make([]byte, 8)
		for j := 0; j < en.Align; j++ {
			buf[j] = key[i*en.Align+j]
		}
		k[i] = binary.LittleEndian.Uint64(buf)
	}

	x := make([]uint64, ls)
	for i := 0; i < ls; i++ {
		buf := make([]byte, 8)
		for j := 0; j < en.Align; j++ {
			buf[j] = s[i*en.Align+j]
		}
		x[i] = binary.LittleEndian.Uint64(buf)
	}

	y := make([]uint64, ls)

	pre := uint64(1)
	if lk < ls {
		for i := 0; i < ls; i++ {
			// d := (x[i] ^ k[i%lk]) ^ pre
			d := (x[i] ^ k[i%lk])
			if pre != 0 {
				d = en.gf.Div(d, pre)
			}
			y[i] = en.gf.SqrR(d)
			pre = y[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			// d := (x[i] ^ k[i]) ^ pre
			d := (x[i] ^ k[i])
			if pre != 0 {
				d = en.gf.Div(d, pre)
			}
			y[i] = en.gf.SqrR(d)
			pre = y[i]
		}
	}

	out := new(bytes.Buffer)
	for i := 0; i < ls; i++ {
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, y[i])
		err := binary.Write(out, binary.LittleEndian, buf[:en.Align])
		if err != nil {
			log.Panicf("convert uint64 to byte error : %v", err)
			return "", nil
		}
	}

	return en.GetHashString(out.Bytes()), out.Bytes()
}

func (en *Encode) DecryptPoSWithVariableLength(key, s []byte) []byte {
	lk := len(key) // assume key is aligned
	for i := 0; i < lk%en.Align; i++ {
		key = append(key, 0x00)
	}

	ls := len(s)
	for i := 0; i < ls%en.Align; i++ {
		s = append(s, 0x00)
	}

	lk = len(key) / en.Align
	ls = len(s) / en.Align

	k := make([]uint64, lk)
	for i := 0; i < lk; i++ {
		buf := make([]byte, 8)
		for j := 0; j < en.Align; j++ {
			buf[j] = key[i*en.Align+j]
		}
		k[i] = binary.LittleEndian.Uint64(buf)
	}

	x := make([]uint64, ls)
	for i := 0; i < ls; i++ {
		buf := make([]byte, 8)
		for j := 0; j < en.Align; j++ {
			buf[j] = s[i*en.Align+j]
		}
		x[i] = binary.LittleEndian.Uint64(buf)
	}

	y := make([]uint64, ls)
	pre := uint64(1)
	if lk < ls {
		for i := 0; i < ls; i++ {
			d := en.gf.Exp(x[i], 2)
			// y[i] = (d ^ pre) ^ k[i%lk]
			if pre != 0 {
				y[i] = en.gf.Mul(d, pre)
			}
			y[i] = y[i] ^ k[i%lk]
			pre = x[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			d := en.gf.Exp(x[i], 2)
			// y[i] = (d ^ pre) ^ k[i]
			if pre != 0 {
				y[i] = en.gf.Mul(d, pre)
			}
			y[i] = y[i] ^ k[i]
			pre = x[i]
		}
	}

	out := new(bytes.Buffer)
	for i := 0; i < ls; i++ {
		buf := make([]byte, 8)
		binary.LittleEndian.PutUint64(buf, y[i])
		err := binary.Write(out, binary.LittleEndian, buf[:en.Align])
		if err != nil {
			log.Panicf("convert uint64 to byte error : %v", err)
			return nil
		}
	}

	return out.Bytes()
}

func (en *Encode) GetHashforPoSKey(key []byte, ls int) string {
	lk := len(key) // assume key is aligned
	for i := 0; i < lk%en.Align; i++ {
		key = append(key, 0x00)
	}

	for i := 0; i < ls%en.Align; i++ {
		ls += 1
	}

	lk = len(key)
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

	return en.GetHashString(d)
}

func (en *Encode) CalculateXorWithAddress(addr, s []byte) []byte {
	lk := len(addr) // assume key is aligned
	for i := 0; i < lk%en.Align; i++ {
		addr = append(addr, 0x00)
	}

	ls := len(s)
	for i := 0; i < ls%en.Align; i++ {
		s = append(s, 0x00)
	}

	lk = len(addr)
	ls = len(s)

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
