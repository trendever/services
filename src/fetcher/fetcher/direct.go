package fetcher

import (
	"fetcher/models"
	"fmt"
	"instagram"
	"sort"
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

	cursor := ""

outer:
	for {
		resp, err := w.api().Inbox(cursor)
		if err != nil {
			return err
		}

		if resp.PendingRequestsTotal > 0 {
			_, err := w.api().DirectThreadApproveAll()
			// do nothing else now
			return err
		}

		for _, thread := range resp.Inbox.Threads {

			info, err := models.GetThreadInfo(thread.ThreadID)
			if err != nil {
				return err
			}

			// check if getting shiet is necessary
			if len(thread.Items) != 1 {
				return fmt.Errorf("Thread (id=%v) got %v msgs, should be 1!", thread.ThreadID, len(thread.Items))
			}
			if thread.Items[0].ItemID == info.LastCheckedID {
				// thread is already crawled; no need to check more
				log.Debug("Skipping not changed threads")
				break outer
			} else {
				log.Debug("Top message differs from saved; do the crawl: %v %v", thread.Items[0].ItemID, info.LastCheckedID)
			}

			err = w.processThread(&info)
			if err != nil {
				return err
			}
		}

		if resp.Inbox.HasOlder {
			cursor = resp.Inbox.OldestCursor
			continue
		}

		break
	}

	return nil
}

func (w *Worker) processThread(info *models.ThreadInfo) error {

	var threadID = info.ThreadID

	log.Debug("Processing thread %v", threadID)
	defer log.Debug("Processing thread %v end", threadID)

	cursor := ""
	lastCrawledID := ""

outer:
	for { // range over thread pages
		resp, err := w.api().DirectThread(threadID, cursor)
		if err != nil {
			return err
		}

		log.Debug("Thread is from %v; got %v messages there", resp.Thread.Inviter.Username, len(resp.Thread.Items))

		msgs := resp.Thread.Items
		sort.Sort(msgs)

		for id, message := range msgs { // range over page messages
			log.Debug("Checking message with id=%v, lastCheckedID=%v", message.ItemID, info.LastCheckedID)
			lastCrawledID = message.ItemID

			if info.LaterThan(message.ItemID) {
				log.Debug("Reached end of the new conversation (%v); exiting", threadID)
				break outer
			}

			// only use messages that are cojoined with media link
			if message.ItemType == "media_share" && message.MediaShare != nil {
				log.Debug("Adding new mediaShare with ID=%v", message.MediaShare.ID)

				// comment is in next follow-up message
				// try to watch in next 2 messages because of service "Notified" msg
				var comment string
				if id-1 >= 0 {
					comment = followUpString(&message, &msgs[id-1])
				}
				if id-2 >= 0 && comment == "" {
					comment = followUpString(&message, &msgs[id-2])
				}

				if err := w.fillDirect(&message, &resp.Thread, comment); err != nil {
					return err
				}
			}
		}

		if !resp.Thread.HasOlder {
			log.Debug("Reached end of the thread %v", threadID)
			break
		}

		cursor = resp.Thread.OldestCursor
	}

	return models.SaveLastCheckedID(threadID, lastCrawledID)
}

func followUpString(mediaShare, followUp *instagram.ThreadItem) string {
	if followUp.ItemType == "text" && mediaShare.UserID == followUp.UserID {
		return followUp.Text
	}
	return ""
}

// fill database model by direct message
func (w *Worker) fillDirect(item *instagram.ThreadItem, thread *instagram.Thread, comment string) error {

	share := item.MediaShare

	// find username
	var username = ""
	for _, user := range thread.Users {
		if user.Pk == item.UserID {
			username = user.Username
			break
		}
	}

	if username == "" {
		return fmt.Errorf("Wut? Could not find username for userID=%v in thread=%v itemID=%v", item.UserID, thread.ThreadID, item.ItemID)
	}

	log.Debug("Filling in new direct story")

	act := &models.Activity{
		Pk:                item.ItemID,
		UserID:            item.UserID,
		UserName:          username,
		UserImageURL:      share.User.ProfilePicURL,
		MentionedUsername: w.api().GetUserName(),
		Type:              "direct",
		Comment:           comment,
		MediaID:           share.ID,
		MediaURL:          fmt.Sprintf("https://instagram.com/p/%v/", share.Code),
		ThreadID:          thread.ThreadID,
	}
	return act.Create()
}

// SendDirectMsg sends text to a new chat
func (w *Worker) SendDirectMsg(threadID, message string) error {

	_, err := w.api().BroadcastText(threadID, message)
	return err
}

// SendDirectMsgToUser sends text to user
func (w *Worker) SendDirectMsgToUser(userID int64, message string) (*instagram.SendTextRespone, error) {
	res, err := w.api().SendText(userID, message)
	return res, err
}
