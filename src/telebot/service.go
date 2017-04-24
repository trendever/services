package main

import (
	"os"
	"os/signal"
	"proto/core"
	"syscall"
	"utils/log"
	"utils/rpc"

	"github.com/codegangsta/cli"
)

type projectService struct{}

var userServer core.UserServiceClient

func main() {

	svc := projectService{}

	app := cli.NewApp()
	app.Name = "Telegram bot"
	app.Usage = "Telegram notify bot"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Run telebot",
			Action: func(c *cli.Context) {
				if err := svc.run(); err != nil {
					log.Fatal(err)
				}
			},
		},
	}
	log.PanicLogger(func() {
		if err := app.Run(os.Args); err != nil {
			log.Fatal(err)
		}
	})
}

// Run bot
func (svc *projectService) run() error {

	settings := GetSettings()
	userServer = core.NewUserServiceClient(rpc.Connect(settings.CoreServer))

	// init Telegram
	t, err := InitBot(settings.Token, settings.Rooms)
	if err != nil {
		return err
	}

	// init api
	InitViews(t)

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	log.Info("Successfully started")

	<-interrupt
	log.Info("Cleanup and terminating...")

	return nil
}
