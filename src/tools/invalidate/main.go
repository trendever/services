package main

import (
	"bufio"
	"golang.org/x/net/proxy"
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
	var proxies []Proxy
	for scanner.Scan() {
		if strings.HasPrefix(scanner.Text(), "#") || scanner.Text() == "" {
			continue
		}
		slice := strings.Split(scanner.Text(), ";")
		switch len(slice) {
		case 1:
			proxies = append(proxies, Proxy{Addr: slice[0]})
		case 3:
			proxies = append(proxies, Proxy{
				Addr: slice[0],
				User: slice[1],
				Pass: slice[2],
			})
		default:
			log.Errorf("unexpected format in '%v', skipping", scanner.Text())
		}
	}
	if scanner.Err() != nil {
		log.Fatal(scanner.Err())
	}

	if len(proxies) == 0 {
		log.Fatalf("no usable proxies are presented")
	}

	rand.Seed(time.Now().UnixNano())
	_, err = instagram.NewInstagram(os.Args[1], os.Args[2], nil)
	if err != nil {
		log.Fatalf("failed to initialize account without proxies %v", err)
	}

	instagram.DoResponseLogging = true

	for _, conf := range proxies {
		dialer, _ := proxy.SOCKS5("tcp", conf.Addr, &proxy.Auth{
			User:     conf.User,
			Password: conf.Pass,
		}, proxy.Direct)
		ig, err := instagram.NewInstagram(os.Args[1], os.Args[2], dialer.Dial)
		if err != nil {
			if ig.CheckpointURL != "" {
				log.Info("checkpoint required!")
				return
			}
			log.Errorf("proxy %v: %v", conf.Addr, err)
			continue
		}
		ret, err := ig.Inbox("")
		log.Debug("%v", ret)
		if ig.CheckpointURL != "" {
			log.Debug("got checkpoint!")
			return
		}
		log.Info("proxy %v passed", conf.Addr)
	}
	log.Info("no more proxies, account is still valid")
}
