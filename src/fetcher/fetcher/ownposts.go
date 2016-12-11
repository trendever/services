package fetcher

import (
	"accountstore/client"
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

	_ = feed
	return nil
}
