package main

import (
	"fmt"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const (
	CONSUMER_COUNT      = 2
	VBUCKET_COUNT       = 10
	STATS_DUMP_DURATION = time.Duration(5) * time.Second
)

type vbProcessingStats map[string]*vbStat

type vbStat struct {
	stats map[string]interface{}
	sync.RWMutex
}

func New() vbProcessingStats {
	vbStats := make(vbProcessingStats, VBUCKET_COUNT)
	for i := 0; i < VBUCKET_COUNT; i++ {
		vbStats[fmt.Sprintf("%d", i)] = &vbStat{
			stats: make(map[string]interface{}),
		}
	}
	return vbStats
}

func (vbs vbProcessingStats) getVBStatDump(vb string) *vbStat {
	return vbs[vb]
}

func (vbs vbProcessingStats) getVBSpecificStatDump(vb, statName string) interface{} {
	vbstat := vbs[vb]
	vbstat.RLock()
	defer vbstat.RUnlock()
	return vbstat.stats[statName]
}

func (vbs vbProcessingStats) updateVBSpecificStat(vb, statName string, data interface{}) {
	vbstat := vbs[vb]
	vbstat.Lock()
	defer vbstat.Unlock()
	vbstat.stats[statName] = data
}

func updater(vbs *vbProcessingStats) {
	var counter uint64
	for {
		counter++

		vbId := rand.Intn(VBUCKET_COUNT)
		vbName := strconv.Itoa(vbId)

		vbs.updateVBSpecificStat(vbName, "last_seq_no", counter)
	}
}

func reader(vbs *vbProcessingStats) {
	for {
		select {
		case <-time.After(STATS_DUMP_DURATION):
			for i := 0; i < VBUCKET_COUNT; i++ {
				vbName := strconv.Itoa(i)
				seqNo := vbs.getVBSpecificStatDump(vbName, "last_seq_no")

				fmt.Printf("vb name: %s seq no: %d\n",
					vbName, seqNo.(uint64))
			}
		}
	}
}

func main() {
	vbs := New()

	go updater(&vbs)

	for i := 0; i < CONSUMER_COUNT; i++ {
		go reader(&vbs)
	}

	time.Sleep(100 * time.Second)
}
