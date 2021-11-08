package main

import (
	"log"
	"time"

	"github.com/junwookheo/bcsos/common/wallet"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

func main2() {
	log.Print("Print log")
	wallet.ValidateAddress([]byte("5J3mBbAH58CpQ3Y5RNJpUKPE62SQ5tfcvU2JpbnkeyhfsYB1Jcn"))
}

func routine1(channel chan string) {
	// 채널에 값을 넣습니다.
	log.Println("sending channel")
	channel <- "data"
	log.Println("sent channel")
}
func routine2(channel chan string) {
	// 채널로부터 값을 꺼내서 출력합니다.
	log.Println("receiving channel")
	time.Sleep(10 * time.Second)
	log.Println(<-channel) // 채널에 값이 들어올때까지 대기
	log.Println("received channel")
}
func main() {
	// string 채널을 위한 메모리를 할당합니다.
	channel := make(chan string)
	go routine1(channel)
	go routine2(channel)
	select {}
}
