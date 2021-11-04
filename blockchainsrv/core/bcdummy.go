package core

import (
	"bufio"
	"encoding/json"
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

func LoadRawdata() {
	file, err := os.Open("../iotdata/IoT_normal_fridge_1.log")
	if err != nil {
		log.Printf("Load Raw data from file error : %v", err)
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)
	i := 0

	for scanner.Scan() {
		var sensordata SensorData
		json.Unmarshal([]byte(scanner.Text()), &sensordata)

		// log.Printf("ID : %v", sensordata.Id)
		// log.Printf("Timestamp : %v", sensordata.Timestamp)
		// log.Printf("Temperature : %v", sensordata.Temperature)
		// log.Printf("Condition : %v", sensordata.Condition)
		// if i > 10 {
		// 	break
		// }
		i += 1
	}
	log.Printf("leng : %v", i)

}
