package main

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/couchbase/cbauth/metakv"
	"github.com/couchbase/eventing/logging"
)

var counter = 0
var path = "/test/eventing/"
var pathToWrite = "/test/eventing/def"
var mod = 3
var size = 100 * 1000

type payload struct {
	Content []byte `json:"content"`
	Ctr     int    `json:"counter"`
}

func callback(path string, value []byte, rev interface{}) error {
	logging.Infof("===============================")
	defer func() {
		logging.Infof("===============================")
	}()
	logging.Debugf("Path: %s", path)

	data := payload{}
	err := json.Unmarshal(value, &data)
	if err != nil {
		logging.Errorf("Unmarshal error: %v", err)
		return nil
	}
	logging.Infof("GET COUNTER CALLBACK: %d", data.Ctr)

mlookup:
	p, _, err := metakv.Get(pathToWrite)
	if err != nil {
		logging.Errorf("Metakv get err: %v", err)
		time.Sleep(time.Second)
		goto mlookup
	}
	err = json.Unmarshal(p, &data)
	if err != nil {
		logging.Errorf("Unmarshal error: %v", err)
		time.Sleep(time.Second)
		goto mlookup
	}
	logging.Infof("GET COUNTER LOOKUP: %d", data.Ctr)
	if data.Ctr != counter%mod {
		time.Sleep(time.Second)
		goto mlookup
	}

	return nil
}

func main() {

	cancelCh := make(chan struct{})
	go func(cancelCh chan struct{}) {
		for {
			err := metakv.RunObserveChildren(path, callback, cancelCh)
			if err != nil {
				time.Sleep(time.Second)
			}
		}
	}(cancelCh)

	data := &payload{
		Content: []byte("abc"),
		Ctr:     counter,
	}
	p, _ := json.Marshal(data)

	err := metakv.Set(pathToWrite, p, nil)
	if err != nil {
		logging.Fatalf("First metakv set, err: %v", err)
	}
	p, _, err = metakv.Get(pathToWrite)
	if err != nil {
		logging.Errorf("Metakv get err: %v", err)
	}

	var abc payload
	err = json.Unmarshal(p, &abc)
	logging.Infof("MAIN LOOP CTR: %d", abc.Ctr)

	tick := time.NewTicker(2 * time.Second)

	for {
		select {
		case <-tick.C:
			data := payload{}
			data.Content = make([]byte, size)
			n, err := rand.Read(data.Content)
			if err != nil {
				logging.Errorf("Rand read err: %v", err)
				return
			}
			logging.Debugf("Read: %d bytes", n)
			counter++
			data.Ctr = counter % mod

			p, err = json.Marshal(&data)
			if err != nil {
				logging.Errorf("Marshal err: %v", err)
				return
			}

			err = metakv.Set(pathToWrite, p, nil)
			if err != nil {
				logging.Errorf("Metakv set err: %v", err)
				return
			}

			logging.Infof("\n\nSTORED COUNTER: %d\n\n", data.Ctr)

			/*p, _, err = metakv.Get(pathToWrite)
			if err != nil {
				logging.Errorf("Metakv get err: %v", err)
			}

			var abc payload
			err = json.Unmarshal(p, &abc)
			logging.Infof("MAIN LOOP CTR: %d", abc.Ctr)*/
		}
	}

	// time.Sleep(10000 * time.Second)
}
