package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
)

func percentileN(latencyStats map[string]uint64, p int) int {
	var samples sort.IntSlice
	var numSamples uint64
	for bin, binCount := range latencyStats {
		sample, err := strconv.Atoi(bin)
		if err == nil {
			samples = append(samples, sample)
			numSamples += binCount
		}
	}
	sort.Sort(samples)
	i := numSamples*uint64(p)/100 - 1

	var counter uint64
	var prevSample int
	for _, sample := range samples {
		if counter > i {
			return prevSample
		}
		counter += latencyStats[strconv.Itoa(sample)]
		prevSample = sample
	}
	return samples[len(samples)-1]
}

func main() {
	filename := os.Args[1]
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		fmt.Println("readFile err:", err)
		return
	}

	var latencyStats map[string]uint64
	err = json.Unmarshal(data, &latencyStats)
	if err != nil {
		fmt.Println("unmarshal err:", err)
		return
	}

	fmt.Println("10:", percentileN(latencyStats, 10))
	fmt.Println("20:", percentileN(latencyStats, 20))
	fmt.Println("30:", percentileN(latencyStats, 30))
	fmt.Println("50:", percentileN(latencyStats, 50))
	fmt.Println("80:", percentileN(latencyStats, 80))
	fmt.Println("90:", percentileN(latencyStats, 90))
	fmt.Println("95:", percentileN(latencyStats, 95))
	fmt.Println("99:", percentileN(latencyStats, 99))
	fmt.Println("100:", percentileN(latencyStats, 100))
}
