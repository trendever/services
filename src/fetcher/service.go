package main

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fetcher/api"
	"fetcher/conf"
	"fetcher/fetcher"
	"fetcher/models"
	"fetcher/views"

	"utils/db"
	"utils/log"

	"github.com/codegangsta/cli"
	"utils/nats"
)

func main() {

	app := cli.NewApp()
	app.Name = "Ig Inbox"
	app.Usage = "instagram fetcher"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Run fetcher",
			Action: func(c *cli.Context) {
				if err := Run(); err != nil {
					log.Fatal(err)
				}
			},
		},
		{
			Name:  "migrate",
			Usage: "Migrate database",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "drop, d",
				},
			},
			Action: func(c *cli.Context) {
				if err := models.AutoMigrate(c.Bool("drop")); err != nil {
					log.Fatal(err)
				}
			},
		},
	}
	log.PanicLogger(func() {
		app.Run(os.Args)
	})
}

// Run main stuff
func Run() error {
	db.Init(&conf.GetSettings().DB)

	// init api
	api.Start()
	nats.Init(&conf.GetSettings().Nats, true)
	views.Init()

	rand.Seed(time.Now().Unix())

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	err := fetcher.Start()
	if err != nil {
		log.Fatal(err)
	}

	// wait for terminating
	<-interrupt
	log.Warn("Cleanup and terminating...")
	return nil
}
