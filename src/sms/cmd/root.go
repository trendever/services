package cmd

import "github.com/spf13/cobra"

//RootCmd is main command
var RootCmd = &cobra.Command{
	Use:   "service",
	Short: "Sms microservice",
}
