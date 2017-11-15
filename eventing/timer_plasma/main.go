package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/couchbase/plasma"
)

var tsLayout = "2006-01-02T15:04:05Z"
var numVbuckets = 1024

type byTimerEntry struct {
	DocID      string
	CallbackFn string
}

func main() {
	os.RemoveAll("teststore.data")

	cfg := plasma.DefaultConfig()
	cfg.File = "teststore.data"
	cfg.AutoLSSCleaning = false

	s, err := plasma.New(cfg)
	if err != nil {
		panic(err)
	}

	numTsToInsert, _ := strconv.Atoi(os.Args[1])
	entriesPerTsPerVb, _ := strconv.Atoi(os.Args[2])

	var appName, callbackFn string
	if len(os.Args) > 3 {
		appName = os.Args[3]
	} else {
		appName = "bucket_op_with_doc_timer.js"
	}

	if len(os.Args) > 4 {
		callbackFn = os.Args[4]
	} else {
		callbackFn = "timerCallback"
	}

	fmt.Printf("numTsToInsert: %v numVbuckets: %v, entriesPerTsPerVb: %v\n",
		numTsToInsert, numVbuckets, entriesPerTsPerVb)

	w := s.NewWriter()
	start := time.Now()
	// Insert in byId
	for i := 0; i < numTsToInsert; i++ {
		for j := 0; j < numVbuckets; j++ {
			for k := 0; k < entriesPerTsPerVb; k++ {
				key := fmt.Sprintf("vb_%d::%v::%s::%s::doc_id_%d", j, appName, start.Format(tsLayout), callbackFn, k)
				w.InsertKV([]byte(key), []byte(""))

			}
			var timerVal string
			for l := 0; l < entriesPerTsPerVb; l++ {
				entry := byTimerEntry{
					DocID:      fmt.Sprintf("doc_id_%d", l),
					CallbackFn: callbackFn,
				}
				data, _ := json.Marshal(&entry)

				if l == 0 {
					timerVal = string(data)
				} else {
					timerVal = fmt.Sprintf("%v,%v", timerVal, string(data))
				}

				byTimerKey := fmt.Sprintf("vb_%v::%s::%s", j, appName, start.Format(tsLayout))
				w.InsertKV([]byte(byTimerKey), []byte(timerVal))
			}
		}
		start = start.Add(time.Second)
	}

	// Insert in byTimer
	fmt.Printf("start: %v\n", start.Format(tsLayout))

	// w := s.NewWriter()
	// for j := 0; j < 5; j++ {
	// 	for i := 0; i < 5; i++ {
	// 		w.InsertKV([]byte(fmt.Sprintf("key-%10d", i)), []byte(fmt.Sprintf("val-%10d", i)))
	// 	}
	// }
	s.PersistAll()

	// rdr := s.NewReader()
	// snapshot := s.NewSnapshot()

	// itr, _ := rdr.NewSnapshotIterator(snapshot)
	// for itr.SeekFirst(); itr.Valid(); itr.Next() {
	// 	fmt.Println("1st key: ", string(itr.Key()), " value: ", string(itr.Value()))
	// 	fmt.Println("2nd key: ", string(itr.Key()), " value: ", string(itr.Value()))
	// }
	// itr.Close()

	s.Close()

}
