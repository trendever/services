package fetcher

import (
	"fetcher/models"
	"fmt"
	"instagram"
	"utils/db"
	"utils/log"
)

// direct activity: accept PM invites; read && parse them
func (w *Worker) directActivity() {

	for {

		err := w.checkNewMessages()
		if err != nil {
			log.Error(err)
		}

		w.next()
	}
}

func (w *Worker) checkNewMessages() error {

	// get non-pending shiet
	// check which threads got updated since last time
	// get them
	// @TODO: use cursors here and in feeds, but must be sufficient for now because it runs pretty often
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

func (w *Worker) processThread(threadID string) error {

	log.Debug("Processing thread %v", threadID)
	defer log.Debug("Processing thread %v end", threadID)

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

		log.Debug("Thread is from %v; got %v messages there", resp.Thread.Inviter.Username, len(resp.Thread.Items))

		// walk in reversed order
		for i := len(resp.Thread.Items) - 1; i >= 0; i-- {
			message := resp.Thread.Items[i]

			if info.GreaterOrEqual(message.ItemID) {
				log.Debug("Reached end of the new conversation (%v;%v;%v); exiting", threadID, message.ItemID, info.LastCheckedID)
				return nil
			}

			if message.ItemType == "media_share" && message.MediaShare != nil {
				log.Debug("Adding new mediaShare with ID=%v", message.MediaShare.ID)

				if err := w.fillDirect(message.MediaShare, threadID, message.Text); err != nil {
					return err
				}
			} else {
				log.Debug("Message with id %v (type %v) does not contain mediaShare", message.ItemID, message.ItemType)
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
			log.Debug("Reached end of the thread %v", threadID)
			break
		}

		cursor = resp.Thread.OldestCursor
	}

	return nil
}

// fill database model by direct message
func (w *Worker) fillDirect(share *instagram.MediaShare, threadID, text string) error {

	log.Debug("Filling in new direct story")

	act := &models.Activity{
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
	}
	return act.Save()
}

// SendDirectMsg sends response to the chat
func (w *Worker) SendDirectMsg(threadID, message string) error {

	_, err := w.api.BroadcastText(threadID, message)
	return err
}
