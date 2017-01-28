package main

import (
	"fmt"

	"github.com/couchbase/indexing/secondary/common"
)

const (
	// EVENTING_ADMIN_SERVICE = "eventing_admin_port"
	EVENTING_ADMIN_SERVICE = "kv"
)

func main() {
	// clusterURL := "http://Administrator:asdasd@172.16.1.198:9000"
	clusterURL := "http://Administrator:asdasd@127.0.0.1:9000"
	//clusterURL := "http://Administrator:asdasd@apple:8091"

	cinfo, err := common.NewClusterInfoCache(clusterURL, "default")
	if err != nil {
		fmt.Printf("new cluster info cache, err: %s\n", err.Error())
		return
	}

	if err := cinfo.Fetch(); err != nil {
		fmt.Printf("cluster info fetch failed, err: %s\n", err.Error())
		return
	}

	eventingAddrs := cinfo.GetNodesByServiceType(EVENTING_ADMIN_SERVICE)

	eventingNodes := []string{}
	for _, eventingAddr := range eventingAddrs {
		addr, _ := cinfo.GetServiceAddress(eventingAddr, EVENTING_ADMIN_SERVICE)
		eventingNodes = append(eventingNodes, addr)
		vbs, _ := cinfo.GetVBuckets(eventingAddr, "default")
		fmt.Printf("addr => %#v vbs => %#v\n", addr, vbs)
	}
	fmt.Printf("Eventing addrs: %#v\n", eventingNodes)
}
