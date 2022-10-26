package simulation

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/junwookheo/bcsos/common/bitcoin"
	"github.com/junwookheo/bcsos/common/config"
	"github.com/junwookheo/bcsos/common/dbagent"
)

// const PATH_TEST = "../iotdata/IoT_normal_fridge_1.log"
const PATH_TEST = "../../blocks.json"
const DB_PATH_TEST = "../bc_dummy.db"

const SIZE_BH = 80

func TestLoadBlockFromJson(t *testing.T) {
	f, err := os.Open(PATH_TEST)
	if err != nil {
		log.Printf("Open error : %v", err)
		return
	}

	scanner := bufio.NewScanner(f)
	const maxCapacity = 16 * 1025 * 1024
	buf := make([]byte, maxCapacity)
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		var rec map[string]interface{}
		err := json.Unmarshal([]byte(scanner.Text()), &rec)
		if err != nil {
			panic(err)
		}

		// var b map[string]interface{}
		b, ok := rec["data"].(map[string]interface{})
		if !ok {
			log.Panicln("Read block type error")
		}

		for key := range b {
			block := bitcoin.NewBlock()
			raw := b[key].(map[string]interface{})["raw_block"].(string)
			rb := bitcoin.NewRawBlock(raw)
			block.Header.Version = rb.ReadUint32()
			block.Header.PreHash = rb.ReverseBuf(rb.ReadBytes(32))
			block.Header.MerkelRoot = rb.ReverseBuf(rb.ReadBytes(32))
			block.Header.Timestamp = rb.ReadUint32()
			block.Header.Difficulty = rb.ReadUint32()
			block.Header.Nonce = rb.ReadUint32()
			log.Printf("Version : %x", block.Header)
			block.SetHash(rb.GetRawBytes(0, 80))
			log.Printf("Version : %v", block.GetHashString())

			txcount := rb.ReadVariant()
			log.Printf("Tx Count : %d", txcount)

			var trs []bitcoin.TransactionHeader

			for i := 0; i < int(txcount); i++ {
				log.Printf("Tx(%v) =================================", i)
				tr := bitcoin.NewTransaction()
				start := rb.GetPosition()
				version := rb.ReadUint32()
				end := rb.GetPosition()
				tr.AppendBuf(rb.GetRawBytes(start, end-start))

				witflag := rb.ReadOptional(2)
				if witflag != nil && witflag[0] == 0x00 && witflag[1] == 0x01 {
					rb.IncreasePos(2)
					tr.Header.Witness = 1
				}
				start = rb.GetPosition()
				incount := rb.ReadVariant()
				log.Printf("Version : %x", version)
				log.Printf("Witness Flag : %v", witflag)
				log.Printf("Input count : %v", incount)

				for j := 0; j < int(incount); j++ {
					log.Printf("Tx(%v) - Input (%v)=================================", i, j)
					preout := rb.ReadBytes(32)
					index := rb.ReadUint32()
					scrlen := rb.ReadVariant()

					log.Printf("Previous Output : %v", preout)
					log.Printf("Index : %v", index)
					log.Printf("Script Length : %v", scrlen)

					if scrlen > 0 {
						script := rb.ReadBytes(uint32(scrlen))
						log.Printf("Unlock Script : %x", script[0])
					}

					seqnumber := rb.ReadUint32()
					log.Printf("Sequence Number : %v", seqnumber)
				}

				outcnt := rb.ReadVariant()
				log.Printf("Output count : %v", outcnt)
				for k := 0; k < int(outcnt); k++ {
					log.Printf("Tx(%v) - Output (%v)=================================", i, k)
					txvalue := rb.ReadUint64()
					scrlen := rb.ReadVariant()
					script := rb.ReadBytes(uint32(scrlen))
					log.Printf("Transaction Value : %v", txvalue)
					log.Printf("Script Length : %v", scrlen)
					log.Printf("Lock Script : %x", script[0])

				}
				end = rb.GetPosition()
				tr.AppendBuf(rb.GetRawBytes(start, end-start))

				if witflag != nil && witflag[0] == 0x00 && witflag[1] == 0x01 {
					for j := 0; j < int(incount); j++ {
						log.Printf("Tx(%v) - Witness (%v)=================================", i, j)
						witcnt := rb.ReadVariant()
						log.Printf("Witness count : %v", witcnt)
						for l := 0; l < int(witcnt); l++ {
							witlen := rb.ReadVariant()
							log.Printf("Witness length : %v", witlen)
							witness := rb.ReadBytes(uint32(witlen))
							log.Printf("Witness : %v", witness)
						}
					}
				}

				start = rb.GetPosition()
				locktime := rb.ReadUint32()
				end = rb.GetPosition()
				tr.AppendBuf(rb.GetRawBytes(start, end-start))
				log.Printf("Lock Time : %v", locktime)

				tr.SetHash()
				trs = append(trs, tr.Header)

			}

			for n, t := range trs {
				log.Printf("%v : %v - %v", n, t.Witness, t.Hash)
			}

			return
		}
	}
}

func TestLoadBtcData(t *testing.T) {
	msg := make(chan bitcoin.BlockPkt)

	go LoadBtcData(PATH_TEST, msg)

	i := 0
	for {
		d, ok := <-msg
		if !ok {
			log.Println("Channle closed")
			break
		}

		if i > 0 {
			close(msg)
			return
		}
		i += 1

		if d.Block == config.END_TEST {
			break
		}

		log.Printf("Block : %v", d.Block[:8])
	}
	close(msg)
}

func TestDBAgent(t *testing.T) {
	dba := dbagent.NewDBAgent(DB_PATH_TEST)
	dba.ShowAllObjets()
	dba.GetLatestBlockHash()
	status := dba.GetDBStatus()
	log.Printf("DB Status : %v", status)
}

func TestDBAgentTime(t *testing.T) {
	ts := time.Now()
	ts1 := time.Now().UnixNano()
	ts2 := time.Unix(ts1/1000000000, ts1%1000000000)

	fmt.Println(ts.String())
	fmt.Println(ts2.String())
	fmt.Println("yyyy-mm-dd HH:mm:ss: ", ts2.Format("2006-01-02 15:04:05.000000"))
}
