package main

import (
	"fmt"
	// "math/rand"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

const (
	NUM_VBUCKETS    = 2
	REPORT_INTERVAL = time.Duration(5)
)

type vbProcessingStats map[string]*vbStats

var store atomic.Value
var mu sync.Mutex

type vbStats struct {
	lastSeqNo    uint64
	streamStatus string
}

func updater() {
	var counter uint64
	for {
		for i := 0; i < NUM_VBUCKETS; i++ {
			vbKey := "vb_" + strconv.Itoa(i)
			counter++
			// time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)

			stats := &vbStats{
				lastSeqNo:    counter,
				streamStatus: "running",
			}

			mu.Lock()
			store1 := store.Load().(vbProcessingStats)
			store1[vbKey] = stats
			store.Store(store1)
			mu.Unlock()
		}
	}
}

func reporter(vbKey string) {
	for {
		select {
		case <-time.After(REPORT_INTERVAL * time.Second):
			store1 := store.Load().(vbProcessingStats)
			fmt.Printf("vbKey: %s stats: %v\n", vbKey, store1[vbKey])
		}
	}
}

func main() {
	vbucketProcStats := make(vbProcessingStats)
	store.Store(vbucketProcStats)

	go updater()

	for i := 0; i < NUM_VBUCKETS; i++ {
		vbKey := "vb_" + strconv.Itoa(i)

		go reporter(vbKey)
	}

	time.Sleep(100 * time.Second)
}
