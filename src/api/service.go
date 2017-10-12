package main

import (
	"common/log"
	"os"

	"github.com/codegangsta/cli"

	"api/cache"
	"api/conf"
	"api/project"
	"net/http"
	//_ "net/http/pprof" // wuut the; could it really slow down?
)

func main() {

	app := cli.NewApp()
	app.Name = "Trendever public api"
	app.Usage = "Shop"
	app.Version = "0.0.1"

	app.Commands = []cli.Command{
		{
			Name:  "start",
			Usage: "Start public api service",
			Action: func(c *cli.Context) {
				settings := conf.GetSettings()
				if settings.Profiler.Web {
					go func() {
						log.Error(http.ListenAndServe(settings.Profiler.Addr, nil))
					}()
				}
				log.Info("Start service")
				svc := project.Service{}

				if err := svc.Run(); err != nil {
					log.Fatal(err)
				}
			},
		},
		{
			Name:  "flushcache",
			Usage: "Flush cache",
			Action: func(c *cli.Context) {
				cache.Init()
				log.Warn("Starting cache flush")
				log.Fatal(cache.Flush())
				log.Info("Cache flushed")
			},
		},
	}

	log.PanicLogger(func() {
		app.Run(os.Args)
	})
}
