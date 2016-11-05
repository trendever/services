package fetcher

import (
	"fetcher/models"
	"fmt"
	"instagram"
	"proto/bot"
	"utils/log"
	"utils/nats"
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

	var threads []models.ThreadInfo

	cursor := ""

collectLoop:
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
			if len(thread.Items) == 0 {
				return fmt.Errorf("Thread (id=%v) got 0 msgs, should be at least 1!", thread.ThreadID)
			}
			if thread.Items[0].ItemID == info.LastCheckedID && !thread.HasNewer {
				log.Debug("Unchanged thread is reached")
				break collectLoop
			}
			threads = append(threads, info)
		}

		if !resp.Inbox.HasOlder {
			break
		}
		cursor = resp.Inbox.OldestCursor
	}

	for it := len(threads) - 1; it >= 0; it-- {
		err := w.processThread(&threads[it])
		if err != nil {
			return err
		}
	}

	return nil
}

func (w *Worker) processThread(info *models.ThreadInfo) error {
	var (
		threadID = info.ThreadID
		resp     *instagram.DirectThreadResponse
		msgs     []instagram.ThreadItem
		err      error
	)
	log.Debug("Processing thread %v, last checked was %v", threadID, info.LastCheckedID)
	defer log.Debug("Processing thread %v end", threadID)

	// populate new messages
	cursor := ""
	for {
		resp, err = w.api().DirectThread(threadID, cursor)
		if err != nil {
			return err
		}

		if info.LaterThan(resp.Thread.OldestCursor) {
			log.Debug("last checked is older then oldest cursor")
			msgs = append(msgs, resp.Thread.Items...)
			cursor = resp.Thread.OldestCursor
			if !resp.Thread.HasOlder {
				log.Debug("Reached end of the thread %v", threadID)
				break
			}
			continue
		}

		if info.LastCheckedID == resp.Thread.OldestCursor {
			log.Debug("last checked match oldest cursor")
			msgs = append(msgs, resp.Thread.Items...)
			break
		}

		log.Debug("last checked should be somewhere in middle")
		for it, msg := range resp.Thread.Items {
			if !info.LaterThan(msg.ItemID) {
				msgs = append(msgs, resp.Thread.Items[:it]...)
				break
			}
		}
		break
	}

	log.Debug("Thread is from %v; got %v new messages there", resp.Thread.Inviter.Username, len(msgs))

	var relatedMedia *instagram.ThreadItem
	// in slice messages are placed from most new to the oldest, so we want to iterate in reverse order
	for it := len(msgs) - 1; it >= 0; it-- {
		message := &msgs[it]
		log.Debug("Checking message with id=%v", message.ItemID)

		switch message.ItemType {
		case "media_share":
			// there was older media without comment
			if relatedMedia != nil {
				if err := w.fillDirect(message, &resp.Thread, ""); err != nil {
					return err
				}
			}
			if message.MediaShare != nil {
				relatedMedia = message
			} else {
				log.Errorf("message %v with type 'media_share' has empty media", message.ItemID)
				relatedMedia = nil
			}

		case "text":
			notify := bot.DirectMessageNotify{
				ThreadId:  threadID,
				MessageId: message.ItemID,
				UserId:    message.UserID,
				Text:      message.Text,
			}

			if relatedMedia != nil {
				comment := ""
				if relatedMedia.UserID == message.UserID {
					comment = message.Text
					notify.RelatedMedia = relatedMedia.MediaShare.ID
				}
				if err := w.fillDirect(relatedMedia, &resp.Thread, comment); err != nil {
					return err
				}
				relatedMedia = nil
			}

			err := nats.StanPublish("direct.new_message", &notify)
			if err != nil {
				return fmt.Errorf("failed to send message notification via stan: %v", err)
			}
		}
		// only if we have no media in progress
		if relatedMedia == nil {
			err := models.SaveLastCheckedID(threadID, message.ItemID)
			if err != nil {
				return err
			}
		}
	}

	// some unfinished stuff
	if relatedMedia != nil {
		if err := w.fillDirect(relatedMedia, &resp.Thread, ""); err != nil {
			return err
		}
		err := models.SaveLastCheckedID(threadID, msgs[0].ItemID)
		if err != nil {
			return err
		}
	}

	return nil
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
		//this is media_share sent by our account
		//ignore it silently
		log.Debug("Wut? Could not find username for userID=%v in thread=%v itemID=%v", item.UserID, thread.ThreadID, item.ItemID)
		return nil
	}

	log.Debug("Filling in new direct story")

	act := &models.Activity{
		Pk:                item.ItemID,
		UserID:            item.UserID,
		UserName:          username,
		UserImageURL:      share.User.ProfilePicURL,
		MentionedUsername: w.username,
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
func (w *Worker) SendDirectMsgToUser(userID uint64, message string) (*instagram.SendTextResponse, error) {
	res, err := w.api().SendText(userID, message)
	return res, err
}
