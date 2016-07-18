package cmd

import "github.com/spf13/cobra"

//RootCmd main command
var RootCmd = &cobra.Command{
	Use:   "serivce",
	Short: "Service for user registration and authorization",
}
