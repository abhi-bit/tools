package main

import (
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/couchbase/eventing/logging"
	"github.com/couchbase/gocb"
)

type userBlob struct {
	Score            int    `json:"score"`
	Followers        int    `json:"followers"`
	Coins            int    `json:"coins"`
	GameCredits      int    `json:"game_credits"`
	MoneySpent       int    `json:"money_spent"`
	LastActiveTs     string `json:"last_active_ts"`
	AccountCreatedTs string `json:"account_created_ts"`
	CurrentLevel     int    `json:"current_level"`
	SocialID         string `json:"social_id"`
	SessionCount     int    `json:"session_count"`
	IsBot            bool   `json:"is_bot"`
}

func main() {
	var connStr, username, password, CBbucket string
	var rate, routines int
	var itemCount uint64
	var err error

	connStr = "couchbase://localhost:12000"
	username = "Administrator"
	password = "asdasd"
	CBbucket = "default"
	rate = 1000
	routines = 100
	itemCount = 1000 * 1000 * 1000

	cluster, err := gocb.Connect(connStr)
	if err != nil {
		logging.Errorf("Cluster connect, err: %v\n", err)
		return
	}

	cluster.Authenticate(gocb.PasswordAuthenticator{
		Username: username,
		Password: password,
	})

	bucket, err := cluster.OpenBucket(CBbucket, "")
	if err != nil {
		logging.Errorf("Open bucket, err: %v\n", err)
		return
	}
	defer bucket.Close()

	var i uint64
	var wg sync.WaitGroup
	wg.Add(routines)

	for r := 0; r < routines; r++ {
		go func(wg *sync.WaitGroup) {
			defer wg.Done()
			ticker := time.NewTicker(time.Second / time.Duration(rate))
			for {
				select {
				case <-ticker.C:
					blob := &userBlob{
						Score:      random(0, 700),
						Followers:  1000,
						Coins:      23456789,
						MoneySpent: 0,
						// LastActiveTs:     time.Now().Add(time.Duration(-1*random(1, 100000)) * time.Second).String(),
						// AccountCreatedTs: time.Now().Add(time.Duration(-1*random(1, 1000000)) * time.Second).String(),
						CurrentLevel: 10,
						SocialID:     "xyz",
						SessionCount: 1,
						IsBot:        false,
					}

					atomic.AddUint64(&i, 1)

					bucket.Upsert(fmt.Sprintf("%d_%s", atomic.LoadUint64(&i), blob.SocialID), blob, 0)

					if atomic.LoadUint64(&i) >= itemCount {
						return
					}
				}
			}
		}(&wg)
	}

	wg.Wait()
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
