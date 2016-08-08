package cmd

import (
	"github.com/spf13/cobra"
	"proto/sms"
	"sms/conf"
	"sms/db"
	"sms/models"
	"sms/server"
	"utils/cli"
	"utils/log"
	"utils/rpc"

	_ "sms/senders"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts service",
	Run: func(cmd *cobra.Command, args []string) {
		// initialize database
		db.InitDB()
		defer db.DB.Close()

		settings := conf.GetSettings()

		db.DB.LogMode(settings.Debug)

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
				models.MakeNewSmsRepository(db.DB),
			),
		)
		cli.Terminate(nil)
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
