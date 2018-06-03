package fetcher

import (
	"accountstore/client"
	"common/log"
	"errors"
	"fetcher/models"
	"fmt"
	"instagram"
	"proto/accountstore"
	"proto/bot"
	"time"
	"utils/nats"
)

const DirectNotifySubject = "direct.notify"
const ThreadsRerCheck = 30

// any direct activity happened before this timeout will be ignored
const DirectExpirationTimeout = time.Hour * 24 * 7

// direct activity: accept PM invites; read && parse them
func checkDirect(meta *client.AccountMeta) error {
	// get non-pending shiet
	// check which threads got updated since last time
	// get them

	threads := []models.ThreadInfo{}

	ig := meta.Get()
	if ig.Debug {
		log.Debug("Checking direct for %v", ig.Username)
	}

	cursor := ""
	var upperTime int64

	cutTime := time.Now().Add(-DirectExpirationTimeout).Unix()
	if meta.AddedAt > cutTime {
		cutTime = meta.AddedAt
	}

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

		// we will not process messages created after first current result to avoid possible shadowing in later calls
		if cursor == "" {
			upperTime = resp.Inbox.Threads[0].LastActivityAt
		}

		for _, thread := range resp.Inbox.Threads {
			info, err := models.GetThreadInfo(thread.ThreadID, ig.UserID)
			if err != nil {
				return err
			}
			// ignore too old activity
			if thread.LastActivityAt/1000000 < cutTime {
				break collectLoop
			}
			// check if getting shiet is necessary
			if len(thread.Items) == 0 {
				return fmt.Errorf("Thread (id=%v) got 0 msgs, should be at least 1!", thread.ThreadID)
			}
			if models.CompareID(thread.Items[0].ItemID, info.LastCheckedID) <= 0 && !thread.HasNewer {
				break collectLoop
			}
			threads = append(threads, info)
		}

		// limit amount of threads for single call
		if len(threads) > ThreadsRerCheck {
			threads = threads[len(threads)-ThreadsRerCheck:]
		}
		if !resp.Inbox.HasOlder {
			break
		}
		cursor = resp.Inbox.OldestCursor
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

	if len(msgs) == 0 {
		return nil
	}

	log.Debug("Got %v new messages for %v from thread %v", len(msgs), meta.Get().Username, threadID)
	sourceID := meta.Get().UserID
	notify := bot.Notify{
		ThreadId: threadID,
		SourceId: sourceID,
	}
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

		if message.ItemType == "media_share" {
			share := message.MediaShare
			if share == nil {
				share = &message.DirectShare.Media
			}
			if meta.Role() != accountstore.Role_User || share.User.Pk == ig.UserID {
				if err := addDirectActivity(message, share, &resp.Thread, meta); err != nil {
					return err
				}
			}
		}
		notify.Messages = append(notify.Messages, MapFromInstagram(message)...)
		info.LastCheckedID = message.ItemID
	}
	if len(notify.Messages) != 0 {
		err = nats.StanPublish(DirectNotifySubject, &notify)
		if err != nil {
			return fmt.Errorf("failed to send message notification via stan: %v", err)
		}
	}

	err = info.Save()
	if err != nil {
		return err
	}

	return nil
}

// populate new messages from direct thread starting with message with passed ID(inclusive)
// returned instagram.DirectThreadResponse is from last page query
func loadThread(meta *client.AccountMeta, threadID, sinceID string) (thread *instagram.DirectThreadResponse, msgs []instagram.ThreadItem, err error) {
	ig := meta.Get()
	// limit loading depth by time
	cutTime := time.Now().Add(-DirectExpirationTimeout).Unix()
	if meta.AddedAt > cutTime {
		cutTime = meta.AddedAt
	}
	cursor := ""
	for {
		thread, err = ig.DirectThread(threadID, cursor)
		if err != nil {
			return
		}

		items := thread.Thread.Items
		// empty page?.. that happens sometimes
		if len(items) == 0 {
			log.Debug("got page without items for direct thread %v with cursor %v", threadID, cursor)
			return
		}
		if models.CompareID(thread.Thread.OldestCursor, sinceID) > 0 && items[len(items)-1].Timestamp/1000000 >= cutTime {
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
			if models.CompareID(msg.ItemID, sinceID) < 0 || msg.Timestamp/1000000 < cutTime {
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
		ret = append(ret, MapFromInstagram(message)...)
	}
	return ret, nil
}

func MapFromInstagram(msg *instagram.ThreadItem) (ret []*bot.Message) {
	type mapped struct {
		kind bot.MessageType
		data string
	}
	for i, item := range (func(msg *instagram.ThreadItem) []mapped {
		switch msg.ItemType {
		case "media_share":
			switch {
			case msg.MediaShare != nil:
				return []mapped{{bot.MessageType_MediaShare, msg.MediaShare.ID}}
			case msg.DirectShare != nil:
				ret := []mapped{{bot.MessageType_MediaShare, msg.DirectShare.Media.ID}}
				if msg.DirectShare.Text != "" {
					ret = append(ret, mapped{bot.MessageType_Text, msg.DirectShare.Text})
				}
				return ret
			default:
				log.Errorf("item %s do not have normal medai share nor direct one", msg.ItemID)
				return nil
			}
		case "media":
			return []mapped{{bot.MessageType_Image, msg.Media.ImageVersions2.Largest().URL}}
		case "text":
			return []mapped{{bot.MessageType_Text, msg.Text}}
		case "link":
			return []mapped{{bot.MessageType_Text, msg.Link.Text}}
		case "like":
			return []mapped{{bot.MessageType_Text, "❤️"}}
		case "profile":
			// we could determinate our local user(if any) probably, but is it necessary?
			return []mapped{{bot.MessageType_Text, fmt.Sprintf("https://www.instagram.com/%s/", msg.Profile.Username)}}
		case "location":
			return []mapped{{bot.MessageType_Text, fmt.Sprintf("https://www.instagram.com/explore/locations/%v/", msg.Location.ID)}}
		case "hashtag":
			return []mapped{{bot.MessageType_Text, fmt.Sprintf("https://www.instagram.com/explore/tags/%s/", msg.HashTag.Name)}}
		case "action_log":
			// @CHECK could there be something useful? afaik it contains topic changes and join|leave notifies
			return nil
		case "reel_share":
			var data string
			if msg.ReelShare.Media.ExpiringAt == 0 {
				data = "*media expired*"
			} else {
				if msg.ReelShare.Media.MediaType == instagram.MediaType_Video {
					data = msg.ReelShare.Media.VideoVersions[0].URL
				} else {
					data = msg.ReelShare.Media.ImageVersions2.Largest().URL
				}
			}
			if msg.ReelShare.Text != "" {
				data += "\n" + msg.ReelShare.Text
			}
			return []mapped{{bot.MessageType_Text, data}}
		case "placeholder":
			return []mapped{{bot.MessageType_System, msg.Placeholder.Message}}
		default:
			log.Debug("unknown type of direct item: %v", msg.ItemType)
			return nil
		}
	}(msg)) {
		id := msg.ItemID
		if i != 0 {
			id = fmt.Sprintf("%s#%d", id, i)
		}
		ret = append(ret, &bot.Message{
			MessageId: id,
			UserId:    msg.UserID,
			Type:      item.kind,
			Data:      item.data,
		})
	}
	return ret
}

func addThreadActivity(item *instagram.ThreadItem, thread *instagram.Thread, meta *client.AccountMeta) error {
	// process private threads only
	if len(thread.Users) != 1 || meta.Role() != accountstore.Role_User {
		return nil
	}

	act := &models.Activity{
		Pk:                item.ItemID,
		UserID:            thread.Users[0].Pk,
		UserName:          thread.Users[0].Username,
		MentionedUsername: meta.Get().Username,
		MentionedRole:     bot.MentionedRole(meta.Role()),
		Type:              "thread",
		ThreadID:          fmt.Sprintf("%v#%v", thread.ThreadID, item.ItemID),
	}
	return act.Create()
}

// fill database model by direct message
func addDirectActivity(item *instagram.ThreadItem, share *instagram.MediaShare, thread *instagram.Thread, meta *client.AccountMeta) error {
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

	act := &models.Activity{
		Pk:                item.ItemID,
		UserID:            item.UserID,
		UserName:          username,
		MentionedUsername: ig.Username,
		MentionedRole:     bot.MentionedRole(meta.Role()),
		Type:              "direct",
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
