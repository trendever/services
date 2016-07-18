package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"proto/sms"
	"utils/cli"
	"utils/log"
	"utils/rpc"
	"sms/conf"
	"sms/db"
	"sms/models"
	"sms/senders"
	"sms/server"
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

		// start rpc server
		log.Info("Starting rpc server...")
		addr := fmt.Sprintf("%v:%v", settings.RPC.Host, settings.RPC.Port)
		log.Info("Listen %s \n", addr)
		grpcServer := rpc.Serve(addr)

		// register SmsServer
		log.Info("Registering sms server...")
		sms.RegisterSmsServiceServer(
			grpcServer,
			server.NewSmsServer(
				senders.NewMTSClient(settings.MTS.Login, settings.MTS.Password, settings.MTS.Naming, settings.MTS.Rates),
				models.MakeNewSmsRepository(db.DB)))

		cli.Terminate(nil)
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
