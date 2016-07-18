package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	protocol "proto/auth"
	"utils/rpc"
	"golang.org/x/net/context"
	"log"
)

var cmdTest = &cobra.Command{
	Use:   "test",
	Short: "Sends test requests to the auth service",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			log.Fatalf("Pass hostname as first argument for command")
		}
		conn := rpc.Connect(args[0])
		defer conn.Close()

		c := protocol.NewAuthServiceClient(conn)
		var resp interface{}
		var err error
		switch method {
		case "create":
			resp, err = create(c)
		case "login":
			resp, err = login(c)
		case "send_pass":
			resp, err = sendPass(c)
		case "token_data":
			resp, err = getTokeData(c)
		case "new_token":
			resp, err = getNewToken(c)
		default:
			log.Fatal("Unknown method")
		}
		fmt.Println(resp, err)
	},
}

var (
	method            string
	phone             string
	instagramUsername string
	username          string
	password          string
	tokeData          string
)

func init() {
	cmdTest.Flags().StringVarP(&method, "method", "m", "create", "One of exests method: create, login, send_pass, token_data")
	cmdTest.Flags().StringVarP(&phone, "phone", "p", "", "User phone")
	cmdTest.Flags().StringVarP(&instagramUsername, "iname", "i", "", "User instagram username")
	cmdTest.Flags().StringVarP(&username, "name", "n", "", "User username")
	cmdTest.Flags().StringVarP(&password, "password", "s", "", "User password")
	cmdTest.Flags().StringVarP(&tokeData, "token", "t", "", "Token data")

	RootCmd.AddCommand(cmdTest)
}

func create(c protocol.AuthServiceClient) (interface{}, error) {
	return c.RegisterNewUser(context.Background(), &protocol.NewUserRequest{
		PhoneNumber:       phone,
		Username:          username,
		InstagramUsername: instagramUsername,
	})
}

func login(c protocol.AuthServiceClient) (interface{}, error) {
	return c.Login(context.Background(), &protocol.LoginRequest{
		PhoneNumber: phone,
		Password:    password,
	})
}

func sendPass(c protocol.AuthServiceClient) (interface{}, error) {
	return c.SendNewSmsPassword(context.Background(), &protocol.SmsPasswordRequest{
		PhoneNumber: phone,
	})
}

func getTokeData(c protocol.AuthServiceClient) (interface{}, error) {
	return c.GetTokenData(context.Background(), &protocol.TokenDataRequest{Token: tokeData})
}

func getNewToken(c protocol.AuthServiceClient) (interface{}, error) {
	return c.GetNewToken(context.Background(), &protocol.NewTokenRequest{PhoneNumber: phone})
}
