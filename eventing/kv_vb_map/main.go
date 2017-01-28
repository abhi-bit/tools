package main

import (
	"fmt"

	"github.com/couchbase/indexing/secondary/common"
)

const (
	DATA_SERVICE = "kv"
)

func main() {
	clusterURL := "http://Administrator:asdasd@127.0.0.1:9001"

	cinfo, err := common.NewClusterInfoCache(clusterURL, "default")
	if err != nil {
		fmt.Printf("new cluster info cache, err: %s\n", err.Error())
		return
	}

	if err := cinfo.Fetch(); err != nil {
		fmt.Printf("cluster info fetch failed, err: %s\n", err.Error())
		return
	}

	kvAddrs := cinfo.GetNodesByServiceType(DATA_SERVICE)

	kvVbMap := make(map[uint16]string)

	for _, kvaddr := range kvAddrs {
		addr, _ := cinfo.GetServiceAddress(kvaddr, DATA_SERVICE)
		vbs, _ := cinfo.GetVBuckets(kvaddr, "default")
		for i := 0; i < len(vbs); i++ {
			kvVbMap[uint16(vbs[i])] = addr
		}
	}

	for k, v := range kvVbMap {
		fmt.Printf("%v\t%s\n", k, v)
	}
}
