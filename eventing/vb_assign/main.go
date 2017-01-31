package main

import (
	"fmt"
	"math/rand"
	"sort"
	"time"
)

const (
	NUM_VBUCKETS = 1024
)

func generateRandomIps(nodeCount int) []string {
	fmt.Println("Num nodes:", nodeCount)
	var addrs []string

	for i := 0; i < nodeCount; i++ {
		firstOctet := rand.Intn(255)
		secondOctet := rand.Intn(255)
		thirdOctet := rand.Intn(255)
		fourthOctet := rand.Intn(255)
		addr := fmt.Sprintf("%d.%d.%d.%d", firstOctet, secondOctet, thirdOctet, fourthOctet)

		addrs = append(addrs, addr)
	}
	return addrs
}

func main() {
	rand.Seed(time.Now().UnixNano())

	numNodes := rand.Intn(19)
	numWorkers := rand.Intn(8)

	eventingAddrs := generateRandomIps(numNodes)

	sort.Strings(eventingAddrs)

	vbucketsPerNode := NUM_VBUCKETS / len(eventingAddrs)

	vbEventingMap := make(map[int]string)
	workerVbMap := make(map[string]map[string]int)

	vbCountPerNode := make([]int, len(eventingAddrs))
	var vbNo int

	for i := 0; i < len(eventingAddrs); i++ {
		vbCountPerNode[i] = vbucketsPerNode
		vbNo += vbucketsPerNode
	}

	remainingVbs := NUM_VBUCKETS - vbNo
	if remainingVbs > 0 {
		for i := 0; i < remainingVbs; i++ {
			vbCountPerNode[i] = vbCountPerNode[i] + 1
		}
	}

	// verification step
	var totalVbs int
	for _, v := range vbCountPerNode {
		totalVbs += v
	}
	fmt.Printf("total vbs: %v\n", totalVbs)

	var startVb int
	for i, v := range vbCountPerNode {
		for j := 0; j < v; j++ {
			vbEventingMap[startVb] = eventingAddrs[i]
			startVb++
		}
	}

	result := make(map[string]int)
	for _, v := range vbEventingMap {
		if _, ok := result[v]; !ok {
			workerVbMap[v] = make(map[string]int)
			result[v] = 1
		} else {
			result[v] = result[v] + 1
		}
	}

	for node := range workerVbMap {
		vbucketCount := result[node]

		vbsPerWorker := vbucketCount / numWorkers

		vbCountPerWorker := make([]int, numWorkers)
		vbNo = 0

		for i := 0; i < numWorkers; i++ {
			vbCountPerWorker[i] = vbsPerWorker
			vbNo += vbsPerWorker
		}

		remainingVbs = vbucketCount - vbNo
		if remainingVbs > 0 {
			for i := 0; i < remainingVbs; i++ {
				vbCountPerWorker[i] = vbCountPerWorker[i] + 1
			}
		}

		for i := 0; i < numWorkers; i++ {
			workerName := fmt.Sprintf("worker_%d", i)
			workerVbMap[node][workerName] = vbCountPerWorker[i]
		}
	}

	for i, v := range eventingAddrs {
		fmt.Printf("node index: %v\t\n addr: %v\tvb count: %v\n", i, v, result[v])
		for k, vbCount := range workerVbMap[v] {
			fmt.Printf("\t\tworker name: %v\tvb count:%v\n", k, vbCount)
		}
	}
}
