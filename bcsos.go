package main

import (
	"fmt"
	"log"

	"github.com/junwookheo/bcsos/common/wallet"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main() {
	log.Print("Print log")
	wallet.ValidateAddress([]byte("5J3mBbAH58CpQ3Y5RNJpUKPE62SQ5tfcvU2JpbnkeyhfsYB1Jcn"))

	//a := [][]int{{1, 9, 5}, {4, 5, 6}, {7, 8, 9}}
	var b [5]int

	a := make([][]int, 4)
	for i := 0; i < 4; i++ {
		a[i] = make([]int, 5)
		for j := 0; j < 5; j++ {
			a[i][j] = j + i*j
		}
	}


	fmt.Println("A:", a)
	fmt.Println("B:", b)

	for i, c := range a[1]{
		b[i] = c
	}

	fmt.Println("A:", a)
	fmt.Println("B:", b)
}
