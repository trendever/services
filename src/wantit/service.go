package main

import (
	"utils/log"
	"os"

	"wantit/wantit"

	"github.com/codegangsta/cli"
)

func main() {

	svc := wantit.ProjectService{}

	app := cli.NewApp()
	app.Name = "Wantit"
	app.Usage = "Instagram wantit bot"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Run wantit",
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
