// Tool receives raw events from dcp-client.
package main

import "flag"
import "fmt"
import "os"
import "strings"
import "time"
import "encoding/binary"

import "github.com/couchbase/cbauth"
import "github.com/couchbase/indexing/secondary/common"
import "github.com/couchbase/indexing/secondary/dcp"
import mcd "github.com/couchbase/indexing/secondary/dcp/transport"
import mc "github.com/couchbase/indexing/secondary/dcp/transport/client"
import (
	"bytes"
	"encoding/json"
	"github.com/couchbase/indexing/secondary/logging"
)

const (
	dcpDatatypeJSON      = uint8(1)
	dcpDatatypeJSONXattr = uint8(5)
)

type xattrMetadata struct {
	Cas    string   `json:"cas"`
	Digest uint32   `json:"digest"`
	Timers []string `json:"timers"`
}

var options struct {
	buckets    []string // buckets to connect with
	maxVbno    int      // maximum number of vbuckets
	kvaddrs    []string
	stats      int // periodic timeout(ms) to print stats, 0 will disable
	printflogs bool
	auth       string
	info       bool
	debug      bool
	trace      bool
}

var done = make(chan bool, 16)
var rch = make(chan []interface{}, 10000)

func argParse() string {
	var buckets string
	var kvaddrs string

	flag.StringVar(&buckets, "buckets", "default",
		"buckets to listen")
	flag.StringVar(&kvaddrs, "kvaddrs", "",
		"list of kv-nodes to connect")
	flag.IntVar(&options.maxVbno, "maxvb", 1024,
		"maximum number of vbuckets")
	flag.IntVar(&options.stats, "stats", 1000,
		"periodic timeout in mS, to print statistics, `0` will disable stats")
	flag.BoolVar(&options.printflogs, "flogs", false,
		"display failover logs")
	flag.StringVar(&options.auth, "auth", "",
		"Auth user and password")
	flag.BoolVar(&options.info, "info", false,
		"display informational logs")
	flag.BoolVar(&options.debug, "debug", false,
		"display debug logs")
	flag.BoolVar(&options.trace, "trace", false,
		"display trace logs")

	flag.Parse()

	options.buckets = strings.Split(buckets, ",")
	if options.debug {
		logging.SetLogLevel(logging.Debug)
	} else if options.trace {
		logging.SetLogLevel(logging.Trace)
	} else {
		logging.SetLogLevel(logging.Info)
	}
	if kvaddrs == "" {
		logging.Fatalf("please provided -kvaddrs")
	}
	options.kvaddrs = strings.Split(kvaddrs, ",")

	args := flag.Args()
	if len(args) < 1 {
		usage()
		os.Exit(1)
	}
	return args[0]
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage : %s [OPTIONS] <cluster-addr> \n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	cluster := argParse()

	// setup cbauth
	if options.auth != "" {
		up := strings.Split(options.auth, ":")
		if _, err := cbauth.InternalRetryDefaultInit(cluster, up[0], up[1]); err != nil {
			logging.Fatalf("Failed to initialize cbauth: %s", err)
		}
	}

	for _, bucket := range options.buckets {
		go startBucket(cluster, bucket, options.kvaddrs)
	}
	receive()
}

func startBucket(cluster, bucketn string, kvaddrs []string) int {
	defer func() {
		if r := recover(); r != nil {
			logging.Errorf("%s:\n%s\n", r, logging.StackTrace())
		}
	}()

	logging.Infof("Connecting with %q\n", bucketn)
	b, err := common.ConnectBucket(cluster, "default", bucketn)
	mf(err, "bucket")

	dcpConfig := map[string]interface{}{
		"genChanSize":    10000,
		"dataChanSize":   10000,
		"numConnections": 1,
		"activeVbOnly":   true,
	}
	dcpFeed, err := b.StartDcpFeedOver(
		couchbase.NewDcpFeedName("rawupr"),
		uint32(0), uint32(4), options.kvaddrs, 0xABCD, dcpConfig)
	mf(err, "- upr")

	vbnos := listOfVbnos(options.maxVbno)

	flogs, err := b.GetFailoverLogs(0xABCD, vbnos, dcpConfig)
	mf(err, "- dcp failoverlogs")

	if options.printflogs {
		printFlogs(vbnos, flogs)
	}

	go startDcp(dcpFeed, flogs)

	for {
		e, ok := <-dcpFeed.C
		if ok == false {
			logging.Infof("Closing for bucket %q\n", b.Name)
		}
		rch <- []interface{}{b.Name, e}
	}
}

func startDcp(dcpFeed *couchbase.DcpFeed, flogs couchbase.FailoverLog) {
	start, end := uint64(0), uint64(0xFFFFFFFFFFFFFFFF)
	snapStart, snapEnd := uint64(0), uint64(0)
	for vbno, flog := range flogs {
		x := flog[len(flog)-1] // map[uint16][][2]uint64
		opaque, flags, vbuuid := uint16(vbno), uint32(0), x[0]
		err := dcpFeed.DcpRequestStream(
			vbno, opaque, flags, vbuuid, start, end, snapStart, snapEnd)
		mf(err, fmt.Sprintf("stream-req for %v failed", vbno))
	}
}

func mf(err error, msg string) {
	if err != nil {
		logging.Fatalf("%v: %v", msg, err)
	}
}

func receive() {
	// bucket -> Opcode -> #count
	counts := make(map[string]map[mcd.CommandCode]int)

	var tick <-chan time.Time
	if options.stats > 0 {
		tick = time.Tick(time.Millisecond * time.Duration(options.stats))
	}

loop:
	for {
		select {
		case msg, ok := <-rch:
			if ok == false {
				break loop
			}
			bucket, e := msg[0].(string), msg[1].(*mc.DcpEvent)
			if e.Opcode == mcd.DCP_MUTATION {
				switch e.Datatype {
				case dcpDatatypeJSON:
					logging.Infof("Datatype: JSON KEY -- %v\n", string(e.Key))
					logging.Infof("     %v\n", string(e.Value))
				case dcpDatatypeJSONXattr:
					totalXattrLen := binary.BigEndian.Uint32(e.Value[0:])
					totalXattrData := e.Value[4 : 4+totalXattrLen-1]

					xattrPrefix := "eventing"
					var xMeta xattrMetadata
					var bytesDecoded uint32
					for bytesDecoded < totalXattrLen {
						frameLength := binary.BigEndian.Uint32(totalXattrData)
						bytesDecoded += 4
						fmt.Printf("uint32 bytesDecoded: %v frameLength: %v dump: %v remain: %v\n",
							bytesDecoded, frameLength, totalXattrData[:4], len(totalXattrData))
						frameData := totalXattrData[4 : 4+frameLength-1]
						bytesDecoded += frameLength
						if bytesDecoded < totalXattrLen {
							totalXattrData = totalXattrData[4+frameLength:]
							fmt.Printf("total bytesDecoded: %v frameLength: %v total: %v dump: %v\n",
								bytesDecoded, frameLength, len(totalXattrData), totalXattrData[:4])
						}

						if len(frameData) > len(xattrPrefix) {
							if bytes.Compare(frameData[:len(xattrPrefix)], []byte(xattrPrefix)) == 0 {
								toParse := frameData[len(xattrPrefix)+1:]

								fmt.Printf("toParse: %v frameData: %v\n",
									toParse, frameData)

								err := json.Unmarshal(toParse, &xMeta)
								if err != nil {
									fmt.Println(err)
									continue
								}
							}
						}

						logging.Infof("Datatype: JSON+XATTR KEY -- %v\n", string(e.Key))
						logging.Infof("totalLength: %v frameLength: %v frameData: %v\n",
							totalXattrLen, frameLength, string(frameData))
					}

					if xMeta.Cas != "" && xMeta.Digest != 0 {
						fmt.Println("Eventing encountered")
					}
				}
			}
			if _, ok := counts[bucket]; !ok {
				counts[bucket] = make(map[mcd.CommandCode]int)
			}
			if _, ok := counts[bucket][e.Opcode]; !ok {
				counts[bucket][e.Opcode] = 0
			}
			counts[bucket][e.Opcode]++

		case <-tick:
			// for bucket, m := range counts {
			// 	logging.Infof("%q %s\n", bucket, sprintCounts(m))
			// }
			// logging.Infof("\n")
		}
	}
}

func sprintCounts(counts map[mcd.CommandCode]int) string {
	line := ""
	for i := 0; i < 256; i++ {
		opcode := mcd.CommandCode(i)
		if n, ok := counts[opcode]; ok {
			line += fmt.Sprintf("%s:%v ", mcd.CommandNames[opcode], n)
		}
	}
	return strings.TrimRight(line, " ")
}

func listOfVbnos(maxVbno int) []uint16 {
	// list of vbuckets
	vbnos := make([]uint16, 0, maxVbno)
	for i := 0; i < maxVbno; i++ {
		vbnos = append(vbnos, uint16(i))
	}
	return vbnos
}

func printFlogs(vbnos []uint16, flogs couchbase.FailoverLog) {
	for i, vbno := range vbnos {
		logging.Infof("Failover log for vbucket %v\n", vbno)
		logging.Infof("   %#v\n", flogs[uint16(i)])
	}
	logging.Infof("\n")
}
