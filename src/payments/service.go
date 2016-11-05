package main

import (
	"payments/cmd"
	"utils/cli"

	// load gateways
	_ "payments/payture"
)

func main() {
	cli.Main(&cmd.Service{})
}
