package fetcher

import (
	"accountstore/client"
	"errors"
	"fetcher/models"
	"fmt"
	"instagram"
	"net/http"
	"proto/bot"
	"utils/db"
	"utils/log"
	"utils/nats"
)

// @TODO delete expired requests and send errors notify with them

func processRequests(meta *client.AccountMeta) error {
	var requests []models.DirectRequest
	err := db.New().Where("user_id = ?", meta.Get().UserID).Order("id").Limit(100).Find(&requests).Error
	if err != nil {
		return fmt.Errorf("failed to load requests: %v", err)
	}

	for _, req := range requests {
		err = sendMessage(meta, &req)
		if err != nil {
			return err
		}
	}
	return nil
}

func sendMessage(meta *client.AccountMeta, req *models.DirectRequest) error {
	if req.Data == "" {
		// @TODO send something with nats?
		log.Warn("skipping empty message %v", req)
		err := db.New().Delete(req).Error
		if err != nil {
			return fmt.Errorf("failed to remove handled request from pending table: %v", err)
		}
		return nil
	}

	ig, err := meta.Delayed()
	if err != nil {
		return err
	}
	reply := bot.DirectNotify{ReplyKey: req.ReplyKey}

	reply.ThreadId, reply.MessageId, err = performSend(ig, req)
	switch {
	case err == nil:

	case err.Error() == "Thread does not exist" || err.Error() == "bad destination":
		reply.Error = err.Error()

	default:
		return err
	}

	err = nats.StanPublish(DirectNotifySubject, &reply)
	if err != nil {
		return fmt.Errorf("failed to send reply via stan: %v", err)
	}
	if req.ID == 0 {
		log.Warn("zero id in DirectRequest")
		return nil
	}
	// @CHECK i can not see any real reason to save request logs. am i wrong?
	err = db.New().Delete(req).Error
	if err != nil {
		return fmt.Errorf("failed to remove handled request from pending table: %v", err)
	}
	return nil
}

func performSend(ig *instagram.Instagram, req *models.DirectRequest) (threadID, messageID string, err error) {
	switch req.Kind {
	case bot.MessageType_CreateThread:
		var bot *instagram.Instagram
		bot, err = global.pubPool.GetRandom()
		if err != nil {
			return
		}
		req.Participants = append(req.Participants, bot.UserID)
		fallthrough

	case bot.MessageType_Text:
		switch {
		case req.ThreadID != "":
			threadID = req.ThreadID
			messageID, err = ig.BroadcastText(req.ThreadID, req.Data)
		case len(req.Participants) != 0:
			threadID, messageID, err = ig.SendText(req.Data, req.Participants...)
		default:
			err = errors.New("bad destination")
		}
		if err != nil {
			return
		}
		if req.Caption != "" {
			_, err := ig.DirectUpdateTitle(threadID, req.Caption)
			if err != nil {
				log.Errorf("set title for thread %v failed: %v", threadID, err)
			}
		}

	case bot.MessageType_MediaShare:
		if req.ThreadID == "" {
			err = errors.New("bad destination")
			return
		}
		threadID = req.ThreadID
		messageID, err = ig.ShareMedia(threadID, req.Data)

	case bot.MessageType_Image:
		if req.ThreadID == "" {
			err = errors.New("bad destination")
			return
		}
		threadID = req.ThreadID
		var resp *http.Response
		resp, err = http.Get(req.Data)
		if err != nil {
			return
		}
		if resp.StatusCode != 200 {
			err = fmt.Errorf("failed to load image for request %v: bad status '%v(%v)'", req, resp.Status, resp.StatusCode)
			return
		}
		contentType := resp.Header.Get("Content-Type")
		switch contentType {
		case "", "application/octet-stream", "image/jpeg", "image/pjpeg": // ignore
		default:
			err = fmt.Errorf("unexpected content type '%v' for request %v", contentType, req)
			return
		}
		messageID, err = ig.SendPhoto(threadID, resp.Body)
		resp.Body.Close()

	case bot.MessageType_ReplyComment:
		if req.ThreadID == "" {
			err = errors.New("bad destination")
			return
		}
		_, err = ig.CommentMedia(req.ThreadID, req.Data)
		// @TODO looks non-nice. Can we return status code inside error or instagram.Message?
		if err != nil && err.Error() == "Sorry, this media has been deleted" {
			err = nil
		}
		return
	}

	return

}
