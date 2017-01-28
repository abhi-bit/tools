package main

import (
	"log"
	"time"

	"github.com/couchbase/indexing/secondary/common"
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}

func watchClusterChanges() {
	selfRestart := func() {
		log.Println("Within self restart function")
		time.Sleep(time.Duration(1) * time.Second)
		go watchClusterChanges()
	}

	clusterURL := "http://Administrator:asdasd@127.0.0.1:9000"

	scn, err := common.NewServicesChangeNotifier(clusterURL, "default")
	if err != nil {
		log.Printf("newserviceschange notifier, err: %v\n", err.Error())
		selfRestart()
		return
	}

	defer scn.Close()

	ch := scn.GetNotifyCh()

	for {
		select {
		case _, ok := <-ch:
			if !ok {
				log.Printf("Not ok message from channel\n")
				selfRestart()
				return
			} else {
				log.Printf("Got message that cluster state has changed\n")
				selfRestart()
				return
			}
		}
	}

}

func main() {
	watchClusterChanges()

	time.Sleep(10000 * time.Second)
}
