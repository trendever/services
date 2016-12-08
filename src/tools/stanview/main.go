package main

import (
	"fmt"
	"os"
	"utils/nats"

	"github.com/davecgh/go-spew/spew"
	"github.com/pborman/uuid"
)

func main() {

	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "Usage: %v nats_addr event_name\n", os.Args[0])
		return
	}

	nats.StanSubscribe(&nats.StanSubscription{
		Subject: os.Args[2],
		Group:   uuid.New(),
		DecodedHandler: func(message interface{}) bool {
			spew.Dump(message)
			return true
		},
	})

	nats.Init(&nats.Config{
		URL:         os.Args[1],
		StanCluster: "stan",
		StanID:      uuid.New(),
	}, true)

	select {}
}
