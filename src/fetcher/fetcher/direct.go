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
	ig, err := meta.Delayed()
	if err != nil {
		return err
	}
	threadID := info.ThreadID

	resp, msgs, err := loadThread(meta, threadID, info.LastCheckedID)
	if err != nil {
		return err
	}
	// strip checked message
	if len(msgs) > 0 && msgs[len(msgs)-1].ItemID == info.LastCheckedID {
		msgs = msgs[:len(msgs)-1]
	}

	log.Debug("Got %v new messages for %v from thread %v", len(msgs), meta.Get().Username, threadID)
	// @TODO send one notify for multiple messages where possible
	sourceID := meta.Get().UserID
	// in slice messages are placed from most new to the oldest, so we want to iterate in reverse order
	for it := len(msgs) - 1; it >= 0; it-- {
		message := &msgs[it]
		// temporarily ignore newest messages to prevent possible shading part of direct activity
		if message.Timestamp > upperTime {
			continue
		}

		if info.LastCheckedID == "" { // new untracked thread
			if err := addThreadActivity(message, &resp.Thread, meta); err != nil {
				return err
			}
		}

		switch message.ItemType {
		// such a special case for shares feels somewhat inconsistent
		// we can use notifications in wantit probably...
		case "media_share":
			if meta.Role() == accountstore.Role_User &&
				message.MediaShare != nil &&
				message.MediaShare.User.Pk == ig.UserID {
				if err := addDirectActivity(message, &resp.Thread, meta, ""); err != nil {
					return err
				}
			}

		case "media":
			notify := bot.Notify{
				ThreadId: threadID,
				SourceId: sourceID,
				Messages: []*bot.Message{
					{
						MessageId: message.ItemID,
						UserId:    message.UserID,
						Type:      bot.MessageType_Image,
						Data:      message.Media.ImageVersions2.Largest(),
					},
				},
			}
			err := nats.StanPublish(DirectNotifySubject, &notify)
			if err != nil {
				return fmt.Errorf("failed to send message notification via stan: %v", err)
			}

		case "link":
			message.Text = message.Link.Text
			fallthrough

		case "text":
			notify := bot.Notify{
				ThreadId: threadID,
				SourceId: sourceID,
				Messages: []*bot.Message{
					{
						MessageId: message.ItemID,
						UserId:    message.UserID,
						Type:      bot.MessageType_Text,
						Data:      message.Text,
					},
				},
			}
			err := nats.StanPublish(DirectNotifySubject, &notify)
			if err != nil {
				return fmt.Errorf("failed to send message notification via stan: %v", err)
			}
		}
		info.LastCheckedID = message.ItemID
		err := info.Save()
		if err != nil {
			return err
		}
	}

	return nil
}

// populate new messages from direct thread starting with message with passed ID(inclusive)
// returned instagram.DirectThreadResponse is from last page query
func loadThread(meta *client.AccountMeta, threadID, sinceID string) (thread *instagram.DirectThreadResponse, msgs []instagram.ThreadItem, err error) {
	ig := meta.Get()
	cursor := ""
	for {
		thread, err = ig.DirectThread(threadID, cursor)
		if err != nil {
			return
		}

		items := thread.Thread.Items
		if models.CompareID(thread.Thread.OldestCursor, sinceID) > 0 && items[len(items)-1].Timestamp/1000000 >= meta.AddedAt {
			msgs = append(msgs, items...)
			cursor = thread.Thread.OldestCursor
			if !thread.Thread.HasOlder {
				return
			}
			continue
		}

		if sinceID == thread.Thread.OldestCursor {
			msgs = append(msgs, items...)
			return
		}

		for it, msg := range items {
			if models.CompareID(msg.ItemID, sinceID) < 0 || msg.Timestamp/1000000 < meta.AddedAt {
				msgs = append(msgs, items[:it]...)
				break
			}
		}
		return
	}
}

func getEncodedThread(meta *client.AccountMeta, threadID, since string) (ret []*bot.Message, err error) {
	_, msgs, err := loadThread(meta, threadID, since)
	ret = []*bot.Message{}
	// in slice messages are placed from most new to the oldest, so we want to iterate in reverse order
	for it := len(msgs) - 1; it >= 0; it-- {
		message := &msgs[it]

		switch message.ItemType {
		case "media_share":
			ret = append(ret, &bot.Message{
				MessageId: message.ItemID,
				UserId:    message.UserID,
				Type:      bot.MessageType_MediaShare,
				Data:      message.MediaShare.ID,
			})

		case "media":
			ret = append(ret, &bot.Message{
				MessageId: message.ItemID,
				UserId:    message.UserID,
				Type:      bot.MessageType_Image,
				Data:      message.Media.ImageVersions2.Largest(),
			})

		case "link":
			message.Text = message.Link.Text
			fallthrough

		case "text":
			ret = append(ret, &bot.Message{
				MessageId: message.ItemID,
				UserId:    message.UserID,
				Type:      bot.MessageType_Text,
				Data:      message.Text,
			})
		}
	}
	return ret, nil
}

func addThreadActivity(item *instagram.ThreadItem, thread *instagram.Thread, meta *client.AccountMeta) error {
	// process private threads only
	if thread.ThreadType != "private" || meta.Role() != accountstore.Role_User {
		return nil
	}

	act := &models.Activity{
		Pk:                item.ItemID,
		UserID:            item.UserID,
		UserName:          thread.Users[0].Username,
		MentionedUsername: meta.Get().Username,
		MentionedRole:     bot.MentionedRole(meta.Role()),
		Type:              "thread",
		ThreadID:          fmt.Sprintf("%v#%v", thread.ThreadID, item.ItemID),
	}
	return act.Create()
}

// fill database model by direct message
func addDirectActivity(item *instagram.ThreadItem, thread *instagram.Thread, meta *client.AccountMeta, comment string) error {
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
		ThreadID:          fmt.Sprintf("%v#%v", thread.ThreadID, item.ItemID),
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
