package main

import (
	"log"
	"net/http"
	"text/template"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.Lshortfile)
}

type Student struct {
	Name       string
	College    string
	RollNumber int
}

func renderTemplate(w http.ResponseWriter, r *http.Request) {
	student := Student{
		Name:       "GB",
		College:    "GolangBlogs",
		RollNumber: 1,
	}
	parsedTemplate, _ := template.ParseFiles("./blockchainsim/static/index.html")
	err := parsedTemplate.Execute(w, student)
	if err != nil {
		log.Println("Error executing template :", err)
		return
	}
}

func main() {
	log.Print("Print log")
	// wallet.ValidateAddress([]byte("5J3mBbAH58CpQ3Y5RNJpUKPE62SQ5tfcvU2JpbnkeyhfsYB1Jcn"))

	fileServer := http.FileServer(http.Dir("./blockchainsim/static"))
	http.Handle("/static/", http.StripPrefix("/static", fileServer))
	http.Handle("/", fileServer)
	err := http.ListenAndServe(":5050", nil)
	if err != nil {
		log.Fatal("Error Starting the HTTP Server :", err)
		return
	}
}
