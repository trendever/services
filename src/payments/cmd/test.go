package cmd

import (
	"fmt"
	"log"
	"proto/payment"
	"utils/rpc"

	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var cmdTest = &cobra.Command{
	Use:   "test",
	Short: "Test RPC",
	Run: func(cmd *cobra.Command, args []string) {

		conn := rpc.Connect(conn)
		defer conn.Close()

		client := payment.NewPaymentServiceClient(conn)

		switch method {
		case "create":

			res, err := client.CreateOrder(context.Background(), &payment.CreateOrderRequest{
				Data: &payment.OrderData{
					Amount:         amount,
					Currency:       payment.Currency(payment.Currency_value[currency]),
					LeadId:         leadID,
					ShopCardNumber: shopCard,
				},
			})

			fmt.Printf("Result: %#v\nError: %v\n", res, err)

		case "buy":
			res, err := client.BuyOrder(context.Background(), &payment.BuyOrderRequest{
				PayId: payID,
				Ip:    ip,
			})

			fmt.Printf("Result: %#v\nError: %v\n", res, err)
		default:
			log.Fatal("Unknown method")
		}

	},
}

var (
	method string
	conn   string

	// create
	amount   uint64
	currency string
	leadID   uint64
	shopCard string

	// buy
	payID uint64
	ip    string
)

func init() {
	RootCmd.AddCommand(cmdTest)
	cmdTest.Flags().StringVarP(&method, "method", "m", "create", "Method to call")
	cmdTest.Flags().StringVarP(&conn, "conn", "s", ":7777", "Connect to")

	cmdTest.Flags().Uint64VarP(&amount, "amount", "a", 100, "Amount to transfer")
	cmdTest.Flags().StringVarP(&currency, "currency", "c", "RUB", "Currency")
	cmdTest.Flags().Uint64VarP(&leadID, "lead", "l", 0, "Connected leadID")
	cmdTest.Flags().StringVarP(&shopCard, "shopCard", "t", "", "Shop card ID")

	cmdTest.Flags().Uint64VarP(&payID, "pay", "p", 0, "Previously generated pay ID")
	cmdTest.Flags().StringVarP(&ip, "ip", "i", "127.0.0.1", "Customer IP addr")
}
