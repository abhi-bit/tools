package main

import (
	"fmt"
	"io"
	"log"
	"math/rand"
	"os"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/abhi-bit/tools/eventing/high_volume_logger/logger"
)

func main() {
	if len(os.Args) < 7 {
		fmt.Println("<program_name> <size of each log entry> <lines_tow_write_per_worker> <worker_count> <max_file_size> <file_count>")
		return
	}

	appFile := os.Args[1]
	size, err := strconv.Atoi(os.Args[2])
	if err != nil {
		log.Println("Size string conversion, err:", err)
		return
	}

	count, err := strconv.Atoi(os.Args[3])
	if err != nil {
		log.Println("Count string conversion, err:", err)
		return
	}

	thrCount, err := strconv.Atoi(os.Args[4])
	if err != nil {
		log.Println("ThrCount string conversion, err:", err)
		return
	}

	fileSize, err := strconv.Atoi(os.Args[5])
	if err != nil {
		log.Println("File size string conversion, err:", err)
		return
	}

	fileCount, err := strconv.Atoi(os.Args[6])
	if err != nil {
		log.Println("File count string conversion, err:", err)
		return
	}

	appLogger, err := logger.OpenAppLog(appFile, 0600, int64(fileSize), int64(fileCount))
	if err != nil {
		log.Println("open app log, err:", err)
		return
	}
	defer appLogger.Close()

	token := make([]byte, size)
	rand.Read(token)

	newLineEntry := []byte(fmt.Sprintf("\n"))
	token = append(token, newLineEntry[0])

	latencyStats := make(map[int64]int64)
	rwMutex := &sync.RWMutex{}

	// log.Printf("App log: %s byte slice size: %d count: %d thrCount: %d\n", appFile, size, count, thrCount)

	var wg sync.WaitGroup
	wg.Add(thrCount)

	for i := 0; i < thrCount; i++ {

		go func(appLogger io.WriteCloser, latencyStats map[int64]int64, rwMutex *sync.RWMutex, wg *sync.WaitGroup) {
			defer wg.Done()

			for j := 0; j < count; j++ {
				startTs := time.Now()
				_, err := appLogger.Write(token)
				if err != nil {
					log.Println("Write err:", err)
				}
				endTs := time.Now()

				elapsedNS := endTs.Sub(startTs).Nanoseconds()
				rwMutex.Lock()
				if _, ok := latencyStats[elapsedNS]; !ok {
					latencyStats[elapsedNS] = 0
				}
				latencyStats[elapsedNS]++
				rwMutex.Unlock()

			}
		}(appLogger, latencyStats, rwMutex, &wg)
	}

	wg.Wait()

	fmt.Println("Percentile\tTime elapsed(in ns)")
	fmt.Println("50\t\t", percentileN(latencyStats, 50))
	fmt.Println("80\t\t", percentileN(latencyStats, 80))
	fmt.Println("90\t\t", percentileN(latencyStats, 90))
	fmt.Println("95\t\t", percentileN(latencyStats, 95))
	fmt.Println("99\t\t", percentileN(latencyStats, 99))
	fmt.Println("100\t\t", percentileN(latencyStats, 100))
}

func percentileN(latencyStats map[int64]int64, p int) int {
	var samples sort.IntSlice
	var numSamples int64
	for bin, binCount := range latencyStats {
		samples = append(samples, int(bin))
		numSamples += binCount
	}
	sort.Sort(samples)
	i := numSamples*int64(p)/100 - 1

	var counter int64
	var prevSample int
	for _, sample := range samples {
		if counter > i {
			return prevSample
		}
		counter += latencyStats[int64(sample)]
		prevSample = sample
	}
	return samples[len(samples)-1]
}
