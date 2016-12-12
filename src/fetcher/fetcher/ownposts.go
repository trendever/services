package fetcher

import (
	"accountstore/client"
	"fetcher/models"
	"fmt"
	"proto/bot"
	"utils/log"
)

func parseOwnPosts(meta *client.AccountMeta) error {

	ig, err := meta.Delayed()
	if err != nil {
		return err
	}

	feed, err := ig.GetUserFeed(ig.UserID)
	if err != nil {
		return err
	}

	for _, story := range feed.Items {
		log.Debug("Parsing own post")

		act := &models.Activity{
			Pk:       fmt.Sprintf("%v", story.Pk), // instagram's post primary key from json
			UserID:   story.User.Pk,
			MediaID:  story.ID,
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
