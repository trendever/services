package fetcher

import (
	"fetcher/models"
	"instagram"
	"strings"
	"time"
	"utils/log"
)

type textField struct {
	userName string
	textType string
	comment  string
}

// get activity: fetch and parse instagram feed
func (w *Worker) getActivity() {

	// little log
	log.Debug("Start getting with timeout: %v", w.timeout)

	for {
		// get recent activity
		ract, err := w.api.GetRecentActivity()
		if err != nil {
			log.Warn("Got error %v while fetching recent activitity with user %v", err, w.api.GetUserName())
			time.Sleep(time.Second)
			continue
		}

		// fetch old stories
		for _, story := range append(ract.OldStories, ract.NewStories...) {
			err := fillFeed(story, w.api.GetUserName())
			if err != nil {
				log.Error(err)
			}
		}

		// sleep
		w.next()
	}
}

// fill database model
func fillFeed(stories instagram.RecentActivityStories, mentionName string) error {

	log.Debug("Fetching new story")

	// parse text field
	txt := parseText(stories.Args.Text)

	act := &models.Activity{
		Pk:           stories.Pk, // instagram's post primary key from json
		UserID:       stories.Args.ProfileID,
		UserImageURL: stories.Args.ProfileImage,

		MentionedUsername: mentionName,

		UserName: txt.userName,
		Type:     txt.textType,
		Comment:  txt.comment,
	}

	// check if Args.Media have items
	if len(stories.Args.Media) > 0 {
		act.MediaID = stories.Args.Media[0].ID
		act.MediaURL = stories.Args.Media[0].Image
	}

	return act.Save()
}

// parse Args.Text field
func parseText(text string) *textField {

	txt := &textField{
		userName: strings.Fields(text)[0],
	}

	switch {
	case strings.Contains(text, "liked your photo"):
		txt.textType = "likit already"
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
