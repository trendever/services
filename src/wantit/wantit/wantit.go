package wantit

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"wantit/api"
	"wantit/conf"

	"accountstore/client"
	"proto/accountstore"
	"proto/bot"
	"utils/log"
	"utils/mandible"
	"utils/nats"
	"utils/rpc"
)

type textField struct {
	userName string
	textType string
	comment  string
}

// ProjectService struct
type ProjectService struct{}

var (
	lastChecked = int64(0)
	pool        *client.AccountsPool
	settings    = conf.GetSettings()
	// @TODO products with big id will have more symbols
	codeRegexp     = regexp.MustCompile("t[a-z]+[0-9]{4}($|[^0-9])")
	avatarUploader = mandible.New(conf.GetSettings().MandibleURL)
)

// ResetLastChecked drops last checked id
func (svc *ProjectService) ResetLastChecked() error {
	return os.Remove(conf.GetSettings().LastCheckedFile)
}

// Run fetching
func (svc *ProjectService) Run() (err error) {

	rand.Seed(time.Now().Unix())
	api.Start()

	nats.Init(&settings.Nats, true)

	conn := rpc.Connect(settings.Instagram.StoreAddr)
	cli := accountstore.NewAccountStoreServiceClient(conn)
	pool, err = client.InitPoll(
		accountstore.Role_AuxPrivate, cli,
		nil, nil,
		&settings.Instagram.Settings,
		settings.Instagram.ResponseLogging,
	)
	if err != nil {
		return fmt.Errorf("failed to init acoounts pool: %v", err)
	}

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// we can safely start main loop even before registering apis
	// because every API request will take a connection from the connection pool
	// with getFreeApi()
	restoreLastChecked()
	go registerOrders()

	// wait for terminating
	<-interrupt
	saveLastChecked()
	log.Warn("Cleanup and terminating...")
	os.Exit(0)

	return nil
}

func restoreLastChecked() {
	bytes, err := ioutil.ReadFile(conf.GetSettings().LastCheckedFile)
	if err != nil {
		log.Error(err)
		return
	}

	res, err := strconv.ParseInt(string(bytes), 10, 64)
	if err != nil {
		log.Error(err)
		return
	}

	lastChecked = res
	log.Debug("Loaded last checked id: %v", lastChecked)
}

func saveLastChecked() {
	ioutil.WriteFile(conf.GetSettings().LastCheckedFile, []byte(strconv.FormatInt(lastChecked, 10)), 0644)
}

func registerOrders() {
	timeout, _ := time.ParseDuration(settings.Instagram.TimeoutMin)
	loopStarted := time.Now()

	for {
		// make some delays in case loops runs too fast
		// startup delay is OK
		if time.Since(loopStarted) < time.Second {
			time.Sleep(timeout)
		}
		loopStarted = time.Now()

		// Step #1: get new entries from fetcher
		res, err := retrieveActivities()

		if err != nil {
			log.Warn("RPC connection error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if len(res.Result) > 0 {
			log.Debug("Got %v results since %v", len(res.Result), lastChecked)
		}

		for _, activity := range res.Result {
			log.Debug("processing activity %+v", activity)
			var err error
			var retry bool

			if activity.Type == "thread" {
				retry, err = processThreadOrder(activity)
			} else {
				retry, err = processPotentialOrder(activity.MediaId, activity)
			}

			if err != nil {
				if retry {
					log.Error(err)
				} else {
					log.Warn(err.Error())
				}
			}

			if !retry {
				lastChecked = activity.Id
			} else {
				break
			}
		}
	}
}

func retrieveActivities() (*bot.RetrieveActivitiesReply, error) {
	ctx, cancel := rpc.DefaultContext()
	defer cancel()
	return api.FetcherClient.RetrieveActivities(ctx, &bot.RetrieveActivitiesRequest{
		Conds: []*bot.RetriveCond{
			{
				Role: bot.MentionedRole_Wantit,
				Type: []string{"mentioned", "direct"},
			},
			{
				Role: bot.MentionedRole_User,
				Type: []string{"commented", "direct", "thread"},
			},
		},
		AfterId: lastChecked,
		Limit:   100, //@CHECK this number
	})
}
