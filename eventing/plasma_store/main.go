package main

import (
	"fmt"
	"os"

	"github.com/couchbase/nitro/plasma"
)

func main() {
	os.RemoveAll("teststore.data")
	cfg := plasma.DefaultConfig()
	cfg.File = "teststore.data"

	s, err := plasma.New(cfg)
	if err != nil {
		panic(err)
	}
	defer s.Close()

	w := s.NewWriter()
	for i := 0; i < 10000; i++ {
		w.InsertKV([]byte(fmt.Sprintf("key-%10d", i)), []byte(fmt.Sprintf("val-%10d", i)))
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
	fmt.Printf("Val: %v\n", string(v))
}
