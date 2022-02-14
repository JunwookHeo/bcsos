package main

import (
	"log"
	"sync"
	"time"

	"github.com/junwookheo/bcsos/common/listener"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

var i int

func work() {
	time.Sleep(250 * time.Millisecond)
	i++
	log.Println(i)
}

func routine(command <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var status = "Play"
	for {
		select {
		case cmd := <-command:
			log.Println(cmd)
			switch cmd {
			case "Stop":
				return
			case "Pause":
				status = "Pause"
			default:
				status = "Play"
			}
		default:
			if status == "Play" {
				work()
			}
		}
	}
}

func routine2(command <-chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	var status = "Play"
	for {
		select {
		case cmd := <-command:
			log.Println("222", cmd)
			switch cmd {
			case "Stop":
				return
			case "Pause":
				status = "Pause"
			default:
				status = "Play"
			}
		default:
			if status == "Play" {
				work()
			}
		}
	}
}

func main() {
	l := listener.EventListener{}
	var wg sync.WaitGroup
	wg.Add(1)
	command := make(chan string)
	wg.Add(1)
	command2 := make(chan string)
	l.AddListener(command)
	l.AddListener(command2)

	go routine(command, &wg)
	go routine2(command2, &wg)

	time.Sleep(1 * time.Second)
	//command <- "Pause"
	l.Notify("Pause")

	time.Sleep(1 * time.Second)
	//command <- "Play"
	l.Notify("Play")

	time.Sleep(1 * time.Second)
	//command <- "Stop"
	l.Notify("Stop")

	wg.Wait()
}
