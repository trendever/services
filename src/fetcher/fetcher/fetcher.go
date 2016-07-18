package fetcher

import (
	"math/rand"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"fetcher/api"
	"fetcher/conf"
	"fetcher/db"
	"fetcher/models"

	"github.com/codegangsta/cli"
	"instagram_api"
	"utils/log"

	_ "github.com/lib/pq"
	"fetcher/views"
)

var (
	modelsList = []interface{}{
		&models.Activity{},
	}
)

type textField struct {
	userName string
	textType string
	comment  string
}

type ProjectService struct{}

// migrate
func (this *ProjectService) AutoMigrate(cli *cli.Context) error {
	// initialize database
	db.InitDB()
	defer db.DB.Close()

	if cli.Bool("drop") {
		err := db.DB.DropTableIfExists(modelsList...).Error
		if err != nil {
			return err
		}

		log.Warn("Drop Tables: success.")
	}

	err := db.DB.AutoMigrate(modelsList...).Error
	if err != nil {
		return err
	}

	log.Info("Migration: success.")

	return nil
}

// run fetching
func (this *ProjectService) Run() error {
	// initialize database
	db.InitDB()
	defer db.DB.Close()

	settings := conf.GetSettings()

	// init api
	api.Start()
	views.Init()

	rand.Seed(time.Now().Unix())

	// connections pool
	var apis []*instagram_api.Instagram

	// interrupt
	interrupt := make(chan os.Signal)
	signal.Notify(interrupt, os.Interrupt, os.Kill, syscall.SIGTERM)

	// open connection and append connections pool
	for _, user := range settings.Instagram.Users {
		api, err := instagram_api.NewInstagram(
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

		time.Sleep(500 * time.Millisecond)
	}

	// wait for terminating
	for {
		select {
		case <-interrupt:
			log.Warn("Cleanup and terminating...")
			os.Exit(0)
		}
	}

	return nil
}

// get activity
func getActivity(api *instagram_api.Instagram, rndTimeout int) {

	// little log
	log.Debug("Start getting with timeout: %v ms.", rndTimeout)

	for {
		// get recent activity
		ract, err := api.GetRecentActivity()
		if err != nil {
			log.Warn("Got error %v while fetching recent activitity with user %v", err, api.GetUserName())
			log.Fatal(err)
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
		time.Sleep(time.Duration(rndTimeout) * time.Millisecond)
	}
}

// fetch data and fill database model
func fetch(stories instagram_api.RecentActivityStories, mentionName string) {

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
	if ok := db.DB.NewRecord(act); ok {

		var count int

		// check by pk if record exist
		if err := db.DB.Model(&act).Where("pk = ?", act.Pk).Count(&count).Error; err == nil && count <= 0 {
			if err := db.DB.Create(&act).Error; err != nil {
				log.Error(err)
			} else {
				log.Debug("Add row: %v", act.Pk)
			}
		} else if err != nil {
			// COUNT(*) error
			log.Error(err)
		}
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
func generateTimeout(min, max int) int {
	return min + rand.Intn(max-min)
}
