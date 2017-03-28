package main

import (
	"fmt"
	"instagram"
)

func main() {
	i, err := instagram.NewInstagram("theta.tnd", "422312", nil)
	if err != nil {
		panic(err)
	}

	postToLike := "1463894800754542930_4594731360"

	res, err := i.CommentMedia(postToLike, "works")

	if err != nil {
		panic(err)
	}

	fmt.Printf("%v", res)
}
