package fetcher

import (
	"accountstore/client"
	"common/log"
	"fetcher/models"
	"fmt"
	"proto/bot"
)

func parseOwnPosts(meta *client.AccountMeta) error {
	ig, err := meta.Delayed()
	if err != nil {
		return err
	}
	if ig.Debug {
		log.Debug("Parsing own posts for %v", ig.Username)
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
