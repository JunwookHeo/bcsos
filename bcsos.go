package main

import (
	"fmt"
	"log"
	"time"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

type Student struct {
	Name       string
	College    string
	RollNumber int
}

// func renderTemplate(w http.ResponseWriter, r *http.Request) {
// 	student := Student{
// 		Name:       "GB",
// 		College:    "GolangBlogs",
// 		RollNumber: 1,
// 	}
// 	parsedTemplate, _ := template.ParseFiles("./blockchainsim/static/index.html")
// 	err := parsedTemplate.Execute(w, student)
// 	if err != nil {
// 		log.Println("Error executing template :", err)
// 		return
// 	}
// }
type timetest struct {
	ts time.Time
}

func main() {
	log.Print("Print log")
	// wallet.ValidateAddress([]byte("5J3mBbAH58CpQ3Y5RNJpUKPE62SQ5tfcvU2JpbnkeyhfsYB1Jcn"))

	// fileServer := http.FileServer(http.Dir("./blockchainsim/static"))
	// http.Handle("/static/", http.StripPrefix("/static", fileServer))
	// http.Handle("/", fileServer)
	// err := http.ListenAndServe(":5050", nil)
	// if err != nil {
	// 	log.Fatal("Error Starting the HTTP Server :", err)
	// 	return
	// }

	ts := timetest{time.Now()}

	for i := 0; i < 5; i++ {
		log.Printf("call test : %v", i)
		ts.ts = time.Now()
		go ts.time_test(fmt.Sprintf("test # %v", i))
		time.Sleep(300 * time.Millisecond)
	}
	time.Sleep(3 * time.Second)
}

func (t *timetest) time_test(title string) {
	log.Printf("test start : %v", title)

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()
	<-ticker.C
	if time.Now().UnixNano()-t.ts.UnixNano() > 1*1000000000 {
		log.Printf("test end: %v", title)
	} else {
		log.Printf("test cancel: %v", title)
	}

}
