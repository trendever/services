package cmd

import (
	"github.com/spf13/cobra"
	"proto/sms"
	"sms/conf"
	"sms/models"
	"sms/server"
	"utils/cli"
	"utils/db"
	"utils/log"
	"utils/nats"
	"utils/rpc"

	_ "sms/senders"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts service",
	Run: func(cmd *cobra.Command, args []string) {
		// initialize database
		db.Init(&conf.GetSettings().DB)

		settings := conf.GetSettings()

		sender, err := server.GetSender(settings.Sender)
		if err != nil {
			log.Fatal(err)
		}

		// start rpc server
		log.Info("Starting rpc server...")
		log.Info("Listen %s \n", settings.Rpc)
		grpcServer := rpc.Serve(settings.Rpc)

		// register SmsServer
		log.Info("Registering sms server...")
		sms.RegisterSmsServiceServer(
			grpcServer,
			server.NewSmsServer(
				sender,
				models.MakeNewSmsRepository(db.New()),
			),
		)
		nats.Init(&settings.Nats, true)

		cli.Terminate(nil)
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
