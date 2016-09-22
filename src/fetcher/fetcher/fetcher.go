package fetcher

import (
	"github.com/codegangsta/cli"
	"instagram"
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"
	"utils/db"
	"utils/log"

	"fetcher/api"
	"fetcher/conf"
	"fetcher/models"
	"fetcher/views"
)

var modelsList = []interface{}{
	&models.Activity{},
	&models.Thread{},
}

type textField struct {
	userName string
	textType string
	comment  string
}

// ProjectService is fetcher service
type ProjectService struct{}

// AutoMigrate used models
func (ps *ProjectService) AutoMigrate(cli *cli.Context) error {
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

// Run fetching
func (ps *ProjectService) Run() error {
	db.Init(&conf.GetSettings().DB)

	settings := conf.GetSettings()

	// init api
	api.Start()
	views.Init()

	rand.Seed(time.Now().Unix())

	// connections pool
	var apis []*instagram.Instagram

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

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
		rndTimeout := generateTimeout(
			settings.Instagram.TimeoutMin,
			settings.Instagram.TimeoutMax,
		)

		go getActivity(api, rndTimeout)
		go directActivity(api, rndTimeout)
	}

	// wait for terminating
	<-interrupt
	log.Warn("Cleanup and terminating...")
	return nil
}

// get activity: fetch and parse instagram feed
func getActivity(api *instagram.Instagram, rndTimeout time.Duration) {

	// little log
	log.Debug("Start getting with timeout: %v ms.", rndTimeout)

	for {
		// get recent activity
		ract, err := api.GetRecentActivity()
		if err != nil {
			log.Warn("Got error %v while fetching recent activitity with user %v", err, api.GetUserName())
			time.Sleep(time.Second)
			continue
		}

		// fetch old stories
		for _, story := range ract.OldStories {
			fetch(story, api.GetUserName())
		}

		// fetch new stories
		for _, story := range ract.NewStories {
			fetch(story, api.GetUserName())
		}

		// sleep
		time.Sleep(rndTimeout)
	}
}

// fetch data and fill database model
func fetch(stories instagram.RecentActivityStories, mentionName string) {

	// parse text field
	txt := parseText(stories.Args.Text)

	act := &models.Activity{
		Pk:           stories.Pk, // instagram's post primary key from json
		UserID:       stories.Args.ProfileID,
		UserImageUrl: stories.Args.ProfileImage,

		MentionedUsername: mentionName,

		UserName: txt.userName,
		Type:     txt.textType,
		Comment:  txt.comment,
	}

	// check if Args.Media have items
	if len(stories.Args.Media) > 0 {
		act.MediaID = stories.Args.Media[0].ID
		act.MediaUrl = stories.Args.Media[0].Image
	}

	// write activity to DB
	// @TODO: just make pk primary key
	if ok := db.New().NewRecord(act); ok {

		var count int

		// check by pk if record exist
		err := db.New().Model(&act).Where("pk = ?", act.Pk).Count(&count).Error
		if err != nil {
			log.Error(err)
			return
		}

		if count >= 0 {
			// skipping dupe
			return
		}

		// now -- create
		err = db.New().Create(&act).Error
		if err != nil {
			log.Error(err)
		}

		log.Debug("Add row: %v", act.Pk)
	}
}

// parse Args.Text field
func parseText(text string) *textField {

	txt := &textField{
		userName: strings.Fields(text)[0],
	}

	switch {
	case strings.Contains(text, "liked your photo"):
		txt.textType = "liked_photo"
	case strings.Contains(text, "started following you"):
		txt.textType = "start_following"
	case strings.Contains(text, "took a photo of you"):
		txt.textType = "took_photo"
	case strings.Contains(text, "mentioned you in a comment:"):
		txt.textType = "mentioned"
		txt.comment = strings.Split(text, "mentioned you in a comment: ")[1]
	case strings.Contains(text, "commented:"):
		txt.textType = "commented"
		txt.comment = strings.Split(text, "commented: ")[1]
	}

	return txt
}

// get random timeout
func generateTimeout(min, max int) time.Duration {
	return time.Duration(min+rand.Intn(max-min)) * time.Millisecond
}
