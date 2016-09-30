package fetcher

import (
	"fetcher/models"
	"fmt"
	"instagram"
	"utils/db"
	"utils/log"
)

// direct activity: accept PM invites; read && parse them
func (w *worker) directActivity() {

	for {

		err := w.checkNewMessages()
		if err != nil {
			log.Error(err)
		}

		w.next()
	}
}

func (w *worker) checkNewMessages() error {

	// get non-pending shiet
	// check which threads got updated since last time
	// get them
	resp, err := w.api.Inbox("")
	if err != nil {
		return err
	}

	if resp.PendingRequestsTotal > 0 {
		_, err := w.api.DirectThreadApproveAll()
		// do nothing else now
		return err
	}

	for _, thread := range resp.Inbox.Threads {

		err := w.processThread(thread.ThreadID)
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *worker) processThread(threadID string) error {

	info, err := models.GetThreadInfo(threadID)
	if err != nil {
		return err
	}

	cursor := ""

	for {
		resp, err := w.api.DirectThread(threadID, cursor)
		if err != nil {
			return err
		}

		for _, message := range resp.Thread.Items {

			if info.GreaterOrEqual(message.ItemID) {
				log.Debug("Reached end of the new conversation (%v;%v); exiting", threadID, message.ItemID)
				return nil
			}

			if message.ItemType == "media_share" && message.MediaShare != nil {
				log.Debug("Adding new mediaShare with ID=%v", message.MediaShare.ID)

				if err := w.fillDirect(message.MediaShare, threadID, message.Text); err != nil {
					return err
				}
			}

			info.LastCheckedID = message.ItemID
			err := db.New().
				Model(&models.ThreadInfo{}).
				Where("thread_id = ?", threadID).
				Update("last_checked_id", info.LastCheckedID).
				Error
			if err != nil {
				return err
			}
		}

		if !resp.Thread.HasOlder {
			break
		}

		cursor = resp.Thread.OldestCursor
	}

	return nil
}

// fill database model by direct message
func (w *worker) fillDirect(share *instagram.MediaShare, threadID, text string) error {

	log.Debug("Filling in new direct story")

	return saveActivity(&models.Activity{
		Pk:                fmt.Sprintf("%v", share.Pk),
		UserID:            share.User.Pk,
		UserImageURL:      share.User.ProfilePicURL,
		MentionedUsername: w.api.GetUserName(),
		UserName:          share.User.Username,
		Type:              "direct",
		Comment:           text,
		MediaID:           share.ID,
		MediaURL:          fmt.Sprintf("https://instagram.com/p/%v/", share.Code),
		ThreadID:          threadID,
	})
}
