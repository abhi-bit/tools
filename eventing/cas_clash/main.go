package main

import (
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/couchbase/eventing/producer"
	"github.com/couchbase/go-couchbase"
)

const (
	WorkerCount   = 2
	DocCount      = 1
	SleepDuration = time.Second * time.Duration(1)
)

type eventingMetaData struct {
	Name    string `json:"name"`
	Counter string `json:"counter"`
}

var getsOpCallback = func(args ...interface{}) error {
	b := args[0].(*couchbase.Bucket)
	cas := args[1].(*uint64)
	keyName := args[2].(string)
	value := args[3].(*eventingMetaData)

	err := b.Gets(keyName, value, cas)
	if err != nil {
		log.Printf("gets err: %v\n", err)
	} else {
		log.Println("Gets success")
	}
	return err
}

var casOpCallback = func(args ...interface{}) error {
	b := args[0].(*couchbase.Bucket)
	cas := args[1].(*uint64)
	keyName := args[2].(string)
	value := args[3].(*eventingMetaData)

	_, err := b.Cas(keyName, 0, *cas, value)
	if err != nil {
		log.Printf("cas err: %v\n", err)
	} else {
		log.Println("Cas success")
	}
	return err
}

func doBucketOps(workerId int, b *couchbase.Bucket) {
	var counter uint64

	keyName := fmt.Sprintf("doc_%d", rand.Intn(DocCount))
	value := &eventingMetaData{
		Name:    keyName,
		Counter: fmt.Sprintf("worker%d_%d", workerId, counter),
	}
	b.Set(keyName, 0, value)
	for {
		counter++
		keyName = fmt.Sprintf("doc_%d", rand.Intn(DocCount))

		var cas uint64

		// producer.Retry(producer.NewFixedBackoff(time.Second), getsOpCallback, b, &cas, keyName, value)
		producer.Retry(producer.NewExponentialBackoff(), getsOpCallback, b, &cas, keyName, value)

		value.Counter = fmt.Sprintf("worker%d_%d", workerId, counter)

		// producer.Retry(producer.NewFixedBackoff(time.Second), casOpCallback, b, &cas, keyName, value)
		producer.Retry(producer.NewExponentialBackoff(), casOpCallback, b, &cas, keyName, value)
	}
}

func main() {
	c, err := couchbase.Connect("http://donut:8091")
	if err != nil {
		log.Fatal(err)
	}

	p, err := c.GetPool("default")
	if err != nil {
		log.Fatal(err)
	}

	b, err := p.GetBucket("eventing")
	if err != nil {
		log.Fatal(err)
	}

	var wg sync.WaitGroup
	for i := 0; i < WorkerCount; i++ {
		wg.Add(1)
		go doBucketOps(i, b)
	}

	wg.Wait()
}
