package cmd

import (
	"github.com/spf13/cobra"
	"google.golang.org/grpc"
	"mail/config"
	"mail/mailers"
	"mail/models"
	"mail/server"
	"net"
	"proto/mail"
	"utils/db"
	"utils/log"
)

var startCmd = &cobra.Command{
	Use:   "start",
	Short: "Starts mail service",
	Run: func(cmd *cobra.Command, args []string) {
		config := config.Get()
		db.Init(&config.DB)
		log.Info("Starting mail microservice...")
		listener, err := net.Listen("tcp", config.RPC)
		if err != nil {
			log.Fatal(err)
		}

		s := grpc.NewServer()
		var mailer server.Mailer
		if config.Debug {
			mailer = mailers.NewMailcatcher(config.MailCatcher.Addr)
		} else {
			mailer = mailers.MakeNewMailgunMailer(
				config.MailGun.Domain,
				config.MailGun.APIKey,
				config.MailGun.PublicAPIKey,
			)
		}

		mail.RegisterMailServiceServer(s, server.MakeNewMailServer(mailer, models.MakeNewMailRepository(db.New())))
		s.Serve(listener)
	},
}

func init() {
	RootCmd.AddCommand(startCmd)
}
