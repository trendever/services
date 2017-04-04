package fetcher

import (
	"accountstore/client"
	"fetcher/models"
	"fmt"
	"proto/bot"
	"utils/log"
)

func parseOwnPosts(meta *client.AccountMeta) error {
	log.Debug("Parsing own posts for %v", meta.Get().Username)
	ig, err := meta.Delayed()
	if err != nil {
		return err
	}

	feed, err := ig.GetUserFeed(ig.UserID)
	if err != nil {
		return err
	}
	for _, story := range feed.Items {

		act := &models.Activity{
			Pk:       fmt.Sprintf("%v", story.Pk), // instagram's post primary key from json
			UserID:   story.User.Pk,
			MediaId:  story.ID,
			MediaURL: fmt.Sprintf("https://instagram.com/p/%v/", story.Code),

			MentionedUsername: meta.Get().Username,
			MentionedRole:     bot.MentionedRole(meta.Role()),

			UserName: meta.Get().Username,
			Type:     "ownfeed",
		}

		log.Error(act.Create())
	}

	return nil
}
