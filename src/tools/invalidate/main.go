package main

import (
	"bufio"
	"instagram"
	"math/rand"
	"os"
	"strings"
	"time"
	"utils/log"
)

type Proxy struct {
	Addr, User, Pass string
}

func main() {
	log.Init(true, "invalidate", "")
	if len(os.Args) != 4 {
		log.Fatalf("Usage: %v username password proxyfile", os.Args[0])
	}

	file, err := os.Open(os.Args[3])
	if err != nil {
		log.Fatalf("failed to open file: %v\n", err)
	}
	scanner := bufio.NewScanner(file)
	var proxies []string
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#") || scanner.Text() == "" {
			continue
		}
		proxies = append(proxies, scanner.Text())
	}
	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}

	if len(proxies) == 0 {
		log.Fatalf("no usable proxies are presented")
	}

	rand.Seed(time.Now().UnixNano())
	_, err = instagram.NewInstagram(os.Args[1], os.Args[2], "")
	if err != nil {
		log.Fatalf("failed to initialize account without proxies %v", err)
	}

INV:
	for _, proxy := range proxies {
		ig, err := instagram.NewInstagram(os.Args[1], os.Args[2], proxy)
		if err != nil {
			if ig.CheckpointURL != "" {
				log.Info("checkpoint required!")
				return
			}
			log.Errorf("proxy %v: %v", proxy, err)
			continue
		}
		_, err = ig.Inbox("")
		if err != nil {
			log.Errorf("proxy %v: %v", proxy, err)
			if ig.CheckpointURL != "" {
				log.Debug("got checkpoint!")
				return
			}
		} else {
			log.Info("proxy %v passed", proxy)
		}
	}
	goto INV
}
