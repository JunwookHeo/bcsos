package poscipher

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"log"
	"math/big"

	"github.com/holiman/uint256"
	"github.com/junwookheo/bcsos/common/galois"
)

const ALIGN = 4
const PALIGN = 8

var GF = galois.GFN(32)
var GFP = galois.NewGFP()
var Ix3 = GetInverseX3()

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

const PSIZE = 31 // Prime field with 32byte so 31bytes is max to convert Prime field
func GetInverseX3() *uint256.Int {
	p := GFP.Prime.ToBig()
	p = p.Mul(p, big.NewInt(2))
	p = p.Sub(p, big.NewInt(1))
	p = p.Div(p, big.NewInt(3))
	p = p.Mod(p, GFP.Prime.ToBig())
	I, _ := uint256.FromBig(p)
	return I
}

func EncryptPoSWithPrimeField(key, s []byte) (string, []byte) {
	lk := len(key) // assume key is aligned
	if lk%PALIGN != 0 {
		log.Panicf("Error Key length ALIGN: %v", lk)
	}
	ls := len(s)
	lpad := ls % PSIZE
	if lpad != 0 {
		for i := 0; i < PSIZE-lpad; i++ { // Add Padding
			s = append(s, 0x0)
		}
	}

	ls = len(s)
	if ls%PSIZE != 0 {
		log.Panicf("Error Source length ALIGN : %v", ls)
	}

	lk /= PALIGN // 4bytes of key is used for encryption
	ls /= PSIZE

	key_reader := bytes.NewReader(key)
	s_reader := bytes.NewReader(s)

	k := make([]uint64, lk)
	err := binary.Read(key_reader, binary.LittleEndian, &k)
	if err != nil {
		log.Panicf("converting key to uint32 error : %v", err)
		return "", nil
	}

	x := make([][PSIZE]byte, ls)
	err = binary.Read(s_reader, binary.LittleEndian, &x)
	if err != nil {
		log.Panicf("converting source to uint32 error : %v", err)
		return "", nil
	}

	y := make([]*uint256.Int, ls)
	pre := uint256.NewInt(1)
	if lk < ls {
		for i := 0; i < ls; i++ {
			xu := uint256.NewInt(0)
			xu.SetBytes31(x[i][:])
			ku := uint256.NewInt(k[i%lk])
			d := GFP.Add(xu, ku)
			// d = GFP.Add(d, pre)
			d = GFP.Div(d, pre)
			y[i] = GFP.Exp(d, Ix3)
			pre = y[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			xu := uint256.NewInt(0)
			xu.SetBytes31(x[i][:])
			ku := uint256.NewInt(k[i])
			d := GFP.Add(xu, ku)
			d = GFP.Add(d, pre)
			y[i] = GFP.Exp(d, Ix3)
			pre = y[i]
		}
	}

	buf := new(bytes.Buffer)
	for i := 0; i < len(y); i++ {
		err = binary.Write(buf, binary.LittleEndian, y[i].Bytes32())
		if err != nil {
			log.Panicf("convert uint32 to byte error : %v", err)
			return "", nil
		}
	}

	return GetHashString(buf.Bytes()), buf.Bytes()
}

func DecryptPoSWithPrimeField(key, s []byte) []byte {
	lk := len(key) // assume key is aligned
	if lk%PALIGN != 0 {
		log.Panicf("Error Key length ALIGN: %v", lk)
	}
	ls := len(s)
	if lk%(PSIZE+1) != 0 {
		log.Panicf("Error Source length ALIGN : %v", ls)
	}

	lk /= PALIGN
	ls /= (PSIZE + 1)

	key_reader := bytes.NewReader(key)
	s_reader := bytes.NewReader(s)

	k := make([]uint64, lk)
	err := binary.Read(key_reader, binary.LittleEndian, &k)
	if err != nil {
		log.Panicf("converting key to uint32 error : %v", err)
		return nil
	}

	x := make([][32]byte, ls)
	err = binary.Read(s_reader, binary.LittleEndian, &x)
	if err != nil {
		log.Panicf("converting source to uint32 error : %v", err)
		return nil
	}

	y := make([]*uint256.Int, ls)
	pre := uint256.NewInt(1)
	if lk < ls {
		for i := 0; i < ls; i++ {
			xu := uint256.NewInt(0)
			xu.SetBytes32(x[i][:])
			ku := uint256.NewInt(k[i%lk])
			d := GFP.Exp(xu, uint256.NewInt(3))
			// d = GFP.Sub(d, pre)
			d = GFP.Mul(d, pre)
			y[i] = GFP.Sub(d, ku)
			pre = xu
		}
	} else {
		for i := 0; i < ls; i++ {
			xu := uint256.NewInt(0)
			xu.SetBytes32(x[i][:])
			ku := uint256.NewInt(k[i%lk])
			d := GFP.Exp(xu, uint256.NewInt(3))
			d = GFP.Sub(d, pre)
			y[i] = GFP.Sub(d, ku)
			pre = xu
		}
	}

	buf := new(bytes.Buffer)
	for i := 0; i < len(y); i++ {
		yu := y[i].Bytes32()
		err = binary.Write(buf, binary.LittleEndian, yu[1:])
		if err != nil {
			log.Panicf("convert uint32 to byte error : %v", err)
			return nil
		}
	}

	return buf.Bytes()
}
