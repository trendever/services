package fetcher

import (
	"accountstore/client"
	"errors"
	"fetcher/models"
	"fmt"
	"instagram"
	"proto/accountstore"
	"proto/bot"
	"utils/log"
	"utils/nats"
)

const DirectMessageSubject = "direct.new_message"

// direct activity: accept PM invites; read && parse them
func checkDirect(meta *client.AccountMeta) error {
	// get non-pending shiet
	// check which threads got updated since last time
	// get them

	var threads []models.ThreadInfo

	cursor := ""

collectLoop:
	for {
		ig, err := meta.Delayed()
		if err != nil {
			return err
		}
		resp, err := ig.Inbox(cursor)
		if err != nil {
			return err
		}

		if resp.PendingRequestsTotal > 0 {
			ig, err := meta.Delayed()
			if err != nil {
				return err
			}
			_, err = ig.DirectThreadApproveAll()
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
		err := processThread(meta, &threads[it])
		if err != nil {
			return err
		}
	}

	return nil
}

func processThread(meta *client.AccountMeta, info *models.ThreadInfo) error {
	var (
		threadID = info.ThreadID
		resp     *instagram.DirectThreadResponse
		msgs     []instagram.ThreadItem
	)
	log.Debug("Processing thread %v, last checked was %v", threadID, info.LastCheckedID)
	defer log.Debug("Processing thread %v end", threadID)

	// populate new messages
	cursor := ""
	for {
		ig, err := meta.Delayed()
		if err != nil {
			return err
		}
		resp, err = ig.DirectThread(threadID, cursor)
		if err != nil {
			return err
		}

		if info.LaterThan(resp.Thread.OldestCursor) {
			msgs = append(msgs, resp.Thread.Items...)
			cursor = resp.Thread.OldestCursor
			if !resp.Thread.HasOlder {
				break
			}
			continue
		}

		if info.LastCheckedID == resp.Thread.OldestCursor {
			msgs = append(msgs, resp.Thread.Items...)
			break
		}

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
				if err := fillDirect(message, &resp.Thread, meta, ""); err != nil {
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
				if err := fillDirect(relatedMedia, &resp.Thread, meta, comment); err != nil {
					return err
				}
				relatedMedia = nil
			}

			err := nats.StanPublish(DirectMessageSubject, &notify)
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
		if err := fillDirect(relatedMedia, &resp.Thread, meta, ""); err != nil {
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
func fillDirect(item *instagram.ThreadItem, thread *instagram.Thread, meta *client.AccountMeta, comment string) error {

	share := item.MediaShare
	ig := meta.Get()

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

	if meta.Role() == accountstore.Role_User && share.User.Pk != ig.UserID {
		// ignore media with someone else's posts for shops
		log.Debug("ignoring medaishare %v with foreign post", item.ItemID)
		return nil
	}

	log.Debug("Filling in new direct story")

	act := &models.Activity{
		Pk:                item.ItemID,
		UserID:            item.UserID,
		UserName:          username,
		UserImageURL:      share.User.ProfilePicURL,
		MentionedUsername: ig.Username,
		MentionedRole:     bot.MentionedRole(meta.Role()),
		Type:              "direct",
		Comment:           comment,
		MediaID:           share.ID,
		MediaURL:          fmt.Sprintf("https://instagram.com/p/%v/", share.Code),
		ThreadID:          thread.ThreadID,
	}
	return act.Create()
}

func CreateThread(inviter uint64, participants []uint64, caption, initMsg string) (threadID string, err error) {
	ig, found := global.usersPool.Get(inviter)
	bot, err := global.pubPool.GetRandom()
	participants = append(participants, bot.UserID)
	if err != nil {
		return
	}
	if !found {
		return "", AccountUnavailable
	}
	tid, _, err := ig.SendText(initMsg, participants...)
	if err != nil {
		return "", err
	}
	_, err = ig.DirectUpdateTitle(tid, caption)
	if err != nil {
		log.Errorf("set title for thread %v failed:", tid, err)
	}
	return tid, nil
}

func leaveAllThreads(meta *client.AccountMeta) error {
	ig, err := meta.Delayed()
	if err != nil {
		return err
	}

	cursor := ""
	for {
		resp, err := ig.Inbox(cursor)
		if err != nil {
			return err
		}

		if resp.PendingRequestsTotal > 0 {
			ig, err := meta.Delayed()
			if err != nil {
				return err
			}
			_, err = ig.DirectThreadApproveAll()
			return err
		}

		for _, thread := range resp.Inbox.Threads {
			if len(thread.Users) < 2 {
				continue
			}
			ig, err := meta.Delayed()
			if err != nil {
				return err
			}
			_, err = ig.DirectThreadAction(thread.ThreadID, instagram.ActionLeave)
			if err != nil {
				return err
			}
		}
		if !resp.Inbox.HasOlder {
			return nil
		}
		cursor = resp.Inbox.OldestCursor
	}

	return errors.New("unreachable point reached in leaveAllThreads()")
}
