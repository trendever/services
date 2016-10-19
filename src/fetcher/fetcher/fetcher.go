package fetcher

import (
	"fetcher/conf"
	"instagram"
	"time"
	"utils/log"
)

// Start starts main fetching duty
func Start() error {
	settings := conf.GetSettings()

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

	// parse timeouts
	min, err := time.ParseDuration(settings.Instagram.TimeoutMin)
	if err != nil {
		return err
	}
	max, err := time.ParseDuration(settings.Instagram.TimeoutMax)
	if err != nil {
		return err
	}

	// run goroutine
	for _, api := range apis {

		pool := instagram.NewPool(&instagram.PoolSettings{
			TimeoutMin:     int(min / time.Millisecond),
			TimeoutMax:     int(max / time.Millisecond),
			ReloginTimeout: int(time.Second * 10 / time.Millisecond),
		})

		pool.Add(api)

		fetcherWorker := &Worker{
			pool:     pool,
			username: api.GetUserName(),
		}

		fetcherWorker.start()
	}

	return nil
}
