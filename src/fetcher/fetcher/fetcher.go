package fetcher

import (
	"math/rand"
	"os"
	"os/signal"
	"syscall"
	"time"
	"utils/db"
	"utils/log"

	"fetcher/api"
	"fetcher/conf"
	"fetcher/models"
	"fetcher/views"
	"instagram"

	"github.com/codegangsta/cli"
)

var modelsList = []interface{}{
	&models.Activity{},
	&models.ThreadInfo{},
}

type textField struct {
	userName string
	textType string
	comment  string
}

// AutoMigrate used models
func AutoMigrate(cli *cli.Context) error {
	// initialize database
	db.Init(&conf.GetSettings().DB)

	if cli.Bool("drop") {
		err := db.New().DropTableIfExists(modelsList...).Error
		if err != nil {
			return err
		}

		log.Warn("Drop Tables: success.")
	}

	err := db.New().AutoMigrate(modelsList...).Error
	if err != nil {
		return err
	}

	log.Info("Migration: success.")

	return nil
}

// Run main stuff
func Run() error {
	db.Init(&conf.GetSettings().DB)

	settings := conf.GetSettings()

	// to prevent service restart too quickly and thus compromise bot
	// also make sure config is ok and we don't get panic in future
	startTimeout, err := generateTimeout(settings)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(startTimeout)

	// init api
	api.Start()
	views.Init()

	rand.Seed(time.Now().Unix())

	// connections pool
	var apis []*instagram.Instagram

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	pool.Lock()

	// open connection and append connections pool
	for _, user := range settings.Instagram.Users {
		api, err := instagram.NewInstagram(
			user.Username,
			user.Password,
		)
		if err != nil {
			log.Warn("Failed to log-in with user %v: %v", user.Username, err)
			return err
		}
		apis = append(apis, api)
	}

	// run goroutine
	for _, api := range apis {

		// random timeout
		rndTimeout, err := generateTimeout(settings)
		if err != nil {
			log.Fatal(err)
		}

		fetcherWorker := &worker{
			api:     api,
			timeout: rndTimeout,
		}

		//	go fetcherWorker.getActivity()
		go fetcherWorker.directActivity()
	}

	pool.Unlock()

	// wait for terminating
	<-interrupt
	log.Warn("Cleanup and terminating...")
	return nil
}

// get random timeout
func generateTimeout(settings *conf.Settings) (time.Duration, error) {

	min, err := time.ParseDuration(settings.Instagram.TimeoutMin)
	if err != nil {
		return time.Duration(0), err
	}
	max, err := time.ParseDuration(settings.Instagram.TimeoutMax)
	if err != nil {
		return time.Duration(0), err
	}

	return min + time.Duration(rand.Intn(int(max-min))), nil
}
