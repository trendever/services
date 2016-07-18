package cmd

import (
	"github.com/spf13/cobra"
	protocol "proto/sms"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Sends test sms",
	Run: func(cmd *cobra.Command, args []string) {
		// Set up a connection to the server.
		conn, err := grpc.Dial(args[0], grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := protocol.NewSmsServiceClient(conn)

		switch method {
		case "send":
			send(c)
		case "status":
			status(c)
		default:
			log.Fatalf("Unknown method %q", method)
		}
	},
}

var (
	method  string
	phone   string
	message string
	id      int64
)

func init() {
	testCmd.Flags().StringVar(&method, "method", "send", "RPC method 'send' or 'status'")
	testCmd.Flags().StringVarP(&phone, "phone", "p", "", "Phone number")
	testCmd.Flags().StringVarP(&message, "message", "m", "", "Message text")
	testCmd.Flags().Int64VarP(&id, "id", "i", 0, "Message ID")

	RootCmd.AddCommand(testCmd)
}

func send(c protocol.SmsServiceClient) {
	res, err := c.SendSMS(context.Background(), &protocol.SendSMSRequest{Phone: phone, Msg: message})
	if err != nil {
		log.Fatalf("Can't send sms: %v", err)
	}
	log.Println(res)
}

func status(c protocol.SmsServiceClient) {
	res, err := c.RetrieveSmsStatus(context.Background(), &protocol.RetrieveSmsStatusRequest{Id: id})
	if err != nil {
		log.Fatalf("Can't retrieve status: %v", err)
	}
	log.Println(res)
}
