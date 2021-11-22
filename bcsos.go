package main

import (
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/junwookheo/bcsos/common/config"
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

	// fileServer := http.FileServer(http.Dir("./blockchainsim/static"))
	// http.Handle("/static/", http.StripPrefix("/static", fileServer))
	// http.Handle("/", fileServer)
	// err := http.ListenAndServe(":5050", nil)
	// if err != nil {
	// 	log.Fatal("Error Starting the HTTP Server :", err)
	// 	return
	// }

	rand.Seed(time.Now().UnixNano())
	for k := 0; k < 100; k++ {
		func() {
			w := config.BASIC_UNIT_TIME * config.RATE_TSC

			ids := func(w int, num int) string {
				ids := []string{}
				for i := 0; i < num; i++ {
					f := rand.ExpFloat64() / float64(config.LAMBDA_ED)
					l := int(f * float64(w))
					ids = append(ids, strconv.Itoa(l))
				}
				return strings.Join(ids, ", ")
			}(w, 10)
			log.Println("EXP :", ids)
		}()
		func() {
			w := config.TOTAL_TRANSACTIONS + config.TOTAL_TRANSACTIONS/config.NUM_TRANSACTION_BLOCK

			ids := func(w int, num int) string {
				ids := []string{}
				for i := 0; i < num; i++ {
					l := rand.Intn(int(w))
					ids = append(ids, strconv.Itoa(l))
				}
				return strings.Join(ids, ", ")
			}(w, 10)

			log.Println("RAND", ids)
		}()
	}
}
