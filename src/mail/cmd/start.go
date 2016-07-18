package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"proto/mail"
	"utils/log"
	"google.golang.org/grpc"
	"net"
	"mail/db"
	"mail/mailers"
	"mail/models"
	"mail/server"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts mail service",
	Run: func(cmd *cobra.Command, args []string) {
		db, err := db.GetPG(viper.GetString("db.config"))
		if err != nil {
			log.Fatal(err)
		}
		port := viper.GetString("port")
		host := viper.GetString("host")
		log.Info("Starting mail microservice on the port: %s \n", port)
		listener, err := net.Listen("tcp", host+":"+port)
		if err != nil {
			log.Fatal(err)
		}

		s := grpc.NewServer()
		var mailer server.Mailer
		if viper.GetBool("debug") {
			mailer = mailers.NewMailcatcher(viper.GetString("mailcatcher.addr"))
		} else {
			mailer = mailers.MakeNewMailgunMailer(
				viper.GetString("mailgun.domain"),
				viper.GetString("mailgun.apiKey"),
				viper.GetString("mailgun.publicApiKey"))
		}

		mail.RegisterMailServiceServer(s, server.MakeNewMailServer(mailer, models.MakeNewMailRepository(db)))
		s.Serve(listener)
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
