package fetcher

import (
	"accountstore/client"
	"errors"
	"fetcher/models"
	"fmt"
	"instagram"
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
		switch req.Type {
		case models.SendMessageRequest, models.ShareMediaRequest:
			err = sendMessage(meta, &req)
		case models.CreateThreadRequest:
			err = createThread(meta, &req)
		default:
			log.Errorf("unknown request type %v", req.Type)
			continue
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func sendMessage(meta *client.AccountMeta, req *models.DirectRequest) error {
	if req.Data == "" {
		// @TODO send something with nats?
		log.Warn("skipping empty message")
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

	case err.Error() == "Thread does not exist":
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
	switch req.Type {
	case models.SendMessageRequest, models.CreateThreadRequest:
		switch {
		case req.ThreadID != "":
			threadID = req.ThreadID
			messageID, err = ig.BroadcastText(req.ThreadID, req.Data)
		case len(req.Participants) != 0:
			threadID, messageID, err = ig.SendText(req.Data, req.Participants...)
		default:
			err = errors.New("destination is unspecified")
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

	case models.ShareMediaRequest:
		if req.ThreadID == "" {
			err = errors.New("bad destination")
			return
		}
		threadID = req.ThreadID
		messageID, err = ig.ShareMedia(threadID, req.Data)
	}

	return
}

func createThread(meta *client.AccountMeta, req *models.DirectRequest) error {
	bot, err := global.pubPool.GetRandom()
	if err != nil {
		return err
	}
	req.Participants = append(req.Participants, bot.UserID)
	return sendMessage(meta, req)
}
