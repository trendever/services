package main

import (
	"common/log"
	"os"

	"savetrend/saver"

	"github.com/codegangsta/cli"
)

func main() {

	svc := saver.ProjectService{}

	app := cli.NewApp()
	app.Name = "Ig Inbox"
	app.Usage = "Instagram savertrend bot"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Run saver",
			Action: func(c *cli.Context) {
				if err := svc.Run(); err != nil {
					log.Fatal(err)
				}
			},
		},
		{
			Name:  "migrate",
			Usage: "Migrate Database",
			Action: func(c *cli.Context) {
				if c.Bool("drop") {
					if err := svc.ResetLastChecked(); err != nil {
						log.Fatal(err)
					}
				}
			},
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "drop, d",
				},
			},
		},
	}
	log.PanicLogger(func() {
		app.Run(os.Args)
	})
}
