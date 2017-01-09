package main

import (
	"fmt"
	"instagram"
	"math/rand"
	"os"
	"time"
	"utils/log"
)

const printStep = 10

func main() {
	log.Init(true, "invalidate", "")
	if len(os.Args) != 3 {
		log.Info("Usage: %v username password", os.Args[0])
		return
	}
	rand.Seed(time.Now().Unix())
	ig, err := instagram.NewInstagram(os.Args[1], os.Args[2])
	if err != nil {
		log.Errorf("failed to initialize account: %v", err)
		return
	}
	var counter uint64
	start := time.Now()
	for {
		_, err = ig.SearchUsers(fmt.Sprintf("%v", rand.Intn(32)))
		if err != nil {
			log.Errorf("request failed: %v", err)
			return
		}
		counter++
		if counter == printStep {
			counter = 0
			now := time.Now()
			log.Debug("%v req/s", float64(printStep)/float64(now.Sub(start))*float64(time.Second))
			start = now
		}
	}
}
