package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
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
	var rate int
	var err error

	if len(os.Args) == 5 {
		connStr = os.Args[1]
		username = os.Args[2]
		password = os.Args[3]
		CBbucket = os.Args[4]
		rate, err = strconv.Atoi(os.Args[5])
		if err != nil {
			logging.Errorf("Failed to convert rate to int, err: %v", err)
			return
		}
	} else {
		connStr = "couchbase://localhost:12000"
		username = "Administrator"
		password = "asdasd"
		CBbucket = "default"
		rate = 10000
	}

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

	ticker := time.NewTicker(time.Second / time.Duration(rate))

	i := 0
	for {
		select {
		case <-ticker.C:
			blob := &userBlob{
				Score:            random(500, 700),
				Followers:        random(1000, 1000000),
				Coins:            random(123456789, 12345678901234),
				MoneySpent:       random(0, 1),
				LastActiveTs:     time.Now().Add(time.Duration(-1*random(1, 100000)) * time.Second).String(),
				AccountCreatedTs: time.Now().Add(time.Duration(-1*random(1, 1000000)) * time.Second).String(),
				CurrentLevel:     random(1000, 100000),
				SocialID:         fmt.Sprintf("xyz_%d", random(100, 123456789)),
				SessionCount:     random(1, 100),
				IsBot:            false,
			}

			i++

			bucket.Upsert(fmt.Sprintf("%d_%s", i, blob.SocialID), blob, 0)
		}
	}
}

func random(min, max int) int {
	rand.Seed(time.Now().Unix())
	return rand.Intn(max-min) + min
}
