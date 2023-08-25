package simulation

import (
	"bufio"
	"encoding/json"
	"log"
	"os"
	"strings"

	"github.com/junwookheo/bcsos/common/blockdata"
	"github.com/junwookheo/bcsos/common/config"
)

func LoadBtcData(path string, msg chan blockdata.BlockPkt) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Load Raw data from file error : %v", err)
		msg <- blockdata.BlockPkt{Timestamp: 0, Block: config.END_TEST}
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	const maxCapacity = 16 * 1025 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	i := 0

	for scanner.Scan() {
		select {
		case <-msg:
			log.Println("Channel closed go rutine")
			return
		default:
			var rec map[string]interface{}
			err := json.Unmarshal([]byte(scanner.Text()), &rec)
			if err != nil {
				log.Panicln(err)
				msg <- blockdata.BlockPkt{Timestamp: 0, Hash: "", Block: config.END_TEST}
			}

			// var b map[string]interface{}
			b, ok := rec["data"].(map[string]interface{})
			if !ok {
				log.Panicln("Read block type error")
				msg <- blockdata.BlockPkt{Timestamp: 0, Hash: "", Block: config.END_TEST}
			}

			for key := range b {
				raw := b[key].(map[string]interface{})["raw_block"].(string)
				msg <- blockdata.BlockPkt{Timestamp: 0, Hash: "", Block: raw}
			}
			log.Printf("Num blocks : %v", i)
			i++
		}
	}

	msg <- blockdata.BlockPkt{Timestamp: 0, Hash: "", Block: config.END_TEST}
	log.Println("LoadBtcData End")
}

func LoadEthData(path string, msg chan blockdata.BlockPkt) {
	file, err := os.Open(path)
	if err != nil {
		log.Printf("Load Raw data from file error : %v", err)
		msg <- blockdata.BlockPkt{Timestamp: 0, Hash: "", Block: config.END_TEST}
		return
	}

	defer file.Close()

	scanner := bufio.NewScanner(file)
	const maxCapacity = 16 * 1025 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	i := 0

	for scanner.Scan() {
		select {
		case <-msg:
			log.Println("Channel closed go rutine")
			return
		default:
			var rec map[string]interface{}
			err := json.Unmarshal([]byte(scanner.Text()), &rec)
			if err != nil {
				log.Panicln(err)
				msg <- blockdata.BlockPkt{Timestamp: 0, Hash: "", Block: config.END_TEST}
			}

			raw := strings.TrimPrefix(rec["result"].(string), "0x")
			hash := strings.TrimPrefix(rec["id"].(string), "0x")
			msg <- blockdata.BlockPkt{Timestamp: 0, Hash: hash, Block: raw}
			log.Printf("Num blocks : %v", i)
			i++
		}
	}

	msg <- blockdata.BlockPkt{Timestamp: 0, Hash: "", Block: config.END_TEST}
	log.Println("LoadBtcData End")
}
