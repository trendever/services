package project

import (
	"core/api"
	"core/conf"
	"core/models"
	"core/qor"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"utils/coins"
	"utils/db"
	"utils/log"
	"utils/nats"
)

// Service with service entry points
type Service struct{}

// AutoMigrate adds new columns to database
func (s *Service) AutoMigrate(cli *cli.Context) {
	log.Info("Start migration")

	// init config
	conf.Init()
	// connect to database
	db.Init(&conf.GetSettings().DB)

	if cli.Bool("drop") {
		err := db.
			New().
			DropTableIfExists(qor.Models...).
			Error

		if err != nil {
			log.Fatal(err)
		}
	}

	err := db.New().AutoMigrate(qor.Models...).Error
	if err != nil {
		log.Fatal(err)
	}

	if err := models.Migrate(); err != nil {
		log.Fatal(err)
	}

	log.Info("Migration: success.")
}

// Run starts it all
func (s *Service) Run(cli *cli.Context) {
	conf.Init()
	settings := conf.GetSettings()
	if settings.Profiler.Web {
		go func() {
			log.Error(http.ListenAndServe(settings.Profiler.Addr, nil))
		}()
	}
	log.Info("Starting service")

	go log.PanicLogger(func() {
		// connect to database
		db.Init(&conf.GetSettings().DB)
		nats.Init(&conf.GetSettings().Nats, true)
		if err := models.LoadOrCreateSystemUser(); err != nil {
			log.Fatal(fmt.Errorf("Failed to load/create system user: %v", err))
		}

		// Initial web server
		r := gin.Default()
		qor.Init(r) //start qor

		if err := models.InitializeMonetization(); err != nil {
			log.Fatal(fmt.Errorf("Failed to load/create initial monetization plan: %v", err))
		}
		models.ReloadAnswers()

		// Start api
		api.Start()
		coins.SetGRPCCli(api.TrendcoinServiceClient)

		// Initial gin web server
		if err := r.Run(settings.AppHost); err != nil {
			log.Fatal(err)
		}
	})

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	ss := <-interrupt
	log.Warn("Service stopped: %v", ss)
}
