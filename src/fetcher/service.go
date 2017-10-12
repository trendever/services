package main

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"

	"fetcher/conf"
	"fetcher/fetcher"
	"fetcher/models"
	"fetcher/views"

	"common/db"
	"common/log"
	"utils/nats"
	"utils/rpc"

	"github.com/codegangsta/cli"
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
	config := conf.GetSettings()
	db.Init(&config.DB)
	nats.Init(&conf.GetSettings().Nats, true)

	rand.Seed(time.Now().Unix())
	err := fetcher.Start()
	if err != nil {
		log.Fatal(err)
	}

	// init api
	views.Init(rpc.Serve(config.RPC))

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// wait for terminating
	<-interrupt
	log.Warn("Cleanup and terminating...")
	return nil
}
