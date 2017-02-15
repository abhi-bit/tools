package main

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/couchbase/cbauth"
	"github.com/couchbase/cbauth/metakv"
)

type TestBlob struct {
	A int
}

func metakvFetch(path string, v interface{}) (bool, error) {
	value, _, err := metakv.Get(path)
	if err != nil {
		fmt.Printf("Path: %s Get failed, err: %v\n", path, err)
		return false, nil
	}

	if value == nil {
		return false, err
	}

	err = json.Unmarshal(value, v)
	if err != nil {
		fmt.Printf("Path: %s Failed to unmarshal val, err: %v\n", path, err)
	}

	return true, nil
}

func metakvUpsert(path string, v interface{}) error {
	value, err := json.Marshal(v)
	if err != nil {
		fmt.Printf("Path: %s Failed to unmarshal, err: %v\n", path, err)
		return err
	}

	err = metakv.Set(path, value, nil)
	if err != nil {
		fmt.Printf("Path: %s failed to perform metakv set, err: %v\n", path, err)
	}
	return err
}

func metakvDel(path string) error {
	err := metakv.Delete(path, nil)
	if err != nil {
		fmt.Printf("Path: %s failed to perform metakv delete, err: %v", path, err)
	}
	return err
}

func metakvRecursiveDel(dirpath string) error {
	err := metakv.RecursiveDelete(dirpath)
	if err != nil {
		fmt.Printf("dirpath: %s failed to perform recursive delete, err: %v\n", err)
	}
	return err
}

var cancelCh chan struct{}

func metaKVCallback(path string, value []byte, rev interface{}) error {
	fmt.Printf("Noticed change in path: %s\n", path)
	return nil
}

func main() {
	_, err := cbauth.InternalRetryDefaultInit("http://127.0.0.1:9000", "Administrator", "asdasd")
	if err != nil {
		fmt.Printf("Failed to init abauth, err: %v\n", err)
		return
	}
	cancelCh = make(chan struct{}, 1)

	go metakv.RunObserveChildren("/", metaKVCallback, cancelCh)

	for i := 0; i < 2; i++ {
		path := "/A/" + strconv.Itoa(i)
		value := &TestBlob{A: i}

		err = metakvUpsert(path, value)
		if err != nil {
			fmt.Printf("Path: %s Upsert failed, err: %v\n", path, err)
			continue
		}
		var v interface{}
		metakvFetch(path, &v)
		fmt.Printf("Fetched from metakv path: %s value: %#v\n", path, v)
	}

	entries, err := metakv.ListAllChildren("/")
	if err != nil {
		fmt.Printf("Failed to list children, err: %v\n", err)
	} else {
		fmt.Printf("Children:\n")
		for _, entry := range entries {
			fmt.Printf("\tK: %s\n\tV: %s\n", entry.Path, string(entry.Value))
		}
	}

	metakvRecursiveDel("/A/")

	path := "/A/1"
	var v interface{}
	metakvFetch(path, &v)
	fmt.Printf("Fetched from metakv path: %s value: %#v\n", path, v)
}
