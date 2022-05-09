package simulation

import (
	"bufio"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/junwookheo/bcsos/common/config"
)

type SensorData struct {
	Id          int     `json:"id"`
	Timestamp   int64   `json:"Timestamp"`
	Temperature float64 `json:"Fridge_Temperature"`
	Condition   string  `json:"Temp_Condition"`
}

func CreateBlock() bool {
	return false
}

func LoadRawdata(path string, msg chan string) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Load Raw data from file error : %v", err)
		return
	}

	defer file.Close()
	//defer close(msg)

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	// cnt := 0
	for scanner.Scan() {
		// jsonstr := scanner.Text()
		// var sensordata SensorData
		// json.Unmarshal([]byte(jsonstr), &sensordata)
		// sensordata.Timestamp = time.Now().String()
		// var buffer bytes.Buffer
		// json.NewEncoder(&buffer).Encode(&sensordata)

		// msg <- buffer.String()

		msg <- scanner.Text()

		// TEST CONDEEEEEEEEEEEE
		// if cnt > 20 {
		// 	break
		// }
		// cnt++
	}

	msg <- config.END_TEST
}

var letterRunes = []rune("0123456789 abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func genRandString() string {
	n := rand.Intn(70) + 30
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func LoadRawdataFromRandom(msg chan string) {
	rand.Seed(time.Now().UnixNano())

	for i := 0; i < config.TOTAL_TRANSACTIONS; i++ {
		sensordata := SensorData{Id: i, Timestamp: time.Now().UnixNano(), Temperature: (rand.Float64()*80. - 30.), Condition: genRandString()}
		jstr, err := json.Marshal(&sensordata)
		if err != nil {
			log.Panicf("gen error : %v", err)
			msg <- config.END_TEST
			return
		}

		msg <- string(jstr)
	}

	msg <- config.END_TEST
}
