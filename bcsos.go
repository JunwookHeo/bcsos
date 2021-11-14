package main

import (
	"fmt"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/junwookheo/bcsos/common/wallet"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	log.Print("Print log")
	wallet.ValidateAddress([]byte("5J3mBbAH58CpQ3Y5RNJpUKPE62SQ5tfcvU2JpbnkeyhfsYB1Jcn"))

	samples := 10
	cnt1 := 0
	cnt2 := 0
	rand.Seed(time.Now().Unix())
	for i := 0; i < samples; i++ {
		f := rand.ExpFloat64() / 0.1
		l := int(f) * 60 * 2
		if l < 828 {
			cnt1++
		} else {
			cnt2++
		}
		log.Printf("%d, %d", i, l)
	}
	log.Printf("%d - %d", cnt1, cnt2)

	var IDs []string
	for _, i := range []int{1, 2, 3, 4} {
		IDs = append(IDs, strconv.Itoa(i))
	}

	fmt.Println(strings.Join(IDs, ", "))
}
