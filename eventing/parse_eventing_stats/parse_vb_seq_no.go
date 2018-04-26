package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
)

func main() {
	if len(os.Args) < 1 {
		fmt.Println("Insufficient arguments provided")
		return
	}

	filename := os.Args[1]
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("Readfile err:", err)
		return
	}

	var statsEntries []interface{}
	err = json.Unmarshal(data, &statsEntries)
	if err != nil {
		fmt.Println("Unmarshal err:", err)
		return
	}

	for _, val := range statsEntries {

		entry := val.(map[string]interface{})

		for entryKey, entryVal := range entry {

			if entryKey == "vb_seq_no_stats" {

				vbSeqNoStats := entryVal.(map[string]interface{})

				var vbs []int
				for vb := range vbSeqNoStats {
					vbNo, _ := strconv.Atoi(vb)
					vbs = append(vbs, vbNo)
				}
				sort.Ints(vbs)

				for _, vb := range vbs {
					fmt.Printf("\nvb: %d\n", vb)

					seqStats := vbSeqNoStats[fmt.Sprintf("%d", vb)].([]interface{})
					for _, entry := range seqStats {

						vbStats := entry.(map[string]interface{})

						startSeqNo := vbStats["start_seq_no"]
						seqCloseStream := vbStats["seq_no_after_close_stream"]
						seqStreamEnd := vbStats["seq_no_at_stream_end"]
						checkpointSeqNo := vbStats["last_checkpointed_seq_no"]
						hostName := vbStats["host_name"]
						workerName := vbStats["worker_name"]

						fmt.Printf("SSN: %-10vSCS: %-10vSSE: %-10vCSN: %-10vHN: %-30vWN: %-40v\n",
							startSeqNo, seqCloseStream, seqStreamEnd, checkpointSeqNo,
							hostName, workerName)
					}
				}
			}

		}
	}
}
