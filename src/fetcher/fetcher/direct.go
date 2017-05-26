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

const DirectNotifySubject = "direct.notify"

// direct activity: accept PM invites; read && parse them
func checkDirect(meta *client.AccountMeta) error {
	// get non-pending shiet
	// check which threads got updated since last time
	// get them

	threads := []models.ThreadInfo{}

	log.Debug("Checking direct for %v", meta.Get().Username)

	cursor := ""
	var upperTime int64

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

		if len(resp.Inbox.Threads) == 0 {
			return nil
		}

		upperTime = resp.Inbox.Threads[0].LastActivityAt

		for _, thread := range resp.Inbox.Threads {
			info, err := models.GetThreadInfo(thread.ThreadID, ig.UserID)
			if err != nil {
				return err
			}
			// ignore too old activity
			if thread.LastActivityAt/1000000 < meta.AddedAt {
				break collectLoop
			}
			// check if getting shiet is necessary
			if len(thread.Items) == 0 {
				return fmt.Errorf("Thread (id=%v) got 0 msgs, should be at least 1!", thread.ThreadID)
			}
			if thread.Items[0].ItemID == info.LastCheckedID && !thread.HasNewer {
				break collectLoop
			}
			threads = append(threads, info)
		}

		if !resp.Inbox.HasOlder {
			break
		}
		cursor = resp.Inbox.OldestCursor
		// process only last page in one call
		threads = threads[:0]
	}

	for it := len(threads) - 1; it >= 0; it-- {
		err := processThread(meta, &threads[it], upperTime)
		if err != nil {
			return err
		}
	}

	return nil
}

func processThread(meta *client.AccountMeta, info *models.ThreadInfo, upperTime int64) error {
	var (
		threadID = info.ThreadID
		resp     *instagram.DirectThreadResponse
		msgs     []instagram.ThreadItem
	)

	ig, err := meta.Delayed()
	if err != nil {
		return err
	}
	// populate new messages
	cursor := ""
	for {
		resp, err = ig.DirectThread(threadID, cursor)
		if err != nil {
			return err
		}

		items := resp.Thread.Items
		if info.LaterThan(resp.Thread.OldestCursor) && items[len(items)-1].Timestamp/1000000 >= meta.AddedAt {
			msgs = append(msgs, items...)
			cursor = resp.Thread.OldestCursor
			if !resp.Thread.HasOlder {
				break
			}
			continue
		}

		if info.LastCheckedID == resp.Thread.OldestCursor {
			msgs = append(msgs, items...)
			break
		}

		for it, msg := range items {
			if !info.LaterThan(msg.ItemID) || msg.Timestamp/1000000 < meta.AddedAt {
				msgs = append(msgs, items[:it]...)
				break
			}
		}
		break
	}

	log.Debug("Got %v new messages for %v from thread %v", len(msgs), meta.Get().Username, threadID)

	var relatedMedia *instagram.ThreadItem
	sourceID := meta.Get().UserID
	// in slice messages are placed from most new to the oldest, so we want to iterate in reverse order
	for it := len(msgs) - 1; it >= 0; it-- {
		message := &msgs[it]
		// temporarily ignore newest messages to prevent possible shading part of direct activity
		if message.Timestamp > upperTime {
			continue
		}

		switch message.ItemType {
		// such a special case for shares feels somewhat inconsistent
		// we can use notifications in wantit probably...
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

		case "media":
			notify := bot.DirectNotify{
				ThreadId:  threadID,
				MessageId: message.ItemID,
				UserId:    message.UserID,
				SourceId:  sourceID,
				Type:      bot.MessageType_Image,
				Data:      message.Media.ImageVersions2.Largest(),
			}
			err := nats.StanPublish(DirectNotifySubject, &notify)
			if err != nil {
				return fmt.Errorf("failed to send message notification via stan: %v", err)
			}

		case "text":
			notify := bot.DirectNotify{
				ThreadId:  threadID,
				MessageId: message.ItemID,
				UserId:    message.UserID,
				SourceId:  sourceID,
				Type:      bot.MessageType_Text,
				Data:      message.Text,
			}

			if relatedMedia != nil {
				comment := ""
				if relatedMedia.UserID == message.UserID {
					comment = message.Text
				}
				if err := fillDirect(relatedMedia, &resp.Thread, meta, comment); err != nil {
					return err
				}
				relatedMedia = nil
			}

			err := nats.StanPublish(DirectNotifySubject, &notify)
			if err != nil {
				return fmt.Errorf("failed to send message notification via stan: %v", err)
			}
		}
		// only if we have no media in progress
		if relatedMedia == nil {
			info.LastCheckedID = message.ItemID
			err := info.Save()
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
		info.LastCheckedID = msgs[0].ItemID
		err := info.Save()
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
		MentionedUsername: ig.Username,
		MentionedRole:     bot.MentionedRole(meta.Role()),
		Type:              "direct",
		Comment:           comment,
		MediaId:           share.ID,
		MediaURL:          fmt.Sprintf("https://instagram.com/p/%v/", share.Code),
		ThreadID:          thread.ThreadID,
	}
	return act.Create()
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
