package main

import (
	"fmt"
	"os"
	// "time"

	"github.com/couchbase/nitro/plasma"
)

func main() {
	os.RemoveAll("teststore.data")
	cfg := plasma.DefaultConfig()
	cfg.File = "teststore.data"
	cfg.AutoLSSCleaning = false
	fmt.Printf("cfg dump: %#v\n", cfg)

	s, err := plasma.New(cfg)
	if err != nil {
		panic(err)
	}

	w := s.NewWriter()
	for j := 0; j < 5; j++ {
		for i := 0; i < 5; i++ {
			w.InsertKV([]byte(fmt.Sprintf("key-%10d", i)), []byte(fmt.Sprintf("val-%10d", i)))
		}
	}

	k := []byte(fmt.Sprintf("key-%10d", 1000))
	w.InsertKV(k, []byte(fmt.Sprintf("%d", 1)))
	w.InsertKV(k, []byte(fmt.Sprintf("%d", 2)))
	w.InsertKV(k, []byte(fmt.Sprintf("%d", 3)))
	w.InsertKV(k, []byte(fmt.Sprintf("%d", 4)))
	w.InsertKV(k, []byte(fmt.Sprintf("%d", 5)))

	v, err := w.LookupKV(k)
	if err != nil {
		panic(err)
	}
	s.PersistAll()

	rdr := s.NewReader()
	snapshot := s.NewSnapshot()

	itr := rdr.NewSnapshotIterator(snapshot)
	for itr.SeekFirst(); itr.Valid(); itr.Next() {
		fmt.Println("1st key: ", string(itr.Key()), " value: ", string(itr.Value()))
		fmt.Println("2nd key: ", string(itr.Key()), " value: ", string(itr.Value()))
	}
	itr.Close()

	// time.Sleep(10 * time.Second)
	fmt.Printf("Val: %v\n", string(v))
	s.Close()

	// time.Sleep(10 * time.Second)
}
