package main

import (
	"common/log"
	"common/proxy"
	"github.com/codegangsta/cli"
	"net/http"
	"os"
	"os/signal"
	"proto/core"
	"syscall"
	"utils/rpc"
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

	transport, err := proxy.TransportFromURL(settings.Proxy)
	log.Fatal(err)
	http.DefaultClient = &http.Client{
		Transport: transport,
	}

	userServer = core.NewUserServiceClient(rpc.Connect(settings.CoreServer))

	// init Telegram
	bot, err := InitBot(settings.Token, settings.Rooms)
	if err != nil {
		return err
	}

	// init api
	InitViews(bot)

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	log.Info("Successfully started")

	<-interrupt
	log.Info("Cleanup and terminating...")

	return nil
}
