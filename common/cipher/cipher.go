package cipher

import (
	"crypto/sha256"
	"encoding/hex"
)

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
