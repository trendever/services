package main

import (
	"bufio"
	"common/log"
	"instagram"
	"os"
	"strings"
)

type Proxy struct {
	Addr, User, Pass string
}

func main() {
	log.Init(true, "invalidate", "")
	if len(os.Args) < 2 {
		log.Fatalf("Usage: %v accountsfile [proxyfile]", os.Args[0])
	}

	proxies := []string{""}

	if len(os.Args) > 2 {
		file, err := os.Open(os.Args[2])
		if err != nil {
			log.Fatalf("failed to open file: %v\n", err)
		}
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			if strings.HasPrefix(scanner.Text(), "#") || scanner.Text() == "" {
				continue
			}
			proxies = append(proxies, scanner.Text())
		}
		if scanner.Err() != nil {
			log.Fatal(scanner.Err())
		}
	}

	type auth struct {
		login string
		pass  string
	}
	var accs []auth

	f, err := os.Open(os.Args[1])
	log.Fatal(err)
	scan := bufio.NewScanner(f)
	for scan.Scan() {
		trimmed := strings.Trim(scan.Text(), " \t\r")
		if strings.HasPrefix(trimmed, "#") || trimmed == "" {
			continue
		}
		split := strings.Split(trimmed, " ")
		if len(split) != 2 {
			log.Warn("bad login pair: '%v'", scan.Text())
			continue
		}
		accs = append(accs, auth{login: split[0], pass: split[1]})
	}

	for _, acc := range accs {
		for _, proxy := range proxies {
			ig, err := instagram.NewInstagram(acc.login, acc.pass, proxy, false)
			if err != nil {
				if ig != nil && ig.CheckpointURL != "" {
					log.Info("checkpoint required for %v!", acc.login)
					return
				}
				log.Errorf("%v, proxy '%v': %v", acc.login, proxy, err)
				continue
			}
			_, err = ig.Inbox("")
			if err != nil {
				log.Errorf("%v, proxy '%v': %v", acc.login, proxy, err)
				if ig.CheckpointURL != "" {
					log.Info("checkpoint required for %v!", acc.login)
					return
				}
			} else {
				log.Info("proxy '%v' passed for %v", proxy, acc.login)
			}
		}
	}
}
