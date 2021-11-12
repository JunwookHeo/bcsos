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

	for i := 0; i < 100; i++ {
		f := rand.ExpFloat64()/0.25
		in := int(f)
		log.Printf("%d, %d", i, in)
	}

}
