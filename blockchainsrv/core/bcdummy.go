package core

import (
	"bufio"
	"log"
	"os"
)

type SensorData struct {
	Id          int     `json:"id"`
	Timestamp   string  `json:"Timestamp"`
	Temperature float32 `json:"Fridge_Temperature"`
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

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		msg <- scanner.Text()
	}
}
