package main

import (
	"common/db"
	"common/log"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"os/signal"
	"proto/push"
	"push/config"
	"push/exteral"
	"push/models"
	"push/pushers"
	"push/server"
	"syscall"
	"time"
	"utils/rpc"
)

func main() {
	var cmd = cobra.Command{
		Use:   "service",
		Short: "push service",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "start",
		Short: "Starts service",
		Run: func(cmd *cobra.Command, args []string) {
			log.Info("Starting service...")

			config.Init()
			exteral.Init()
			db.Init(&config.Get().DB)
			pushers.Init()
			rpc := rpc.MakeServer(config.Get().RPC)
			server := server.NewPushServer()

			log.Info("Registering server...")
			push.RegisterPushServiceServer(
				rpc.Server,
				server,
			)
			rpc.StartServe()

			interrupt := make(chan os.Signal)
			signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

			<-interrupt
			rpc.Stop()
			server.Stop()
		},
	})

	var drop bool
	migrateCmd := &cobra.Command{
		Use:   "migrate",
		Short: "Runs database migration",
		Run: func(cmd *cobra.Command, args []string) {
			log.Warn("Starting database migration for elasticsync service")
			config.Init()
			c := config.Get()
			db.Init(&c.DB)

			models.Migrate(drop)

			log.Info("Migration done")
		},
	}
	migrateCmd.Flags().BoolVarP(&drop, "drop", "d", false, "Drops tables before migration")
	cmd.AddCommand(migrateCmd)

	var serviceName, token, data, title, body string
	pushCmd := &cobra.Command{
		Use:   "push",
		Short: "Pushs one notification from arguments",
		Run: func(cmd *cobra.Command, args []string) {
			config.Init()
			pushers.Init()
			if token == "" || (data == "" && body == "") {
				log.Fatal(errors.New("empty token or (body and data) argument"))
			}
			service, ok := push.ServiceType_value[serviceName]
			if !ok {
				log.Fatal(fmt.Errorf("unknown service %v", serviceName))
			}
			pusher, err := pushers.GetPusher(push.ServiceType(service))
			if err != nil {
				log.Fatal(err)
			}
			res, err := pusher.Push(&models.PushNotify{
				Data:       data,
				Body:       body,
				Title:      title,
				Priority:   push.Priority_HING,
				Expiration: time.Now().Add(time.Second * 10),
			}, []string{token})
			if err != nil {
				log.Fatal(fmt.Errorf("failed to push msg: %v", err))
			}
			if res.Invalids == nil && res.Updates == nil && res.NeedRetry == nil {
				fmt.Println("success")
			}
		},
	}
	pushCmd.Flags().StringVarP(&serviceName, "service", "s", "FCM", "Type of push service")
	pushCmd.Flags().StringVarP(&token, "token", "t", "", "Token")
	pushCmd.Flags().StringVarP(&data, "data", "d", "", "Data, valid json")
	pushCmd.Flags().StringVarP(&title, "caption", "c", "", "Notification caption")
	pushCmd.Flags().StringVarP(&body, "body", "b", "", "Notification body")
	cmd.AddCommand(pushCmd)

	log.PanicLogger(func() {
		if err := cmd.Execute(); err != nil {
			log.Fatal(err)
		}
	})
}
