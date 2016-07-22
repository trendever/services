package main

import (
	"os"
	"utils/log"

	"fetcher/fetcher"

	"github.com/codegangsta/cli"
)

func main() {

	svc := fetcher.ProjectService{}

	app := cli.NewApp()
	app.Name = "Ig Inbox"
	app.Usage = "instagram fetcher"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Run fetcher",
			Action: func(c *cli.Context) {
				if err := svc.Run(); err != nil {
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
				if err := svc.AutoMigrate(c); err != nil {
					log.Fatal(err)
				}
			},
		},
	}
	log.PanicLogger(func() {
		app.Run(os.Args)
	})
}
