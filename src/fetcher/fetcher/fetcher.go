package fetcher

import (
	"math/rand"
	"time"
	"utils/log"

	"fetcher/conf"
	"instagram"
)

// Start starts main fetching duty
func Start() error {
	settings := conf.GetSettings()

	// to prevent service restart too quickly and thus compromise bot
	// also make sure config is ok and we don't get panic in future
	startTimeout, err := generateTimeout(settings)
	if err != nil {
		log.Fatal(err)
	}
	time.Sleep(startTimeout)

	// connections pool
	var apis []*instagram.Instagram

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

		fetcherWorker := &Worker{
			api:     api,
			timeout: rndTimeout,
		}

		fetcherWorker.start()
	}

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
