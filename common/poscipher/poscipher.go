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

// Binary Field
const ALIGN = 4
const Ix2 = 2147483648 //6442450943  // (3P-2)/2 Mod (P-1)
var GF = galois.GFN(32)

// Prime Field
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

	// pre := uint32(0)
	pre := uint32(1)
	if lk < ls {
		for i := 0; i < ls; i++ {
			// d := (x[i] ^ k[i%lk]) ^ pre
			d := (x[i] ^ k[i%lk])
			d = uint32(GF.Div(uint64(d), uint64(pre)))
			y[i] = uint32(GF.Exp(uint64(d), Ix2))

			if y[i] == 0 {
				pre = uint32(1)
			} else {
				pre = y[i]
			}
		}
	} else if lk > ls {
		for i := 0; i < ls; i++ {
			// y[i] = k[i] ^ x[i]
			d := (x[i] ^ k[i]) ^ pre
			y[i] = uint32(GF.Exp(uint64(d), Ix2))
			pre = y[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			// y[i] = k[i] ^ x[i]
			d := (x[i] ^ k[i]) ^ pre
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
			if pre == 0 {
				y[i] = uint32(GF.Mul(uint64(d), uint64(1)))
			} else {
				y[i] = uint32(GF.Mul(uint64(d), uint64(pre)))
			}

			y[i] = y[i] ^ k[i%lk]
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
		d[i] = s[i] ^ addr[i%lk]
	}

	return d
}

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
	ks := GFP.LoadUint256FromKey(key)
	xs := GFP.LoadUint256FromStream31(s)
	lk := len(ks)
	ls := len(xs)

	y := make([]*uint256.Int, ls)
	pre := uint256.NewInt(1)
	if lk < ls {
		for i := 0; i < ls; i++ {
			d := xs[i].Clone()
			if !ks[i%lk].IsZero() {
				d = GFP.Div(xs[i], ks[i%lk])
			}
			if !pre.IsZero() {
				d = GFP.Div(d, pre)
			}

			y[i] = GFP.Exp(d, Ix3)
			pre = y[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			d := xs[i].Clone()
			if !ks[i%lk].IsZero() {
				d = GFP.Div(xs[i], ks[i])
			}
			if !pre.IsZero() {
				d = GFP.Div(d, pre)
			}

			y[i] = GFP.Exp(d, Ix3)
			pre = y[i]
		}
	}

	buf := new(bytes.Buffer)
	for i := 0; i < len(y); i++ {
		err := binary.Write(buf, binary.LittleEndian, y[i].Bytes32())
		if err != nil {
			log.Panicf("convert uint32 to byte error : %v", err)
			return "", nil
		}
	}

	return GetHashString(buf.Bytes()), buf.Bytes()
}

func DecryptPoSWithPrimeField(key, s []byte) []byte {
	ks := GFP.LoadUint256FromKey(key)
	xs := GFP.LoadUint256FromStream32(s)
	lk := len(ks)
	ls := len(xs)

	y := make([]*uint256.Int, ls)
	pre := uint256.NewInt(1)
	if lk < ls {
		for i := 0; i < ls; i++ {
			d := GFP.Exp(xs[i], uint256.NewInt(3))
			// d = GFP.Sub(d, pre)
			if !pre.IsZero() {
				d = GFP.Mul(d, pre)
			}
			if !ks[i%lk].IsZero() {
				y[i] = GFP.Mul(d, ks[i%lk])
			}

			pre = xs[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			d := GFP.Exp(xs[i], uint256.NewInt(3))
			if !pre.IsZero() {
				d = GFP.Mul(d, pre)
			}
			if !ks[i].IsZero() {
				y[i] = GFP.Mul(d, ks[i])
			}

			pre = xs[i]
		}
	}

	buf := new(bytes.Buffer)
	for i := 0; i < len(y); i++ {
		yu := y[i].Bytes32()
		err := binary.Write(buf, binary.LittleEndian, yu[1:])
		if err != nil {
			log.Panicf("convert uint32 to byte error : %v", err)
			return nil
		}
	}

	return buf.Bytes()
}

func EncryptPoSWithPrimeFieldPreKey(key, s []byte) (string, []byte) {
	ks := GFP.LoadUint256FromStream32(key)
	xs := GFP.LoadUint256FromStream31(s)
	lk := len(ks)
	ls := len(xs)

	y := make([]*uint256.Int, ls)
	pre := uint256.NewInt(1)
	if lk < ls {
		for i := 0; i < ls; i++ {
			d := xs[i].Clone()
			if !ks[i%lk].IsZero() {
				d = GFP.Div(xs[i], ks[i%lk])
			}
			if !pre.IsZero() {
				d = GFP.Div(d, pre)
			}

			y[i] = GFP.Exp(d, Ix3)
			pre = y[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			d := xs[i].Clone()
			if !ks[i%lk].IsZero() {
				d = GFP.Div(xs[i], ks[i])
			}
			if !pre.IsZero() {
				d = GFP.Div(d, pre)
			}

			y[i] = GFP.Exp(d, Ix3)
			pre = y[i]
		}
	}

	buf := new(bytes.Buffer)
	for i := 0; i < len(y); i++ {
		err := binary.Write(buf, binary.LittleEndian, y[i].Bytes32())
		if err != nil {
			log.Panicf("convert uint32 to byte error : %v", err)
			return "", nil
		}
	}

	return GetHashString(buf.Bytes()), buf.Bytes()
}

func DecryptPoSWithPrimeFieldPreKey(key, s []byte) []byte {
	ks := GFP.LoadUint256FromStream32(key)
	xs := GFP.LoadUint256FromStream32(s)
	lk := len(ks)
	ls := len(xs)

	y := make([]*uint256.Int, ls)
	pre := uint256.NewInt(1)
	if lk < ls {
		for i := 0; i < ls; i++ {
			d := GFP.Exp(xs[i], uint256.NewInt(3))
			// d = GFP.Sub(d, pre)
			if !pre.IsZero() {
				d = GFP.Mul(d, pre)
			}
			if !ks[i%lk].IsZero() {
				y[i] = GFP.Mul(d, ks[i%lk])
			} else {
				y[i] = d
			}

			pre = xs[i]
		}
	} else {
		for i := 0; i < ls; i++ {
			d := GFP.Exp(xs[i], uint256.NewInt(3))
			if !pre.IsZero() {
				d = GFP.Mul(d, pre)
			}
			if !ks[i].IsZero() {
				y[i] = GFP.Mul(d, ks[i])
			} else {
				y[i] = d
			}

			pre = xs[i]
		}
	}

	buf := new(bytes.Buffer)
	for i := 0; i < len(y); i++ {
		yu := y[i].Bytes32()
		err := binary.Write(buf, binary.LittleEndian, yu[1:])
		if err != nil {
			log.Panicf("convert uint32 to byte error : %v", err)
			return nil
		}
	}

	return buf.Bytes()
}
