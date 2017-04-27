package main

import (
	"encoding/json"
	"fmt"

	"github.com/couchbase/go-couchbase"
)

func main() {
	connStr := fmt.Sprintf("http://127.0.0.1:9000")

	conn, err := couchbase.ConnectWithAuthCreds(connStr, "eventing", "asdasd")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	pool, err := conn.GetPool("default")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}

	bucketHandle, err := pool.GetBucketWithAuth("default", "eventing", "asdasd")
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Println("auth done")

	data := make(map[string]interface{})
	data["city"] = "BLR"
	data["pincode"] = 560001

	content, err := json.Marshal(&data)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Println("marshal done")

	err = bucketHandle.Set("key", 10000, data)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Println("bucket set done")

	err = bucketHandle.SetRaw("key", 10000, content)
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	fmt.Println("bucket setraw done")

}
