package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	protocol "proto/mail"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"log"
	"os"
	"strings"
	"time"
)

var testCmd = &cobra.Command{
	Use:   "test",
	Short: "Sends test email to your address",
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) != 1 {
			fmt.Println("Pass host:port as first program argument")
			os.Exit(-1)
		}
		// Set up a connection to the server.
		conn, err := grpc.Dial(args[0], grpc.WithInsecure())
		if err != nil {
			log.Fatalf("did not connect: %v", err)
		}
		defer conn.Close()
		c := protocol.NewMailServiceClient(conn)

		switch method {
		case "send":
			send(c)
		case "status":
			status(c)
		default:
			log.Fatalf("Unknown test method: %s", method)
		}
		time.Sleep(time.Second)

	},
}

var (
	from    string
	subject string
	message string
	to      string
	method  string
	id      uint64
)

func init() {
	testCmd.Flags().StringVarP(&from, "from", "f", "", "From field")
	testCmd.Flags().StringVarP(&subject, "subject", "s", "", "Subject field")
	testCmd.Flags().StringVarP(&message, "message", "m", "", "Message field")
	testCmd.Flags().StringVarP(&to, "to", "t", "", "To field")
	testCmd.Flags().StringVar(&method, "method", "send", "Method 'send' or 'status'")
	testCmd.Flags().Uint64Var(&id, "id", 1, "Email id")
	RootCmd.AddCommand(testCmd)
}

func send(c protocol.MailServiceClient) {
	msg := &protocol.MessageRequest{From: from, To: strings.Split(to, ","), Subject: subject, Message: message}
	log.Println(msg)
	r, err := c.Send(context.Background(), msg)
	if err != nil {
		log.Fatalf("could not send: %v", err)
	}
	log.Printf("Send response: %v %s", r.Id, r.Status)
}

func status(c protocol.MailServiceClient) {
	msg := &protocol.StatusRequest{Id: id}
	r, err := c.Status(context.Background(), msg)

	if err != nil {
		log.Fatalf("could not send: %v", err)
	}
	log.Printf("Status response: %v %s %s", r.Id, r.Status, r.Error)
}
