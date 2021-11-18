package main

import (
	"log"
	"math/rand"

	"github.com/junwookheo/bcsos/common/wallet"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	log.Print("Print log")
	wallet.ValidateAddress([]byte("5J3mBbAH58CpQ3Y5RNJpUKPE62SQ5tfcvU2JpbnkeyhfsYB1Jcn"))

	maxn := 0
	for i := 0; i < 10000000000; i++ {
		f := rand.ExpFloat64() / 0.1 * 600
		if int(f) > maxn {
			log.Println(int(f))
			maxn = int(f)
		}
	}
}

func BubbleSort(data []int) {
	for i := 0; i < len(data); i++ {
		for j := 1; j < len(data)-i; j++ {
			if data[j] < data[j-1] {
				data[j], data[j-1] = data[j-1], data[j]
			}
		}
	}
	log.Println(data)
}
