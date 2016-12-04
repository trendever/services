package fetcher

import (
	"accountstore/client"
	"fetcher/models"
	"instagram"
	"strings"
	"utils/log"
)

type textField struct {
	userName string
	textType string
	comment  string
}

// get activity: fetch and parse instagram feed
func getActivity(meta *client.AccountMeta) error {
	ig, err := meta.Delayed()
	if err != nil {
		return err
	}
	// get recent activity
	ract, err := ig.GetRecentActivity()
	if err != nil {
		return err
	}

	// fetch old stories
	for _, story := range append(ract.OldStories, ract.NewStories...) {
		err := fillFeed(story, ig.Username)
		if err != nil {
			return err
		}
	}
	return nil
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

	return act.Create()
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
