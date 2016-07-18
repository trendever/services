package main

import (
	"flag"
	"fmt"
	"github.com/timehop/go-bitly"
	"os"
)

func main() {
	var (
		accessToken string
		login       string
		apiKey      string
		domain      string
		url         string
	)

	flag.StringVar(&accessToken, "a", "", "access token")
	flag.StringVar(&login, "l", "", "login")
	flag.StringVar(&apiKey, "k", "", "api key")
	flag.StringVar(&domain, "d", "", "custom domain")

	flag.Parse()

	args := flag.Args()
	if len(args) != 1 {
		flag.Usage()
		os.Exit(1)
	}

	url = args[0]

	var (
		s   bitly.ShortenResult
		err error
	)

	c := bitly.Client{AccessToken: accessToken, Login: login, APIKey: apiKey}
	if domain == "" {
		s, err = c.Shorten(url)
	} else {
		s, err = c.ShortenUsingDomain(url, domain)
	}

	fmt.Printf("GlobalHash is %v\n", s.GlobalHash)
	fmt.Printf("Hash is %v\n", s.Hash)
	fmt.Printf("Long URL is %v\n", s.LongURL)
	fmt.Printf("New Hash is %v\n", s.NewHash)
	fmt.Printf("URL is %v\n", s.URL)
	if err != nil {
		fmt.Printf("Error is %v\n", err.Error())
	}
}
