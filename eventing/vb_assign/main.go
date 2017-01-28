package main

import (
	"fmt"
	"sort"
)

const (
	NUM_VBUCKETS = 1024
)

func main() {
	eventingAddrs := []string{
		"10.4.2.1:8091",
		"10.1.2.3:8091",
		"10.4.5.2:8091",
		"10.1.2.2:8091",
		"10.1.1.1:8091",
	}

	sort.Strings(eventingAddrs)

	vbucketsPerNode := NUM_VBUCKETS / len(eventingAddrs)

	vbEventingMap := make(map[int]string)
	// vbEventingMap := make(map[uint16]string)
	// var startVB uint16
	var startVB int

	for i := 0; i < len(eventingAddrs); i++ {
		for j := 0; j <= vbucketsPerNode && startVB < NUM_VBUCKETS; j++ {
			vbEventingMap[startVB] = eventingAddrs[i]
			startVB++
		}
	}

	var keys []int
	for k := range vbEventingMap {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		fmt.Printf("%v\t%v\n", k, vbEventingMap[k])
	}

	result := make(map[string]int)
	for _, v := range vbEventingMap {
		if _, ok := result[v]; !ok {
			result[v] = 1
		} else {
			result[v] = result[v] + 1
		}
	}
	// fmt.Printf("vbEventingMap dump: %#v\n", result)
}
