package network

import (
	"log"
	"testing"
)

func TestXOR(t *testing.T) {
	d1 := xordistance("aaf0aa63ef4fb4d8933524ebcfd97c33e7c2cd7c31ccbda894aa792a7d63cc95", "1e78c84fce7b59fe96f8267989cb96a1ebd70dc83111eecfd73b4c21ab5d84e3")
	d2 := xordistance("aaf0aa63ef4fb4d8933524ebcfd97c33e7c2cd7c31ccbda894aa792a7d63cc95", "ac15c0aa170e705b104dd2b1000f8650a4255f6b813cc9559843a3d580454048")
	log.Printf("distance1 : %v", d1)
	log.Printf("distance2 : %v", d2)
	log.Printf("d1 - d2 : %v", d2.Cmp(d1))
}
